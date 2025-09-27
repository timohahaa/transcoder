package splitter

import (
	"context"
	"net/url"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/internal/composer/modules/analyze"
	"github.com/timohahaa/transcoder/internal/composer/modules/task"
	"github.com/timohahaa/transcoder/pkg/errors"
	"github.com/timohahaa/transcoder/pkg/ffmpeg"
	"github.com/timohahaa/transcoder/pkg/ffprobe"
	pb "github.com/timohahaa/transcoder/proto/composer"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Splitter) process(t task.Task) (task.Task, error) {
	var (
		lg          = s.l.WithFields(log.Fields{"task_id": t.ID})
		ctx, cancel = context.WithCancel(context.Background())
		taskDir     = filepath.Join(s.cfg.WorkDir, t.ID.String())
		err         error
	)
	defer cancel()

	var cleanFull, skipTask bool
	defer func() {
		if cleanFull {
			_ = s.cleanFull(t, taskDir)
		} else {
			// @todo: clean only necessary files
		}

		if skipTask {
			_ = s.mod.queue.Skip(context.Background(), t.ID)
		}
	}()

	// check cancel
	{
		done := make(chan struct{})
		defer close(done)
		go func() {
			tic := time.NewTicker(5 * time.Second)
			defer tic.Stop()

			for range tic.C {
				select {
				case <-done:
					return
				default:
					if taskCanceled, _ := s.mod.task.TaskCanceled(t.ID); taskCanceled {
						lg.Warn("task canceled")
						cancel()
						cleanFull = true
						skipTask = true
					}
				}
			}
		}()
	}

	// download source
	var sourcePath string
	if sourcePath, err = s.downloadSource(ctx, t, filepath.Join(taskDir, "source")); err != nil {
		cleanFull = true
		return t, err
	}

	var sourceInfo *ffprobe.Info
	if sourceInfo, err = ffprobe.GetInfo(ctx, sourcePath); err != nil {
		cleanFull = true
		return t, errors.Splitter(err)
	}
	println(sourceInfo)

	// unmux audio/video
	var (
		videoFile  string
		audioFiles []string
	)
	if videoFile, err = ffmpeg.UnmuxVideo(
		ctx,
		sourceInfo,
		sourcePath,
		filepath.Join(taskDir, "videos"),
	); err != nil {
		cleanFull = true
		return t, errors.Unmux(err)
	}

	if audioFiles, err = ffmpeg.UnmuxAudios(
		ctx,
		sourceInfo,
		sourcePath,
		filepath.Join(taskDir, "audios"),
	); err != nil {
		cleanFull = true
		return t, errors.Unmux(err)
	}

	if err := preValidate(ctx, videoFile, audioFiles); err != nil {
		cleanFull = true
		return t, err
	}

	var chunks []ffmpeg.Chunk
	if chunks, err = s.split(ctx, sourceInfo, videoFile, filepath.Join(taskDir, "chunks")); err != nil {
		cleanFull = true
		return t, err
	}

	// presets
	var chunkPresets map[string]analyze.ChunkPresets
	if chunkPresets, err = analyze.CalcChunkPresets(sourceInfo, chunks); err != nil {
		cleanFull = true
		return t, errors.Splitter(err)
	}

	if err := validateChunks(chunkPresets, sourceInfo); err != nil {
		cleanFull = true
		return t, err
	}

	// write subtasks to redis queue
	if err := s.writeToRedis(ctx, t, sourceInfo, chunks, chunkPresets); err != nil {
		cleanFull = true
		skipTask = true
		return t, err
	}

	// update db

	return t, nil
}

func (s *Splitter) writeToRedis(
	ctx context.Context,
	t task.Task,
	info *ffprobe.Info,
	chunks []ffmpeg.Chunk,
	chunkPresets map[string]analyze.ChunkPresets,
) error {
	var lg = s.l.WithFields(log.Fields{"task_id": t.ID})

	if err := s.mod.queue.PrepareTaskMeta(ctx, t.ID); err != nil {
		return err
	}

	var (
		high = info.GetHighestVideo()
		vPb  = pb.Video{
			Codec: high.CodecName,
			// BitRate:  0, // changes from chunk to chunk
			// Duration: 0, // changes from chunk to chunk
			Quality: high.GetQuality(),
			Presets: nil,
			PixFmt:  high.PixFmt,
		}
		tPb = pb.Task{
			ID: t.ID[:],
			// Part: 0, // changes from chunk to chunk
			PartsTotal: int32(len(chunks)),
			Video:      &vPb,
			Audio:      nil,
			// Source: "", // changes from chunk to chunk
			PushTo:    s.cfg.HttpAddr,
			CreatedAt: timestamppb.Now(),
		}
		baseChunkUrl string
		queueKey     = t.Routing
	)
	{
		parsedUrl, err := url.Parse("http://" + s.cfg.HttpAddr + "/v1/chunk")
		if err != nil {
			return errors.Splitter(err)
		}
		q := url.Values{}
		q.Add("task_id", t.ID.String())
		parsedUrl.RawQuery = q.Encode()
		baseChunkUrl = parsedUrl.String()
	}

	lg.Debugf("adding chunks to queue: %v", queueKey)

	for i, chunk := range chunks {
		chunkSource := baseChunkUrl + "&filepath=" + chunk.Path
		chunkInfo := chunkPresets[chunk.Name]

		tPb.Part = int32(i)
		tPb.Source = chunkSource
		tPb.Video.Presets = chunkInfo.Presets
		tPb.Video.BitRate = chunkInfo.Ffprobe.Format.BitRate
		tPb.Video.Duration = float32(chunkInfo.Ffprobe.GetDuration())
		tPb.Video.CreatePoster = chunk.Num == 0

		if err := s.mod.queue.AddSubtask(ctx, queueKey, &tPb); err != nil {
			return err
		}
	}

	return nil
}

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
			_ = s.clean(t, taskDir)
		}

		if skipTask {
			_ = s.mod.queue.SkipTask(context.Background(), t.ID)
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

	var progress = func(point int64) { _ = s.mod.task.UpdateProgress(ctx, t.ID, point) }

	// download source
	var sourcePath string
	if sourcePath, err = s.downloadSource(ctx, t, filepath.Join(taskDir, "source")); err != nil {
		cleanFull = true
		return t, err
	}
	progress(task.ProgressAfterDownloadSource)

	var sourceInfo *ffprobe.Info
	if sourceInfo, err = ffprobe.GetInfo(ctx, sourcePath); err != nil {
		cleanFull = true
		return t, errors.Splitter(err)
	}

	// unmux audio/video
	var (
		videoFile  string
		audioFiles []string
	)
	if videoFile, err = ffmpeg.UnmuxVideo(
		ctx,
		sourceInfo,
		sourcePath,
		filepath.Join(taskDir, "video"),
	); err != nil {
		cleanFull = true
		return t, errors.Unmux(err)
	}

	// Need to update ffprobe info after unmux.
	// Cause for some containers like mkv there is no stream duration in ffprobe,
	// It means that video and audio files can have different duration
	// only format duration.
	// and we wont know it unless we unmux the file
	if sourceInfo, err = ffprobe.GetInfo(ctx, videoFile); err != nil {
		cleanFull = true
		return t, errors.Splitter(err)
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

	progress(task.ProgressAfterUnmux)

	if err := preValidate(ctx, videoFile, audioFiles); err != nil {
		cleanFull = true
		return t, err
	}

	var chunks []ffmpeg.Chunk
	if chunks, err = s.split(ctx, sourceInfo, videoFile, filepath.Join(taskDir, "chunks")); err != nil {
		cleanFull = true
		return t, err
	}

	progress(task.ProgressAfterSplit)

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

	var audioPresets map[string]analyze.AudioPreset
	if audioPresets, err = analyze.CalcAudioPresets(ctx, sourceInfo, audioFiles); err != nil {
		cleanFull = true
		return t, errors.Splitter(err)
	}

	// write subtasks to redis queue
	if err := s.writeToRedis(
		ctx,
		t,
		sourceInfo,
		chunks,
		chunkPresets,
		audioFiles,
		audioPresets,
	); err != nil {
		cleanFull = true
		skipTask = true
		return t, err
	}

	progress(task.ProgressAfterCreateSubtasks)

	// update db
	if err := s.mod.task.UpdateStatus(ctx, t.ID, task.StatusEncoding, nil); err != nil {
		return t, errors.DB(err)
	}

	return t, nil
}

func (s *Splitter) writeToRedis(
	ctx context.Context,
	t task.Task,
	info *ffprobe.Info,
	chunks []ffmpeg.Chunk,
	chunkPresets map[string]analyze.ChunkPresets,
	audioFiles []string,
	audioPresets map[string]analyze.AudioPreset,
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
			PartsTotal: int32(len(chunks) + len(audioFiles)),
			Video:      &vPb,
			Audio:      nil,
			// Source: "", // changes from chunk to chunk
			PushTo:    s.cfg.HttpAddr,
			CreatedAt: timestamppb.Now(),
		}
		queueKey                   = t.Routing
		baseChunkUrl, baseAudioUrl string
	)
	{
		parsedUrl, err := url.Parse("http://" + s.cfg.HttpAddr + "/v1/files/chunk")
		if err != nil {
			return errors.Splitter(err)
		}
		q := url.Values{}
		q.Add("task_id", t.ID.String())
		parsedUrl.RawQuery = q.Encode()
		baseChunkUrl = parsedUrl.String()
	}
	{
		parsedUrl, err := url.Parse("http://" + s.cfg.HttpAddr + "/v1/files/audio")
		if err != nil {
			return errors.Splitter(err)
		}
		q := url.Values{}
		q.Add("task_id", t.ID.String())
		parsedUrl.RawQuery = q.Encode()
		baseAudioUrl = parsedUrl.String()
	}

	lg.Debugf("adding subtasks to queue: %v", queueKey)

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

	tPb.Video = nil
	for i, filePath := range audioFiles {
		audioSource := baseAudioUrl + "&filepath=" + filePath
		audioPreset := audioPresets[filePath]

		tPb.Part = int32(i + len(chunks))
		tPb.Source = audioSource
		tPb.Audio = &pb.Audio{
			Codec:    audioPreset.Ffprobe.GetAllAudios()[0].CodecName,
			BitRate:  audioPreset.Ffprobe.Format.BitRate,
			Duration: float32(audioPreset.Ffprobe.GetDuration()),
			TrackNum: int32(i),
			Preset:   audioPreset.Preset,
		}

		if err := s.mod.queue.AddSubtask(ctx, queueKey, &tPb); err != nil {
			return err
		}
	}

	return nil
}

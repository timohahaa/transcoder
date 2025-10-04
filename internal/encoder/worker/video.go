package worker

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/pkg/errors"
	"github.com/timohahaa/transcoder/pkg/ffmpeg"
	pb "github.com/timohahaa/transcoder/proto/composer"
	"google.golang.org/protobuf/types/known/durationpb"
)

var (
	zeroTime, _ = time.Parse(time.TimeOnly, "00:00:00")
)

func (w *Worker) video(task *pb.Task, taskID uuid.UUID) error {
	var (
		lg           = w.l.WithFields(log.Fields{"task_id": taskID})
		assetsFolder = filepath.Join(
			w.opts.WorkDir,
			taskID.String(),
			"assets",
			strconv.FormatInt(int64(task.Part), 10),
		)
	)

	defer func() {
		if err := os.RemoveAll(assetsFolder); err != nil {
			lg.Errorf("clean assets: %v", err)
		}
		if err := os.RemoveAll(task.Source); err != nil {
			lg.Errorf("clean assets: %v", err)
		}
	}()

	var (
		out        *ffmpeg.Output
		err        error
		progCB     = w.getProgressCallback(task, taskID)
		posterPath string
	)

	out, err = ffmpeg.EncodeCPU(
		context.Background(),
		[]int{w.opts.CpuIdx},
		task.Source,
		assetsFolder,
		task.Presets(),
		progCB,
	)
	if err != nil {
		return errors.Ffmpeg(err)
	}

	if task.Video.CreatePoster {
		if posterPath, err = w.poster(task, assetsFolder, out); err != nil {
			lg.Errorf("create poster: %v", err)
			posterPath = ""
		}
	}

	if err := w.uploadChunks(task, taskID, out.Qualities); err != nil {
		return err
	}

	if _, err = os.Stat(posterPath); err == nil {
		if err := w.uploadPoster(task, taskID, posterPath); err != nil {
			return err
		}
	}

	return nil
}

func (w *Worker) getProgressCallback(task *pb.Task, taskID uuid.UUID) ffmpeg.ProgressCallback {
	var (
		prevProgress = ffmpeg.Progress{
			Time: zeroTime,
		}
		prevSentTime = time.Now()
	)

	return func(progress ffmpeg.Progress) {
		if time.Since(prevSentTime).Seconds() >= 2 {
			if err := w.composer.UpdateProgress(context.Background(), &pb.UpdateProgressReq{
				ID:    task.ID,
				Delta: durationpb.New(progress.Time.Sub(prevProgress.Time)),
			}); err != nil {
				w.l.WithFields(log.Fields{
					"task_id": taskID,
				}).Errorf("send progress: %v", err)
			}
			prevProgress = progress
			prevSentTime = time.Now()
		}
	}
}

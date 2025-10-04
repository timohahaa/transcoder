package worker

import (
	"context"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/pkg/errors"
	"github.com/timohahaa/transcoder/pkg/ffmpeg"
	pb "github.com/timohahaa/transcoder/proto/composer"
)

func (w *Worker) audio(task *pb.Task, taskID uuid.UUID) error {
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
		err       error
		audioPath string
	)

	audioPath, err = ffmpeg.EncodeAudio(
		context.Background(),
		[]int{w.opts.CpuIdx},
		task.Source,
		assetsFolder,
		task.Audio.Preset.Preset(),
	)

	if err != nil {
		return errors.Ffmpeg(err)
	}

	return w.uploadAudio(task, taskID, audioPath)
}

package encoder

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/timohahaa/transcoder/pkg/errors"
	"github.com/timohahaa/transcoder/pkg/request"
	pb "github.com/timohahaa/transcoder/proto/composer"
)

const (
	retryAttempts = 3
)

func (srv *Service) prefetch(task *pb.Task, taskID uuid.UUID) (path string, retErr error) {
	var srcFolder = filepath.Join(
		srv.cfg.WorkDir,
		taskID.String(),
		"source",
	)
	if err := os.MkdirAll(srcFolder, os.ModePerm); err != nil {
		return "", errors.Encoder(err)
	}

	switch {
	case task.Audio != nil:
		path = filepath.Join(
			srcFolder,
			fmt.Sprintf("audio_%d", task.Audio.TrackNum)+filepath.Ext(task.Source),
		)
	case task.Video != nil:
		path = filepath.Join(
			srcFolder,
			fmt.Sprintf("chunk_%d", task.Part)+filepath.Ext(task.Source),
		)
	default:
		// should never
		return "", fmt.Errorf("invalid task: no video or audio found")
	}

	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return "", errors.Encoder(err)
	}
	if err := request.Download(context.Background(), task.Source, path, retryAttempts); err != nil {
		return "", err
	}

	return path, nil
}

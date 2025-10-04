package worker

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/google/uuid"
	"github.com/timohahaa/transcoder/pkg/errors"
	"github.com/timohahaa/transcoder/pkg/ffmpeg"
	"github.com/timohahaa/transcoder/pkg/request"
	pb "github.com/timohahaa/transcoder/proto/composer"
	"golang.org/x/sync/errgroup"
)

const (
	retryAttempts = 3
)

func (w *Worker) uploadChunks(task *pb.Task, taskID uuid.UUID, assets []ffmpeg.Quality) error {
	var baseURL string
	{
		u, err := url.Parse("http://" + task.PushTo + "/v1/files/chunk")
		if err != nil {
			return errors.Encoder(err)
		}

		q := url.Values{}
		q.Add("task_id", taskID.String())
		q.Add("chunk_num", strconv.FormatInt(int64(task.Part), 10))
		u.RawQuery = q.Encode()
		baseURL = u.String()
	}

	var eGroup = &errgroup.Group{}
	for _, asset := range assets {
		eGroup.Go(func() error {
			var (
				url       = baseURL + "&quality=" + asset.Name
				resp, err = request.Upload(context.Background(), url, asset.Path, retryAttempts)
			)
			if err != nil {
				return errors.Network(fmt.Errorf("sending request (url = %v): %v", url, err))
			}
			if resp.StatusCode != http.StatusOK {
				return errors.Network(fmt.Errorf("expected 200 OK (url = %v): %v", url, resp.StatusCode))
			}
			return err
		})
	}
	if err := eGroup.Wait(); err != nil {
		return err
	}

	return nil
}

func (w *Worker) uploadPoster(task *pb.Task, taskID uuid.UUID, posterPath string) error {
	var baseURL string
	{
		u, err := url.Parse("http://" + task.PushTo + "/v1/files/poster")
		if err != nil {
			return errors.Generic(err)
		}

		q := url.Values{}
		q.Add("task_id", taskID.String())
		u.RawQuery = q.Encode()
		baseURL = u.String()
	}

	var resp, err = request.Upload(context.Background(), baseURL, posterPath, retryAttempts)
	if err != nil {
		return errors.Network(fmt.Errorf("sending request (url = %v): %v", baseURL, err))
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Network(fmt.Errorf("expected 200 OK (url = %v): %v", baseURL, resp.StatusCode))
	}

	return nil
}

func (w *Worker) uploadAudio(task *pb.Task, taskID uuid.UUID, audioPath string) error {
	var baseURL string
	{
		u, err := url.Parse("http://" + task.PushTo + "/v1/files/audio")
		if err != nil {
			return errors.Generic(err)
		}

		q := url.Values{}
		q.Add("task_id", taskID.String())
		q.Add("track_num", strconv.FormatInt(int64(task.Audio.TrackNum), 10))
		u.RawQuery = q.Encode()
		baseURL = u.String()
	}

	var resp, err = request.Upload(context.Background(), baseURL, audioPath, retryAttempts)
	if err != nil {
		return errors.Network(fmt.Errorf("sending request (url = %v): %v", baseURL, err))
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Network(fmt.Errorf("expected 200 OK (url = %v): %v", baseURL, resp.StatusCode))
	}

	return nil
}

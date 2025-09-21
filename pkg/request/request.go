package request

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/timohahaa/transcoder/pkg/errors"
)

var (
	putClient = http.Client{
		Timeout: 10 * time.Minute,
		Transport: &http.Transport{
			DisableKeepAlives:  true,
			DisableCompression: true,
		},
	}
	getClient = http.Client{
		Timeout: 10 * time.Minute,
		Transport: &http.Transport{
			DisableKeepAlives:  true,
			DisableCompression: true,
		},
	}
)

func Upload(ctx context.Context, url, path string, maxRetries uint) (resp *http.Response, err error) {
	if maxRetries == 0 {
		maxRetries = 1
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Generic(err)
	}
	defer f.Close()

	i, err := f.Stat()
	if err != nil {
		return nil, errors.Generic(err)
	}

	for n := range maxRetries {
		if n > 0 {
			time.Sleep(time.Duration(n+1) * (1_00 * time.Millisecond))
			if _, err := f.Seek(0, io.SeekStart); err != nil {
				return nil, errors.Generic(err)
			}
		}

		r, err := http.NewRequestWithContext(ctx, http.MethodPost, url, f)
		if err != nil {
			return nil, errors.Generic(err)
		}

		r.ContentLength = i.Size()

		if resp, err = putClient.Do(r); err == nil {
			return resp, nil
		}
	}
	if err == nil {
		return nil, fmt.Errorf("request upload bug")
	}
	return nil, err
}

func Download(ctx context.Context, src, dst string, maxRetries uint) (retErr error) {
	f, err := os.Create(dst)
	if err != nil {
		return errors.Generic(err)
	}
	defer func() {
		f.Sync()
		f.Close()
		if retErr != nil {
			os.Remove(dst)
		}
	}()

	r, err := http.NewRequestWithContext(ctx, http.MethodGet, src, nil)
	if err != nil {
		return errors.Generic(err)
	}

	download := func() error {
		resp, err := getClient.Do(r)
		if err != nil {
			return errors.Network(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return errors.Network(fmt.Errorf(
				"expected 200 OK (url = %v): %v",
				src, resp.StatusCode,
			))
		}

		switch n, err := io.Copy(f, resp.Body); {
		case err != nil:
			return err
		case n != resp.ContentLength:
			return errors.Network(fmt.Errorf(
				"content length header didn't match with file size: %v vs %v",
				resp.ContentLength, n,
			))
		}
		return nil
	}

	for n := range maxRetries {
		if n > 0 {
			time.Sleep(time.Duration(n+1) * (1_00 * time.Millisecond))
			if _, err := f.Seek(0, io.SeekStart); err != nil {
				return errors.Generic(err)
			}
		}
		if err = download(); err == nil {
			return nil
		}
	}
	if err == nil {
		return fmt.Errorf("request download bug")
	}
	return err
}

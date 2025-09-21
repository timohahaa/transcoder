package ffmpeg

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var uid = strconv.Itoa(os.Getuid())

func scope(
	ctx context.Context,
	cpu []int,
	path string,
	env []string,
	args []string,
	progCB ProgressCallback,
) (string, error) {
	allowedCPUs := make([]string, 0, len(cpu))
	for _, idx := range cpu {
		allowedCPUs = append(allowedCPUs, strconv.Itoa(idx))
	}
	scopeArgs := []string{
		//"-q",
		"-S",
		"systemd-run",
		"--uid", uid,
		"--scope", "-p", fmt.Sprintf("AllowedCPUs=%s", strings.Join(allowedCPUs, ",")),
	}

	for _, e := range env {
		scopeArgs = append(scopeArgs,
			"--setenv", e,
		)
	}

	scopeArgs = append(scopeArgs, bin)
	scopeArgs = append(scopeArgs, args...)

	cmd := exec.CommandContext(ctx, "sudo", scopeArgs...)
	cmd.Env = env

	var wrapErr = func(err error) error {
		switch err.(type) {
		case *Error:
			return err
		}
		return &Error{
			Cmd:     cmd.String(),
			Code:    ErrUnknown,
			Path:    path,
			Message: err.Error(),
		}
	}

	stderr, err := cmd.StderrPipe()

	if err != nil {
		return "", wrapErr(err)
	}

	if err := cmd.Start(); err != nil {
		return "", wrapErr(err)
	}
	defer cmd.Cancel()

	var lastOutput string
	if lastOutput, err = readFFmpegStderr(stderr, path, progCB); err != nil {
		return "", wrapErr(err)
	}

	if err := cmd.Wait(); err != nil {
		switch err.(type) {
		case *Error:
			return "", err
		}

		return "", &Error{
			Cmd:     cmd.String(),
			Code:    ErrUnknown,
			Path:    path,
			Message: fmt.Errorf("%v: %v", err, lastOutput).Error(),
		}
	}

	return cmd.String(), nil
}

func readFFmpegStderr(r io.Reader, src string, progCB ProgressCallback) (string, error) {
	var (
		buf          = make([]byte, 1024)
		lastOutput   string
		lastProgress Progress
	)
	if progCB == nil {
		progCB = DiscardProgress
	}

	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			lastOutput = string(buf[:n])

			if prog, err := parseProgress(lastOutput); err == nil {
				lastProgress = prog
				progCB(lastProgress)
			}

			if err := parseTranscodingError(lastOutput, src, lastProgress.Time); err != nil {
				return lastOutput, err
			}
		}

		if err != nil {
			if err == io.EOF {
				return lastOutput, nil
			}
			return lastOutput, err
		}
	}
}

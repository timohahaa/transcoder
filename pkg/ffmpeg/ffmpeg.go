package ffmpeg

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
)

var bin = "ffmpeg"

func SetBin(v string) {
	bin = v
}

var versionRe = regexp.MustCompile("version (.*?) Copyright")

func Version() (string, error) {
	out, err := exec.Command(bin, "-version").Output()
	if err != nil {
		return "", err
	}
	result := versionRe.FindStringSubmatch(string(out))
	if len(result) != 2 {
		return "", errors.New("failed to get version")
	}
	return result[1], nil
}

func Check(ctx context.Context, path string) error {
	args := []string{
		"-xerror",
		"-v", "error",
		"-i", path,
		"-f", "null",
		"-",
	}
	return Execute(ctx, path, args)
}

func Execute(ctx context.Context, src string, args []string) error {
	log.WithFields(log.Fields{
		"mod":  "ffmpeg",
		"args": args,
		"file": src,
	}).Debug("")

	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &bytes.Buffer{}

	if err := cmd.Run(); err != nil {
		log.WithFields(log.Fields{
			"mod":    "ffmpeg",
			"cmd":    cmd.String(),
			"stdout": cmd.Stdout.(*bytes.Buffer).String(),
			"stderr": cmd.Stderr.(*bytes.Buffer).String(),
		}).Error(err)
		return parseError(cmd.Stderr.(*bytes.Buffer).String(), src, time.Time{})
	}

	return nil
}

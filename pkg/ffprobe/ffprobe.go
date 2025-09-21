package ffprobe

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

var bin = "ffprobe"

func SetBin(v string) {
	bin = v
}

func GetInfo(ctx context.Context, path string) (*Info, error) {
	args := []string{
		"-hide_banner",
		//"-v", "error",
		"-print_format", "json",
		"-show_streams",
		"-show_format",
		"-i", path,
	}
	out, err := execute(ctx, args)
	if err != nil {
		return nil, err
	}
	var info Info
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func execute(ctx context.Context, args []string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &bytes.Buffer{}

	if err := cmd.Run(); err != nil {
		log.WithFields(log.Fields{
			"mod":    "ffprobe",
			"cmd":    cmd.String(),
			"stdout": cmd.Stdout.(*bytes.Buffer).String(),
			"stderr": cmd.Stderr.(*bytes.Buffer).String(),
		}).Error(err)
		return nil, err
	}

	return cmd.Stdout.(*bytes.Buffer).Bytes(), nil
}

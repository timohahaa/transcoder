package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// all cases:
// 1) trim start
// 2) trim end
// 3) trim start + trim end
// 4) trim start + pad end
// 5) pad start
// 6) pad end
// 7) pad start + pad end
// 8) pad start + trim end
// 9) do nothing

type AudioPreset struct {
	Channels     int
	Bitrate      int64
	SampleRate   int64
	PadBefore    float64
	PadAfter     float64
	TrimBefore   float64
	TrimDuration float64
}

func (a *AudioPreset) prepareCmd(src string) (input []string, filter string) {
	switch {
	case a.TrimBefore > 0 && a.PadAfter == 0 || // case 1
		a.PadBefore == 0 && a.TrimDuration > 0 || // case 2
		a.TrimBefore > 0 && a.TrimDuration > 0: // case 3
		return a.trimOnly(src)

	case a.TrimBefore > 0 && a.PadAfter > 0: // case 4
		return a.trimPad(src)

	case a.PadBefore > 0 && a.TrimDuration == 0 || // case 5
		a.TrimBefore == 0 && a.PadAfter > 0 || // case 6
		a.PadBefore > 0 && a.PadAfter > 0: // case 7
		return a.padOnly(src)

	case a.PadBefore > 0 && a.TrimDuration > 0: // case 8
		return a.padTrim(src)

	default: // case 9
		input = append(input, "-i", src)
	}
	return
}

func (a *AudioPreset) trimOnly(src string) (input []string, filter string) {
	input = append(input, "-i", src)

	filterParams := []string{}
	if a.TrimBefore > 0 {
		filterParams = append(filterParams, "start=0")
	}
	if a.TrimDuration > 0 {
		filterParams = append(filterParams, fmt.Sprintf("duration=%v", a.TrimDuration))
	}

	filter = "atrim=" + strings.Join(filterParams, ":")

	return
}

func (a *AudioPreset) padOnly(src string) (input []string, filter string) {
	// input
	{
		if a.PadBefore > 0 {
			input = append(input,
				"-f", "lavfi",
				"-t", strconv.FormatFloat(a.PadBefore, 'f', -1, 64),
				"-i", fmt.Sprintf("anullsrc=channel_layout=%d:sample_rate=%d", a.Channels, a.SampleRate),
			)
		}

		input = append(input, "-i", src)

		if a.PadAfter > 0 {
			input = append(input,
				"-f", "lavfi",
				"-t", strconv.FormatFloat(a.PadAfter, 'f', -1, 64),
				"-i", fmt.Sprintf("anullsrc=channel_layout=%d:sample_rate=%d", a.Channels, a.SampleRate),
			)
		}
	}
	// filter
	{
		var numSources = 1
		if a.PadBefore > 0 {
			numSources += 1
		}
		if a.PadAfter > 0 {
			numSources += 1
		}

		for i := range numSources {
			filter += fmt.Sprintf("[%d:a]", i)
		}

		filter += fmt.Sprintf("concat=n=%d:v=0:a=1", numSources)
	}
	return
}

func (a *AudioPreset) padTrim(src string) (input []string, filter string) {
	// input
	{
		input = append(input,
			"-f", "lavfi",
			"-t", strconv.FormatFloat(a.PadBefore, 'f', -1, 64),
			"-i", fmt.Sprintf("anullsrc=channel_layout=%d:sample_rate=%d", a.Channels, a.SampleRate),
		)
		input = append(input, "-i", src)
	}
	// filter
	{
		filter = fmt.Sprintf(
			"[0:a][1:a]concat=n=2:v=0:a=1[concat_out];[concat_out]atrim=duration=%v",
			a.TrimDuration,
		)
	}
	return
}

func (a *AudioPreset) trimPad(src string) (input []string, filter string) {
	// input
	{
		input = append(input, "-i", src)
		input = append(input,
			"-f", "lavfi",
			"-t", strconv.FormatFloat(a.PadAfter, 'f', -1, 64),
			"-i", fmt.Sprintf("anullsrc=channel_layout=%d:sample_rate=%d", a.Channels, a.SampleRate),
		)
	}
	// filter
	{
		filter = "[0:a]atrim=start=0[trim_out];[trim_out][1:a]concat=n=2:v=0:a=1"
	}
	return
}

func EncodeAudio(
	ctx context.Context,
	cpuIdx []int,
	src, dst string,
	preset AudioPreset,
) (string, error) {
	if err := os.MkdirAll(dst, os.ModePerm); err != nil {
		return "", err
	}

	var (
		output = filepath.Join(dst, "audio.mp4")
		args   = []string{
			"-xerror",
			"-hide_banner",
			"-y",
		}
		input, filter = preset.prepareCmd(src)
	)

	args = append(args, input...)

	if filter != "" {
		args = append(args, "-filter_complex", filter)
	}

	args = append(args,
		"-c:a", "libfdk_aac",
		"-vn",
		"-b:a", strconv.FormatInt(preset.Bitrate, 10),
		"-ar", strconv.FormatInt(preset.SampleRate, 10),
		"-ac", strconv.Itoa(preset.Channels),
		"-movflags", "+faststart",
		output,
	)

	if _, err := scope(ctx, cpuIdx, src, nil, args, DiscardProgress); err != nil {
		return "", err
	}

	return output, nil
}

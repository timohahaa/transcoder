package ffmpeg

import (
	"errors"
	"regexp"
	"strconv"
	"time"
)

var (
	progressRe = regexp.MustCompile(`(.)*size=\s*(\d*)(?:kB|KiB)\s+time=\s*(\d*:\d*:\d*.\d+)\s+bitrate=\s*(\d*.\d*)kbits/s\s+.*?speed=\s*(\d*.\d*)x`)
)

type ProgressCallback func(progress Progress)

var DiscardProgress ProgressCallback = func(progress Progress) {}

type Progress struct {
	Size     int // in Kb
	Time     time.Time
	Bitrate  float64
	Speed    float64
	Finished bool
}

func parseProgress(stderr string) (Progress, error) {
	var (
		result Progress
		err    error
	)

	parsedInfo := progressRe.FindStringSubmatch(stderr)
	if len(parsedInfo) == 0 {
		return result, errors.New("no data")
	}

	if parsedInfo[1] == "L" {
		result.Finished = true
	}

	if result.Size, err = strconv.Atoi(parsedInfo[2]); err != nil {
		return result, err
	}
	if result.Time, err = time.Parse(time.TimeOnly, parsedInfo[3]); err != nil {
		return result, err
	}
	if result.Bitrate, err = strconv.ParseFloat(parsedInfo[4], 64); err != nil {
		return result, err
	}
	if result.Speed, err = strconv.ParseFloat(parsedInfo[5], 64); err != nil {
		return result, err
	}

	return result, nil
}

package ffmpeg

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Preset struct {
	Quality        string // 360, 720, 1440, etc...
	MaxBitRate     int64
	MinBitRate     int64
	FPS            string
	Codec          string
	Bufsize        int64
	GOPSeconds     int32
	Profile        string
	Level          string
	CRF            int32 // -cq for GPU and -crf for CPU
	ColorTrc       string
	ColorSpace     string
	ColorPrimaries string
	Tune           string // psnr/ssim/grain/zerolatency/animation/film - content type
	Transpose      string // clock/cclock/flip
	IsVertical     bool
	Width          int
	Height         int
}

type (
	Output struct {
		Cmd       string
		Qualities []Quality
	}
	Quality struct {
		Name string
		Path string
	}
)

func EncodeCPU(
	ctx context.Context,
	cpuIdx []int,
	src, dst string,
	ps []Preset,
	progCB ProgressCallback,
) (*Output, error) {
	var (
		args = []string{
			"-xerror",
			"-hide_banner", "-y",
			"-i", src,
		}
		qualities = make([]Quality, 0, len(ps))
	)

	for _, p := range ps {

		var qualityBasedDst = filepath.Join(dst, p.Quality)
		if err := os.MkdirAll(qualityBasedDst, os.ModePerm); err != nil {
			return nil, fmt.Errorf("make dir: %v", err)
		}

		var path = filepath.Join(qualityBasedDst, fmt.Sprintf("%s.mp4", p.Quality))

		args = append(args,
			"-c:v", p.Codec,
			"-profile:v", p.Profile,
			"-vf", vfOptsCPU(p),
		)

		if p.MaxBitRate != 0 {
			args = append(args, "-maxrate", strconv.FormatInt(p.MaxBitRate, 10))
		}
		if p.MinBitRate != 0 {
			args = append(args, "-minrate", strconv.FormatInt(p.MinBitRate, 10))
		}

		args = append(args,
			"-bufsize", strconv.FormatInt(p.Bufsize, 10),
			"-crf", strconv.FormatInt(int64(p.CRF), 10),
			// "-colorspace", p.ColorSpace,
			// "-color_primaries", p.ColorPrimaries,
			// "-color_trc", p.ColorTrc,
		)

		gop, err := gopOpts(p)
		if err != nil {
			return nil, err
		}
		args = append(args, gop...)

		args = append(args,
			"-an",
			"-movflags", "+faststart",
			path,
		)

		qualities = append(qualities, Quality{
			Name: p.Quality,
			Path: path,
		})
	}

	cmd, err := scope(ctx, cpuIdx, src, nil, args, progCB)
	if err != nil {
		return nil, err
	}

	return &Output{
		Cmd:       cmd,
		Qualities: qualities,
	}, nil
}

func vfOptsCPU(p Preset) string {
	var opts = []string{
		fmt.Sprintf("fps=%s", p.FPS),
		"format=yuv420p",
		"setsar=1/1",
	}

	if p.IsVertical {
		p.Width, p.Height = p.Height, p.Width
	}

	opts = append(opts, fmt.Sprintf("scale=%d:%d", p.Width, p.Height))

	// ffmpeg does transpose on cpu automatically
	// if p.Transpose != "" {
	// opts = append(opts, fmt.Sprintf("transpose=%s", p.Transpose))
	// }

	return strings.Join(opts, ",")
}

func gopOpts(p Preset) ([]string, error) {
	var (
		fps, err = floatFPS(p.FPS)
		gop      = int64(math.Round(float64(p.GOPSeconds) * fps))
		opts     = []string{
			"-g", strconv.FormatInt(gop, 10),
			"-keyint_min", strconv.FormatInt(gop, 10),
			"-sc_threshold", "0",
		}
	)
	return opts, err
}

func floatFPS(fps string) (float64, error) {
	parts := strings.Split(fps, "/")
	switch len(parts) {
	case 1:
		result, err := strconv.ParseFloat(fps, 64)
		if err != nil {
			return 0, fmt.Errorf("can't parse frame rate: %v", err)
		}
		return result, nil
	case 2:
		num1, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return 0, fmt.Errorf("can't parse frame rate: %v", err)
		}
		num2, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return 0, fmt.Errorf("can't parse frame rate: %v", err)
		}
		if num2 == 0 {
			return 0, nil
		}

		return math.Round(num1/num2*100) / 100.0, nil
	default:
		return 0, fmt.Errorf("can't split string correctly (frameRate = %s)", fps)
	}
}

package ffprobe

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/timohahaa/transcoder/pkg/consts"
)

const quickTimeCompatableBrand = "qt  "

func (info Info) GetFileExt() string {
	switch info.Format.FormatName {
	case consts.FormatAVI:
		return consts.ExtAVI
	case consts.FormatMPEGTS:
		return consts.ExtTS
	case consts.FormatMP3:
		return consts.ExtMP3
	case consts.FormatMatroska:
		return consts.ExtMKV
	case consts.FormatMP4:
		if strings.Contains(info.Format.Tags.CompatibleBrands, quickTimeCompatableBrand) ||
			strings.Contains(info.Format.Tags.MajorBrand, quickTimeCompatableBrand) {
			return consts.ExtMOV
		}
		return consts.ExtMP4
	default:
		return consts.ExtMP4
	}
}

func (info Info) GetDuration() float64 {
	var dur = info.GetHighestVideo().Duration
	if dur == 0 {
		dur = info.Format.Duration
	}
	return dur
}

func (info Info) GetHighestVideo() Stream {
	if len(info.Streams) == 0 {
		return Stream{}
	}

	sortCodecs := map[string]int{
		"vp9":  1,
		"h264": 2,
	}

	maxVideo := Stream{}
	for _, stream := range info.Streams {
		if stream.CodecType == consts.CodecTypeAudio {
			// skip audio :)
			continue
		}

		if stream.Height >= maxVideo.Height && !stream.IsPicture() {

			maxCodecIdx, _ := sortCodecs[maxVideo.CodecName]
			videoCodecIdx, _ := sortCodecs[stream.CodecName]

			if videoCodecIdx >= maxCodecIdx {
				maxVideo = stream
			}
		}
	}

	return maxVideo
}

func (info Info) GetAllAudios() []Stream {
	var a []Stream
	for _, s := range info.Streams {
		if s.CodecType == consts.CodecTypeAudio {
			a = append(a, s)
		}
	}
	return a
}

func (info Info) GetAllVideos() []Stream {
	var a []Stream
	for _, s := range info.Streams {
		if s.CodecType == consts.CodecTypeVideo {
			a = append(a, s)
		}
	}
	return a
}

func (s Stream) IsPicture() bool {
	switch {
	case
		s.Disposition.AttachedPic == 1, // for covers of audio
		s.CodecName == "png",           // for PNG
		s.CodecName == "mjpeg",
		s.DurationTs == 1: // for videos by 1 sec
		return true
	default:
		return false
	}
}

func (s Stream) GetFrameRate() (float64, error) {
	if s.CodecType == consts.CodecTypeAudio {
		return 0, errors.New("audio file has no framerate")
	}

	parts := strings.Split(s.RFrameRate, "/")
	switch len(parts) {
	case 1:
		result, err := strconv.ParseFloat(s.RFrameRate, 64)
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
		return 0, fmt.Errorf("can't split string correctly (frameRate = %s)", s.RFrameRate)
	}
}

func (s Stream) GetRotate() int {
	rotate := s.Tags.Rotate
	for _, item := range s.SideDataList {
		if item.Rotation != 0 {
			rotate = item.Rotation
			break
		}
	}
	return rotate
}

func (s Stream) GetResolution() (width, height int) {
	switch int(math.Abs(float64(s.GetRotate()))) {
	case 90, 270:
		width = s.Height
		height = s.Width
	default:
		width = s.Width
		height = s.Height
	}
	return
}

func (s Stream) GetPixelAspectRatio() (width, height int, err error) {
	if s.SampleAspectRatio == "" || s.SampleAspectRatio == "N/A" {
		return 0, 0, fmt.Errorf("not available")
	}

	params := strings.Split(s.SampleAspectRatio, ":")
	if len(params) != 2 {
		return 0, 0, fmt.Errorf("invalid aspect ratio value")
	}

	if width, err = strconv.Atoi(params[0]); err != nil {
		return
	}

	if height, err = strconv.Atoi(params[1]); err != nil {
		return
	}

	return
}

func (s Stream) GetRenderResolution() (width, height int) {
	metaW, metaH := s.GetResolution()
	asW, asH, err := s.GetPixelAspectRatio()
	if err != nil {
		return metaW, metaH
	}

	var aspectRatio float64 = float64(asW) / float64(asH)

	w := int(float64(metaW) * aspectRatio)
	h := metaH
	if w%2 != 0 {
		w -= 1
	}

	switch int(math.Abs(float64(s.GetRotate()))) {
	case 90, 270:
		w = metaW
		h = int(float64(metaH) * aspectRatio)
		if h%2 != 0 {
			h -= 1
		}
	}

	return w, h
}

func (s Stream) GetQuality() string {
	if s.CodecType == consts.CodecTypeAudio {
		return "audio"
	}

	w, h := s.GetRenderResolution()
	q := min(h, w)

	if 0 < q && q <= 360 {
		return consts.Q360p
	}
	if 360 < q && q <= 480 {
		return consts.Q480p
	}
	if 480 < q && q <= 720 {
		return consts.Q720p
	}
	if 720 < q && q <= 1080 {
		return consts.Q1080p
	}
	if 1080 < q && q <= 1440 {
		return consts.Q1440p
	}

	return consts.Q2160p
}

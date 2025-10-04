package composer

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/timohahaa/transcoder/pkg/ffmpeg"
	"google.golang.org/protobuf/proto"
)

func (t *Task) Marshal() ([]byte, error) { return proto.Marshal(t) }
func (t *Task) Unmarshal(b []byte) error { return proto.Unmarshal(b, t) }

func (t *Task) Presets() []ffmpeg.Preset {
	if t == nil {
		return nil
	}
	if t.Video == nil {
		return nil
	}

	var presets []ffmpeg.Preset
	for _, p := range t.Video.Presets {
		presets = append(presets, ffmpeg.Preset{
			Quality:        p.Quality,
			MaxBitRate:     p.MaxBitRate,
			MinBitRate:     p.MinBitRate,
			FPS:            p.FPS,
			Codec:          p.Codec,
			Bufsize:        p.Bufsize,
			GOPSeconds:     p.GOPSeconds,
			Profile:        p.Profile,
			Level:          p.Level,
			CRF:            p.CRF,
			ColorTrc:       p.ColorTrc,
			ColorSpace:     p.ColorSpace,
			ColorPrimaries: p.ColorPrimaries,
			Tune:           p.Tune,
			Transpose:      p.Transpose,
			IsVertical:     p.IsVertical,
			Width:          int(p.Width),
			Height:         int(p.Height),
		})
	}
	return presets
}

func (q *Preset) Marshal() ([]byte, error) { return proto.Marshal(q) }
func (q *Preset) Unmarshal(b []byte) error { return proto.Unmarshal(b, q) }

func (q *Preset) Copy() *Preset {
	if q == nil {
		return nil
	}
	return &Preset{
		Quality:        q.Quality,
		MaxBitRate:     q.MaxBitRate,
		MinBitRate:     q.MinBitRate,
		FPS:            q.FPS,
		Codec:          q.Codec,
		Bufsize:        q.Bufsize,
		GOPSeconds:     q.GOPSeconds,
		Profile:        q.Profile,
		Level:          q.Level,
		CRF:            q.CRF,
		ColorTrc:       q.ColorTrc,
		ColorSpace:     q.ColorSpace,
		ColorPrimaries: q.ColorPrimaries,
		Tune:           q.Tune,
		Transpose:      q.Transpose,
		IsVertical:     q.IsVertical,
		Width:          q.Width,
		Height:         q.Height,
	}
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	j, _ := json.Marshal(e)
	return string(j)
}

func (e *Error) Value() (driver.Value, error) {
	return json.Marshal(e)
}

func (ap *AudioPreset) Copy() *AudioPreset {
	if ap == nil {
		return nil
	}
	return &AudioPreset{
		Channels:     ap.Channels,
		Bitrate:      ap.Bitrate,
		SampleRate:   ap.SampleRate,
		PadBefore:    ap.PadBefore,
		PadAfter:     ap.PadAfter,
		TrimBefore:   ap.TrimBefore,
		TrimDuration: ap.TrimDuration,
	}
}

func (a *AudioPreset) Preset() ffmpeg.AudioPreset {
	if a == nil {
		return ffmpeg.AudioPreset{}
	}
	return ffmpeg.AudioPreset{
		Channels:     int(a.Channels),
		Bitrate:      a.Bitrate,
		SampleRate:   a.SampleRate,
		PadBefore:    float64(a.PadBefore),
		PadAfter:     float64(a.PadAfter),
		TrimBefore:   float64(a.TrimBefore),
		TrimDuration: float64(a.TrimDuration),
	}
}

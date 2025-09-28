package composer

import (
	"database/sql/driver"
	"encoding/json"

	"google.golang.org/protobuf/proto"
)

func (t *Task) Marshal() ([]byte, error) { return proto.Marshal(t) }
func (t *Task) Unmarshal(b []byte) error { return proto.Unmarshal(b, t) }

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
		Channels:   ap.Channels,
		Bitrate:    ap.Bitrate,
		SampleRate: ap.SampleRate,
		PadBefore:  ap.PadBefore,
		PadAfter:   ap.PadAfter,
	}
}

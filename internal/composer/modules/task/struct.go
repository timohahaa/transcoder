package task

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/google/uuid"
	pb "github.com/timohahaa/transcoder/proto/composer"
)

const (
	AutoEncoder     = "auto"
	SoftwareEncoder = "cpu"
	HardwareEncoder = "gpu"
)

const (
	StatusPending           = "pending"
	StatusWaitingSplitting  = "waiting-splitting"
	StatusSplitting         = "splitting"
	StatusEncoding          = "encoding"
	StatusWaitingAssembling = "waiting-assembling"
	StatusAssembling        = "assembling"
	StatusDone              = "done"
	StatusError             = "error"
	StatusCanceled          = "canceled"
)

type Task struct {
	ID       uuid.UUID `db:"task_id"   json:"id"`
	Source   Source    `db:"source"    json:"source"`
	Status   string    `db:"status"    json:"status"`
	Encoder  string    `db:"encoder"   json:"encoder"`
	Routing  string    `db:"routing"   json:"routing"`
	Duration float64   `db:"duration"  json:"duration"`
	FileSize int64     `db:"file_size" json:"file_size"`
	Settings Settings  `db:"settings"  json:"settings"`
	Error    *pb.Error `db:"error"     json:"error"`
}

type Source struct {
	HTTP *SourceHTTP `json:"http"`
	FS   *SourceFS   `json:"fs"`
}

func (s *Source) Scan(value any) error {
	var source []byte
	switch v := value.(type) {
	case []byte:
		source = v
	case string:
		source = []byte(v)
	}

	return json.Unmarshal(source, &s)
}

func (s Source) Value() (driver.Value, error) {
	return json.Marshal(s)
}

type SourceHTTP struct {
	URL string `json:"url"`
}

type SourceFS struct {
	Path string `json:"path"`
}

type Settings struct {
	Encrypt bool `json:"encrypt"`
}

func (s *Settings) Scan(value any) error {
	var source []byte
	switch v := value.(type) {
	case []byte:
		source = v
	case string:
		source = []byte(v)
	}

	return json.Unmarshal(source, &s)
}

func (s Settings) Value() (driver.Value, error) {
	return json.Marshal(s)
}

type CreateForm struct {
	Source   Source   `db:"source"    json:"source"`
	Duration float64  `db:"duration"  json:"duration"  validate:"gt=0"`
	FileSize int64    `db:"file_size" json:"file_size" validate:"gt=0"`
	Settings Settings `db:"settings"  json:"settings"`
}

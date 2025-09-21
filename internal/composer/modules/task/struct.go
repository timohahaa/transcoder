package task

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/google/uuid"
)

const (
	AutoEncoder     = "auto"
	SoftwareEncoder = "cpu"
	HardwareEncoder = "gpu"
)

const (
	StatusSplitting         = "splitting"
	StatusAssembling        = "assembling"
	StatusError             = "error"
	StatusWaitingSplitting  = "waiting-splitting"
	StatusWaitingAssembling = "waiting-assembling"
)

type Task struct {
	ID       uuid.UUID `db:"task_id"`
	Source   Source    `db:"source"`
	Encoder  string    `db:"encoder"`
	Routing  string    `db:"routing"`
	Duration float64   `db:"duration"`
	FileSize int64     `db:"file_size"`
	Settings Settings  `db:"settings"`
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

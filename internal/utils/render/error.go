package render

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/pkg/validate"
)

type HTTPError struct {
	Status        int                        `json:"-"`
	Message       string                     `json:"message"`
	Detail        string                     `json:"detail,omitempty"`
	InvalidParams []validate.InvalidParamErr `json:"invalid_params,omitempty"`
}

func (e *HTTPError) Error() string {
	return e.Message
}

func Error(w http.ResponseWriter, err error) error {
	switch {
	case strings.Contains(err.Error(), "invalid UUID"):
		return writeError(w, HTTPError{
			Status:  http.StatusBadRequest,
			Message: "invalid uuid format",
			Detail:  err.Error(),
		})
	case errors.Is(err, pgx.ErrNoRows):
		return writeError(w, HTTPError{
			Status:  http.StatusNotFound,
			Message: "resource not found",
		})
	}

	switch err := err.(type) {
	case validate.InvalidParamsErr:
		return writeError(w, HTTPError{
			Status:        http.StatusBadRequest,
			Message:       "invalid params found",
			InvalidParams: err,
		})
	case *json.SyntaxError:
		return writeError(w, HTTPError{
			Status:  http.StatusBadRequest,
			Message: "json syntax error",
			Detail:  err.Error(),
		})
	case *json.UnmarshalTypeError:
		return writeError(w, HTTPError{
			Status:  http.StatusBadRequest,
			Message: "json unmarshal error",
			Detail:  err.Error(),
		})
	case *time.ParseError:
		return writeError(w, HTTPError{
			Status:  http.StatusBadRequest,
			Message: "invalid time format",
			Detail:  err.Error(),
		})
	case *HTTPError:
		return writeError(w, *err)
	}

	log.Errorf("render: unrecognized error: %v", err)
	return writeError(w, HTTPError{
		Status:  http.StatusInternalServerError,
		Message: "internal server error",
		Detail:  err.Error(),
	})
}

func writeError(w http.ResponseWriter, httpErr HTTPError) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpErr.Status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	return enc.Encode(map[string]any{
		"error": httpErr,
	})
}

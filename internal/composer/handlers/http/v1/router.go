package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/timohahaa/transcoder/internal/composer/handlers/http/v1/files"
)

func New(workDir string) *chi.Mux {
	var (
		mux = chi.NewMux()
	)

	mux.Mount("/files", files.New(workDir))

	return mux
}

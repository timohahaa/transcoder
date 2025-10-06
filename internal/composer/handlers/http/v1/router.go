package v1

import (
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	_ "github.com/timohahaa/transcoder/docs"
	"github.com/timohahaa/transcoder/internal/composer/handlers/http/v1/files"
)

func New(workDir string) *chi.Mux {
	var (
		mux = chi.NewMux()
	)

	mux.Mount("/swagger", httpSwagger.Handler(
		httpSwagger.UIConfig(map[string]string{
			"supportedSubmitMethods": "[]",
			//"supportedSubmitMethods": `["get", "head"]`,
		}),
	))
	mux.Mount("/files", files.New(workDir))

	return mux
}

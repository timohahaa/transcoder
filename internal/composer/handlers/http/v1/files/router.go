package files

import "github.com/go-chi/chi/v5"

type handlers struct {
	workDir string
}

func New(workDir string) *chi.Mux {
	var (
		mux = chi.NewMux()
		h   = &handlers{
			workDir: workDir,
		}
	)

	mux.Get("/chunk", h.getChunk)
	mux.Post("/chunk", h.pushChunk)
	mux.Get("/audio", h.getAudio)
	mux.Post("/audio", h.pushAudio)
	mux.Post("/poster", h.pushPoster)

	return mux
}

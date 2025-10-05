package tasks

import "github.com/go-chi/chi/v5"

type handlers struct {
}

func New() *chi.Mux {
	var (
		mux = chi.NewMux()
		h   = &handlers{}
	)

	mux.Post("/", h.create)
	mux.Route("/{task_id}", func(mux chi.Router) {
		mux.Get("/", h.get)
		mux.Post("/", h.delete)
		mux.Post("/cancel", h.cancel)
		mux.Post("/progress", h.progress)
	})

	return mux
}

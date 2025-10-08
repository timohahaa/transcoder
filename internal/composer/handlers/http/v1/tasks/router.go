package tasks

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/timohahaa/transcoder/internal/composer/modules/queue"
	"github.com/timohahaa/transcoder/internal/composer/modules/task"
)

type (
	handlers struct {
		mod mod
	}
	mod struct {
		task  *task.Module
		queue *queue.Module
	}
)

func New(conn *pgxpool.Pool, redis redis.UniversalClient) *chi.Mux {
	var (
		mux = chi.NewMux()
		h   = &handlers{
			mod: mod{
				task:  task.New(conn, redis),
				queue: queue.New(conn, redis),
			},
		}
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

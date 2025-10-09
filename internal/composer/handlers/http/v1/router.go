package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	_ "github.com/timohahaa/transcoder/docs"
	"github.com/timohahaa/transcoder/internal/composer/handlers/http/v1/files"
	"github.com/timohahaa/transcoder/internal/composer/handlers/http/v1/tasks"
)

func New(
	conn *pgxpool.Pool,
	redis redis.UniversalClient,
	workDir string,
) *chi.Mux {
	var (
		mux = chi.NewMux()
	)

	mux.Mount("/swagger", httpSwagger.Handler(
		httpSwagger.DefaultModelsExpandDepth(httpSwagger.HideModel),
		httpSwagger.UIConfig(map[string]string{
			"supportedSubmitMethods": "[]",
			//"supportedSubmitMethods": `["get", "head"]`,
		}),
	))
	mux.Mount("/files", files.New(workDir))
	mux.Mount("/tasks", tasks.New(conn, redis))

	return mux
}

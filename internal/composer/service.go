package composer

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/internal/composer/assembler"
	"github.com/timohahaa/transcoder/internal/composer/splitter"
)

type Service struct {
	cfg    Config
	signal chan os.Signal

	mux    *chi.Mux
	server http.Server
	redis  redis.UniversalClient
	conn   *pgxpool.Pool
}

func New(cfg Config) (*Service, error) {
	cfg.setDefaults()

	var (
		mux = chi.NewMux()
		s   = &Service{
			cfg:    cfg,
			signal: make(chan os.Signal),
			mux:    mux,
			server: http.Server{
				Addr:         cfg.HttpAddr,
				Handler:      mux,
				ReadTimeout:  15 * time.Second, // so bit, cause sending and receiving video-files
				WriteTimeout: 15 * time.Second,
			},
		}
		err error
	)

	s.redis = redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    cfg.Redis.Addrs,
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
	})
	if err = s.redis.Ping(context.Background()).Err(); err != nil {
		// return nil, err
	}

	if s.conn, err = pgxpool.New(context.Background(), cfg.PostgresDSN); err != nil {
		return nil, err
	}

	if err = s.conn.Ping(context.Background()); err != nil {
		// return nil, err
	}

	return s, nil
}

func (srv *Service) Run() error {
	var (
		signals = []os.Signal{
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGKILL,
		}
	)

	log.Infof("HTTP server listening on: %s", srv.cfg.HttpAddr)
	go func() {
		if err := srv.server.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Warn("HTTP server: closed.")
			} else {
				log.Fatalf("HTTP server: %v", err)
			}
		}
	}()

	splitter := splitter.New(srv.conn, srv.redis, splitter.Config{
		HttpAddr: srv.cfg.HttpAddr,
		WorkDir:  srv.cfg.WorkDir,
	})
	splitter.Run(srv.cfg.Splitter.Workers, srv.cfg.Splitter.Watchers)

	assembler := assembler.New(srv.conn, assembler.Config{
		WorkDir: srv.cfg.WorkDir,
	})
	assembler.Run(srv.cfg.Assembler.Workers, srv.cfg.Assembler.Watchers)

	signal.Notify(srv.signal, signals...)
	signal := <-srv.signal
	log.Infof("got signal: %s", signal)

	if err := srv.server.Shutdown(context.Background()); err != nil {
		log.Errorf("shutdown HTTP server: %v", err)
	}

	{
		wg := sync.WaitGroup{}
		wg.Go(splitter.Shutdown)
		wg.Go(assembler.Shutdown)
		wg.Wait()
	}

	srv.conn.Close()
	_ = srv.redis.Close()

	return nil
}

package composer

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/internal/composer/assembler"
	"github.com/timohahaa/transcoder/internal/composer/splitter"
)

type Service struct {
	cfg    Config
	signal chan os.Signal

	redis redis.UniversalClient
	conn  *pgxpool.Pool
}

func New(cfg Config) (*Service, error) {
	cfg.setDefaults()

	var (
		s = &Service{
			cfg:    cfg,
			signal: make(chan os.Signal),
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

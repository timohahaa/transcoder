package composer

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

type (
	Service struct {
		cfg    Config
		signal chan os.Signal

		redis redis.UniversalClient
		conn  *pgxpool.Pool
	}
)
type (
	Config struct {
		PostgresDSN string `arg:"required,-,--,env:POSTGRES_DSN"`
		Redis
	}
	Redis struct {
		Addrs    []string `arg:"required,-,--,env:REDIS_ADDRS"`
		Username string   `arg:"required,-,--,env:REDIS_USERNAME"`
		Password string   `arg:"required,-,--,env:REDIS_PASSWORD"`
	}
)

func New(cfg Config) (*Service, error) {
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

	signal.Notify(srv.signal, signals...)
	signal := <-srv.signal
	log.Infof("got signal: %s", signal)

	return nil
}

package worker

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

type (
	Service struct {
		cfg    Config
		signal chan os.Signal
	}

	Config struct {
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
	return s, err
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

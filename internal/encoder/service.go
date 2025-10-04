package encoder

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

type Service struct {
	cfg    Config
	signal chan os.Signal
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

	fmt.Printf("%+v\n", cfg)

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

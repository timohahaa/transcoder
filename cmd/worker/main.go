package main

import (
	"github.com/alexflint/go-arg"
	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/internal/worker"
)

func main() {
	var cfg worker.Config

	arg.MustParse(&cfg)

	service, err := worker.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"github.com/alexflint/go-arg"
	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/internal/composer"
)

func main() {
	var cfg composer.Config

	arg.MustParse(&cfg)

	service, err := composer.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}

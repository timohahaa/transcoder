package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/alexflint/go-arg"
	log "github.com/sirupsen/logrus"
	"github.com/timohahaa/transcoder/internal/composer"
	"golang.org/x/term"
)

func init() {
	log.SetReportCaller(true)
	callerPrettyfier := func(f *runtime.Frame) (function string, file string) {
		if parts := strings.SplitAfter(f.File, "github.com/timohahaa"); len(parts) > 1 {
			return "", fmt.Sprintf(" %s:%d", parts[1], f.Line)
		}
		return "", fmt.Sprintf(" %s:%d", f.File, f.Line)
	}
	if term.IsTerminal(int(syscall.Stdin)) {
		log.SetFormatter(&log.TextFormatter{
			TimestampFormat:  time.DateTime,
			CallerPrettyfier: callerPrettyfier,
		})
	} else {
		log.SetFormatter(&log.JSONFormatter{
			TimestampFormat:  time.DateTime,
			CallerPrettyfier: callerPrettyfier,
		})
	}
	if debug, _ := strconv.ParseBool(os.Getenv("DEBUG")); debug {
		log.SetLevel(log.DebugLevel)
	}
}

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

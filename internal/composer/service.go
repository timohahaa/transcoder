package composer

import (
	"context"
	"errors"
	"net"
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
	"github.com/timohahaa/transcoder/internal/composer/handlers/grpc/composer"
	v1 "github.com/timohahaa/transcoder/internal/composer/handlers/http/v1"
	"github.com/timohahaa/transcoder/internal/composer/splitter"
	pb "github.com/timohahaa/transcoder/proto/composer"
	"google.golang.org/grpc"
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
		mux        = chi.NewMux()
		httpServer = http.Server{
			Addr:         srv.cfg.HttpAddr,
			Handler:      mux,
			ReadTimeout:  15 * time.Second, // so big, cause sending and receiving video-files
			WriteTimeout: 15 * time.Second,
		}
	)

	mux.Mount("/v1", v1.New(srv.cfg.WorkDir))

	log.Infof("HTTP server listening on: %s", srv.cfg.HttpAddr)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				log.Warn("HTTP server: closed.")
			} else {
				log.Fatalf("HTTP server: %v", err)
			}
		}
	}()

	var (
		grpcServer  = grpc.NewServer()
		grpcHandler = composer.New(srv.conn, srv.redis)
	)

	pb.RegisterComposerServer(grpcServer, grpcHandler)

	log.Infof("gRPC server listening on: %s", srv.cfg.GrpcAddr)
	go func() {
		lis, err := net.Listen("tcp", srv.cfg.GrpcAddr)
		if err != nil {
			log.Fatalf("gRPC server: failed to listen on %v: %v", srv.cfg.GrpcAddr, err)
		}
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server: %v", err)
		}
	}()

	// workers
	splitter := splitter.New(srv.conn, srv.redis, splitter.Config{
		HttpAddr: srv.cfg.HttpAddr,
		WorkDir:  srv.cfg.WorkDir,
	})
	splitter.Run(srv.cfg.Splitter.Workers, srv.cfg.Splitter.Watchers)

	assembler := assembler.New(srv.conn, srv.redis, assembler.Config{
		WorkDir: srv.cfg.WorkDir,
	})
	assembler.Run(srv.cfg.Assembler.Workers, srv.cfg.Assembler.Watchers)

	var signals = []os.Signal{
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGKILL,
	}
	signal.Notify(srv.signal, signals...)
	signal := <-srv.signal
	log.Infof("got signal: %s", signal)

	if err := httpServer.Shutdown(context.Background()); err != nil {
		log.Errorf("shutdown HTTP server: %v", err)
	}

	grpcServer.GracefulStop()

	var wg = sync.WaitGroup{}
	wg.Go(splitter.Shutdown)
	wg.Go(assembler.Shutdown)
	wg.Wait()

	srv.conn.Close()
	_ = srv.redis.Close()

	return nil
}

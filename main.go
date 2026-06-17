package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"go.tknz.dev/internal/db"
	"go.tknz.dev/internal/kms"
	"go.tknz.dev/internal/mail"
	"go.tknz.dev/internal/srv"
)

var (
	addr       = flag.String("addr", ":8080", "tcp address to listen on")
	debug      = flag.Bool("debug", false, "enable debug logs")
	mailAddr   = flag.String("mail-addr", "", "mail grpc address")
	mailDomain = flag.String("mail-domain", "tknz.local", "domain to use for outbound mail")
)

func _init() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	zerolog.LevelFieldName = "severity"
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05.000000",
		})
	}

	err := db.Init(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialise db, err: %w", err)
	}

	if err = kms.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialise kms, err: %w", err)
	}

	if err = mail.Init(ctx, *mailAddr, *mailDomain); err != nil {
		return fmt.Errorf("failed to initialise mail, err: %w", err)
	}

	return nil
}

func main() {
	flag.Parse()

	if err := _init(); err != nil {
		log.Fatal().Err(err).Msg("initialisation failed")
	}

	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}

	handler := srv.New()
	srv := http.Server{
		Handler: h2c.NewHandler(handler, &http2.Server{}),
	}

	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		log.Info().Str("addr", l.Addr().String()).Msg("start")
		if err := srv.Serve(l); !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil {
			log.Fatal().Err(err).Msg("http server failed")
		}

	case <-sigCh:
		log.Info().Msg("stop")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal().Err(err).Msg("failed to stop gracefully")
		}
	}
}

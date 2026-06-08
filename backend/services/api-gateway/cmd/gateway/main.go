package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	identityv1 "github.com/venturez/backend/gen/go/identity/v1"
	"github.com/venturez/backend/services/api-gateway/internal/auth"
	"github.com/venturez/backend/services/api-gateway/internal/config"
	"github.com/venturez/backend/services/api-gateway/internal/server"
)

func main() {
	cfg := config.Load()

	pubPEM, err := os.ReadFile(cfg.JWTPublicKeyPath)
	if err != nil {
		slog.Error("read public key", "err", err)
		os.Exit(1)
	}

	verifier, err := auth.NewVerifier(pubPEM)
	if err != nil {
		slog.Error("parse public key", "err", err)
		os.Exit(1)
	}

	conn, err := grpc.NewClient(cfg.IdentityAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("dial identity", "err", err)
		os.Exit(1)
	}

	defer conn.Close()
	identity := identityv1.NewIdentityClient(conn)

	srv := &http.Server{
		Addr: cfg.HTTPAddr,
		Handler: server.New(cfg, verifier, identity),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout: 15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout: 60 * time.Second,
	}

	go func() {
		slog.Info("api-gateway listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	
	slog.Info("shutting down api-gateway")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

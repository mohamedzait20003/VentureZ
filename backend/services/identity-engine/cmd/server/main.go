package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	identityv1 "github.com/venturez/backend/gen/go/identity/v1"
	"github.com/venturez/backend/services/identity-engine/internal/config"
	"github.com/venturez/backend/services/identity-engine/internal/mailer"
	"github.com/venturez/backend/services/identity-engine/internal/service"
	"github.com/venturez/backend/services/identity-engine/internal/store"
	"github.com/venturez/backend/services/identity-engine/internal/token"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("db connect", "err", err)
		os.Exit(1)
	}

	defer st.Close()

	privPEM, err := os.ReadFile(cfg.JWTPrivateKeyPath)
	if err != nil {
		slog.Error("read private key", "err", err)
		os.Exit(1)
	}

	priv, err := token.LoadPrivateKey(privPEM)
	if err != nil {
		slog.Error("parse private key", "err", err)
		os.Exit(1)
	}

	signer := token.NewSigner(priv, cfg.AccessTTL, cfg.Issuer)

	var m mailer.Mailer = mailer.LogMailer{}
	if cfg.SMTPHost != "" {
		m = mailer.NewSMTP(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.MailFrom)
		slog.Info("using SMTP mailer", "host", cfg.SMTPHost)
	} else {
		slog.Warn("SMTP not configured; emails will be logged, not sent")
	}

	svc := service.New(st, signer, cfg.RefreshTTL, m, cfg.AppBaseURL)

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		slog.Error("listen", "err", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	identityv1.RegisterIdentityServer(grpcServer, svc)

	go func() {
		slog.Info("identity-engine listening", "addr", cfg.GRPCAddr)
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("serve", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	
	slog.Info("shutting down identity-engine")
	grpcServer.GracefulStop()
}

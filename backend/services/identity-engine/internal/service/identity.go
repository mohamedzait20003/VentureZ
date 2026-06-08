package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	identityv1 "github.com/venturez/backend/gen/go/identity/v1"
	"github.com/venturez/backend/services/identity-engine/internal/mailer"
	"github.com/venturez/backend/services/identity-engine/internal/password"
	"github.com/venturez/backend/services/identity-engine/internal/store"
	"github.com/venturez/backend/services/identity-engine/internal/token"
)

type Identity struct {
	identityv1.UnimplementedIdentityServer
	store *store.Store
	signer *token.Signer
	refreshTTL time.Duration
	mailer mailer.Mailer
	baseURL string
}

func New(st *store.Store, signer *token.Signer, refreshTTL time.Duration, m mailer.Mailer, baseURL string) *Identity {
	return &Identity{store: st, signer: signer, refreshTTL: refreshTTL, mailer: m, baseURL: baseURL}
}

func (s *Identity) Register(ctx context.Context, req *identityv1.RegisterRequest) (*identityv1.AuthResponse, error) {
	if req.GetEmail() == "" || req.GetUsername() == "" || len(req.GetPassword()) < 8 {
		return nil, status.Error(codes.InvalidArgument, "email, username and 8+ char password are required")
	}

	hash, err := password.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, "hashing failed")
	}
	
	u, err := s.store.CreateUser(ctx, req.GetEmail(), req.GetUsername(), req.GetFirstName(), req.GetLastName(), hash)
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			return nil, status.Error(codes.AlreadyExists, "email or username already taken")
		}
		
		return nil, status.Error(codes.Internal, "could not create user")
	}

	s.issueEmailVerification(ctx, u)
	return s.issueAuth(ctx, u, req.GetDevice())
}

func (s *Identity) Login(ctx context.Context, req *identityv1.LoginRequest) (*identityv1.AuthResponse, error) {
	u, err := s.store.UserByIdentifier(ctx, req.GetIdentifier())
	
	if err != nil || !password.VerifyPassword(u.PasswordHash, req.GetPassword()) {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	return s.issueAuth(ctx, u, req.GetDevice())
}

func (s *Identity) Refresh(ctx context.Context, req *identityv1.RefreshRequest) (*identityv1.AuthResponse, error) {
	h := token.Hash(req.GetRefreshToken())
	
	userID, err := s.store.UserIDByActiveRefresh(ctx, h)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
	}
	
	if err := s.store.RevokeRefresh(ctx, h); err != nil {
		return nil, status.Error(codes.Internal, "rotation failed")
	}

	u, err := s.store.UserByID(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "user load failed")
	}

	return s.issueAuth(ctx, u, req.GetDevice())
}

func (s *Identity) Logout(ctx context.Context, req *identityv1.LogoutRequest) (*identityv1.SimpleResponse, error) {
	_ = s.store.RevokeRefresh(ctx, token.Hash(req.GetRefreshToken()))
	return &identityv1.SimpleResponse{Success: true}, nil
}

func (s *Identity) RequestPasswordReset(ctx context.Context, req *identityv1.RequestPasswordResetRequest) (*identityv1.SimpleResponse, error) {
	if u, err := s.store.UserByIdentifier(ctx, req.GetEmail()); err == nil {
		if raw, h, gerr := token.New(); gerr == nil {
			if err := s.store.UpsertUserToken(ctx, u.ID, "password_reset", h, time.Now().Add(time.Hour)); err == nil {
				link := fmt.Sprintf("%s/reset-password?token=%s", s.baseURL, raw)
				subject, html, text := mailer.PasswordReset(link)
				s.sendAsync(u.Email, subject, html, text)
			}
		}
	}
	
	return &identityv1.SimpleResponse{Success: true}, nil
}

func (s *Identity) ResetPassword(ctx context.Context, req *identityv1.ResetPasswordRequest) (*identityv1.SimpleResponse, error) {
	if len(req.GetNewPassword()) < 8 {
		return nil, status.Error(codes.InvalidArgument, "password too short")
	}

	userID, err := s.store.ConsumeUserToken(ctx, "password_reset", token.Hash(req.GetToken()))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid or expired token")
	}

	hash, err := password.HashPassword(req.GetNewPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, "hashing failed")
	}

	if err := s.store.UpdatePassword(ctx, userID, hash); err != nil {
		return nil, status.Error(codes.Internal, "update failed")
	}

	_ = s.store.RevokeAllUserSessions(ctx, userID)
	return &identityv1.SimpleResponse{Success: true}, nil
}

func (s *Identity) VerifyEmail(ctx context.Context, req *identityv1.VerifyEmailRequest) (*identityv1.SimpleResponse, error) {
	userID, err := s.store.ConsumeUserToken(ctx, "email_verify", token.Hash(req.GetToken()))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid or expired token")
	}

	if err := s.store.SetEmailVerified(ctx, userID); err != nil {
		return nil, status.Error(codes.Internal, "update failed")
	}

	return &identityv1.SimpleResponse{Success: true}, nil
}

func (s *Identity) issueEmailVerification(ctx context.Context, u *store.User) {
	raw, h, err := token.New()
	if err != nil {
		return
	}

	if err := s.store.UpsertUserToken(ctx, u.ID, "email_verify", h, time.Now().Add(24*time.Hour)); err == nil {
		link := fmt.Sprintf("%s/verify-email?token=%s", s.baseURL, raw)
		subject, html, text := mailer.VerifyEmail(link)
		s.sendAsync(u.Email, subject, html, text)
	}
}

func (s *Identity) sendAsync(to, subject, html, text string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := s.mailer.Send(ctx, to, subject, html, text); err != nil {
			slog.Error("send email failed", "to", to, "err", err)
		}
	}()
}

func (s *Identity) issueAuth(ctx context.Context, u *store.User, dev *identityv1.DeviceInfo) (*identityv1.AuthResponse, error) {
	access, ttl, err := s.signer.Issue(u.ID, u.Role)
	if err != nil {
		return nil, status.Error(codes.Internal, "token issue failed")
	}

	raw, h, err := token.New()
	if err != nil {
		return nil, status.Error(codes.Internal, "refresh gen failed")
	}
	
	d := store.Device{}
	if dev != nil {
		d = store.Device{UserAgent: dev.GetUserAgent(), IP: dev.GetIpAddress(), Name: dev.GetDeviceName()}
	}

	if err := s.store.CreateSession(ctx, u.ID, h, d, s.refreshTTL); err != nil {
		return nil, status.Error(codes.Internal, "session create failed")
	}

	return &identityv1.AuthResponse{
		AccessToken: access,
		RefreshToken: raw,
		ExpiresIn: int64(ttl.Seconds()),
		User: &identityv1.User{
			Id:            u.ID,
			Email:         u.Email,
			Username:      u.Username,
			FirstName:     u.FName,
			LastName:      u.LName,
			Role:          u.Role,
			EmailVerified: u.EmailVerified,
		},
	}, nil
}

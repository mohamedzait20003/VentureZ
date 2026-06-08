// Package handler holds the gateway's REST handlers.
package handler

import (
	"encoding/json"
	"net"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	identityv1 "github.com/venturez/backend/gen/go/identity/v1"
	"github.com/venturez/backend/services/api-gateway/internal/respond"
)

const refreshCookieName = "refresh_token"

type Auth struct {
	rpc identityv1.IdentityClient
	cookieSecure bool
	cookieMaxAge int
}

func NewAuth(rpc identityv1.IdentityClient, cookieSecure bool, cookieMaxAge int) *Auth {
	return &Auth{rpc: rpc, cookieSecure: cookieSecure, cookieMaxAge: cookieMaxAge}
}

type registerReq struct {
	Email string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Password string `json:"password"`
}

type loginReq struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type resetRequestReq struct {
	Email string `json:"email"`
}

type resetConfirmReq struct {
	Token string `json:"token"`
	NewPassword string `json:"new_password"`
}

type verifyReq struct {
	Token string `json:"token"`
}

type userJSON struct {
	ID string `json:"id"`
	Email string `json:"email"`
	Username string `json:"username"`
	FirstName string `json:"first_name"`
	LastName string `json:"last_name"`
	Role string `json:"role"`
	EmailVerified bool `json:"email_verified"`
}

type authJSON struct {
	AccessToken string `json:"access_token"`
	ExpiresIn int64 `json:"expires_in"`
	User *userJSON `json:"user"`
}

func (h *Auth) Register(w http.ResponseWriter, r *http.Request) {
	var in registerReq
	if !decode(w, r, &in) {
		return
	}

	res, err := h.rpc.Register(r.Context(), &identityv1.RegisterRequest{
		Email:     in.Email,
		Username:  in.Username,
		FirstName: in.FirstName,
		LastName:  in.LastName,
		Password:  in.Password,
		Device:    device(r),
	})

	h.writeAuth(w, res, err, http.StatusCreated)
}

func (h *Auth) Login(w http.ResponseWriter, r *http.Request) {
	var in loginReq
	if !decode(w, r, &in) {
		return
	}

	res, err := h.rpc.Login(r.Context(), &identityv1.LoginRequest{
		Identifier: in.Identifier,
		Password:   in.Password,
		Device:     device(r),
	})

	h.writeAuth(w, res, err, http.StatusOK)
}

func (h *Auth) Refresh(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(refreshCookieName)
	if err != nil {
		respond.Error(w, http.StatusUnauthorized, "missing refresh token")
		return
	}

	res, gerr := h.rpc.Refresh(r.Context(), &identityv1.RefreshRequest{
		RefreshToken: c.Value,
		Device:       device(r),
	})

	h.writeAuth(w, res, gerr, http.StatusOK)
}

func (h *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(refreshCookieName); err == nil {
		_, _ = h.rpc.Logout(r.Context(), &identityv1.LogoutRequest{RefreshToken: c.Value})
	}

	h.clearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Auth) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var in resetRequestReq
	if !decode(w, r, &in) {
		return
	}

	_, err := h.rpc.RequestPasswordReset(r.Context(), &identityv1.RequestPasswordResetRequest{Email: in.Email})
	h.writeSimple(w, err)
}

func (h *Auth) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var in resetConfirmReq
	if !decode(w, r, &in) {
		return
	}

	_, err := h.rpc.ResetPassword(r.Context(), &identityv1.ResetPasswordRequest{
		Token:       in.Token,
		NewPassword: in.NewPassword,
	})

	h.writeSimple(w, err)
}

func (h *Auth) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var in verifyReq
	if !decode(w, r, &in) {
		return
	}

	_, err := h.rpc.VerifyEmail(r.Context(), &identityv1.VerifyEmailRequest{Token: in.Token})
	h.writeSimple(w, err)
}

func (h *Auth) writeAuth(w http.ResponseWriter, res *identityv1.AuthResponse, err error, okStatus int) {
	if err != nil {
		respond.Error(w, httpStatus(err), grpcMsg(err))
		return
	}

	h.setRefreshCookie(w, res.GetRefreshToken())
	u := res.GetUser()
	respond.JSON(w, okStatus, authJSON{
		AccessToken: res.GetAccessToken(),
		ExpiresIn:   res.GetExpiresIn(),
		User: &userJSON{
			ID:            u.GetId(),
			Email:         u.GetEmail(),
			Username:      u.GetUsername(),
			FirstName:     u.GetFirstName(),
			LastName:      u.GetLastName(),
			Role:          u.GetRole(),
			EmailVerified: u.GetEmailVerified(),
		},
	})
}

func (h *Auth) writeSimple(w http.ResponseWriter, err error) {
	if err != nil {
		respond.Error(w, httpStatus(err), grpcMsg(err))
		return
	}

	respond.JSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *Auth) setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name: refreshCookieName,
		Value: token,
		Path: "/auth",
		HttpOnly: true,
		Secure: h.cookieSecure,
		SameSite: http.SameSiteStrictMode,
		MaxAge: h.cookieMaxAge,
	})
}

func (h *Auth) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: refreshCookieName,
		Value: "",
		Path: "/auth",
		HttpOnly: true,
		Secure: h.cookieSecure,
		SameSite: http.SameSiteStrictMode,
		MaxAge: -1,
	})
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}

	return true
}

func device(r *http.Request) *identityv1.DeviceInfo {
	return &identityv1.DeviceInfo{
		UserAgent:  r.UserAgent(),
		IpAddress:  clientIP(r),
		DeviceName: r.Header.Get("X-Device-Name"),
	}
}

func clientIP(r *http.Request) string {
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

func httpStatus(err error) int {
	switch status.Code(err) {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func grpcMsg(err error) string {
	if s, ok := status.FromError(err); ok {
		return s.Message()
	}
	return "internal error"
}

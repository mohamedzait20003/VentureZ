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

// Auth turns REST calls into identity-engine gRPC calls and manages the
type Auth struct {
	rpc          identityv1.IdentityClient
	cookieSecure bool
	cookieMaxAge int
}

func NewAuth(rpc identityv1.IdentityClient, cookieSecure bool, cookieMaxAge int) *Auth {
	return &Auth{rpc: rpc, cookieSecure: cookieSecure, cookieMaxAge: cookieMaxAge}
}

// ---- request shapes ----

type registerReq struct {
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
}

type loginReq struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type resetRequestReq struct {
	Email string `json:"email"`
}

type resetConfirmReq struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type verifyReq struct {
	Token string `json:"token"`
}

// ---- response shapes ----

type loginEnvelope struct {
	Message string    `json:"message"`
	Data    loginData `json:"data"`
}

type loginData struct {
	Role        string      `json:"role"`
	AccessToken string      `json:"accessToken"`
	Profile     profileJSON `json:"profile"`
}

type profileJSON struct {
	Email         string         `json:"email"`
	Username      string         `json:"username"`
	FirstName     string         `json:"first_name"`
	LastName      string         `json:"last_name"`
	Sessions      []sessionJSON  `json:"sessions,omitempty"`
	Subscription  *subJSON       `json:"subscription,omitempty"`
	Subscriptions []subJSON      `json:"subscriptions,omitempty"`
	Businesses    []businessJSON `json:"businesses,omitempty"`
}

type sessionJSON struct {
	ID         string `json:"id"`
	DeviceName string `json:"device_name"`
	IPAddress  string `json:"ip_address"`
	UserAgent  string `json:"user_agent"`
	CreatedAt  string `json:"created_at"`
	LastUsedAt string `json:"last_used_at"`
	ExpiresAt  string `json:"expires_at"`
}

type subJSON struct {
	ID               string `json:"id"`
	Status           string `json:"status"`
	TierCode         string `json:"tier_code"`
	TierName         string `json:"tier_name"`
	PriceCents       int64  `json:"price_cents"`
	CurrentPeriodEnd string `json:"current_period_end,omitempty"`
}

type businessJSON struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Industry string `json:"industry"`
	Country  string `json:"country"`
	Currency string `json:"currency"`
}

// ---- handlers ----

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
	if err != nil {
		respond.Error(w, httpStatus(err), grpcMsg(err))
		return
	}
	respond.JSON(w, http.StatusCreated, map[string]string{"message": res.GetMessage()})
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
	h.writeLogin(w, res, err, "Logged in successfully", http.StatusOK)
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
	h.writeLogin(w, res, gerr, "Token refreshed", http.StatusOK)
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

// ---- helpers ----

func (h *Auth) writeLogin(w http.ResponseWriter, res *identityv1.LoginResponse, err error, message string, okStatus int) {
	if err != nil {
		respond.Error(w, httpStatus(err), grpcMsg(err))
		return
	}
	// The refresh token goes ONLY into the HttpOnly cookie, never the JSON body.
	h.setRefreshCookie(w, res.GetRefreshToken())
	respond.JSON(w, okStatus, loginEnvelope{
		Message: message,
		Data: loginData{
			Role:        res.GetRole(),
			AccessToken: res.GetAccessToken(),
			Profile:     toProfileJSON(res.GetProfile()),
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

func toProfileJSON(p *identityv1.Profile) profileJSON {
	out := profileJSON{
		Email:     p.GetEmail(),
		Username:  p.GetUsername(),
		FirstName: p.GetFirstName(),
		LastName:  p.GetLastName(),
	}
	for _, sn := range p.GetSessions() {
		out.Sessions = append(out.Sessions, sessionJSON{
			ID:         sn.GetId(),
			DeviceName: sn.GetDeviceName(),
			IPAddress:  sn.GetIpAddress(),
			UserAgent:  sn.GetUserAgent(),
			CreatedAt:  sn.GetCreatedAt(),
			LastUsedAt: sn.GetLastUsedAt(),
			ExpiresAt:  sn.GetExpiresAt(),
		})
	}
	if sub := p.GetSubscription(); sub != nil {
		s := toSubJSON(sub)
		out.Subscription = &s
	}
	for _, sub := range p.GetSubscriptions() {
		out.Subscriptions = append(out.Subscriptions, toSubJSON(sub))
	}
	for _, b := range p.GetBusinesses() {
		out.Businesses = append(out.Businesses, businessJSON{
			ID:       b.GetId(),
			Name:     b.GetName(),
			Industry: b.GetIndustry(),
			Country:  b.GetCountry(),
			Currency: b.GetCurrency(),
		})
	}
	return out
}

func toSubJSON(s *identityv1.Subscription) subJSON {
	return subJSON{
		ID:               s.GetId(),
		Status:           s.GetStatus(),
		TierCode:         s.GetTierCode(),
		TierName:         s.GetTierName(),
		PriceCents:       s.GetPriceCents(),
		CurrentPeriodEnd: s.GetCurrentPeriodEnd(),
	}
}

func (h *Auth) setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     "/auth",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   h.cookieMaxAge,
	})
}

func (h *Auth) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/auth",
		HttpOnly: true,
		Secure:   h.cookieSecure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
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

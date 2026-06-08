package store

import (
	"context"
	"time"
)

type SessionInfo struct {
	ID, DeviceName, IP, UserAgent    string
	CreatedAt, LastUsedAt, ExpiresAt time.Time
}

type SubscriptionInfo struct {
	ID, Status, TierCode, TierName string
	PriceCents                     int64
	CurrentPeriodEnd               *time.Time
}

type BusinessInfo struct {
	ID, Name, Industry, Country, Currency string
}

func (s *Store) ActiveSessions(ctx context.Context, userID string) ([]SessionInfo, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, COALESCE(device_name,''), COALESCE(host(ip_address),''), COALESCE(user_agent,''), created_at, last_used_at, expires_at
		FROM sessions
		WHERE user_id=$1 AND revoked_at IS NULL AND expires_at > now()
		ORDER BY last_used_at DESC`, userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var out []SessionInfo
	for rows.Next() {
		var x SessionInfo

		if err := rows.Scan(&x.ID, &x.DeviceName, &x.IP, &x.UserAgent,
			&x.CreatedAt, &x.LastUsedAt, &x.ExpiresAt); err != nil {
			return nil, err
		}

		out = append(out, x)
	}

	return out, rows.Err()
}

func (s *Store) Subscriptions(ctx context.Context, clientID string) ([]SubscriptionInfo, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT s.id, s.status, t.code, t.name, t.price_cents, s.current_period_end
		FROM subscriptions s
		JOIN plan_tiers t ON t.id = s.tier_id
		WHERE s.client_id=$1
		ORDER BY s.created_at DESC`, clientID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var out []SubscriptionInfo
	for rows.Next() {
		var x SubscriptionInfo
		if err := rows.Scan(&x.ID, &x.Status, &x.TierCode, &x.TierName,
			&x.PriceCents, &x.CurrentPeriodEnd); err != nil {
			return nil, err
		}

		out = append(out, x)
	}

	return out, rows.Err()
}

func (s *Store) BusinessesByClient(ctx context.Context, clientID string) ([]BusinessInfo, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, COALESCE(industry,''), COALESCE(country,''), currency
		FROM businesses
		WHERE client_id=$1
		ORDER BY created_at`, clientID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var out []BusinessInfo
	for rows.Next() {
		var x BusinessInfo
		if err := rows.Scan(&x.ID, &x.Name, &x.Industry, &x.Country, &x.Currency); err != nil {
			return nil, err
		}

		out = append(out, x)
	}

	return out, rows.Err()
}

package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dbURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dbURL)
	
	if err != nil {
		return nil, err
	}

	return &Store{pool: pool}, nil
}

func (s *Store) Close() { s.pool.Close() }

type User struct {
	ID, Email, PasswordHash, Role, Username, FName, LName string
	EmailVerified bool
}

type Device struct {
	UserAgent, IP, Name string
}

const userCols = `id, email, username, COALESCE(first_name,''), COALESCE(last_name,''), password_hash, role, emailed_verified`

func scanUser(row pgx.Row) (*User, error) {
	u := &User{}
	err := row.Scan(&u.ID, &u.Email, &u.Username, &u.FName, &u.LName, &u.PasswordHash, &u.Role, &u.EmailVerified)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

// --- users ---

func (s *Store) CreateUser(ctx context.Context, email, username, fName, lName, passwordHash string) (*User, error) {
	u, err := scanUser(s.pool.QueryRow(ctx, `
		INSERT INTO users (email, username, first_name, last_name, password_hash)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+userCols,
		email, username, fName, lName, passwordHash))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return nil, ErrConflict
		}
		return nil, err
	}
	return u, nil
}

// UserByIdentifier looks a user up by either email or username.
func (s *Store) UserByIdentifier(ctx context.Context, identifier string) (*User, error) {
	return scanUser(s.pool.QueryRow(ctx,
		`SELECT `+userCols+` FROM users WHERE email = $1 OR username = $1`, identifier))
}

func (s *Store) UserByID(ctx context.Context, id string) (*User, error) {
	return scanUser(s.pool.QueryRow(ctx,
		`SELECT `+userCols+` FROM users WHERE id = $1`, id))
}

func (s *Store) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE users SET password_hash=$2, updated_at=now() WHERE id=$1`, userID, passwordHash)
	return err
}

func (s *Store) SetEmailVerified(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE users SET emailed_verified=true, email_verified_at=now(), updated_at=now() WHERE id=$1`, userID)
	return err
}

// --- sessions ---

func (s *Store) CreateSession(ctx context.Context, userID, refreshHash string, d Device, ttl time.Duration) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO sessions (user_id, refresh_token_hash, user_agent, ip_address, device_name, expires_at)
		VALUES ($1, $2, $3, NULLIF($4,'')::inet, $5, $6)`,
		userID, refreshHash, d.UserAgent, d.IP, d.Name, time.Now().Add(ttl))
	return err
}

func (s *Store) UserIDByActiveRefresh(ctx context.Context, refreshHash string) (string, error) {
	var userID string
	err := s.pool.QueryRow(ctx, `
		SELECT user_id FROM sessions
		WHERE refresh_token_hash=$1 AND revoked_at IS NULL AND expires_at > now()`,
		refreshHash).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return userID, err
}

func (s *Store) RevokeRefresh(ctx context.Context, refreshHash string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE sessions SET revoked_at=now() WHERE refresh_token_hash=$1`, refreshHash)
	return err
}

func (s *Store) RevokeAllUserSessions(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE sessions SET revoked_at=now() WHERE user_id=$1 AND revoked_at IS NULL`, userID)
	return err
}

// --- user_tokens (reset / verify): one row per (user,type) via upsert ---

func (s *Store) UpsertUserToken(ctx context.Context, userID, typ, tokenHash string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO user_tokens (user_id, type, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, type) DO UPDATE
		  SET token_hash=EXCLUDED.token_hash, expires_at=EXCLUDED.expires_at,
		      consumed_at=NULL, created_at=now()`,
		userID, typ, tokenHash, expiresAt)
	return err
}

func (s *Store) ConsumeUserToken(ctx context.Context, typ, tokenHash string) (string, error) {
	var userID string
	err := s.pool.QueryRow(ctx, `
		UPDATE user_tokens SET consumed_at=now()
		WHERE type=$1 AND token_hash=$2 AND consumed_at IS NULL AND expires_at > now()
		RETURNING user_id`, typ, tokenHash).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return userID, err
}

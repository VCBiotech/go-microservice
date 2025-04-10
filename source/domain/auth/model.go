package auth

import "time"

type OAuthProvider struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type UserOAuth struct {
	ID             int       `db:"id"`
	UserID         int       `db:"user_id"`
	ProviderID     int       `db:"provider_id"`
	ProviderUserID string    `db:"provider_user_id"`
	AccessToken    string    `db:"access_token"`
	RefreshToken   string    `db:"refresh_token"`
	TokenExpires   time.Time `db:"token_expires"`
}

type Session struct {
	ID           int       `db:"id"`
	UserID       int       `db:"user_id"`
	SessionToken string    `db:"session_token"`
	ExpiresAt    time.Time `db:"expires_at"`
	CreatedAt    time.Time `db:"created_at"`
}

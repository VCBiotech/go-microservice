package user

import "time"

type User struct {
	ID           int       `json:"user_id,omitempty"`
	Email        string    `json:"user_email,omitempty"`
	PasswordHash string    `json:"user_password_hash,omitempty"` // Optional, depending on your auth strategy
	CreatedAt    time.Time `json:"user_created_at,omitempty"`
	UpdatedAt    time.Time `json:"user_updated_at,omitempty"`
}

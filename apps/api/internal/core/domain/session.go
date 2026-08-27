package domain

import "time"

// Session represents a logged-in user's session data
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"` // e.g., "admin", "viewer"
	CSRFToken         string    `json:"csrf_token"`
	ExpiresAt         time.Time `json:"expires_at"`
	DBType            string    `json:"db_type"`
	EncryptedDBConfig string    `json:"encrypted_db_config"`
}

// User represents a user of the Dboke application
type User struct {
	ID           string
	Username     string
	PasswordHash string
	Role         string
}

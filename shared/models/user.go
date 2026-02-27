package models

import "time"

// User represents a user in the system
type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // Never serialize password hash
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserFromConfig is loaded from config.yaml
type UserFromConfig struct {
	Username     string `yaml:"username"`
	PasswordHash string `yaml:"password_hash"`
	ShareGroup   string `yaml:"share_group"`
}

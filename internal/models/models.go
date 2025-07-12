package models

import (
	"time"
)

// User represents a Slack user
type User struct {
	ID       string `db:"id"`
	Username string `db:"username"`
	RealName string `db:"real_name"`
	Email    string `db:"email"`
}

// Karma represents a user's karma score
type Karma struct {
	ID        int       `db:"id"`
	UserID    string    `db:"user_id"`
	Username  string    `db:"username"`
	Score     int       `db:"score"`
	UpdatedAt time.Time `db:"updated_at"`
}

// KarmaLog represents individual karma changes
type KarmaLog struct {
	ID        int       `db:"id"`
	UserID    string    `db:"user_id"`
	GivenBy   string    `db:"given_by"`
	Reason    string    `db:"reason"`
	Change    int       `db:"change"` // +1 or -1
	Timestamp time.Time `db:"timestamp"`
	Channel   string    `db:"channel"`
}

// Birthday represents a user's birthday
type Birthday struct {
	ID       int    `db:"id"`
	UserID   string `db:"user_id"`
	Username string `db:"username"`
	Month    int    `db:"month"`    // 1-12
	Day      int    `db:"day"`      // 1-31
	Year     int    `db:"year"`     // Optional, can be 0 if not provided
	Timezone string `db:"timezone"` // Optional timezone
}

// Anniversary represents a user's work anniversary
type Anniversary struct {
	ID       int    `db:"id"`
	UserID   string `db:"user_id"`
	Username string `db:"username"`
	Month    int    `db:"month"`    // 1-12
	Day      int    `db:"day"`      // 1-31
	Year     int    `db:"year"`     // Year they started
	Timezone string `db:"timezone"` // Optional timezone
}

// SassyResponse represents pre-defined sassy responses
type SassyResponse struct {
	ID       int    `db:"id"`
	Response string `db:"response"`
	Category string `db:"category"` // e.g., "thank_you", "karma_given"
	Active   bool   `db:"active"`
}

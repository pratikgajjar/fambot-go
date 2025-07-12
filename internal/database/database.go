package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pratikgajjar/fambot-go/internal/models"
)

// Database wraps the sql.DB connection and provides methods
type Database struct {
	db *sql.DB
}

// New creates a new database connection and initializes tables
func New(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &Database{db: db}

	// Initialize tables
	if err := database.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	// Insert default sassy responses
	if err := database.insertDefaultSassyResponses(); err != nil {
		log.Printf("Warning: failed to insert default sassy responses: %v", err)
	}

	return database, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// createTables creates all necessary tables
func (d *Database) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT NOT NULL,
			real_name TEXT,
			email TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS karma (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			username TEXT NOT NULL,
			score INTEGER DEFAULT 0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS karma_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			given_by TEXT NOT NULL,
			reason TEXT,
			change INTEGER NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			channel TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS birthdays (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			username TEXT NOT NULL,
			month INTEGER NOT NULL,
			day INTEGER NOT NULL,
			year INTEGER DEFAULT 0,
			timezone TEXT DEFAULT 'UTC',
			UNIQUE(user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS anniversaries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			username TEXT NOT NULL,
			month INTEGER NOT NULL,
			day INTEGER NOT NULL,
			year INTEGER NOT NULL,
			timezone TEXT DEFAULT 'UTC',
			UNIQUE(user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS sassy_responses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			response TEXT NOT NULL,
			category TEXT NOT NULL,
			active BOOLEAN DEFAULT 1
		)`,
	}

	for _, query := range queries {
		if _, err := d.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query %s: %w", query, err)
		}
	}

	return nil
}

// User operations
func (d *Database) UpsertUser(user *models.User) error {
	query := `INSERT OR REPLACE INTO users (id, username, real_name, email) VALUES (?, ?, ?, ?)`
	_, err := d.db.Exec(query, user.ID, user.Username, user.RealName, user.Email)
	return err
}

func (d *Database) GetUser(userID string) (*models.User, error) {
	query := `SELECT id, username, real_name, email FROM users WHERE id = ?`
	row := d.db.QueryRow(query, userID)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.RealName, &user.Email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Karma operations
func (d *Database) GetKarma(userID string) (*models.Karma, error) {
	query := `SELECT id, user_id, username, score, updated_at FROM karma WHERE user_id = ?`
	row := d.db.QueryRow(query, userID)

	var karma models.Karma
	err := row.Scan(&karma.ID, &karma.UserID, &karma.Username, &karma.Score, &karma.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &karma, nil
}

func (d *Database) IncrementKarma(userID, username, givenBy, reason, channel string) error {
	// Start transaction
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update or insert karma
	_, err = tx.Exec(`
		INSERT INTO karma (user_id, username, score, updated_at)
		VALUES (?, ?, 1, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			score = score + 1,
			updated_at = ?`,
		userID, username, time.Now(), time.Now())
	if err != nil {
		return err
	}

	// Log the karma change
	_, err = tx.Exec(`
		INSERT INTO karma_log (user_id, given_by, reason, change, timestamp, channel)
		VALUES (?, ?, ?, 1, ?, ?)`,
		userID, givenBy, reason, time.Now(), channel)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (d *Database) GetTopKarma(limit int) ([]models.Karma, error) {
	query := `SELECT id, user_id, username, score, updated_at FROM karma ORDER BY score DESC LIMIT ?`
	rows, err := d.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var karmas []models.Karma
	for rows.Next() {
		var karma models.Karma
		err := rows.Scan(&karma.ID, &karma.UserID, &karma.Username, &karma.Score, &karma.UpdatedAt)
		if err != nil {
			return nil, err
		}
		karmas = append(karmas, karma)
	}

	return karmas, nil
}

// Birthday operations
func (d *Database) SetBirthday(birthday *models.Birthday) error {
	query := `INSERT OR REPLACE INTO birthdays (user_id, username, month, day, year, timezone) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := d.db.Exec(query, birthday.UserID, birthday.Username, birthday.Month, birthday.Day, birthday.Year, birthday.Timezone)
	return err
}

func (d *Database) GetBirthday(userID string) (*models.Birthday, error) {
	query := `SELECT id, user_id, username, month, day, year, timezone FROM birthdays WHERE user_id = ?`
	row := d.db.QueryRow(query, userID)

	var birthday models.Birthday
	err := row.Scan(&birthday.ID, &birthday.UserID, &birthday.Username, &birthday.Month, &birthday.Day, &birthday.Year, &birthday.Timezone)
	if err != nil {
		return nil, err
	}
	return &birthday, nil
}

func (d *Database) GetTodaysBirthdays() ([]models.Birthday, error) {
	now := time.Now()
	month, day := int(now.Month()), now.Day()

	query := `SELECT id, user_id, username, month, day, year, timezone FROM birthdays WHERE month = ? AND day = ?`
	rows, err := d.db.Query(query, month, day)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var birthdays []models.Birthday
	for rows.Next() {
		var birthday models.Birthday
		err := rows.Scan(&birthday.ID, &birthday.UserID, &birthday.Username, &birthday.Month, &birthday.Day, &birthday.Year, &birthday.Timezone)
		if err != nil {
			return nil, err
		}
		birthdays = append(birthdays, birthday)
	}

	return birthdays, nil
}

// Anniversary operations
func (d *Database) SetAnniversary(anniversary *models.Anniversary) error {
	query := `INSERT OR REPLACE INTO anniversaries (user_id, username, month, day, year, timezone) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := d.db.Exec(query, anniversary.UserID, anniversary.Username, anniversary.Month, anniversary.Day, anniversary.Year, anniversary.Timezone)
	return err
}

func (d *Database) GetAnniversary(userID string) (*models.Anniversary, error) {
	query := `SELECT id, user_id, username, month, day, year, timezone FROM anniversaries WHERE user_id = ?`
	row := d.db.QueryRow(query, userID)

	var anniversary models.Anniversary
	err := row.Scan(&anniversary.ID, &anniversary.UserID, &anniversary.Username, &anniversary.Month, &anniversary.Day, &anniversary.Year, &anniversary.Timezone)
	if err != nil {
		return nil, err
	}
	return &anniversary, nil
}

func (d *Database) GetTodaysAnniversaries() ([]models.Anniversary, error) {
	now := time.Now()
	month, day := int(now.Month()), now.Day()

	query := `SELECT id, user_id, username, month, day, year, timezone FROM anniversaries WHERE month = ? AND day = ?`
	rows, err := d.db.Query(query, month, day)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var anniversaries []models.Anniversary
	for rows.Next() {
		var anniversary models.Anniversary
		err := rows.Scan(&anniversary.ID, &anniversary.UserID, &anniversary.Username, &anniversary.Month, &anniversary.Day, &anniversary.Year, &anniversary.Timezone)
		if err != nil {
			return nil, err
		}
		anniversaries = append(anniversaries, anniversary)
	}

	return anniversaries, nil
}

// Sassy response operations
func (d *Database) GetRandomSassyResponse(category string) (*models.SassyResponse, error) {
	query := `SELECT id, response, category, active FROM sassy_responses WHERE category = ? AND active = 1 ORDER BY RANDOM() LIMIT 1`
	row := d.db.QueryRow(query, category)

	var response models.SassyResponse
	err := row.Scan(&response.ID, &response.Response, &response.Category, &response.Active)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (d *Database) insertDefaultSassyResponses() error {
	responses := []models.SassyResponse{
		{Response: "Oh, you're being polite now? How refreshing! Here's some karma for good manners. 💫", Category: "thank_you", Active: true},
		{Response: "Look who remembered their manners! Take some karma, you well-behaved human. ✨", Category: "thank_you", Active: true},
		{Response: "Gratitude detected! Don't get used to this generosity though... 😏", Category: "thank_you", Active: true},
		{Response: "Thank you? In THIS economy? Fine, here's your karma. 💸", Category: "thank_you", Active: true},
		{Response: "Well well well, someone said thank you. I'm impressed. Have some karma! 🎭", Category: "thank_you", Active: true},
		{Response: "Karma delivered with a side of sass! You're welcome. 💅", Category: "karma_given", Active: true},
		{Response: "Another karma point hits the bank! Keep spreading those good vibes. 🏦", Category: "karma_given", Active: true},
		{Response: "Karma level up! Someone's been a good human today. 📈", Category: "karma_given", Active: true},
		{Response: "Ding! Karma deposited. Your account is looking mighty fine! 💰", Category: "karma_given", Active: true},
		{Response: "Karma inflation is real, but you earned this one! 📊", Category: "karma_given", Active: true},
	}

	for _, response := range responses {
		// Check if response already exists
		var exists bool
		err := d.db.QueryRow("SELECT 1 FROM sassy_responses WHERE response = ?", response.Response).Scan(&exists)
		if err == sql.ErrNoRows {
			// Insert new response
			_, err = d.db.Exec("INSERT INTO sassy_responses (response, category, active) VALUES (?, ?, ?)",
				response.Response, response.Category, response.Active)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

package repository

import "time"

// Rule represents a rule entity in the database
type Rule struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Env       string    `db:"env"`
	Version   string    `db:"version"`
	Content   string    `db:"content"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

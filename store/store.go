package store

import (
	"database/sql/driver"

	"github.com/flarco/g"
	"github.com/spf13/cast"
)

const Directory = "./store"

type User struct {
	ID       string `db:"id" json:"id"`
	Email    string `db:"email" json:"email"`
	LegacyID string `db:"legacy_id" json:"legacy_id"`
	Team     Team   `db:"team" json:"team"`
}

type Team struct {
	ID      string   `db:"id" json:"id"`
	Name    string   `db:"name" json:"name"`
	Active  bool     `db:"active" json:"active"`
	Blocked bool     `db:"blocked" json:"blocked"`
	Cities  []string `db:"cities" json:"cities"`
	Export  int      `json:"export"`
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (t *Team) Scan(value interface{}) error {
	val := cast.ToString(value)
	return g.JSONScanner(t, val)
}

// Value return json value, implement driver.Valuer interface
func (t Team) Value() (driver.Value, error) {
	return g.JSONValuer(t, "{}")
}

package migrations

import (
	"strings"

	"github.com/pocketbase/dbx"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	upQueries := []string{
		`CREATE TABLE IF NOT EXISTS obras_export_item (
			id TEXT PRIMARY KEY,
			team_id TEXT NOT NULL,
			obra_id TEXT NOT NULL,
			user_id TEXT,
			exported_at TEXT NOT NULL
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_obras_export_item_team_obra
			ON obras_export_item (team_id, obra_id)`,
		`CREATE INDEX IF NOT EXISTS idx_obras_export_item_team
			ON obras_export_item (team_id)`,
	}

	downQueries := []string{
		"DROP INDEX IF EXISTS idx_obras_export_item_team",
		"DROP INDEX IF EXISTS idx_obras_export_item_team_obra",
		"DROP TABLE IF EXISTS obras_export_item",
	}

	m.Register(func(db dbx.Builder) error {
		for _, q := range upQueries {
			if _, err := db.NewQuery(q).Execute(); err != nil {
				return err
			}
		}
		return nil
	}, func(db dbx.Builder) error {
		for _, q := range downQueries {
			if _, err := db.NewQuery(q).Execute(); err != nil {
				if strings.Contains(strings.ToLower(err.Error()), "no such table") {
					continue
				}
				return err
			}
		}
		return nil
	})
}

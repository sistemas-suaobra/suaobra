package migrations

import (
	"strings"

	"github.com/pocketbase/dbx"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	upQueries := []string{
		`CREATE TABLE IF NOT EXISTS obras_export_log (
			id TEXT PRIMARY KEY,
			team_id TEXT NOT NULL,
			user_id TEXT,
			rows_count INTEGER NOT NULL DEFAULT 0,
			created TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_obras_export_log_team_created ON obras_export_log (team_id, created)`,
	}

	downQueries := []string{
		"DROP INDEX IF EXISTS idx_obras_export_log_team_created",
		"DROP TABLE IF EXISTS obras_export_log",
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

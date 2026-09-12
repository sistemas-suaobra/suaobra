package migrations

import (
	"strings"

	"github.com/pocketbase/dbx"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	upQueries := []string{
		`CREATE TABLE IF NOT EXISTS obras_export_limit_alert (
			id TEXT PRIMARY KEY,
			team_id TEXT NOT NULL,
			day_key TEXT NOT NULL,
			created TEXT NOT NULL
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_obras_export_limit_alert_team_day
			ON obras_export_limit_alert (team_id, day_key)`,
	}

	downQueries := []string{
		"DROP INDEX IF EXISTS idx_obras_export_limit_alert_team_day",
		"DROP TABLE IF EXISTS obras_export_limit_alert",
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

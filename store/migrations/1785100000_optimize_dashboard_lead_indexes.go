package migrations

import (
	"strings"

	"github.com/pocketbase/dbx"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Índices compostos para o dashboard (funnel/history/leads):
// as queries filtram por team_id (+ owner_id) e intervalo de
// favorited_at/visited_at. Sem (team_id, data), o SQLite faz
// full scan em lead mesmo com índices single-column nas datas.
func init() {
	upQueries := []string{
		"CREATE INDEX IF NOT EXISTS idx_lead_team_favorited ON lead (team_id, favorited_at)",
		"CREATE INDEX IF NOT EXISTS idx_lead_team_visited ON lead (team_id, visited_at)",
		"CREATE INDEX IF NOT EXISTS idx_lead_team_owner_favorited ON lead (team_id, owner_id, favorited_at)",
		"CREATE INDEX IF NOT EXISTS idx_lead_team_owner_visited ON lead (team_id, owner_id, visited_at)",
		"CREATE INDEX IF NOT EXISTS idx_list_lead_lead_id ON list_lead (lead_id)",
	}

	downQueries := []string{
		"DROP INDEX IF EXISTS idx_lead_team_favorited",
		"DROP INDEX IF EXISTS idx_lead_team_visited",
		"DROP INDEX IF EXISTS idx_lead_team_owner_favorited",
		"DROP INDEX IF EXISTS idx_lead_team_owner_visited",
		"DROP INDEX IF EXISTS idx_list_lead_lead_id",
	}

	m.Register(func(db dbx.Builder) error {
		for _, q := range upQueries {
			if _, err := db.NewQuery(q).Execute(); err != nil {
				if strings.Contains(strings.ToLower(err.Error()), "no such table") {
					continue
				}
				return err
			}
		}
		return nil
	}, func(db dbx.Builder) error {
		for _, q := range downQueries {
			if _, err := db.NewQuery(q).Execute(); err != nil {
				return err
			}
		}
		return nil
	})
}

package migrations

import (
	"strings"

	"github.com/pocketbase/dbx"
	m "github.com/pocketbase/pocketbase/migrations"
)

// O funil busca o último histórico de cada lead numa passagem só.
// O índice antigo era só em lead_id, então cada lead refazia a busca.
func init() {
	const up = "CREATE INDEX IF NOT EXISTS idx_lead_activity_team_type_lead_created ON lead_activity (team_id, type, lead_id, created)"
	const down = "DROP INDEX IF EXISTS idx_lead_activity_team_type_lead_created"

	m.Register(func(db dbx.Builder) error {
		if _, err := db.NewQuery(up).Execute(); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "no such table") {
				return nil
			}
			return err
		}
		return nil
	}, func(db dbx.Builder) error {
		_, err := db.NewQuery(down).Execute()
		return err
	})
}

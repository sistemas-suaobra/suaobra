package server

import (
	"strings"
	"time"

	"github.com/flarco/g"
	"github.com/pocketbase/dbx"
	"github.com/suaobra/suaobra-app/store"
)

const maxObrasExportLeadsPerDay = 500

func brtDayStart(now time.Time) time.Time {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.FixedZone("BRT", -3*60*60)
	}

	t := now.In(loc)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
}

func dailyObrasExportCount(teamID string, now time.Time) (int, error) {
	if teamID == "" {
		return 0, nil
	}

	var count int
	err := store.MainDbNewQuery(`
		SELECT COALESCE(SUM(rows_count), 0)
		FROM obras_export_log
		WHERE team_id = {:teamId}
		  AND created >= {:dayStart}
	`).Bind(dbx.Params{
		"teamId":   teamID,
		"dayStart": brtDayStart(now).UTC().Format(time.RFC3339),
	}).Row(&count)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such table") {
			return 0, nil
		}
		return 0, g.Error(err, "erro ao consultar exportações do dia")
	}

	return count, nil
}

func capObrasExportRows(teamID string, requested int, now time.Time) (allowed int, err error) {
	if requested <= 0 {
		return 0, nil
	}

	used, err := dailyObrasExportCount(teamID, now)
	if err != nil {
		return 0, err
	}

	remaining := maxObrasExportLeadsPerDay - used
	if remaining <= 0 {
		return 0, nil
	}

	if requested > remaining {
		return remaining, nil
	}

	return requested, nil
}

func recordObrasExport(teamID, userID string, rowsCount int, now time.Time) error {
	if teamID == "" || rowsCount <= 0 {
		return nil
	}

	_, err := store.MainDB.Exec(
		`INSERT INTO obras_export_log (id, team_id, user_id, rows_count, created) VALUES (?, ?, ?, ?, ?)`,
		store.ID("obras_export_log", ""),
		teamID,
		userID,
		rowsCount,
		now.UTC().Format(time.RFC3339),
	)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such table") {
			return nil
		}
		return g.Error(err, "erro ao registrar exportação")
	}

	return nil
}

// recordObrasExportItems marca obras já exportadas pela equipe (primeira vez).
// INSERT OR IGNORE preserva a data da primeira exportação.
func recordObrasExportItems(teamID, userID string, obraIDs []string, now time.Time) error {
	if teamID == "" || len(obraIDs) == 0 {
		return nil
	}

	exportedAt := now.UTC().Format(time.RFC3339)
	for _, obraID := range obraIDs {
		obraID = strings.TrimSpace(obraID)
		if obraID == "" {
			continue
		}
		_, err := store.MainDB.Exec(
			`INSERT OR IGNORE INTO obras_export_item (id, team_id, obra_id, user_id, exported_at) VALUES (?, ?, ?, ?, ?)`,
			store.ID("obras_export_item", ""),
			teamID,
			obraID,
			userID,
			exportedAt,
		)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "no such table") {
				return nil
			}
			return g.Error(err, "erro ao registrar obra exportada")
		}
	}

	return nil
}

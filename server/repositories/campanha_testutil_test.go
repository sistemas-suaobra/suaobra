package repositories

import (
	"testing"

	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
	"github.com/suaobra/suaobra-app/server/testutil/pbtest"
	"github.com/stretchr/testify/require"
)

func newCampanhaTestDAO(t *testing.T) (*daos.Dao, func()) {
	t.Helper()
	return pbtest.NewCampanhaDAO(t)
}

func seedCampanha(t *testing.T, dao *daos.Dao, teamID, campanhaID string, canais []string) *models.Record {
	t.Helper()
	return pbtest.SeedCampanha(t, dao, teamID, campanhaID, canais)
}

func seedDestinatario(
	t *testing.T,
	repo *CampanhaRepo,
	teamID, campanhaID, obraID, contatoTipo, status string,
) *models.Record {
	t.Helper()

	rec, err := repo.CreateDestinatario(map[string]any{
		"team_id":       teamID,
		"campanha_id":   campanhaID,
		"obra_id":       obraID,
		"contato_tipo":  contatoTipo,
		"status":        status,
		"telefone_e164": "5511999999999",
		"nome_contato":  "João Teste",
		"tentativas":    0,
	})
	require.NoError(t, err)
	return rec
}

package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExistsContatoEnviado_SemRegistros(t *testing.T) {
	dao, cleanup := newCampanhaTestDAO(t)
	defer cleanup()

	repo := NewCampanhaRepo(dao)

	ok, err := repo.ExistsContatoEnviado("team_1", "obra_1", "OWNER")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestExistsContatoEnviado_StatusAtivos(t *testing.T) {
	statuses := []string{
		DestStatusPendente,
		DestStatusEmFila,
		DestStatusEnviado,
	}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			dao, cleanup := newCampanhaTestDAO(t)
			defer cleanup()

			repo := NewCampanhaRepo(dao)
			seedCampanha(t, dao, "team_1", "camp_1", []string{"WHATSAPP"})
			seedDestinatario(t, repo, "team_1", "camp_1", "obra_abc", "OWNER", status)

			ok, err := repo.ExistsContatoEnviado("team_1", "obra_abc", "OWNER")
			require.NoError(t, err)
			assert.True(t, ok, "status %s deve bloquear reenvio", status)
		})
	}
}

func TestExistsContatoEnviado_StatusInativos(t *testing.T) {
	statuses := []string{
		DestStatusFalhou,
		DestStatusIgnorado,
	}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			dao, cleanup := newCampanhaTestDAO(t)
			defer cleanup()

			repo := NewCampanhaRepo(dao)
			seedCampanha(t, dao, "team_1", "camp_1", []string{"WHATSAPP"})
			seedDestinatario(t, repo, "team_1", "camp_1", "obra_abc", "OWNER", status)

			ok, err := repo.ExistsContatoEnviado("team_1", "obra_abc", "OWNER")
			require.NoError(t, err)
			assert.False(t, ok, "status %s não deve bloquear reenvio", status)
		})
	}
}

func TestExistsContatoEnviado_PendenteEmOutraCampanha(t *testing.T) {
	// Cenário do bug: campanha 1 ainda processando (PENDENTE), campanha 2 tenta o mesmo contato.
	dao, cleanup := newCampanhaTestDAO(t)
	defer cleanup()

	repo := NewCampanhaRepo(dao)
	seedCampanha(t, dao, "team_1", "camp_1", []string{"WHATSAPP"})
	seedCampanha(t, dao, "team_1", "camp_2", []string{"WHATSAPP"})
	seedDestinatario(t, repo, "team_1", "camp_1", "obra_xyz", "OWNER", DestStatusPendente)

	ok, err := repo.ExistsContatoEnviado("team_1", "obra_xyz", "OWNER")
	require.NoError(t, err)
	assert.True(t, ok, "PENDENTE na campanha 1 deve impedir campanha 2")
}

func TestExistsContatoEnviado_IsolamentoPorTeamObraTipo(t *testing.T) {
	dao, cleanup := newCampanhaTestDAO(t)
	defer cleanup()

	repo := NewCampanhaRepo(dao)
	seedCampanha(t, dao, "team_1", "camp_1", []string{"WHATSAPP"})
	seedDestinatario(t, repo, "team_1", "camp_1", "obra_1", "OWNER", DestStatusPendente)

	cases := []struct {
		name        string
		teamID      string
		obraID      string
		contatoTipo string
		want        bool
	}{
		{"mesmo contato", "team_1", "obra_1", "OWNER", true},
		{"outro team", "team_2", "obra_1", "OWNER", false},
		{"outra obra", "team_1", "obra_2", "OWNER", false},
		{"outro tipo", "team_1", "obra_1", "PROFISSIONAL", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, err := repo.ExistsContatoEnviado(tc.teamID, tc.obraID, tc.contatoTipo)
			require.NoError(t, err)
			assert.Equal(t, tc.want, ok)
		})
	}
}

func TestExistsEnviadoByCampanhaTelefone_DedupNaMesmaCampanha(t *testing.T) {
	dao, cleanup := newCampanhaTestDAO(t)
	defer cleanup()

	repo := NewCampanhaRepo(dao)
	seedCampanha(t, dao, "team_1", "camp_1", []string{"WHATSAPP"})
	dest := seedDestinatario(t, repo, "team_1", "camp_1", "obra_1", "OWNER", DestStatusEnviado)

	ok, err := repo.ExistsEnviadoByCampanhaTelefone("camp_1", "5511999999999", dest.Id)
	require.NoError(t, err)
	assert.False(t, ok, "não deve contar o próprio destinatário")

	ok, err = repo.ExistsEnviadoByCampanhaTelefone("camp_1", "5511999999999", "outro_id")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestFindUltimoContatoCampanha_RetornaTelefoneEnviado(t *testing.T) {
	dao, cleanup := newCampanhaTestDAO(t)
	defer cleanup()

	repo := NewCampanhaRepo(dao)
	seedCampanha(t, dao, "team_1", "camp_antiga", []string{"WHATSAPP"})

	_, err := repo.CreateDestinatario(map[string]any{
		"team_id":       "team_1",
		"campanha_id":   "camp_antiga",
		"obra_id":       "obra_lucas",
		"contato_tipo":  "OWNER",
		"status":        DestStatusEnviado,
		"telefone_e164": "5535998877665",
		"nome_contato":  "Lucas Pelisson",
		"tentativas":    1,
	})
	require.NoError(t, err)

	tel, email, err := repo.FindUltimoContatoCampanha("team_1", "obra_lucas", "OWNER")
	require.NoError(t, err)
	assert.Equal(t, "5535998877665", tel)
	assert.Equal(t, "", email)
}

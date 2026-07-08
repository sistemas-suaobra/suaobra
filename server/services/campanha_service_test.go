package services

import (
	"testing"

	"github.com/pocketbase/pocketbase/daos"
	"github.com/suaobra/suaobra-app/server/repositories"
	"github.com/suaobra/suaobra-app/server/testutil/pbtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizarContatoTipo(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"OWNER", ContatoTipoOwner},
		{"owner", ContatoTipoOwner},
		{"PROPRIETARIO", ContatoTipoOwner},
		{"PROPRIETÁRIO", ContatoTipoOwner},
		{"PROFISSIONAL", ContatoTipoProfessional},
		{"PROFESSIONAL", ContatoTipoProfessional},
		{"", ""},
		{"INVALIDO", ""},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			assert.Equal(t, tc.want, normalizarContatoTipo(tc.in))
		})
	}
}

func TestPersonalizarMensagem_SubstituiVariaveis(t *testing.T) {
	svc := &CampanhaService{}

	msg := svc.PersonalizarMensagem(
		"Olá {{primeiroNome}}, obra em {{bairro}} - {{cidade}}/{{UF}}",
		map[string]string{
			"nome":   "Maria Silva",
			"bairro": "Centro",
			"cidade": "São Paulo",
			"uf":     "SP",
		},
	)

	assert.Contains(t, msg, "Maria")
	assert.Contains(t, msg, "Centro")
	assert.Contains(t, msg, "São Paulo")
	assert.Contains(t, msg, "SP")
	assert.NotContains(t, msg, "{{")
}

func TestJaFoiContactado_DetectaPendenteEmOutraCampanha(t *testing.T) {
	dao, cleanup := pbtest.NewCampanhaDAO(t)
	defer cleanup()

	repo := repositories.NewCampanhaRepo(dao)
	svc := NewCampanhaService(repo, nil, nil, nil)

	teamID := "team_test"
	pbtest.SeedCampanha(t, dao, teamID, "camp_1", []string{"WHATSAPP"})

	_, err := repo.CreateDestinatario(map[string]any{
		"team_id":       teamID,
		"campanha_id":   "camp_1",
		"obra_id":       "obra_x",
		"contato_tipo":  ContatoTipoOwner,
		"status":        repositories.DestStatusPendente,
		"telefone_e164": "5511666666666",
		"tentativas":    0,
	})
	require.NoError(t, err)

	ok, err := svc.jaFoiContactado(teamID, "obra_x", ContatoTipoOwner)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestJaFoiContactado_IgnoraParametrosVazios(t *testing.T) {
	svc := &CampanhaService{}

	ok, err := svc.jaFoiContactado("", "obra", ContatoTipoOwner)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestPermiteReenvioAposFalhou_ExistsContatoEnviado(t *testing.T) {
	dao, cleanup := pbtest.NewCampanhaDAO(t)
	defer cleanup()

	repo := repositories.NewCampanhaRepo(dao)

	teamID := "team_test"
	pbtest.SeedCampanha(t, dao, teamID, "camp_antiga", []string{"WHATSAPP"})

	_, err := repo.CreateDestinatario(map[string]any{
		"team_id":      teamID,
		"campanha_id":  "camp_antiga",
		"obra_id":      "obra_retry",
		"contato_tipo": ContatoTipoOwner,
		"status":       repositories.DestStatusFalhou,
		"erro":         "sem telefone",
		"tentativas":   1,
	})
	require.NoError(t, err)

	ok, err := repo.ExistsContatoEnviado(teamID, "obra_retry", ContatoTipoOwner)
	require.NoError(t, err)
	assert.False(t, ok, "FALHOU não deve bloquear nova campanha")
}

func TestGetCampaignChannels(t *testing.T) {
	dao, cleanup := pbtest.NewCampanhaDAO(t)
	defer cleanup()

	svc := NewCampanhaService(repositories.NewCampanhaRepo(dao), nil, nil, nil)

	rec := pbtest.SeedCampanha(t, dao, "team_1", "camp_channels", []string{"WHATSAPP", "EMAIL"})

	wa, email := svc.getCampaignChannels(rec)
	assert.True(t, wa)
	assert.True(t, email)

	rec.Set("canal", []string{"WHATSAPP"})
	wa, email = svc.getCampaignChannels(rec)
	assert.True(t, wa)
	assert.False(t, email)
}

func seedServiceCampanha(t *testing.T, dao *daos.Dao, teamID, campanhaID string, canais []string) {
	t.Helper()
	pbtest.SeedCampanha(t, dao, teamID, campanhaID, canais)
}

func TestComplementarContatosDoHistorico_ReutilizaTelefone(t *testing.T) {
	dao, cleanup := pbtest.NewCampanhaDAO(t)
	defer cleanup()

	repo := repositories.NewCampanhaRepo(dao)
	svc := NewCampanhaService(repo, nil, nil, nil)

	teamID := "team_test"
	pbtest.SeedCampanha(t, dao, teamID, "camp_hist", []string{"WHATSAPP"})

	_, err := repo.CreateDestinatario(map[string]any{
		"team_id":       teamID,
		"campanha_id":   "camp_hist",
		"obra_id":       "obra_lucas",
		"contato_tipo":  ContatoTipoOwner,
		"status":        repositories.DestStatusEnviado,
		"telefone_e164": "5535998877665",
		"nome_contato":  "Lucas Pelisson",
		"tentativas":    1,
	})
	require.NoError(t, err)

	contato := &ObraContatoCampanha{
		ObraID:        "obra_lucas",
		ContatoTipo:   ContatoTipoOwner,
		NomeContato:   "Lucas Pelisson",
		TelefonesE164: []string{},
	}

	svc.complementarContatosDoHistorico(teamID, contato)

	require.Len(t, contato.TelefonesE164, 1)
	assert.Equal(t, "5535998877665", contato.TelefonesE164[0])
}

func TestUniqueNonEmpty(t *testing.T) {
	out := uniqueNonEmpty([]string{"a", "", "a", "b", "b"})
	assert.Equal(t, []string{"a", "b"}, out)
}

func TestNormalizeEmail(t *testing.T) {
	assert.Equal(t, "teste@email.com", normalizeEmail("  Teste@Email.COM  "))
}

func TestFormatPhone(t *testing.T) {
	svc := &CampanhaService{}
	assert.Equal(t, "5511999887766", svc.formatPhone("(11) 99988-7766"))
	assert.Equal(t, "", svc.formatPhone("abc"))
}

func TestContatoStatusBloqueiaReenvio_DocumentacaoSQL(t *testing.T) {
	bloqueia := func(status string) bool {
		return status != repositories.DestStatusFalhou && status != repositories.DestStatusIgnorado
	}

	for _, s := range []string{repositories.DestStatusPendente, repositories.DestStatusEmFila, repositories.DestStatusEnviado} {
		assert.True(t, bloqueia(s), "SQL deve tratar %s como contactado", s)
	}
	for _, s := range []string{repositories.DestStatusFalhou, repositories.DestStatusIgnorado} {
		assert.False(t, bloqueia(s), "SQL não deve tratar %s como contactado", s)
	}
}

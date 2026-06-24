package pbtest

import (
	"testing"

	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tests"
	"github.com/stretchr/testify/require"
)

// NewCampanhaDAO retorna um DAO PocketBase pronto para testes de campanha.
func NewCampanhaDAO(t *testing.T) (*daos.Dao, func()) {
	t.Helper()

	testApp, err := tests.NewTestApp()
	require.NoError(t, err)

	dao := testApp.Dao()
	require.NoError(t, EnsureCampanhaCollections(dao))

	return dao, testApp.Cleanup
}

// EnsureCampanhaCollections cria as collections campanhas e campanha_destinatarios se não existirem.
func EnsureCampanhaCollections(dao *daos.Dao) error {
	if _, err := dao.FindCollectionByNameOrId("campanhas"); err != nil {
		if err := createCampanhasCollection(dao); err != nil {
			return err
		}
	}

	if _, err := dao.FindCollectionByNameOrId("campanha_destinatarios"); err != nil {
		if err := createCampanhaDestinatariosCollection(dao); err != nil {
			return err
		}
	}

	return nil
}

func createCampanhasCollection(dao *daos.Dao) error {
	col := &models.Collection{Name: "campanhas", Type: models.CollectionTypeBase}
	for _, f := range []schema.SchemaField{
		{Name: "team_id", Type: schema.FieldTypeText},
		{Name: "nome", Type: schema.FieldTypeText},
		{Name: "canal", Type: schema.FieldTypeJson},
		{Name: "status", Type: schema.FieldTypeText},
		{Name: "mensagem_template", Type: schema.FieldTypeText},
		{Name: "assunto_email", Type: schema.FieldTypeText},
		{Name: "manter_ia", Type: schema.FieldTypeBool},
	} {
		col.Schema.AddField(&f)
	}
	return dao.SaveCollection(col)
}

func createCampanhaDestinatariosCollection(dao *daos.Dao) error {
	col := &models.Collection{Name: "campanha_destinatarios", Type: models.CollectionTypeBase}
	for _, f := range []schema.SchemaField{
		{Name: "team_id", Type: schema.FieldTypeText},
		{Name: "campanha_id", Type: schema.FieldTypeText},
		{Name: "obra_id", Type: schema.FieldTypeText},
		{Name: "contato_tipo", Type: schema.FieldTypeText},
		{Name: "status", Type: schema.FieldTypeText},
		{Name: "telefone_e164", Type: schema.FieldTypeText},
		{Name: "email", Type: schema.FieldTypeText},
		{Name: "nome_contato", Type: schema.FieldTypeText},
		{Name: "cidade", Type: schema.FieldTypeText},
		{Name: "bairro", Type: schema.FieldTypeText},
		{Name: "uf", Type: schema.FieldTypeText},
		{Name: "address", Type: schema.FieldTypeText},
		{Name: "erro", Type: schema.FieldTypeText},
		{Name: "tentativas", Type: schema.FieldTypeNumber},
		{Name: "enviado_em", Type: schema.FieldTypeDate},
	} {
		col.Schema.AddField(&f)
	}
	return dao.SaveCollection(col)
}

// SeedCampanha insere uma campanha mínima para testes.
func SeedCampanha(t *testing.T, dao *daos.Dao, teamID, campanhaID string, canais []string) *models.Record {
	t.Helper()

	col, err := dao.FindCollectionByNameOrId("campanhas")
	require.NoError(t, err)

	rec := models.NewRecord(col)
	rec.Set("id", campanhaID)
	rec.Set("team_id", teamID)
	rec.Set("nome", "Campanha teste")
	rec.Set("canal", canais)
	rec.Set("status", "RASCUNHO")
	rec.Set("mensagem_template", "Olá {{nome}}")
	rec.Set("manter_ia", false)
	require.NoError(t, dao.SaveRecord(rec))

	return rec
}

package repositories

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

type ConexaoRepo struct {
	dao *daos.Dao
}

func NewConexaoRepo(dao *daos.Dao) *ConexaoRepo {
	return &ConexaoRepo{dao: dao}
}

func (r *ConexaoRepo) FindActiveWhatsappByTeam(teamID string) (*models.Record, error) {
	con, err := r.dao.FindFirstRecordByFilter(
		"conexoes",
		`team_id = {:team} && canal = "WHATSAPP" && ativo = true`,
		dbx.Params{"team": teamID},
	)
	if err != nil {
		return nil, err
	}
	if con == nil || con.Id == "" {
		return nil, nil
	}
	return con, nil
}

// FindActiveWhatsappByOwner busca a conexão de WhatsApp do PRÓPRIO usuário
// (sem fallback). Retorna (nil, nil) quando o usuário ainda não tem a sua.
// Usado na criação para garantir idempotência por usuário sem "herdar" a
// conexão legada/compartilhada do time.
func (r *ConexaoRepo) FindActiveWhatsappByOwner(teamID, userID string) (*models.Record, error) {
	if userID == "" {
		return nil, nil
	}
	con, err := r.dao.FindFirstRecordByFilter(
		"conexoes",
		`team_id = {:team} && canal = "WHATSAPP" && ativo = true && user_id = {:user}`,
		dbx.Params{"team": teamID, "user": userID},
	)
	if err != nil {
		// not found vira (nil, nil) para simplificar os checks de idempotência
		return nil, nil
	}
	if con == nil || con.Id == "" {
		return nil, nil
	}
	return con, nil
}

// FindActiveWhatsappLegacy busca a conexão legada/compartilhada do time
// (user_id vazio), usada como fallback para quem ainda não conectou o seu número.
func (r *ConexaoRepo) FindActiveWhatsappLegacy(teamID string) (*models.Record, error) {
	con, err := r.dao.FindFirstRecordByFilter(
		"conexoes",
		`team_id = {:team} && canal = "WHATSAPP" && ativo = true && user_id = ""`,
		dbx.Params{"team": teamID},
	)
	if err != nil {
		return nil, nil
	}
	if con == nil || con.Id == "" {
		return nil, nil
	}
	return con, nil
}

func (r *ConexaoRepo) FindAllActiveWhatsappByTeam(teamID string) ([]*models.Record, error) {
	return r.dao.FindRecordsByFilter(
		"conexoes",
		`team_id = {:team} && canal = "WHATSAPP" && ativo = true`,
		"-updated",
		0,
		0,
		dbx.Params{"team": teamID},
	)
}

// FindActiveWhatsappForUser resolve a conexão a ser USADA pelo usuário:
// primeiro a dele (user_id), e se não houver, cai no fallback legado do time.
func (r *ConexaoRepo) FindActiveWhatsappForUser(teamID, userID string) (*models.Record, error) {
	if con, _ := r.FindActiveWhatsappByOwner(teamID, userID); con != nil {
		return con, nil
	}
	return r.FindActiveWhatsappLegacy(teamID)
}

func (r *ConexaoRepo) CreateWhatsapp(teamID, userID, name string) (*models.Record, error) {
	col, err := r.dao.FindCollectionByNameOrId("conexoes")
	if err != nil {
		return nil, err
	}

	con := models.NewRecord(col)
	con.Set("team_id", teamID)
	con.Set("user_id", userID)
	con.Set("canal", "WHATSAPP")
	con.Set("nome", name)
	con.Set("ativo", true)

	if err := r.dao.SaveRecord(con); err != nil {
		return nil, err
	}
	return con, nil
}

func (r *ConexaoRepo) FindActiveEmailByTeam(teamID string) (*models.Record, error) {
	con, err := r.dao.FindFirstRecordByFilter(
		"conexoes",
		`team_id = {:team} && canal = "EMAIL" && ativo = true`,
		dbx.Params{"team": teamID},
	)
	if err != nil {
		return nil, err
	}
	if con == nil || con.Id == "" {
		return nil, nil
	}
	return con, nil
}

func (r *ConexaoRepo) CreateEmail(teamID, nome string) (*models.Record, error) {
	col, err := r.dao.FindCollectionByNameOrId("conexoes")
	if err != nil {
		return nil, err
	}

	con := models.NewRecord(col)
	con.Set("team_id", teamID)
	con.Set("canal", "EMAIL")
	con.Set("nome", nome)
	con.Set("ativo", true)

	if err := r.dao.SaveRecord(con); err != nil {
		return nil, err
	}
	return con, nil
}

func (r *ConexaoRepo) Delete(rec *models.Record) {
	_ = r.dao.DeleteRecord(rec)
}

// FindByID busca uma conexao por ID
func (r *ConexaoRepo) FindByID(id string) (*models.Record, error) {
	return r.dao.FindRecordById("conexoes", id)
}

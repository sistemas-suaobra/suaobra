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

func (r *ConexaoRepo) CreateWhatsapp(teamID, name string) (*models.Record, error) {
	col, err := r.dao.FindCollectionByNameOrId("conexoes")
	if err != nil {
		return nil, err
	}

	con := models.NewRecord(col)
	con.Set("team_id", teamID)
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

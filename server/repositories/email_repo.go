package repositories

import (
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

type EmailRepo struct {
	dao *daos.Dao
}

func NewEmailRepo(dao *daos.Dao) *EmailRepo {
	return &EmailRepo{dao: dao}
}

func (r *EmailRepo) FindByConexao(conexaoID string) (*models.Record, error) {
	email, err := r.dao.FindFirstRecordByFilter(
		"conexoes_email",
		`conexao = {:conexao}`,
		dbx.Params{"conexao": conexaoID},
	)
	if err != nil {
		// Se a collection não existe, retorna nil sem erro
		if err.Error() == "sql: no rows in result set" || 
		   err.Error() == "the requested resource wasn't found" ||
		   strings.Contains(err.Error(), "no such table") {
			return nil, nil
		}
		return nil, err
	}
	return email, nil
}

func (r *EmailRepo) Create(conexaoID string, fields map[string]any) (*models.Record, error) {
	col, err := r.dao.FindCollectionByNameOrId("conexoes_email")
	if err != nil {
		return nil, err
	}

	email := models.NewRecord(col)
	email.Set("conexao", conexaoID)
	for k, v := range fields {
		email.Set(k, v)
	}

	if err := r.dao.SaveRecord(email); err != nil {
		return nil, err
	}
	return email, nil
}

func (r *EmailRepo) Update(email *models.Record, fields map[string]any) error {
	for k, v := range fields {
		email.Set(k, v)
	}
	return r.dao.SaveRecord(email)
}

package repositories

import (
	"github.com/flarco/g"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

type IntencaoRepo struct {
	dao *daos.Dao
}

func NewIntencaoRepo(dao *daos.Dao) *IntencaoRepo {
	return &IntencaoRepo{dao: dao}
}

// FindAtivasByTeamID retorna todas as intenções ativas de uma team, ordenadas por prioridade
func (r *IntencaoRepo) FindAtivasByTeamID(teamID string) ([]*models.Record, error) {
	records, err := r.dao.FindRecordsByFilter(
		"agente_ia_intencoes",
		"team_id = {:teamId} && ativa = true",
		"-prioridade",
		0,
		0,
		dbx.Params{"teamId": teamID},
	)
	if err != nil {
		return nil, g.Error(err, "erro ao buscar intenções ativas")
	}
	return records, nil
}

// FindByID busca uma intenção por ID
func (r *IntencaoRepo) FindByID(id string) (*models.Record, error) {
	record, err := r.dao.FindRecordById("agente_ia_intencoes", id)
	if err != nil {
		return nil, g.Error(err, "intenção não encontrada")
	}
	return record, nil
}

// ConversaRepo gerencia conversas de IA
type ConversaRepo struct {
	dao *daos.Dao
}

func NewConversaRepo(dao *daos.Dao) *ConversaRepo {
	return &ConversaRepo{dao: dao}
}

// FindByTelefone busca conversa ativa por telefone
func (r *ConversaRepo) FindByTelefone(teamID, telefone string) (*models.Record, error) {
	records, err := r.dao.FindRecordsByFilter(
		"conversas_ia",
		"team_id = {:teamId} && telefone = {:telefone}",
		"-updated",
		1,
		0,
		dbx.Params{
			"teamId":   teamID,
			"telefone": telefone,
		},
	)
	if err != nil {
		return nil, g.Error(err, "erro ao buscar conversa por telefone")
	}

	if len(records) == 0 {
		return nil, nil
	}

	return records[0], nil
}

// FindByCampanhaAndTelefone busca conversa por campanha e telefone
func (r *ConversaRepo) FindByCampanhaAndTelefone(campanhaID, telefone string) (*models.Record, error) {
	record, err := r.dao.FindFirstRecordByFilter(
		"conversas_ia",
		"campanha_id = {:campanhaId} && telefone = {:telefone}",
		dbx.Params{
			"campanhaId": campanhaID,
			"telefone":   telefone,
		},
	)
	if err != nil {
		return nil, err
	}
	return record, nil
}

// Create cria uma nova conversa
func (r *ConversaRepo) Create(data map[string]interface{}) (*models.Record, error) {
	collection, err := r.dao.FindCollectionByNameOrId("conversas_ia")
	if err != nil {
		return nil, g.Error(err, "collection conversas_ia não encontrada")
	}

	record := models.NewRecord(collection)
	for k, v := range data {
		record.Set(k, v)
	}

	if err := r.dao.SaveRecord(record); err != nil {
		return nil, g.Error(err, "erro ao criar conversa")
	}

	return record, nil
}

// Update atualiza uma conversa existente
func (r *ConversaRepo) Update(record *models.Record, data map[string]interface{}) error {
	for k, v := range data {
		record.Set(k, v)
	}

	if err := r.dao.SaveRecord(record); err != nil {
		return g.Error(err, "erro ao atualizar conversa")
	}

	return nil
}

// Save salva um record de conversa
func (r *ConversaRepo) Save(record *models.Record) error {
	if err := r.dao.SaveRecord(record); err != nil {
		return g.Error(err, "erro ao salvar conversa")
	}
	return nil
}
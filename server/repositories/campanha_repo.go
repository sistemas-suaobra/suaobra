package repositories

import (
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

// CampanhaStatus constantes de status da campanha
const (
	CampanhaStatusRascunho    = "RASCUNHO"
	CampanhaStatusAgendada    = "AGENDADA"
	CampanhaStatusEmAndamento = "EM_ANDAMENTO"
	CampanhaStatusPausada     = "PAUSADA"
	CampanhaStatusConcluida   = "CONCLUIDA"
	CampanhaStatusCancelada   = "CANCELADA"
)

// DestinatarioStatus constantes de status do destinatário
const (
	DestStatusPendente = "PENDENTE"
	DestStatusEmFila   = "EM_FILA"
	DestStatusEnviado  = "ENVIADO"
	DestStatusFalhou   = "FALHOU"
	DestStatusIgnorado = "IGNORADO"
)

type CampanhaRepo struct {
	dao *daos.Dao
}

func NewCampanhaRepo(dao *daos.Dao) *CampanhaRepo {
	return &CampanhaRepo{dao: dao}
}

// FindByID busca uma campanha pelo ID
func (r *CampanhaRepo) FindByID(id string) (*models.Record, error) {
	return r.dao.FindRecordById("campanhas", id)
}

// FindConexaoByID busca uma conexão pelo ID (para descobrir o canal)
func (r *CampanhaRepo) FindConexaoByID(id string) (*models.Record, error) {
	return r.dao.FindRecordById("conexoes", id)
}

// FindLeadByID busca um lead pelo ID do PocketBase
func (r *CampanhaRepo) FindLeadByID(id string) (*models.Record, error) {
	return r.dao.FindRecordById("lead", id)
}

// FindByTeam busca campanhas de um team
func (r *CampanhaRepo) FindByTeam(teamID string, limit, offset int) ([]*models.Record, error) {
	return r.dao.FindRecordsByFilter(
		"campanhas",
		"team_id = {:teamId}",
		"-created",
		limit, offset,
		dbx.Params{"teamId": teamID},
	)
}

// FindByTeamAndStatus busca campanhas de um team com status específico
func (r *CampanhaRepo) FindByTeamAndStatus(teamID string, status string) ([]*models.Record, error) {
	return r.dao.FindRecordsByFilter(
		"campanhas",
		"team_id = {:teamId} && status = {:status}",
		"-created",
		0, 0,
		dbx.Params{"teamId": teamID, "status": status},
	)
}

// Create cria uma nova campanha
func (r *CampanhaRepo) Create(teamID, nome, mensagemTemplate, canal string) (*models.Record, error) {
	col, err := r.dao.FindCollectionByNameOrId("campanhas")
	if err != nil {
		return nil, err
	}

	rec := models.NewRecord(col)
	rec.Set("team_id", teamID)
	rec.Set("nome", nome)
	rec.Set("mensagem_template", mensagemTemplate)
	rec.Set("canal", canal)
	rec.Set("status", CampanhaStatusRascunho)

	if err := r.dao.SaveRecord(rec); err != nil {
		return nil, err
	}
	return rec, nil
}

// UpdateStatus atualiza o status de uma campanha
func (r *CampanhaRepo) UpdateStatus(rec *models.Record, status string) error {
	rec.Set("status", status)
	
	now := time.Now().UTC().Format(time.RFC3339)
	switch status {
	case CampanhaStatusEmAndamento:
		rec.Set("iniciado_em", now)
	case CampanhaStatusConcluida, CampanhaStatusCancelada:
		rec.Set("finalizado_em", now)
	}
	
	return r.dao.SaveRecord(rec)
}

// Delete exclui uma campanha
func (r *CampanhaRepo) Delete(rec *models.Record) error {
	return r.dao.DeleteRecord(rec)
}

// ---------------- Destinatários ----------------

// FindDestinatariosByCampanha busca destinatários de uma campanha
func (r *CampanhaRepo) FindDestinatariosByCampanha(campanhaID string) ([]*models.Record, error) {
	return r.dao.FindRecordsByFilter(
		"campanha_destinatarios",
		"campanha_id = {:campanhaId}",
		"-created",
		0, 0,
		dbx.Params{"campanhaId": campanhaID},
	)
}

// FindDestinatariosPendentes busca destinatários pendentes de uma campanha
func (r *CampanhaRepo) FindDestinatariosPendentes(campanhaID string) ([]*models.Record, error) {
	return r.dao.FindRecordsByFilter(
		"campanha_destinatarios",
		"campanha_id = {:campanhaId} && status = {:status}",
		"-created",
		0, 0,
		dbx.Params{"campanhaId": campanhaID, "status": DestStatusPendente},
	)
}

// CreateDestinatario cria um destinatário para a campanha
func (r *CampanhaRepo) CreateDestinatario(campanhaID, leadID, nomeContato string) (*models.Record, error) {
	col, err := r.dao.FindCollectionByNameOrId("campanha_destinatarios")
	if err != nil {
		return nil, err
	}

	rec := models.NewRecord(col)
	rec.Set("campanha_id", campanhaID)
	rec.Set("lead_id", leadID)
	rec.Set("nome_contato", nomeContato)
	rec.Set("status", DestStatusPendente)
	rec.Set("tentativas", 0)

	if err := r.dao.SaveRecord(rec); err != nil {
		return nil, err
	}
	return rec, nil
}

// UpdateDestinatarioStatus atualiza o status de um destinatário
func (r *CampanhaRepo) UpdateDestinatarioStatus(rec *models.Record, status string, erro string) error {
	rec.Set("status", status)
	if erro != "" {
		rec.Set("erro", erro)
	}
	if status == DestStatusFalhou {
		rec.Set("tentativas", rec.GetInt("tentativas")+1)
	}
	if status == DestStatusEnviado {
		rec.Set("enviado_em", time.Now().UTC().Format(time.RFC3339))
	}
	return r.dao.SaveRecord(rec)
}

// SaveDestinatario salva as alterações de um destinatário (ex: após enriquecimento)
func (r *CampanhaRepo) SaveDestinatario(rec *models.Record) error {
	return r.dao.SaveRecord(rec)
}

// CountDestinatariosByStatus conta destinatários por status
func (r *CampanhaRepo) CountDestinatariosByStatus(campanhaID string) (map[string]int, error) {
	destinatarios, err := r.FindDestinatariosByCampanha(campanhaID)
	if err != nil {
		return nil, err
	}

	stats := map[string]int{
		"total":    len(destinatarios),
		"pendente": 0,
		"em_fila":  0,
		"enviado":  0,
		"falhou":   0,
		"ignorado": 0,
	}

	for _, d := range destinatarios {
		status := d.GetString("status")
		switch status {
		case DestStatusPendente:
			stats["pendente"]++
		case DestStatusEmFila:
			stats["em_fila"]++
		case DestStatusEnviado:
			stats["enviado"]++
		case DestStatusFalhou:
			stats["falhou"]++
		case DestStatusIgnorado:
			stats["ignorado"]++
		}
	}

	return stats, nil
}

// CountEnviadosHojeByTeam conta quantas mensagens foram enviadas hoje para um team
func (r *CampanhaRepo) CountEnviadosHojeByTeam(teamID string) (int, error) {
	today := time.Now().UTC().Format("2006-01-02") + " 00:00:00.000Z"
	records, err := r.dao.FindRecordsByFilter(
		"campanha_destinatarios",
		"team_id = {:teamId} && status = {:status} && enviado_em >= {:today}",
		"", 0, 0,
		dbx.Params{"teamId": teamID, "status": DestStatusEnviado, "today": today},
	)
	if err != nil {
		return 0, err
	}
	return len(records), nil
}

// CountRespostas conta as respostas (campanha_respostas) para uma campanha
func (r *CampanhaRepo) CountRespostas(campanhaID string) (int, error) {
	// Verifica se a collection existe (pode não ter sido criada ainda)
	_, err := r.dao.FindCollectionByNameOrId("campanha_respostas")
	if err != nil {
		return 0, nil // Retorna 0 se a tabela ainda não existir
	}

	respostas, err := r.dao.FindRecordsByFilter(
		"campanha_respostas",
		"campanha_id = {:campanhaId} && tipo = 'RESPOSTA'",
		"",
		0, 0,
		dbx.Params{"campanhaId": campanhaID},
	)
	if err != nil {
		return 0, err
	}
	return len(respostas), nil
}

// GetDashboardStats retorna estatísticas agregadas para o dashboard
func (r *CampanhaRepo) GetDashboardStats(teamID string) (map[string]any, error) {
	stats := map[string]any{
		"enviadas":  0,
		"lidas":     0,
		"respostas": 0,
		"campanhas": []map[string]any{},
	}

	// 1. Total enviadas
	var totalEnviadas int
	err := r.dao.DB().NewQuery("SELECT COUNT(*) FROM campanha_destinatarios WHERE team_id={:teamId} AND status='ENVIADO'").
		Bind(dbx.Params{"teamId": teamID}).
		Row(&totalEnviadas)
	if err == nil {
		stats["enviadas"] = totalEnviadas
	}

	// 2. Total respostas
	_, err = r.dao.FindCollectionByNameOrId("campanha_respostas")
	if err == nil {
		var totalRespostas int
		err = r.dao.DB().NewQuery("SELECT COUNT(*) FROM campanha_respostas WHERE team_id={:teamId} AND tipo='RESPOSTA'").
			Bind(dbx.Params{"teamId": teamID}).
			Row(&totalRespostas)
		if err == nil {
			stats["respostas"] = totalRespostas
		}
	}

	// 3. Campanhas recentes (para gráfico)
	type CampanhaStat struct {
		Nome     string `db:"nome"`
		Enviados int    `db:"enviados"`
	}
	var campanhasStats []CampanhaStat

	query := `
		SELECT 
			c.nome, 
			COUNT(cd.id) as enviados
		FROM campanhas c
		LEFT JOIN campanha_destinatarios cd ON c.id = cd.campanha_id AND cd.status = 'ENVIADO'
		WHERE c.team_id = {:teamId}
		GROUP BY c.id, c.nome, c.created
		ORDER BY c.created DESC
		LIMIT 5
	`
	err = r.dao.DB().NewQuery(query).Bind(dbx.Params{"teamId": teamID}).All(&campanhasStats)
	if err == nil {
		campanhasList := []map[string]any{}
		for _, cs := range campanhasStats {
			campanhasList = append(campanhasList, map[string]any{
				"nome":     cs.Nome,
				"enviados": cs.Enviados,
			})
		}
		stats["campanhas"] = campanhasList
	}

	return stats, nil
}

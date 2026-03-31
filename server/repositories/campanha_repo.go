package repositories

import (
	"strings"
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

// ---------------- Campanhas ----------------

// FindByID busca uma campanha pelo ID
func (r *CampanhaRepo) FindByID(id string) (*models.Record, error) {
	return r.dao.FindRecordById("campanhas", id)
}

// FindConexaoByID busca uma conexão pelo ID (para descobrir o canal)
func (r *CampanhaRepo) FindConexaoByID(id string) (*models.Record, error) {
	return r.dao.FindRecordById("conexoes", id)
}

// LEGADO: mantém por compatibilidade caso algum ponto ainda chame.
// O fluxo novo de campanhas não deve depender mais de lead.
func (r *CampanhaRepo) FindLeadByID(id string) (*models.Record, error) {
	return r.dao.FindRecordById("lead", id)
}

// FindByTeam busca campanhas de um team
func (r *CampanhaRepo) FindByTeam(teamID string, limit, offset int) ([]*models.Record, error) {
	return r.dao.FindRecordsByFilter(
		"campanhas",
		"team_id = {:teamId}",
		"-created",
		limit,
		offset,
		dbx.Params{"teamId": teamID},
	)
}

// FindByTeamAndStatus busca campanhas de um team com status específico
func (r *CampanhaRepo) FindByTeamAndStatus(teamID string, status string) ([]*models.Record, error) {
	return r.dao.FindRecordsByFilter(
		"campanhas",
		"team_id = {:teamId} && status = {:status}",
		"-created",
		0,
		0,
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
		0,
		0,
		dbx.Params{"campanhaId": campanhaID},
	)
}

// FindDestinatariosPendentes busca destinatários pendentes de uma campanha
func (r *CampanhaRepo) FindDestinatariosPendentes(campanhaID string) ([]*models.Record, error) {
	return r.dao.FindRecordsByFilter(
		"campanha_destinatarios",
		"campanha_id = {:campanhaId} && status = {:status}",
		"-created",
		0,
		0,
		dbx.Params{
			"campanhaId": campanhaID,
			"status":     DestStatusPendente,
		},
	)
}

// FindDestinatarioByID busca um destinatário pelo ID
func (r *CampanhaRepo) FindDestinatarioByID(id string) (*models.Record, error) {
	return r.dao.FindRecordById("campanha_destinatarios", id)
}

// LEGADO: dedup antigo por campanha + obra + tipo.
// Mantido por compatibilidade, mas o fluxo novo deve usar
// FindDestinatarioByCampanhaContatoValor.
func (r *CampanhaRepo) FindDestinatarioByCampanhaObraContato(campanhaID, obraID, contatoTipo string) (*models.Record, error) {
	records, err := r.dao.FindRecordsByExpr(
		"campanha_destinatarios",
		dbx.NewExp(
			"campanha_id = {:campanhaId} AND obra_id = {:obraId} AND contato_tipo = {:contatoTipo}",
			dbx.Params{
				"campanhaId":  campanhaID,
				"obraId":      obraID,
				"contatoTipo": contatoTipo,
			},
		),
	)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, nil
	}

	return records[0], nil
}

// FindDestinatarioByCampanhaContatoValor verifica se já existe destinatário
// para a mesma campanha/obra/tipo e mesmo contato específico.
//
// Regras:
// - se telefone vier preenchido, deduplica pelo telefone e email vazio
// - se email vier preenchido, deduplica pelo email e telefone vazio
// - se ambos vierem vazios, procura um registro "sem contato"
func (r *CampanhaRepo) FindDestinatarioByCampanhaContatoValor(
	campanhaID, obraID, contatoTipo, telefone, email string,
) (*models.Record, error) {
	params := dbx.Params{
		"campanhaId":  campanhaID,
		"obraId":      obraID,
		"contatoTipo": contatoTipo,
	}

	expr := `
		campanha_id = {:campanhaId}
		AND obra_id = {:obraId}
		AND contato_tipo = {:contatoTipo}
	`

	switch {
	case telefone != "" && email == "":
		expr += ` AND telefone_e164 = {:telefone} AND (email = '' OR email IS NULL)`
		params["telefone"] = telefone

	case email != "" && telefone == "":
		expr += ` AND email = {:email} AND (telefone_e164 = '' OR telefone_e164 IS NULL)`
		params["email"] = email

	case telefone == "" && email == "":
		expr += ` AND (telefone_e164 = '' OR telefone_e164 IS NULL) AND (email = '' OR email IS NULL)`

	default:
		expr += ` AND telefone_e164 = {:telefone} AND email = {:email}`
		params["telefone"] = telefone
		params["email"] = email
	}

	records, err := r.dao.FindRecordsByExpr(
		"campanha_destinatarios",
		dbx.NewExp(expr, params),
	)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, nil
	}

	return records[0], nil
}

// CreateDestinatario cria um destinatário
func (r *CampanhaRepo) CreateDestinatario(data map[string]any) (*models.Record, error) {
	col, err := r.dao.FindCollectionByNameOrId("campanha_destinatarios")
	if err != nil {
		return nil, err
	}

	rec := models.NewRecord(col)

	for k, v := range data {
		rec.Set(k, v)
	}

	if rec.GetString("status") == "" {
		rec.Set("status", DestStatusPendente)
	}

	if rec.Get("tentativas") == nil {
		rec.Set("tentativas", 0)
	}

	if rec.GetString("team_id") == "" && rec.GetString("campanha_id") != "" {
		campanha, err := r.FindByID(rec.GetString("campanha_id"))
		if err == nil && campanha != nil {
			rec.Set("team_id", campanha.GetString("team_id"))
		}
	}

	if err := r.dao.SaveRecord(rec); err != nil {
		return nil, err
	}

	return rec, nil
}

// ExistsContatoEnviado verifica se já houve envio anterior para a mesma obra/tipo no time
func (r *CampanhaRepo) ExistsContatoEnviado(teamID, obraID, contatoTipo string) (bool, error) {
	recs, err := r.dao.FindRecordsByFilter(
		"campanha_destinatarios",
		"team_id = {:team_id} && obra_id = {:obra_id} && contato_tipo = {:contato_tipo} && status = {:status}",
		"-enviado_em,-updated",
		1,
		0,
		dbx.Params{
			"team_id":      teamID,
			"obra_id":      obraID,
			"contato_tipo": contatoTipo,
			"status":       DestStatusEnviado,
		},
	)
	if err != nil {
		return false, err
	}

	return len(recs) > 0, nil
}

// UpdateDestinatarioStatus atualiza o status de um destinatário
func (r *CampanhaRepo) UpdateDestinatarioStatus(dest *models.Record, status, errMsg string) error {
	if dest == nil {
		return nil
	}

	dest.Set("status", status)
	dest.Set("erro", strings.TrimSpace(errMsg))

	if status == DestStatusEnviado {
		dest.Set("enviado_em", time.Now().UTC())
	}

	if status == DestStatusEnviado || status == DestStatusFalhou {
		dest.Set("tentativas", dest.GetInt("tentativas")+1)
	}

	return r.dao.SaveRecord(dest)
}

// SaveDestinatario salva as alterações de um destinatário
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
		switch d.GetString("status") {
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

// CountEnviadosHojeByTeam conta quantas mensagens foram enviadas hoje para um team.
// Aqui usa JOIN com campanhas para não depender 100% do team_id denormalizado
// em campanha_destinatarios.
func (r *CampanhaRepo) CountEnviadosHojeByTeam(teamID string) (int, error) {
	today := time.Now().UTC().Format("2006-01-02T00:00:00Z")

	var total int
	query := `
		SELECT COUNT(cd.id)
		FROM campanha_destinatarios cd
		INNER JOIN campanhas c ON c.id = cd.campanha_id
		WHERE c.team_id = {:teamId}
		  AND cd.status = {:status}
		  AND cd.enviado_em >= {:today}
	`

	err := r.dao.DB().
		NewQuery(query).
		Bind(dbx.Params{
			"teamId": teamID,
			"status": DestStatusEnviado,
			"today":  today,
		}).
		Row(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// CountRespostas conta mensagens recebidas do lead (webhook) ligadas à campanha.
// Fonte: collection campanha_lead_respostas.
func (r *CampanhaRepo) CountRespostas(campanhaID string) (int, error) {
	var totalRespostas int
	query := `
		SELECT COUNT(clr.id)
		FROM campanha_lead_respostas clr
		WHERE clr.campanha_id = {:campanhaId}
	`

	err := r.dao.DB().
		NewQuery(query).
		Bind(dbx.Params{"campanhaId": campanhaID}).
		Row(&totalRespostas)
	if err != nil {
		return 0, err
	}

	return totalRespostas, nil
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
	queryEnviadas := `
		SELECT COUNT(cd.id)
		FROM campanha_destinatarios cd
		INNER JOIN campanhas c ON c.id = cd.campanha_id
		WHERE c.team_id = {:teamId}
		  AND cd.status = 'ENVIADO'
	`
	err := r.dao.DB().
		NewQuery(queryEnviadas).
		Bind(dbx.Params{"teamId": teamID}).
		Row(&totalEnviadas)
	if err == nil {
		stats["enviadas"] = totalEnviadas
	}

	// 2. Total respostas (mensagens recebidas — campanha_lead_respostas)
	var totalRespostas int
	queryRespostas := `
		SELECT COUNT(clr.id)
		FROM campanha_lead_respostas clr
		INNER JOIN campanhas c ON c.id = clr.campanha_id
		WHERE c.team_id = {:teamId}
	`
	err = r.dao.DB().
		NewQuery(queryRespostas).
		Bind(dbx.Params{"teamId": teamID}).
		Row(&totalRespostas)
	if err == nil {
		stats["respostas"] = totalRespostas
	}

	// 3. Campanhas recentes (para gráfico)
	type CampanhaStat struct {
		Nome     string `db:"nome"`
		Enviados int    `db:"enviados"`
	}

	var campanhasStats []CampanhaStat
	queryCampanhas := `
		SELECT
			c.nome,
			COUNT(cd.id) as enviados
		FROM campanhas c
		LEFT JOIN campanha_destinatarios cd
			ON c.id = cd.campanha_id
		   AND cd.status = 'ENVIADO'
		WHERE c.team_id = {:teamId}
		GROUP BY c.id, c.nome, c.created
		ORDER BY c.created DESC
		LIMIT 5
	`

	err = r.dao.DB().
		NewQuery(queryCampanhas).
		Bind(dbx.Params{"teamId": teamID}).
		All(&campanhasStats)
	if err == nil {
		campanhasList := make([]map[string]any, 0, len(campanhasStats))
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

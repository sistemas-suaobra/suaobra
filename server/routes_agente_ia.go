package server

import (
	"github.com/flarco/g"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
)

// ═══════════════════════════════════════════════════════════════════════════
// AGENTE IA ROUTES
// ═══════════════════════════════════════════════════════════════════════════

func RegisterAgenteIARoutes(e *echo.Echo) {
	e.GET("/agente-ia/intencoes", ListarIntencoes)
	e.POST("/agente-ia/intencoes", CriarIntencao)
	e.PATCH("/agente-ia/intencoes/:id", AtualizarIntencao)
	e.DELETE("/agente-ia/intencoes/:id", DeletarIntencao)

	e.GET("/agente-ia/conversas", ListarConversas)
	e.GET("/agente-ia/conversas/:id", ObterConversa)
	e.PATCH("/agente-ia/conversas/:id/status", AtualizarStatusConversa)
}

func getAuthTeamID(c echo.Context) (*pocketbase.PocketBase, string, error) {
	app := c.Get("app").(*pocketbase.PocketBase)

	req := NewRequest(c)
	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" || user.Team.ID == "" {
		return app, "", g.Error("unauthorized")
	}

	return app, user.Team.ID, nil
}

func findTeamScopedRecord(app *pocketbase.PocketBase, collectionName, id, teamID string) (*models.Record, error) {
	return app.Dao().FindFirstRecordByFilter(
		collectionName,
		"id = {:id} && team_id = {:teamId}",
		dbx.Params{
			"id":     id,
			"teamId": teamID,
		},
	)
}

// GET /agente-ia/intencoes
// Lista apenas as intenções da team do usuário autenticado
func ListarIntencoes(c echo.Context) error {
	app, teamID, err := getAuthTeamID(c)
	if err != nil {
		return c.JSON(401, g.M("error", "unauthorized"))
	}

	intencoes, err := app.Dao().FindRecordsByFilter(
		"agente_ia_intencoes",
		"team_id = {:teamId}",
		"-prioridade,-created",
		0,
		0,
		dbx.Params{
			"teamId": teamID,
		},
	)
	if err != nil {
		g.Error(err, "erro ao listar intenções")
		return c.JSON(500, g.M("error", "erro ao listar intenções"))
	}

	result := make([]map[string]any, 0, len(intencoes))
	for _, record := range intencoes {
		result = append(result, record.PublicExport())
	}

	return c.JSON(200, g.M("intencoes", result))
}

// POST /agente-ia/intencoes
// Cria uma nova intenção para a team do usuário autenticado
func CriarIntencao(c echo.Context) error {
	app, teamID, err := getAuthTeamID(c)
	if err != nil {
		return c.JSON(401, g.M("error", "unauthorized"))
	}

	var payload struct {
		Nome          string   `json:"nome"`
		Descricao     string   `json:"descricao"`
		PalavrasChave []string `json:"palavras_chave"`
		Resposta      string   `json:"resposta"`
		Ativa         bool     `json:"ativa"`
		Prioridade    int      `json:"prioridade"`
	}

	if err := c.Bind(&payload); err != nil {
		return c.JSON(400, g.M("error", "invalid request body"))
	}

	if payload.Nome == "" || len(payload.PalavrasChave) == 0 || payload.Resposta == "" {
		return c.JSON(400, g.M("error", "nome, palavras_chave e resposta são obrigatórios"))
	}

	collection, err := app.Dao().FindCollectionByNameOrId("agente_ia_intencoes")
	if err != nil {
		return c.JSON(500, g.M("error", "collection não encontrada"))
	}

	record := models.NewRecord(collection)
	record.Set("team_id", teamID)
	record.Set("nome", payload.Nome)
	record.Set("descricao", payload.Descricao)
	record.Set("palavras_chave", payload.PalavrasChave)
	record.Set("resposta", payload.Resposta)
	record.Set("ativa", payload.Ativa)

	prioridade := payload.Prioridade
	if prioridade <= 0 {
		prioridade = 50
	}
	record.Set("prioridade", prioridade)

	if err := app.Dao().SaveRecord(record); err != nil {
		g.Error(err, "erro ao criar intenção")
		return c.JSON(500, g.M("error", "erro ao criar intenção"))
	}

	return c.JSON(201, g.M("intencao", record.PublicExport()))
}

// PATCH /agente-ia/intencoes/:id
// Atualiza uma intenção da própria team
func AtualizarIntencao(c echo.Context) error {
	app, teamID, err := getAuthTeamID(c)
	if err != nil {
		return c.JSON(401, g.M("error", "unauthorized"))
	}

	id := c.PathParam("id")
	if id == "" {
		return c.JSON(400, g.M("error", "id é obrigatório"))
	}

	record, err := findTeamScopedRecord(app, "agente_ia_intencoes", id, teamID)
	if err != nil || record == nil {
		return c.JSON(404, g.M("error", "intenção não encontrada"))
	}

	var payload struct {
		Nome          *string   `json:"nome"`
		Descricao     *string   `json:"descricao"`
		PalavrasChave *[]string `json:"palavras_chave"`
		Resposta      *string   `json:"resposta"`
		Ativa         *bool     `json:"ativa"`
		Prioridade    *int      `json:"prioridade"`
	}

	if err := c.Bind(&payload); err != nil {
		return c.JSON(400, g.M("error", "invalid request body"))
	}

	if payload.Nome != nil {
		record.Set("nome", *payload.Nome)
	}
	if payload.Descricao != nil {
		record.Set("descricao", *payload.Descricao)
	}
	if payload.PalavrasChave != nil {
		record.Set("palavras_chave", *payload.PalavrasChave)
	}
	if payload.Resposta != nil {
		record.Set("resposta", *payload.Resposta)
	}
	if payload.Ativa != nil {
		record.Set("ativa", *payload.Ativa)
	}
	if payload.Prioridade != nil {
		record.Set("prioridade", *payload.Prioridade)
	}

	if err := app.Dao().SaveRecord(record); err != nil {
		g.Error(err, "erro ao atualizar intenção")
		return c.JSON(500, g.M("error", "erro ao atualizar intenção"))
	}

	return c.JSON(200, g.M("intencao", record.PublicExport()))
}

// DELETE /agente-ia/intencoes/:id
// Deleta uma intenção da própria team
func DeletarIntencao(c echo.Context) error {
	app, teamID, err := getAuthTeamID(c)
	if err != nil {
		return c.JSON(401, g.M("error", "unauthorized"))
	}

	id := c.PathParam("id")
	if id == "" {
		return c.JSON(400, g.M("error", "id é obrigatório"))
	}

	record, err := findTeamScopedRecord(app, "agente_ia_intencoes", id, teamID)
	if err != nil || record == nil {
		return c.JSON(404, g.M("error", "intenção não encontrada"))
	}

	if err := app.Dao().DeleteRecord(record); err != nil {
		g.Error(err, "erro ao deletar intenção")
		return c.JSON(500, g.M("error", "erro ao deletar intenção"))
	}

	return c.JSON(200, g.M("success", true))
}

// GET /agente-ia/conversas
// Lista conversas apenas da team do usuário autenticado
func ListarConversas(c echo.Context) error {
	app, teamID, err := getAuthTeamID(c)
	if err != nil {
		return c.JSON(401, g.M("error", "unauthorized"))
	}

	status := c.QueryParam("status")

	filter := "team_id = {:teamId}"
	params := dbx.Params{
		"teamId": teamID,
	}

	if status != "" {
		filter += " && status = {:status}"
		params["status"] = status
	}

	conversas, err := app.Dao().FindRecordsByFilter(
		"conversas_ia",
		filter,
		"-ultima_mensagem_em",
		100,
		0,
		params,
	)
	if err != nil {
		g.Error(err, "erro ao listar conversas")
		return c.JSON(500, g.M("error", "erro ao listar conversas"))
	}

	result := make([]map[string]any, 0, len(conversas))
	for _, record := range conversas {
		result = append(result, record.PublicExport())
	}

	return c.JSON(200, g.M("conversas", result))
}

// GET /agente-ia/conversas/:id
// Obtém uma conversa específica da própria team
func ObterConversa(c echo.Context) error {
	app, teamID, err := getAuthTeamID(c)
	if err != nil {
		return c.JSON(401, g.M("error", "unauthorized"))
	}

	id := c.PathParam("id")
	if id == "" {
		return c.JSON(400, g.M("error", "id é obrigatório"))
	}

	record, err := findTeamScopedRecord(app, "conversas_ia", id, teamID)
	if err != nil || record == nil {
		return c.JSON(404, g.M("error", "conversa não encontrada"))
	}

	return c.JSON(200, g.M("conversa", record.PublicExport()))
}

// PATCH /agente-ia/conversas/:id/status
// Atualiza o status de uma conversa da própria team
func AtualizarStatusConversa(c echo.Context) error {
	app, teamID, err := getAuthTeamID(c)
	if err != nil {
		return c.JSON(401, g.M("error", "unauthorized"))
	}

	id := c.PathParam("id")
	if id == "" {
		return c.JSON(400, g.M("error", "id é obrigatório"))
	}

	var payload struct {
		Status string `json:"status"`
	}
	if err := c.Bind(&payload); err != nil {
		return c.JSON(400, g.M("error", "invalid request body"))
	}

	validStatus := map[string]bool{
		"ATIVA":      true,
		"PAUSADA":    true,
		"FINALIZADA": true,
	}
	if !validStatus[payload.Status] {
		return c.JSON(400, g.M("error", "status inválido (use ATIVA, PAUSADA ou FINALIZADA)"))
	}

	record, err := findTeamScopedRecord(app, "conversas_ia", id, teamID)
	if err != nil || record == nil {
		return c.JSON(404, g.M("error", "conversa não encontrada"))
	}

	record.Set("status", payload.Status)

	if err := app.Dao().SaveRecord(record); err != nil {
		g.Error(err, "erro ao atualizar status da conversa")
		return c.JSON(500, g.M("error", "erro ao atualizar status"))
	}

	return c.JSON(200, g.M("conversa", record.PublicExport()))
}
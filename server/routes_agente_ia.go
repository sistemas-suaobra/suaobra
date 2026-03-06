package server

import (
	"github.com/flarco/g"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
)

// ═══════════════════════════════════════════════════════════════════════════
// AGENTE IA ROUTES
// ═══════════════════════════════════════════════════════════════════════════

// RegisterAgenteIARoutes registra todas as rotas de agente IA
func RegisterAgenteIARoutes(e *echo.Echo) {
	e.GET("/agente-ia/intencoes", ListarIntencoes)
	e.POST("/agente-ia/intencoes", CriarIntencao)
	e.PATCH("/agente-ia/intencoes/:id", AtualizarIntencao)
	e.DELETE("/agente-ia/intencoes/:id", DeletarIntencao)

	e.GET("/agente-ia/conversas", ListarConversas)
	e.GET("/agente-ia/conversas/:id", ObterConversa)
	e.PATCH("/agente-ia/conversas/:id/status", AtualizarStatusConversa)
}

// GET /agente-ia/intencoes
// Lista todas as intenções (sem autenticação por enquanto)
func ListarIntencoes(c echo.Context) error {
	app := c.Get("app").(*pocketbase.PocketBase)
	
	// Busca todas as intenções
	intencoes, err := app.Dao().FindRecordsByFilter(
		"agente_ia_intencoes",
		"1=1",
		"-prioridade",
		0,
		0,
	)
	if err != nil {
		g.Error(err, "erro ao listar intenções")
		return c.JSON(500, g.M("error", "erro ao listar intenções"))
	}

	result := []map[string]any{}
	for _, record := range intencoes {
		result = append(result, record.PublicExport())
	}

	return c.JSON(200, g.M("intencoes", result))
}

// POST /agente-ia/intencoes
// Cria uma nova intenção (sem autenticação por enquanto)
func CriarIntencao(c echo.Context) error {
	app := c.Get("app").(*pocketbase.PocketBase)

	var payload struct {
		Nome          string   `json:"nome"`
		Descricao     string   `json:"descricao"`
		PalavrasChave []string `json:"palavras_chave"`
		Resposta      string   `json:"resposta"`
		Ativa         bool     `json:"ativa"`
		Prioridade    int      `json:"prioridade"`
		TeamID        string   `json:"team_id"`
	}

	if err := c.Bind(&payload); err != nil {
		return c.JSON(400, g.M("error", "invalid request body"))
	}

	if payload.Nome == "" || len(payload.PalavrasChave) == 0 || payload.Resposta == "" {
		return c.JSON(400, g.M("error", "nome, palavras_chave e resposta são obrigatórios"))
	}

	// Team ID padrão se não fornecido
	if payload.TeamID == "" {
		payload.TeamID = "team_zlvE6stSf3IB2wO" // TODO: pegar do contexto
	}

	// Cria record
	collection, err := app.Dao().FindCollectionByNameOrId("agente_ia_intencoes")
	if err != nil {
		return c.JSON(500, g.M("error", "collection não encontrada"))
	}

	record := models.NewRecord(collection)
	record.Set("team_id", payload.TeamID)
	record.Set("nome", payload.Nome)
	record.Set("descricao", payload.Descricao)
	record.Set("palavras_chave", payload.PalavrasChave)
	record.Set("resposta", payload.Resposta)
	record.Set("ativa", payload.Ativa)

	prioridade := payload.Prioridade
	if prioridade <= 0 {
		prioridade = 50 // default medium priority
	}
	record.Set("prioridade", prioridade)

	if err := app.Dao().SaveRecord(record); err != nil {
		g.Error(err, "erro ao criar intenção")
		return c.JSON(500, g.M("error", "erro ao criar intenção"))
	}

	return c.JSON(201, g.M("intencao", record.PublicExport()))
}

// PATCH /agente-ia/intencoes/:id
// Atualiza uma intenção existente (sem autenticação por enquanto)
func AtualizarIntencao(c echo.Context) error {
	app := c.Get("app").(*pocketbase.PocketBase)

	id := c.PathParam("id")
	if id == "" {
		return c.JSON(400, g.M("error", "id é obrigatório"))
	}

	record, err := app.Dao().FindRecordById("agente_ia_intencoes", id)
	if err != nil {
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

	// Atualiza apenas campos fornecidos
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
// Deleta uma intenção (sem autenticação por enquanto)
func DeletarIntencao(c echo.Context) error {
	app := c.Get("app").(*pocketbase.PocketBase)
	
	id := c.PathParam("id")
	if id == "" {
		return c.JSON(400, g.M("error", "id é obrigatório"))
	}

	record, err := app.Dao().FindRecordById("agente_ia_intencoes", id)
	if err != nil {
		return c.JSON(404, g.M("error", "intenção não encontrada"))
	}

	if err := app.Dao().DeleteRecord(record); err != nil {
		g.Error(err, "erro ao deletar intenção")
		return c.JSON(500, g.M("error", "erro ao deletar intenção"))
	}

	return c.JSON(200, g.M("success", true))
}

// GET /agente-ia/conversas
// Lista conversas de IA da team
func ListarConversas(c echo.Context) error {
	app := c.Get("app").(*pocketbase.PocketBase)

	status := c.QueryParam("status")
	teamID := c.QueryParam("team_id")

	filter := "1=1"
	params := map[string]any{}

	if teamID != "" {
		filter += " && team_id = {:teamId}"
		params["teamId"] = teamID
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

	result := []map[string]any{}
	for _, record := range conversas {
		result = append(result, record.PublicExport())
	}

	return c.JSON(200, g.M("conversas", result))
}

// GET /agente-ia/conversas/:id
// Obtém uma conversa específica com todo o histórico
func ObterConversa(c echo.Context) error {
	app := c.Get("app").(*pocketbase.PocketBase)

	id := c.PathParam("id")
	if id == "" {
		return c.JSON(400, g.M("error", "id é obrigatório"))
	}

	record, err := app.Dao().FindRecordById("conversas_ia", id)
	if err != nil {
		return c.JSON(404, g.M("error", "conversa não encontrada"))
	}

	return c.JSON(200, g.M("conversa", record.PublicExport()))
}

// PATCH /agente-ia/conversas/:id/status
// Atualiza o status de uma conversa (ATIVA, PAUSADA, FINALIZADA)
func AtualizarStatusConversa(c echo.Context) error {
	app := c.Get("app").(*pocketbase.PocketBase)

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

	validStatus := map[string]bool{"ATIVA": true, "PAUSADA": true, "FINALIZADA": true}
	if !validStatus[payload.Status] {
		return c.JSON(400, g.M("error", "status inválido (use ATIVA, PAUSADA ou FINALIZADA)"))
	}

	record, err := app.Dao().FindRecordById("conversas_ia", id)
	if err != nil {
		return c.JSON(404, g.M("error", "conversa não encontrada"))
	}

	record.Set("status", payload.Status)

	if err := app.Dao().SaveRecord(record); err != nil {
		g.Error(err, "erro ao atualizar status da conversa")
		return c.JSON(500, g.M("error", "erro ao atualizar status"))
	}

	return c.JSON(200, g.M("conversa", record.PublicExport()))
}

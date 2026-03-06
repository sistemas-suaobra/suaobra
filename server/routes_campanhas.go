package server

import (
	"errors"
	"fmt"
	"github.com/flarco/g"
	"github.com/labstack/echo/v5"

	"github.com/suaobra/suaobra-app/server/clients/wuzapi"
	"github.com/suaobra/suaobra-app/server/config"
	"github.com/suaobra/suaobra-app/server/repositories"
	"github.com/suaobra/suaobra-app/server/services"
	"time"
)

// helper: monta o CampanhaService por request
func newCampanhaService(req Request) *services.CampanhaService {
	campRepo := repositories.NewCampanhaRepo(req.Dao())
	waSvc := newWhatsAppServiceFromRequest(req)

	conRepo := repositories.NewConexaoRepo(req.Dao())
	emailRepo := repositories.NewEmailRepo(req.Dao())
	emailSvc := services.NewEmailService(conRepo, emailRepo)

	conversaRepo := repositories.NewConversaRepo(req.Dao())

	return services.NewCampanhaService(campRepo, waSvc, emailSvc, conversaRepo)
}

// helper interno: monta WhatsAppService
func newWhatsAppServiceFromRequest(req Request) *services.WhatsAppService {
	cfg := config.NewWhatsMeowConfig()
	wuzClient := wuzapi.NewClient(cfg)

	conRepo := repositories.NewConexaoRepo(req.Dao())
	waRepo := repositories.NewWhatsAppRepo(req.Dao())
	tokenSvc := services.NewTokenService()

	return services.NewWhatsAppService(conRepo, waRepo, wuzClient, tokenSvc)
}

// POST /campanhas/:id/iniciar - Inicia uma campanha
func IniciarCampanha(c echo.Context) error {
	req := NewRequest(c)
	campanhaID := c.PathParam("id")

	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	svc := newCampanhaService(req)

	if err := svc.IniciarCampanha(user.Team.ID, campanhaID); err != nil {
		return ErrJSON(400, err)
	}

	return c.JSON(200, g.M("success", true, "message", "Campanha iniciada. O envio está sendo processado."))
}

// PATCH /campanha-destinatarios/marcar-enviado - Marca destinatários como ENVIADO
func MarcarDestinatariosEnviados(c echo.Context) error {
	req := NewRequest(c)
	var body struct {
		CampanhaID string   `json:"campanha_id"`
		LeadIDs    []string `json:"lead_ids"`
	}
	if err := c.Bind(&body); err != nil {
		return ErrJSON(400, err, "Erro ao ler body")
	}
	if body.CampanhaID == "" || len(body.LeadIDs) == 0 {
		return ErrJSON(400, errors.New("Informe campanha_id e lead_ids"))
	}

	da := req.Dao()
	for _, leadID := range body.LeadIDs {
		filter := fmt.Sprintf(`{"campanha_id": "%s", "lead_id": "%s"}`, body.CampanhaID, leadID)
		dest, err := da.FindFirstRecordByFilter("campanha_destinatarios", filter)
		if err != nil || dest == nil {
			continue
		}
		dest.Set("status", "ENVIADO")
		dest.Set("enviado_em", time.Now().UTC())
		da.SaveRecord(dest)
	}
	return c.JSON(200, g.M("success", true))
}

// GET /campanhas/:id/status - Retorna status detalhado de uma campanha
func StatusCampanha(c echo.Context) error {
	req := NewRequest(c)
	campanhaID := c.PathParam("id")

	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	svc := newCampanhaService(req)

	campanha, stats, err := svc.GetStatus(user.Team.ID, campanhaID)
	if err != nil {
		return ErrJSON(400, err)
	}

	return c.JSON(200, g.M(
		"campanha", campanha.PublicExport(),
		"stats", stats,
	))
}

// POST /campanhas/:id/pausar - Pausa uma campanha em andamento
func PausarCampanha(c echo.Context) error {
	req := NewRequest(c)
	campanhaID := c.PathParam("id")

	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	svc := newCampanhaService(req)

	if err := svc.PausarCampanha(user.Team.ID, campanhaID); err != nil {
		return ErrJSON(400, err)
	}

	return c.JSON(200, g.M("success", true))
}

// POST /campanhas/:id/cancelar - Cancela uma campanha
func CancelarCampanha(c echo.Context) error {
	req := NewRequest(c)
	campanhaID := c.PathParam("id")

	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	svc := newCampanhaService(req)

	if err := svc.CancelarCampanha(user.Team.ID, campanhaID); err != nil {
		return ErrJSON(400, err)
	}

	return c.JSON(200, g.M("success", true))
}

// POST /campanhas/:id/enriquecer - Preenche nome_contato, telefone_e164 e email dos destinatários
func EnriquecerDestinatarios(c echo.Context) error {
	req := NewRequest(c)
	campanhaID := c.PathParam("id")

	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	svc := newCampanhaService(req)

	enriquecidos, semContato, err := svc.EnriquecerDestinatarios(user.Team.ID, campanhaID)
	if err != nil {
		return ErrJSON(400, err)
	}

	return c.JSON(200, g.M(
		"success", true,
		"enriquecidos", enriquecidos,
		"sem_contato", semContato,
	))
}

// POST /campanhas/gerar-mensagem-ia - Gera uma mensagem de campanha com IA
func GerarMensagemCampanhaIA(c echo.Context) error {
	req := NewRequest(c)
	var body struct {
		Objetivo    string  `json:"objetivo"`
		Temperatura float64 `json:"temperatura"`
	}
	if err := c.Bind(&body); err != nil {
		return ErrJSON(400, err, "Erro ao ler body")
	}
	if body.Objetivo == "" {
		return ErrJSON(400, errors.New("O campo 'objetivo' é obrigatório"))
	}

	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	geminiSvc := services.NewGeminiService()
	mensagem, err := geminiSvc.GenerateCampaignMessage(body.Objetivo, body.Temperatura)
	if err != nil {
		g.LogError(err, "Erro ao gerar mensagem com IA para o team "+user.Team.ID)
		return ErrJSON(500, err, "Erro ao gerar mensagem com IA")
	}

	return c.JSON(200, g.M("mensagem", mensagem))
}

// GET /campanhas/dashboard - Retorna estatísticas agregadas
func CampanhasDashboard(c echo.Context) error {
	req := NewRequest(c)
	g.Info("CampanhasDashboard", "user_id", req.UserID(), "auth_error", req.Error)
	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" {
		g.Error(err, "CampanhasDashboard: unauthorized", "user_id", req.UserID())
		return ErrJSON(401, g.Error("unauthorized"))
	}

	svc := newCampanhaService(req)
	stats, err := svc.GetDashboardStats(user.Team.ID)
	if err != nil {
		return ErrJSON(400, err)
	}

	return c.JSON(200, g.M("stats", stats))
}

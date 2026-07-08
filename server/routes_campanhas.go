package server

import (
	"errors"
	"fmt"
	"time"

	"github.com/flarco/g"
	"github.com/labstack/echo/v5"

	"github.com/suaobra/suaobra-app/server/clients/wuzapi"
	"github.com/suaobra/suaobra-app/server/config"
	"github.com/suaobra/suaobra-app/server/repositories"
	"github.com/suaobra/suaobra-app/server/services"
)

func newCampanhaService(req Request) *services.CampanhaService {
	campRepo := repositories.NewCampanhaRepo(req.Dao())
	waSvc := newWhatsAppServiceFromRequest(req)

	conRepo := repositories.NewConexaoRepo(req.Dao())
	emailRepo := repositories.NewEmailRepo(req.Dao())
	emailSvc := services.NewEmailService(conRepo, emailRepo)

	conversaRepo := repositories.NewConversaRepo(req.Dao())

	return services.NewCampanhaService(campRepo, waSvc, emailSvc, conversaRepo)
}

func newWhatsAppServiceFromRequest(req Request) *services.WhatsAppService {
	cfg := config.NewWhatsMeowConfig()
	wuzClient := wuzapi.NewClient(cfg)

	conRepo := repositories.NewConexaoRepo(req.Dao())
	waRepo := repositories.NewWhatsAppRepo(req.Dao())
	tokenSvc := services.NewTokenService()

	return services.NewWhatsAppService(conRepo, waRepo, wuzClient, tokenSvc)
}

// POST /campanhas/:id/destinatarios/obras-plus
func AdicionarDestinatariosObrasPlus(c echo.Context) error {
	req := NewRequest(c)
	campanhaID := c.PathParam("id")

	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	var body struct {
		Destinatarios           []services.CampanhaDestinatarioInput `json:"destinatarios"`
		OcultarJaContactados    bool                                 `json:"ocultar_ja_contactados"`
		OcultarJaContactadosAlt bool                                 `json:"ocultarJaContactados"`
	}

	if err := c.Bind(&body); err != nil {
		return ErrJSON(400, err, "Erro ao ler body")
	}

	if len(body.Destinatarios) == 0 {
		return ErrJSON(400, errors.New("Informe os destinatários"))
	}

	if len(body.Destinatarios) > 50 {
		return ErrJSON(400, errors.New("O limite de disparos por campanha é de 50 leads"))
	}

	// Mantemos as duas chaves por compatibilidade de payload.
	// No fluxo atual de campanha, os destinatários chegam por seleção manual e
	// não são mais filtrados novamente por histórico de contato nesta etapa.
	ocultarJaContactados := body.OcultarJaContactados || body.OcultarJaContactadosAlt

	svc := newCampanhaService(req)

	criados, ignorados, err := svc.AdicionarDestinatariosObrasPlus(
		user.Team.ID,
		campanhaID,
		body.Destinatarios,
		ocultarJaContactados,
	)
	if err != nil {
		return ErrJSON(400, err)
	}

	resp := g.M(
		"success", true,
		"criados", criados,
		"ignorados", ignorados,
	)
	if criados == 0 && ignorados > 0 {
		resp["motivo"] = "Não foi possível obter telefone ou e-mail válido para os canais selecionados. Verifique se o contato possui WhatsApp/e-mail cadastrado ou tente incluir E-mail como canal."
	}

	return c.JSON(200, resp)
}

// POST /campanhas/:id/iniciar
func IniciarCampanha(c echo.Context) error {
	req := NewRequest(c)
	campanhaID := c.PathParam("id")

	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	repo := repositories.NewCampanhaRepo(req.Dao())
	svc := newCampanhaService(req)

	campanha, err := repo.FindByID(campanhaID)
	if err != nil {
		return ErrJSON(404, g.Error(err, "campanha não encontrada"))
	}

	if campanha.GetString("team_id") != user.Team.ID {
		return ErrJSON(403, g.Error("não autorizado"))
	}

	statusAtual := campanha.GetString("status")
	if statusAtual == repositories.CampanhaStatusEmAndamento {
		return ErrJSON(400, g.Error("campanha já está em andamento"))
	}

	if statusAtual == repositories.CampanhaStatusCancelada {
		return ErrJSON(400, g.Error("campanha cancelada não pode ser iniciada"))
	}

	if err := repo.UpdateStatus(campanha, repositories.CampanhaStatusEmAndamento); err != nil {
		return ErrJSON(400, g.Error(err, "erro ao iniciar campanha"))
	}

	go svc.ProcessarCampanhaAsync(campanhaID)

	return c.JSON(200, g.M(
		"success", true,
		"message", "Campanha iniciada. O envio está sendo processado.",
	))
}

// PATCH /campanha-destinatarios/marcar-enviado
// Agora trabalha com destinatario_ids.
// Mantive fallback de lead_ids só para não quebrar legado de uma vez.
func MarcarDestinatariosEnviados(c echo.Context) error {
	req := NewRequest(c)

	var body struct {
		CampanhaID      string   `json:"campanha_id"`
		DestinatarioIDs []string `json:"destinatario_ids"`
		LeadIDs         []string `json:"lead_ids"` // legado
	}

	if err := c.Bind(&body); err != nil {
		return ErrJSON(400, err, "Erro ao ler body")
	}

	if body.CampanhaID == "" {
		return ErrJSON(400, errors.New("Informe campanha_id"))
	}

	da := req.Dao()
	repo := repositories.NewCampanhaRepo(da)

	// fluxo novo
	if len(body.DestinatarioIDs) > 0 {
		for _, destID := range body.DestinatarioIDs {
			dest, err := repo.FindDestinatarioByID(destID)
			if err != nil || dest == nil {
				continue
			}

			if dest.GetString("campanha_id") != body.CampanhaID {
				continue
			}

			dest.Set("status", repositories.DestStatusEnviado)
			dest.Set("enviado_em", time.Now().UTC())
			_ = da.SaveRecord(dest)
		}

		return c.JSON(200, g.M("success", true))
	}

	// fallback legado
	if len(body.LeadIDs) == 0 {
		return ErrJSON(400, errors.New("Informe destinatario_ids"))
	}

	for _, leadID := range body.LeadIDs {
		filter := fmt.Sprintf(`campanha_id = "%s" && lead_id = "%s"`, body.CampanhaID, leadID)
		dest, err := da.FindFirstRecordByFilter("campanha_destinatarios", filter)
		if err != nil || dest == nil {
			continue
		}
		dest.Set("status", repositories.DestStatusEnviado)
		dest.Set("enviado_em", time.Now().UTC())
		_ = da.SaveRecord(dest)
	}

	return c.JSON(200, g.M("success", true))
}

// GET /campanhas/:id/status
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

// POST /campanhas/:id/pausar
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

// POST /campanhas/:id/cancelar
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

// POST /campanhas/:id/enriquecer
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

// POST /campanhas/gerar-mensagem-ia
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

// GET /campanhas/dashboard
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

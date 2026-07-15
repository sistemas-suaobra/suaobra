package server

import (
	"strings"

	"github.com/flarco/g"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/spf13/cast"

	"github.com/suaobra/suaobra-app/server/clients/wuzapi"
	"github.com/suaobra/suaobra-app/server/config"
	"github.com/suaobra/suaobra-app/server/repositories"
	"github.com/suaobra/suaobra-app/server/services"
)

func newWhatsAppService(req Request) *services.WhatsAppService {
	return newWhatsAppServiceFromDao(req.Dao())
}

func newWhatsAppServiceFromDao(dao *daos.Dao) *services.WhatsAppService {
	cfg := config.NewWhatsMeowConfig()
	wuzClient := wuzapi.NewClient(cfg)
	conRepo := repositories.NewConexaoRepo(dao)
	waRepo := repositories.NewWhatsAppRepo(dao)
	tokenSvc := services.NewTokenService()
	return services.NewWhatsAppService(conRepo, waRepo, wuzClient, tokenSvc)
}

// SyncWebhookURLs atualiza a URL de webhook em todos os usuários wuzapi (boot).
func SyncWebhookURLs(app *pocketbase.PocketBase) {
	newWhatsAppServiceFromDao(app.Dao()).SyncWebhookURLs()
}

// POST /conexoes/whatsapp
func CriarConexaoWhatsapp(c echo.Context) error {
	req := NewRequest(c)

	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	var payload struct {
		Nome string `json:"nome"`
	}
	_ = c.Bind(&payload)

	cfg := config.NewWhatsMeowConfig()
	svc := newWhatsAppService(req)

	already, con, wa, err := svc.CreateConnection(
		user.Team.ID,
		user.ID,
		payload.Nome,
		cfg.APIKey,
	)
	if err != nil {
		return ErrJSON(502, err)
	}

	status := 201
	if already {
		status = 200
	}

	return c.JSON(status, g.M(
		"alreadyExists", already,
		"conexao", con.PublicExport(),
		"whatsapp", wa.PublicExport(),
	))
}

// POST /conexoes/whatsapp/connect
func ConectarSessaoWhatsapp(c echo.Context) error {
	req := NewRequest(c)

	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	svc := newWhatsAppService(req)

	raw, err := svc.ConnectByUser(user.Team.ID, user.ID)
	if err != nil {
		return ErrJSON(502, err, "error connecting session")
	}

	return c.JSON(200, g.M("success", true, "raw", raw))
}

// GET /conexoes/whatsapp/qr
func ObterQRCodeWhatsapp(c echo.Context) error {
	req := NewRequest(c)

	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	svc := newWhatsAppService(req)

	code, qr, err := svc.GetQRByUser(user.Team.ID, user.ID)
	if err != nil {
		return ErrJSON(502, err, "error getting qr")
	}

	return c.JSON(200, g.M(
		"success", true,
		"code", code,
		"data", g.M("QRCode", qr),
	))
}

// POST /conexoes/whatsapp/disconnect
func DisconnectConexaoWhatsapp(c echo.Context) error {
	req := NewRequest(c)

	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	svc := newWhatsAppService(req)

	if err := svc.Disconnect(user.Team.ID, user.ID); err != nil {
		g.Warn("disconnect error: %v", err)
		return c.JSON(200, g.M("ok", false, "error", err.Error()))
	}

	return c.JSON(200, g.M("ok", true))
}

// POST /conexoes/whatsapp/delete
func ExcluirConexaoWhatsapp(c echo.Context) error {
	req := NewRequest(c)

	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	svc := newWhatsAppService(req)

	if err := svc.DeleteConnection(user.Team.ID, user.ID); err != nil {
		g.Warn("delete instance error: %v", err)
		return c.JSON(200, g.M("ok", false, "error", err.Error()))
	}

	return c.JSON(200, g.M("ok", true))
}

// GET /conexoes/whatsapp/status
func StatusConexaoWhatsapp(c echo.Context) error {
	req := NewRequest(c)

	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	svc := newWhatsAppService(req)

	connected, owned, jid, err := svc.CheckAndUpdateStatus(user.Team.ID, user.ID)
	if err != nil {
		g.Warn("status check error: %v", err)
		return c.JSON(200, g.M("connected", false, "owned", false, "jid", "", "error", err.Error()))
	}

	return c.JSON(200, g.M("connected", connected, "owned", owned, "jid", jid, "error", ""))
}

// POST /conexoes/whatsapp/send-test
func EnviarMensagemTesteWhatsapp(c echo.Context) error {
	req := NewRequest(c)

	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	type ReqBody struct {
		Phone string `json:"phone"`
		Body  string `json:"body"`
	}

	var body ReqBody
	if err := c.Bind(&body); err != nil {
		return ErrJSON(400, g.Error("invalid request body"))
	}

	if body.Phone == "" || body.Body == "" {
		return ErrJSON(400, g.Error("phone and body are required"))
	}

	svc := newWhatsAppService(req)

	resp, err := svc.SendTestMessage(user.Team.ID, user.ID, body.Phone, body.Body)
	if err != nil {
		g.Warn("send test message error: %v", err)
		return c.JSON(200, g.M("success", false, "error", err.Error()))
	}

	return c.JSON(200, g.M("success", true, "data", resp))
}

// GET /conexoes/whatsapp
func ObterConexaoWhatsapp(c echo.Context) error {
	req := NewRequest(c)

	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	svc := newWhatsAppService(req)

	exists, owned, con, wa, err := svc.GetByUser(user.Team.ID, user.ID)
	if err != nil {
		return ErrJSON(500, err)
	}

	if !exists {
		return c.JSON(200, g.M(
			"exists", false,
			"owned", false,
			"conexao", nil,
			"whatsapp", nil,
		))
	}

	if wa == nil || wa.Id == "" {
		return c.JSON(200, g.M(
			"exists", true,
			"owned", owned,
			"conexao", con.PublicExport(),
			"whatsapp", nil,
		))
	}

	return c.JSON(200, g.M(
		"exists", true,
		"owned", owned,
		"conexao", con.PublicExport(),
		"whatsapp", wa.PublicExport(),
	))
}

// POST /conexoes/whatsapp/fix-webhook
func FixWebhookWhatsapp(c echo.Context) error {
	req := NewRequest(c)

	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	svc := newWhatsAppService(req)

	webhookURL, err := svc.FixWebhookForUser(user.Team.ID, user.ID)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "não encontrada") || strings.Contains(msg, "não encontrados") {
			return ErrJSON(404, err)
		}
		if strings.Contains(msg, "não configurado") {
			return ErrJSON(400, err)
		}
		g.Error(err, "fix-webhook: falha ao atualizar webhook")
		return ErrJSON(502, err)
	}

	return c.JSON(200, g.M("ok", true, "webhook_url", webhookURL))
}

// POST /webhooks/whatsmeow
// Connected → WhatsAppService; Message → fluxo de IA existente (helpers no pacote).
func WebhookWhatsmeow(c echo.Context) error {
	app := c.Get("app").(*pocketbase.PocketBase)

	var body map[string]any
	if err := c.Bind(&body); err != nil {
		g.Warn("webhook/whatsmeow: invalid body: %v", err)
		return c.JSON(400, g.M("error", "invalid body"))
	}

	g.Info("webhook/whatsmeow: raw_body=%s", truncateWebhookLog(g.Marshal(body), 8000))

	eventType := strings.ToLower(strings.TrimSpace(cast.ToString(body["type"])))
	userID := strings.TrimSpace(cast.ToString(body["userID"]))
	tokenHeader := strings.TrimSpace(c.Request().Header.Get("token"))

	if eventType == "" {
		eventType = strings.ToLower(strings.TrimSpace(cast.ToString(body["event"])))
	}
	if userID == "" {
		userID = strings.TrimSpace(cast.ToString(body["token"]))
		if userID == "" {
			userID = tokenHeader
		}
	}

	g.Info("webhook/whatsmeow: type=%q userID=%q tokenHeader=%q", eventType, userID, maskWebhookPhone(tokenHeader))

	svc := newWhatsAppServiceFromDao(app.Dao())

	switch eventType {
	case "connected":
		jid := ""
		if evt, ok := body["event"].(map[string]any); ok {
			jid = strings.TrimSpace(cast.ToString(evt["JID"]))
		}
		if err := svc.HandleConnectedEvent(userID, tokenHeader, jid); err != nil {
			g.Error(err, "webhook/whatsmeow: erro ao processar Connected jid=%q", jid)
		}

	case "message":
		handleWebhookMessage(app, svc, userID, tokenHeader, body)
	}

	return c.JSON(200, g.M("ok", true))
}

func handleWebhookMessage(
	app *pocketbase.PocketBase,
	svc *services.WhatsAppService,
	userID, tokenHeader string,
	body map[string]any,
) {
	wa, _ := svc.ResolveWARecord(userID, tokenHeader)
	if wa == nil || wa.Id == "" {
		g.Warn("webhook/whatsmeow: Message — nenhum registro encontrado userID=%q tokenHeader=%q", userID, maskWebhookPhone(tokenHeader))
		return
	}

	conexaoID := resolveWAConexaoID(wa)
	g.Info(
		"webhook/whatsmeow: conexoes_whatsapp encontrado wa_id=%s conexoes_raw=%s conexao_id=%s",
		wa.Id,
		truncateWebhookLog(g.Marshal(wa.Get("conexoes")), 300),
		conexaoID,
	)

	if conexaoID == "" {
		g.Warn("webhook/whatsmeow: conexão relation vazia para wa=%s", wa.Id)
		return
	}

	conRepo := repositories.NewConexaoRepo(app.Dao())
	con, err := conRepo.FindByID(conexaoID)
	if err != nil || con == nil {
		g.Warn("webhook/whatsmeow: conexão não encontrada para wa=%s conexao_id=%s err=%v", wa.Id, conexaoID, err)
		return
	}

	teamID := strings.TrimSpace(con.GetString("team_id"))
	if teamID == "" {
		g.Warn("webhook/whatsmeow: team_id vazio para conexao=%s", con.Id)
		return
	}

	evt, ok := body["event"].(map[string]any)
	if !ok {
		if data, ok := body["data"].(map[string]any); ok {
			evt = map[string]any{
				"Info": map[string]any{
					"Sender":   data["from"],
					"IsFromMe": data["fromMe"],
					"PushName": data["pushName"],
					"Chat":     data["chat"],
				},
				"Message": map[string]any{
					"conversation": data["body"],
				},
			}
		} else {
			g.Warn("webhook/whatsmeow: event inválido no payload")
			return
		}
	}

	info, _ := evt["Info"].(map[string]any)
	if info == nil {
		g.Warn("webhook/whatsmeow: Info ausente no evento")
		return
	}

	if msgSrc, ok := info["MessageSource"].(map[string]any); ok {
		if info["Sender"] == nil {
			info["Sender"] = msgSrc["Sender"]
		}
		if info["Chat"] == nil {
			info["Chat"] = msgSrc["Chat"]
		}
		if info["IsFromMe"] == nil {
			info["IsFromMe"] = msgSrc["IsFromMe"]
		}
		if info["IsGroup"] == nil {
			info["IsGroup"] = msgSrc["IsGroup"]
		}
	}

	isFromMe := cast.ToBool(info["IsFromMe"])
	if isFromMe {
		pauseConversationalIAOnHumanMessage(app.Dao(), teamID, info, evt)
		g.Debug("webhook/whatsmeow: mensagem enviada por nós (ia pausada para atendimento manual quando aplicável)")
		return
	}

	if cast.ToBool(info["IsGroup"]) {
		g.Debug("webhook/whatsmeow: ignorando mensagem de grupo")
		return
	}

	chatStr := strings.TrimSpace(cast.ToString(info["Chat"]))
	if chatStr == "" {
		if chatMap, ok := info["Chat"].(map[string]any); ok {
			chatStr = strings.TrimSpace(cast.ToString(chatMap["Server"]))
		}
	}
	if strings.Contains(chatStr, "@newsletter") || strings.Contains(chatStr, "newsletter") {
		g.Debug("webhook/whatsmeow: ignorando mensagem de newsletter/canal")
		return
	}

	telefone := extractBestPhoneFromInfo(info)
	if telefone == "" {
		g.Warn("webhook/whatsmeow: telefone vazio após extração de JID Sender=%v SenderAlt=%v Chat=%v", info["Sender"], info["SenderAlt"], info["Chat"])
		return
	}

	nomeContato := strings.TrimSpace(cast.ToString(info["PushName"]))
	if nomeContato == "" {
		nomeContato = telefone
	}

	mensagem := extractWAMessageText(evt)
	if mensagem == "" {
		g.Debug("webhook/whatsmeow: mensagem vazia, ignorando (pode ser mídia)")
		return
	}

	msgIDExterno := extractWAInfoMessageID(info)

	g.Info(
		"webhook/whatsmeow: Mensagem recebida team=%s de %s (%s): %s",
		teamID,
		nomeContato,
		maskWebhookPhone(telefone),
		truncateWebhookLog(mensagem, 1200),
	)

	candidatos := buildTelefoneCandidates(telefone)
	g.Info("webhook/whatsmeow: candidatos telefone=%v", candidatos)

	dest := services.FindDestinatarioEnviadoRecente(app.Dao(), teamID, candidatos)
	if dest != nil {
		g.Info(
			"webhook/whatsmeow: destinatário encontrado dest_id=%s campanha_id=%s",
			dest.Id,
			dest.GetString("campanha_id"),
		)
	}

	conversaRepo := repositories.NewConversaRepo(app.Dao())
	conversaAtual := conversaRepo.FindByTelefoneCandidates(teamID, candidatos)

	iaAtiva := services.IAAtivaParaTelefone(app.Dao(), teamID, candidatos)

	if !iaAtiva {
		services.PausarConversaIA(conversaRepo, teamID, candidatos)

		if err := registrarRespostaWhatsAppSemIA(app.Dao(), conversaRepo, conversaAtual, teamID, telefone, nomeContato, mensagem, dest, msgIDExterno); err != nil {
			g.Warn(
				"webhook/whatsmeow: falha ao registrar resposta sem IA telefone=%s err=%v",
				maskWebhookPhone(telefone),
				err,
			)
		}

		g.Info(
			"webhook/whatsmeow: IA desativada para telefone=%s team=%s (manter_ia=false ou sem campanha)",
			maskWebhookPhone(telefone),
			teamID,
		)
		return
	}

	g.Info("webhook/whatsmeow: Lead %s encontrado em campanha com IA ativa, processando...", maskWebhookPhone(telefone))

	ownerToken := strings.TrimSpace(wa.GetString("numero_e164"))

	go func(teamID, telefone, mensagem, nomeContato, msgID, ownerToken string) {
		dao := app.Dao()

		intencaoRepo := repositories.NewIntencaoRepo(dao)
		conversaRepo := repositories.NewConversaRepo(dao)
		geminiSvc := services.NewGeminiService()
		whatsappSvc := newWhatsAppServiceFromDao(dao)

		iaSvc := services.NewIAConversacionalService(dao, intencaoRepo, conversaRepo, whatsappSvc, geminiSvc)

		g.Info("webhook/whatsmeow: goroutine IA start team=%s telefone=%s", teamID, maskWebhookPhone(telefone))
		if err := iaSvc.ProcessarMensagemRecebida(teamID, telefone, mensagem, nomeContato, msgID, ownerToken); err != nil {
			g.Error(err, "webhook/whatsmeow: erro ao processar mensagem telefone=%s", maskWebhookPhone(telefone))
			return
		}
		g.Info("webhook/whatsmeow: goroutine IA end telefone=%s", maskWebhookPhone(telefone))
	}(teamID, telefone, mensagem, nomeContato, msgIDExterno, ownerToken)
}

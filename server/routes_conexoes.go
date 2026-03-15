package server

import (
	"strings"
	"time"

	"github.com/flarco/g"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
	"github.com/spf13/cast"

	"github.com/suaobra/suaobra-app/server/clients/wuzapi"
	"github.com/suaobra/suaobra-app/server/config"
	"github.com/suaobra/suaobra-app/server/repositories"
	"github.com/suaobra/suaobra-app/server/services"
)

// helper: monta o service por request (usa req.Dao())
func newWhatsAppService(req Request) *services.WhatsAppService {
	cfg := config.NewWhatsMeowConfig()
	wuzClient := wuzapi.NewClient(cfg)

	conRepo := repositories.NewConexaoRepo(req.Dao())
	waRepo := repositories.NewWhatsAppRepo(req.Dao())
	tokenSvc := services.NewTokenService()

	return services.NewWhatsAppService(conRepo, waRepo, wuzClient, tokenSvc)
}

func newEmailService(req Request) *services.EmailService {
	conRepo := repositories.NewConexaoRepo(req.Dao())
	emailRepo := repositories.NewEmailRepo(req.Dao())

	return services.NewEmailService(conRepo, emailRepo)
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
		"conexao",       con.PublicExport(),
		"whatsapp",      wa.PublicExport(),
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

	raw, err := svc.ConnectByTeam(user.Team.ID)
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

	code, qr, err := svc.GetQRByTeam(user.Team.ID)
	if err != nil {
		return ErrJSON(502, err, "error getting qr")
	}

	return c.JSON(200, g.M(
		"success", true,
		"code",    code,
		"data",    g.M("QRCode", qr),
	))
}

// POST /conexoes/whatsapp/disconnect
// Limpa conectado_em e device_jid — "desconecta" no banco.
func DisconnectConexaoWhatsapp(c echo.Context) error {
	req := NewRequest(c)

	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	svc := newWhatsAppService(req)

	if err := svc.Disconnect(user.Team.ID); err != nil {
		g.Warn("disconnect error: %v", err)
		return c.JSON(200, g.M("ok", false, "error", err.Error()))
	}

	return c.JSON(200, g.M("ok", true))
}

// GET /conexoes/whatsapp/status
// Consulta o wuzapi diretamente para saber se está connected e atualiza o banco.
func StatusConexaoWhatsapp(c echo.Context) error {
	req := NewRequest(c)

	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	svc := newWhatsAppService(req)

	connected, jid, err := svc.CheckAndUpdateStatus(user.Team.ID)
	if err != nil {
		g.Warn("status check error: %v", err)
		return c.JSON(200, g.M("connected", false, "jid", ""))
	}

	return c.JSON(200, g.M("connected", connected, "jid", jid))
}

// POST /conexoes/whatsapp/send-test
// Envia uma mensagem de teste do WhatsApp.
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

	resp, err := svc.SendTestMessage(user.Team.ID, body.Phone, body.Body)
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

	exists, con, wa, err := svc.GetByTeam(user.Team.ID)
	if err != nil {
		return ErrJSON(500, err)
	}

	if !exists {
		return c.JSON(200, g.M(
			"exists",   false,
			"conexao",  nil,
			"whatsapp", nil,
		))
	}

	// pode existir conexão mas ainda não ter registro em conexoes_whatsapp
	if wa == nil || wa.Id == "" {
		return c.JSON(200, g.M(
			"exists",   true,
			"conexao",  con.PublicExport(),
			"whatsapp", nil,
		))
	}

	return c.JSON(200, g.M(
		"exists",   true,
		"conexao",  con.PublicExport(),
		"whatsapp", wa.PublicExport(),
	))
}

// SyncWebhookURLs atualiza a URL de webhook em todos os usuários wuzapi ativos.
// Chamado no startup para garantir que o wuzapi aponte para o servidor correto.
func SyncWebhookURLs(app *pocketbase.PocketBase) {
	cfg := config.NewWhatsMeowConfig()
	webhookURL := cfg.WebhookURL()
	if webhookURL == "" {
		g.Warn("SyncWebhookURLs: WHATSMEOW_WEBHOOK_BASE não configurado, pulando sync")
		return
	}

	wuzClient := wuzapi.NewClient(cfg)
	waRepo := repositories.NewWhatsAppRepo(app.Dao())

	records, err := waRepo.FindAll()
	if err != nil || len(records) == 0 {
		g.Info("SyncWebhookURLs: nenhuma conexão WhatsApp para sincronizar")
		return
	}

	for _, wa := range records {
		instanciaID := strings.TrimSpace(wa.GetString("instancia_id"))
		if instanciaID == "" {
			continue
		}

		currentWebhook := strings.TrimSpace(wa.GetString("webhook"))
		if currentWebhook == webhookURL {
			continue
		}

		if err := wuzClient.UpdateAdminUser(instanciaID, map[string]any{
			"webhook": webhookURL,
			"events":  "All",
		}); err != nil {
			g.Warn("SyncWebhookURLs: falha ao atualizar instancia=%s: %v", instanciaID, err)
			continue
		}

		wa.Set("webhook", webhookURL)
		if saveErr := app.Dao().SaveRecord(wa); saveErr != nil {
			g.Warn("SyncWebhookURLs: falha ao salvar webhook local instancia=%s: %v", instanciaID, saveErr)
		} else {
			g.Info("SyncWebhookURLs: webhook atualizado para %s instancia=%s", webhookURL, instanciaID)
		}
	}
}

// POST /conexoes/whatsapp/fix-webhook
// Atualiza a URL de webhook e os eventos no wuzapi para a conexão ativa do usuário.
// Útil quando WHATSMEOW_WEBHOOK_BASE foi alterado ou o webhook estava desconfigurado.
func FixWebhookWhatsapp(c echo.Context) error {
	req := NewRequest(c)

	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	cfg := config.NewWhatsMeowConfig()
	wuzClient := wuzapi.NewClient(cfg)
	conRepo := repositories.NewConexaoRepo(req.Dao())
	waRepo := repositories.NewWhatsAppRepo(req.Dao())

	con, err := conRepo.FindActiveWhatsappByTeam(user.Team.ID)
	if err != nil || con == nil || con.Id == "" {
		return ErrJSON(404, g.Error("nenhuma conexão WhatsApp ativa encontrada"))
	}

	wa, err := waRepo.FindByConexao(con.Id)
	if err != nil || wa == nil || wa.Id == "" {
		return ErrJSON(404, g.Error("dados do WhatsApp não encontrados"))
	}

	instanciaID := strings.TrimSpace(wa.GetString("instancia_id"))
	if instanciaID == "" {
		return ErrJSON(400, g.Error("instancia_id não configurado para esta conexão"))
	}

	webhookURL := cfg.WebhookURL()
	if webhookURL == "" {
		return ErrJSON(400, g.Error("WHATSMEOW_WEBHOOK_BASE não configurado no servidor"))
	}

	if err := wuzClient.UpdateAdminUser(instanciaID, map[string]any{
		"webhook": webhookURL,
		"events":  "All",
	}); err != nil {
		g.Error(err, "fix-webhook: falha ao atualizar usuário wuzapi id=%s", instanciaID)
		return ErrJSON(502, err)
	}

	wa.Set("webhook", webhookURL)
	if saveErr := req.Dao().SaveRecord(wa); saveErr != nil {
		g.Warn("fix-webhook: falha ao salvar webhook no registro local: %v", saveErr)
	}

	g.Info("fix-webhook: webhook atualizado para %s instancia=%s", webhookURL, instanciaID)
	return c.JSON(200, g.M("ok", true, "webhook_url", webhookURL))
}

// POST /webhooks/whatsmeow
// Recebe eventos do wuzapi (sem autenticação de usuário — chamada vem do servidor wuzapi).
// Formato WUZAPI v2: {"type": "Message", "userID": "...", "event": {...}}
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

	// Fallback para formato antigo (testes manuais)
	if eventType == "" {
		eventType = strings.ToLower(strings.TrimSpace(cast.ToString(body["event"])))
	}
	if userID == "" {
		userID = strings.TrimSpace(cast.ToString(body["token"]))
		if userID == "" {
			userID = strings.TrimSpace(c.Request().Header.Get("token"))
		}
	}

	g.Info("webhook/whatsmeow: type=%q userID=%q", eventType, userID)

	switch eventType {
	case "connected":
		if userID == "" {
			g.Warn("webhook/whatsmeow: Connected sem userID")
			return c.JSON(200, g.M("ok", true))
		}

		waRepo := repositories.NewWhatsAppRepo(app.Dao())
		wa, err := waRepo.FindByInstanciaID(userID)
		if err != nil || wa == nil || wa.Id == "" {
			g.Warn("webhook/whatsmeow: nenhum conexoes_whatsapp para userID=%q err=%v", userID, err)
			return c.JSON(200, g.M("ok", true))
		}

		jid := ""
		if evt, ok := body["event"].(map[string]any); ok {
			jid = strings.TrimSpace(cast.ToString(evt["JID"]))
		}

		wa.Set("device_jid", jid)
		wa.Set("conectado_em", time.Now().UTC())

		if saveErr := app.Dao().SaveRecord(wa); saveErr != nil {
			g.Error(saveErr, "webhook/whatsmeow: erro ao salvar Connected jid=%q", jid)
		} else {
			g.Info("webhook/whatsmeow: Connected OK — device_jid=%q waId=%s", jid, wa.Id)
		}

	case "message":
		if userID == "" {
			g.Warn("webhook/whatsmeow: Message sem userID")
			return c.JSON(200, g.M("ok", true))
		}

		waRepo := repositories.NewWhatsAppRepo(app.Dao())
		wa, err := waRepo.FindByInstanciaID(userID)
		if err != nil || wa == nil || wa.Id == "" {
			g.Warn("webhook/whatsmeow: nenhum conexoes_whatsapp para userID=%q", userID)
			return c.JSON(200, g.M("ok", true))
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
			return c.JSON(200, g.M("ok", true))
		}

		conRepo := repositories.NewConexaoRepo(app.Dao())
		con, err := conRepo.FindByID(conexaoID)
		if err != nil || con == nil {
			g.Warn("webhook/whatsmeow: conexão não encontrada para wa=%s conexao_id=%s err=%v", wa.Id, conexaoID, err)
			return c.JSON(200, g.M("ok", true))
		}

		teamID := strings.TrimSpace(con.GetString("team_id"))
		if teamID == "" {
			g.Warn("webhook/whatsmeow: team_id vazio para conexao=%s", con.Id)
			return c.JSON(200, g.M("ok", true))
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
				return c.JSON(200, g.M("ok", true))
			}
		}

		info, _ := evt["Info"].(map[string]any)
		if info == nil {
			g.Warn("webhook/whatsmeow: Info ausente no evento")
			return c.JSON(200, g.M("ok", true))
		}

		isFromMe := cast.ToBool(info["IsFromMe"])
		if isFromMe {
			g.Debug("webhook/whatsmeow: ignorando mensagem enviada por nós")
			return c.JSON(200, g.M("ok", true))
		}

		sender := strings.TrimSpace(cast.ToString(info["Sender"]))
		if sender == "" {
			sender = strings.TrimSpace(cast.ToString(info["Chat"]))
		}
		if sender == "" {
			g.Warn("webhook/whatsmeow: Sender/Chat vazio no evento")
			return c.JSON(200, g.M("ok", true))
		}

		telefone := normalizeWASender(sender)
		if telefone == "" {
			g.Warn("webhook/whatsmeow: telefone vazio após normalização sender=%q", sender)
			return c.JSON(200, g.M("ok", true))
		}

		nomeContato := strings.TrimSpace(cast.ToString(info["PushName"]))
		if nomeContato == "" {
			nomeContato = telefone
		}

		mensagem := extractWAMessageText(evt)
		if mensagem == "" {
			g.Debug("webhook/whatsmeow: mensagem vazia, ignorando (pode ser mídia)")
			return c.JSON(200, g.M("ok", true))
		}

		g.Info(
			"webhook/whatsmeow: Mensagem recebida team=%s de %s (%s): %s",
			teamID,
			nomeContato,
			maskWebhookPhone(telefone),
			truncateWebhookLog(mensagem, 1200),
		)

		var dest *models.Record
		var destErr error
		candidatos := buildTelefoneCandidates(telefone)

		g.Info("webhook/whatsmeow: candidatos telefone=%v", candidatos)

		for _, candidato := range candidatos {
			dest, destErr = app.Dao().FindFirstRecordByFilter(
				"campanha_destinatarios",
				"team_id = {:teamId} && telefone_e164 = {:telefone} && status = 'ENVIADO'",
				dbx.Params{
					"teamId":   teamID,
					"telefone": candidato,
				},
			)
			if destErr == nil && dest != nil {
				g.Info(
					"webhook/whatsmeow: destinatário encontrado dest_id=%s campanha_id=%s telefone_match=%s",
					dest.Id,
					dest.GetString("campanha_id"),
					candidato,
				)
				break
			}
		}

		if destErr != nil || dest == nil {
			g.Warn(
				"webhook/whatsmeow: telefone %s não é lead contatado para team %s candidatos=%v err=%v",
				maskWebhookPhone(telefone),
				teamID,
				candidatos,
				destErr,
			)
			return c.JSON(200, g.M("ok", true))
		}

		campanhaID := strings.TrimSpace(dest.GetString("campanha_id"))
		if campanhaID != "" {
			campanha, campErr := app.Dao().FindRecordById("campanhas", campanhaID)
			if campErr == nil && campanha != nil {
				manterIA := campanha.GetBool("manter_ia")
				g.Info("webhook/whatsmeow: campanha=%s manter_ia=%v", campanhaID, manterIA)
				if !manterIA {
					g.Warn("webhook/whatsmeow: campanha %s não tem IA ativa, ignorando", campanhaID)
					return c.JSON(200, g.M("ok", true))
				}
			} else {
				g.Warn("webhook/whatsmeow: erro ao buscar campanha=%s err=%v", campanhaID, campErr)
			}
		}

		g.Info("webhook/whatsmeow: Lead %s encontrado em campanha com IA ativa, processando...", maskWebhookPhone(telefone))

		go func(teamID, telefone, mensagem, nomeContato string) {
			dao := app.Dao()

			intencaoRepo := repositories.NewIntencaoRepo(dao)
			conversaRepo := repositories.NewConversaRepo(dao)
			geminiSvc := services.NewGeminiService()

			cfg := config.NewWhatsMeowConfig()
			wuzClient := wuzapi.NewClient(cfg)
			tokenSvc := services.NewTokenService()

			conRepo := repositories.NewConexaoRepo(dao)
			waRepo := repositories.NewWhatsAppRepo(dao)
			whatsappSvc := services.NewWhatsAppService(conRepo, waRepo, wuzClient, tokenSvc)

			iaSvc := services.NewIAConversacionalService(dao, intencaoRepo, conversaRepo, whatsappSvc, geminiSvc)

			g.Info("webhook/whatsmeow: goroutine IA start team=%s telefone=%s", teamID, maskWebhookPhone(telefone))
			if err := iaSvc.ProcessarMensagemRecebida(teamID, telefone, mensagem, nomeContato); err != nil {
				g.Error(err, "webhook/whatsmeow: erro ao processar mensagem telefone=%s", maskWebhookPhone(telefone))
				return
			}
			g.Info("webhook/whatsmeow: goroutine IA end telefone=%s", maskWebhookPhone(telefone))
		}(teamID, telefone, mensagem, nomeContato)
	}

	return c.JSON(200, g.M("ok", true))
}

// ═══════════════════════════════════════════════════════════════════════════
// EMAIL ROUTES
// ═══════════════════════════════════════════════════════════════════════════

// POST /conexoes/email
// Salva ou atualiza a configuração de e-mail.
func SalvarConexaoEmail(c echo.Context) error {
	req := NewRequest(c)

	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	type ReqBody struct {
		Nome           string `json:"nome"`
		RemetenteEmail string `json:"remetente_email"`
		ReplyTo        string `json:"reply_to"`
		SMTPHost       string `json:"smtp_host"`
		SMTPPort       int    `json:"smtp_port"`
		SMTPUsuario    string `json:"smtp_usuario"`
		SMTPSenha      string `json:"smtp_senha"`
		Criptografia   string `json:"criptografia"`
		TaxaPorMin     int    `json:"taxa_por_min"`
		TamanhoLote    int    `json:"tamanho_lote"`
		LimiteDiario   int    `json:"limite_diario"`
	}

	var body ReqBody
	if err := c.Bind(&body); err != nil {
		return ErrJSON(400, g.Error("invalid request body"))
	}

	fields := map[string]any{
		"conexoes_email":  body.Nome,
		"remetente_email": body.RemetenteEmail,
		"reply_to":        body.ReplyTo,
		"smtp_host":       body.SMTPHost,
		"smtp_port":       body.SMTPPort,
		"smtp_usuario":    body.SMTPUsuario,
		"criptografia":    body.Criptografia,
	}

	if body.SMTPSenha != "" {
		fields["smtp_senha"] = body.SMTPSenha
	}

	svc := newEmailService(req)

	con, email, err := svc.SaveOrUpdateConfig(user.Team.ID, fields)
	if err != nil {
		g.Warn("save email config error: %v", err)
		return ErrJSON(500, g.Error("erro ao salvar configuração"))
	}

	return c.JSON(200, g.M(
		"success", true,
		"conexao", con,
		"email",   email,
	))
}

// GET /conexoes/email
// Retorna a configuração de e-mail atual.
func ObterConexaoEmail(c echo.Context) error {
	req := NewRequest(c)

	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	svc := newEmailService(req)

	exists, con, email, err := svc.GetConfig(user.Team.ID)
	if err != nil {
		g.Warn("get email config error: %v", err)
		return c.JSON(200, g.M(
			"exists",  false,
			"conexao", nil,
			"email",   nil,
		))
	}

	return c.JSON(200, g.M(
		"exists",  exists,
		"conexao", con,
		"email",   email,
	))
}

// POST /conexoes/email/send-test
// Envia um e-mail de teste.
func EnviarEmailTeste(c echo.Context) error {
	req := NewRequest(c)

	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	type ReqBody struct {
		To      string `json:"to"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}

	var body ReqBody
	if err := c.Bind(&body); err != nil {
		return ErrJSON(400, g.Error("invalid request body"))
	}

	if body.To == "" || body.Subject == "" || body.Body == "" {
		return ErrJSON(400, g.Error("to, subject e body são obrigatórios"))
	}

	svc := newEmailService(req)

	err = svc.SendTestEmail(user.Team.ID, body.To, body.Subject, body.Body)
	if err != nil {
		g.Warn("send test email error teamID=%s to=%s err=%v", user.Team.ID, body.To, err)
		return c.JSON(200, g.M("success", false, "error", err.Error()))
	}

	g.Info("test email sent successfully teamID=%s to=%s", user.Team.ID, body.To)
	return c.JSON(200, g.M("success", true))
}

func resolveWAConexaoID(wa *models.Record) string {
	if wa == nil {
		return ""
	}

	ids := wa.GetStringSlice("conexoes")
	if len(ids) > 0 {
		return strings.TrimSpace(ids[0])
	}

	raw := wa.Get("conexoes")
	switch v := raw.(type) {
	case []any:
		for _, item := range v {
			id := strings.TrimSpace(cast.ToString(item))
			if id != "" {
				return id
			}
		}
	case []string:
		for _, item := range v {
			id := strings.TrimSpace(item)
			if id != "" {
				return id
			}
		}
	}

	return strings.TrimSpace(wa.GetString("conexoes"))
}

func normalizeWASender(sender string) string {
	sender = strings.TrimSpace(sender)
	if sender == "" {
		return ""
	}

	if idx := strings.Index(sender, "@"); idx > 0 {
		sender = sender[:idx]
	}
	if idx := strings.Index(sender, ":"); idx > 0 {
		sender = sender[:idx]
	}

	var b strings.Builder
	b.Grow(len(sender))
	for _, r := range sender {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func extractWAMessageText(evt map[string]any) string {
	msg, ok := evt["Message"].(map[string]any)
	if !ok || msg == nil {
		return ""
	}

	if conv := strings.TrimSpace(cast.ToString(msg["conversation"])); conv != "" {
		return conv
	}

	if extMsg, ok := msg["extendedTextMessage"].(map[string]any); ok {
		if text := strings.TrimSpace(cast.ToString(extMsg["text"])); text != "" {
			return text
		}
	}

	if imgMsg, ok := msg["imageMessage"].(map[string]any); ok {
		if caption := strings.TrimSpace(cast.ToString(imgMsg["caption"])); caption != "" {
			return caption
		}
	}

	if vidMsg, ok := msg["videoMessage"].(map[string]any); ok {
		if caption := strings.TrimSpace(cast.ToString(vidMsg["caption"])); caption != "" {
			return caption
		}
	}

	return ""
}

func buildTelefoneCandidates(telefone string) []string {
	telefone = normalizeWASender(telefone)
	if telefone == "" {
		return nil
	}

	seen := map[string]struct{}{}
	out := make([]string, 0, 4)

	add := func(v string) {
		v = normalizeWASender(v)
		if v == "" {
			return
		}
		if _, ok := seen[v]; ok {
			return
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}

	add(telefone)

	if strings.HasPrefix(telefone, "55") && len(telefone) > 10 {
		add(telefone[2:])
	}

	if !strings.HasPrefix(telefone, "55") && len(telefone) >= 10 {
		add("55" + telefone)
	}

	return out
}

func truncateWebhookLog(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "...(truncated)"
}

func maskWebhookPhone(phone string) string {
	phone = normalizeWASender(phone)
	if len(phone) <= 4 {
		return phone
	}
	return strings.Repeat("*", len(phone)-4) + phone[len(phone)-4:]
}
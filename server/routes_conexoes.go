package server

import (
	"strings"
	"time"

	"github.com/flarco/g"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"

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
		"code", code,
		"data", g.M("QRCode", qr),
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
			"exists", false,
			"conexao", nil,
			"whatsapp", nil,
		))
	}

	// pode existir conexão mas ainda não ter registro em conexoes_whatsapp
	if wa == nil || wa.Id == "" {
		return c.JSON(200, g.M(
			"exists", true,
			"conexao", con.PublicExport(),
			"whatsapp", nil,
		))
	}

	return c.JSON(200, g.M(
		"exists", true,
		"conexao", con.PublicExport(),
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

		currentWebhook := wa.GetString("webhook")
		if currentWebhook == webhookURL {
			continue // já está correto
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

	// Atualiza também o campo webhook no registro local
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

	// WUZAPI v2 usa "type" para o tipo de evento e "userID" para identificar a instância
	eventType, _ := body["type"].(string)
	eventType = strings.ToLower(strings.TrimSpace(eventType))
	userID, _ := body["userID"].(string)

	// Fallback para formato antigo (testes manuais)
	if eventType == "" {
		if evt, ok := body["event"].(string); ok {
			eventType = strings.ToLower(strings.TrimSpace(evt))
		}
	}
	if userID == "" {
		userID, _ = body["token"].(string)
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

		// Extrai JID do payload
		jid := ""
		if evt, ok := body["event"].(map[string]any); ok {
			if v, ok := evt["JID"].(string); ok {
				jid = v
			}
		}

		wa.Set("device_jid", jid)
		wa.Set("conectado_em", time.Now().UTC())

		if saveErr := app.Dao().SaveRecord(wa); saveErr != nil {
			g.Error(saveErr, "webhook/whatsmeow: erro ao salvar Connected jid=%q", jid)
		} else {
			g.Info("webhook/whatsmeow: Connected OK — device_jid=%q waId=%s", jid, wa.Id)
		}

	case "message":
		// Mensagem recebida - processa com IA conversacional
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

		g.Info("webhook/whatsmeow: ✅ conexoes_whatsapp encontrado wa_id=%s conexoes=%s", wa.Id, wa.GetString("conexoes"))

		// Busca conexão para pegar team_id
		conRepo := repositories.NewConexaoRepo(app.Dao())
		con, err := conRepo.FindByID(wa.GetString("conexoes"))
		if err != nil || con == nil {
			g.Warn("webhook/whatsmeow: conexão não encontrada para wa=%s conexoes=%s", wa.Id, wa.GetString("conexoes"))
			return c.JSON(200, g.M("ok", true))
		}

		teamID := con.GetString("team_id")
		if teamID == "" {
			g.Warn("webhook/whatsmeow: team_id vazio para conexao=%s", con.Id)
			return c.JSON(200, g.M("ok", true))
		}

		// Extrai dados da mensagem - formato WUZAPI v2
		// body["event"] contém Info e Message
		evt, ok := body["event"].(map[string]any)
		if !ok {
			// Fallback para formato de teste manual
			if data, ok := body["data"].(map[string]any); ok {
				evt = map[string]any{
					"Info": map[string]any{
						"Sender":   data["from"],
						"IsFromMe": data["fromMe"],
						"PushName": data["pushName"],
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

		// Extrai Info
		info, _ := evt["Info"].(map[string]any)
		if info == nil {
			g.Warn("webhook/whatsmeow: Info ausente no evento")
			return c.JSON(200, g.M("ok", true))
		}

		// Verifica se é mensagem enviada por nós (ignora)
		isFromMe, _ := info["IsFromMe"].(bool)
		if isFromMe {
			g.Debug("webhook/whatsmeow: ignorando mensagem enviada por nós")
			return c.JSON(200, g.M("ok", true))
		}

		// Extrai telefone do remetente (Sender pode ser "5511999999999@s.whatsapp.net")
		sender, _ := info["Sender"].(string)
		if sender == "" {
			sender, _ = info["Chat"].(string)
		}
		if sender == "" {
			g.Warn("webhook/whatsmeow: Sender/Chat vazio no evento")
			return c.JSON(200, g.M("ok", true))
		}

		// Remove qualquer sufixo @... (ex: @s.whatsapp.net, @lid, etc)
		telefone := sender

		// corta @...
		if idx := strings.Index(telefone, "@"); idx > 0 {
			telefone = telefone[:idx]
		}

		// corta :5 etc
		if idx := strings.Index(telefone, ":"); idx > 0 {
			telefone = telefone[:idx]
		}

		// só dígitos
		var b strings.Builder
		b.Grow(len(telefone))
		for _, r := range telefone {
			if r >= '0' && r <= '9' {
				b.WriteRune(r)
			}
		}
		telefone = b.String()

		// Extrai nome do contato (se disponível)
		nomeContato, _ := info["PushName"].(string)
		if nomeContato == "" {
			nomeContato = telefone
		}

		// Extrai texto da mensagem - pode estar em Message.conversation ou Message.extendedTextMessage.text
		mensagem := ""
		if msg, ok := evt["Message"].(map[string]any); ok {
			if conv, ok := msg["conversation"].(string); ok && conv != "" {
				mensagem = conv
			} else if extMsg, ok := msg["extendedTextMessage"].(map[string]any); ok {
				if text, ok := extMsg["text"].(string); ok {
					mensagem = text
				}
			}
		}

		if mensagem == "" {
			g.Debug("webhook/whatsmeow: mensagem vazia, ignorando (pode ser mídia)")
			return c.JSON(200, g.M("ok", true))
		}

		g.Info("webhook/whatsmeow: Mensagem recebida de %s (%s): %s", nomeContato, telefone, mensagem)

		// Verifica se o telefone pertence a um lead que já foi contatado por uma campanha com IA ativa
		// Tenta buscar com o telefone exato ou com variações (com/sem código do país)
		var dest *models.Record
		var destErr error

		// Tenta primeiro com o telefone exato
		dest, destErr = app.Dao().FindFirstRecordByFilter(
			"campanha_destinatarios",
			"team_id = {:teamId} && telefone_e164 = {:telefone} && status = 'ENVIADO'",
			dbx.Params{
				"teamId":   teamID,
				"telefone": telefone,
			},
		)

		// Se não encontrou e o telefone tem código do país, tenta sem
		if (destErr != nil || dest == nil) && strings.HasPrefix(telefone, "55") && len(telefone) > 10 {
			telefoneS := telefone[2:] // Remove "55" do início
			dest, destErr = app.Dao().FindFirstRecordByFilter(
				"campanha_destinatarios",
				"team_id = {:teamId} && telefone_e164 = {:telefone} && status = 'ENVIADO'",
				dbx.Params{
					"teamId":   teamID,
					"telefone": telefoneS,
				},
			)
		}

		// Se não encontrou e o telefone não tem código do país, tenta com 55
		if (destErr != nil || dest == nil) && !strings.HasPrefix(telefone, "55") && len(telefone) >= 10 {
			telefone55 := "55" + telefone
			dest, destErr = app.Dao().FindFirstRecordByFilter(
				"campanha_destinatarios",
				"team_id = {:teamId} && telefone_e164 = {:telefone} && status = 'ENVIADO'",
				dbx.Params{
					"teamId":   teamID,
					"telefone": telefone55,
				},
			)
		}

		if destErr != nil || dest == nil {
			g.Debug("webhook/whatsmeow: telefone %s não é um lead contatado para team %s, ignorando", telefone, teamID)
			g.Debug("webhook/whatsmeow: telefone %s não é um lead contatado, ignorando", telefone)
			return c.JSON(200, g.M("ok", true))
		}

		// Verifica se a campanha tem IA ativa (manter_ia = true)
		campanhaID := dest.GetString("campanha_id")
		if campanhaID != "" {
			campanha, campErr := app.Dao().FindRecordById("campanhas", campanhaID)
			if campErr == nil && campanha != nil {
				manterIA := campanha.GetBool("manter_ia")
				if !manterIA {
					g.Debug("webhook/whatsmeow: campanha %s não tem IA ativa, ignorando", campanhaID)
					return c.JSON(200, g.M("ok", true))
				}
			}
		}

		g.Info("webhook/whatsmeow: Lead %s encontrado em campanha com IA ativa, processando...", telefone)

		// Processa mensagem com IA (em goroutine para não bloquear webhook)
		go func() {
			// Cria services necessários
			intencaoRepo := repositories.NewIntencaoRepo(app.Dao())
			conversaRepo := repositories.NewConversaRepo(app.Dao())
			geminiSvc := services.NewGeminiService()

			cfg := config.NewWhatsMeowConfig()
			wuzClient := wuzapi.NewClient(cfg)
			tokenSvc := services.NewTokenService()
			whatsappSvc := services.NewWhatsAppService(conRepo, waRepo, wuzClient, tokenSvc)

			iaSvc := services.NewIAConversacionalService(app.Dao(), intencaoRepo, conversaRepo, geminiSvc, whatsappSvc)

			if err := iaSvc.ProcessarMensagemRecebida(teamID, telefone, mensagem, nomeContato); err != nil {
				g.Error(err, "Erro ao processar mensagem de %s", telefone)
			}
		}()
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
		Nome          string `json:"nome"`
		RemetenteEmail string `json:"remetente_email"`
		ReplyTo       string `json:"reply_to"`
		SMTPHost      string `json:"smtp_host"`
		SMTPPort      int    `json:"smtp_port"`
		SMTPUsuario   string `json:"smtp_usuario"`
		SMTPSenha     string `json:"smtp_senha"`
		Criptografia  string `json:"criptografia"`
		TaxaPorMin    int    `json:"taxa_por_min"`
		TamanhoLote   int    `json:"tamanho_lote"`
		LimiteDiario  int    `json:"limite_diario"`
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

	// Só atualiza senha se foi enviada
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
		"email", email,
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
		// Se a collection não existe, retorna exists=false ao invés de erro
		return c.JSON(200, g.M(
			"exists", false,
			"conexao", nil,
			"email", nil,
		))
	}

	return c.JSON(200, g.M(
		"exists", exists,
		"conexao", con,
		"email", email,
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

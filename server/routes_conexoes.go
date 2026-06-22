package server

import (
	"strings"
	"time"

	"github.com/flarco/g"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/daos"
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

// POST /conexoes/whatsapp/delete
// Remove a instância no WuzAPI e exclui os registros locais da conexão.
func ExcluirConexaoWhatsapp(c echo.Context) error {
	req := NewRequest(c)

	user, err := getUser(c, req.UserID())
	if err != nil || user.ID == "" {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	svc := newWhatsAppService(req)

	if err := svc.DeleteConnection(user.Team.ID); err != nil {
		g.Warn("delete instance error: %v", err)
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

// SyncWebhookURLs atualiza a URL de webhook em todos os usuários wuzapi.
// Chamado no startup para garantir que o wuzapi aponte para o servidor correto.
// IMPORTANTE: Sempre empurra para o wuzapi, NUNCA confia apenas no DB local,
// pois o wuzapi pode limpar o campo Events durante reconexões.
func SyncWebhookURLs(app *pocketbase.PocketBase) {
	cfg := config.NewWhatsMeowConfig()
	webhookURL := cfg.WebhookURL()
	if webhookURL == "" {
		g.Warn("SyncWebhookURLs: WHATSMEOW_WEBHOOK_BASE não configurado, pulando sync")
		return
	}

	wuzClient := wuzapi.NewClient(cfg)
	waRepo := repositories.NewWhatsAppRepo(app.Dao())

	// ── 1. Sincronizar as instâncias que estão no nosso DB ──
	records, err := waRepo.FindAll()
	if err != nil {
		g.Warn("SyncWebhookURLs: erro ao buscar conexões locais: %v", err)
	}

	knownInstanciaIDs := map[string]bool{}

	for _, wa := range records {
		instanciaID := strings.TrimSpace(wa.GetString("instancia_id"))
		if instanciaID == "" {
			continue
		}
		knownInstanciaIDs[instanciaID] = true

		// SEMPRE empurrar para o wuzapi — o DB local pode estar desatualizado
		// (wuzapi limpa Events ao reconectar, mas nosso DB continua mostrando "All")
		g.Info("SyncWebhookURLs: forçando sync instancia=%s webhook=%q events=All", instanciaID, webhookURL)

		if err := wuzClient.UpdateAdminUser(instanciaID, map[string]any{
			"webhook": webhookURL,
			"events":  "All",
		}); err != nil {
			g.Warn("SyncWebhookURLs: falha ao atualizar instancia=%s: %v", instanciaID, err)
			continue
		}

		wa.Set("webhook", webhookURL)
		wa.Set("events", "All")
		if saveErr := app.Dao().SaveRecord(wa); saveErr != nil {
			g.Warn("SyncWebhookURLs: falha ao salvar local instancia=%s: %v", instanciaID, saveErr)
		} else {
			g.Info("SyncWebhookURLs: sync OK instancia=%s", instanciaID)
		}
	}

	// ── 2. Verificar TODAS as instâncias no wuzapi (inclusive órfãs) ──
	// Instâncias órfãs são aquelas que existem no wuzapi mas não no nosso DB,
	// o que pode acontecer por criações duplicadas ou erros.
	allUsers, listErr := wuzClient.ListAllAdminUsers()
	if listErr != nil {
		g.Warn("SyncWebhookURLs: falha ao listar todos os users wuzapi: %v", listErr)
		return
	}

	for _, user := range allUsers {
		if knownInstanciaIDs[user.ID] {
			continue // já sincronizado acima
		}

		// Instância órfã — só atualizar se o webhook aponta para um dos nossos domínios
		if user.Webhook == "" || user.Webhook == webhookURL ||
			strings.Contains(user.Webhook, "suaobra") ||
			strings.Contains(user.Webhook, "ngrok") {

			g.Info("SyncWebhookURLs: sync instância órfã id=%s name=%s connected=%v events=%q webhook=%q",
				user.ID, user.Name, user.Connected, user.Events, user.Webhook)

			if err := wuzClient.UpdateAdminUser(user.ID, map[string]any{
				"webhook": webhookURL,
				"events":  "All",
			}); err != nil {
				g.Warn("SyncWebhookURLs: falha ao sync órfã id=%s: %v", user.ID, err)
			}
		}
	}

	g.Info("SyncWebhookURLs: sync completo — %d instâncias locais, %d total no wuzapi", len(records), len(allUsers))
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
	wa.Set("events", "All")
	if saveErr := req.Dao().SaveRecord(wa); saveErr != nil {
		g.Warn("fix-webhook: falha ao salvar webhook no registro local: %v", saveErr)
	}

	g.Info("fix-webhook: webhook atualizado para %s instancia=%s", webhookURL, instanciaID)
	return c.JSON(200, g.M("ok", true, "webhook_url", webhookURL))
}

// resolveWARecord tenta encontrar o registro conexoes_whatsapp usando vários métodos:
// 1. Por instancia_id (userID do wuzapi no body)
// 2. Por token/numero_e164 (fallback — algumas versões do wuzapi enviam o token)
// 3. Pelo token no header (fallback — wuzapi pode enviar no header "token")
func resolveWARecord(waRepo *repositories.WhatsAppRepo, userID string, tokenHeader string) (*models.Record, string) {
	// Tentar por instancia_id primeiro
	if userID != "" {
		wa, err := waRepo.FindByInstanciaID(userID)
		if err == nil && wa != nil && wa.Id != "" {
			g.Debug("webhook/whatsmeow: resolve por instancia_id=%q → wa=%s", userID, wa.Id)
			return wa, strings.TrimSpace(wa.GetString("instancia_id"))
		}

		// Fallback: userID pode ser o token (numero_e164)
		wa, err = waRepo.FindByToken(userID)
		if err == nil && wa != nil && wa.Id != "" {
			instID := strings.TrimSpace(wa.GetString("instancia_id"))
			g.Info("webhook/whatsmeow: resolve por token(body)=%q → wa=%s instancia=%s", maskWebhookPhone(userID), wa.Id, instID)
			return wa, instID
		}
	}

	// Fallback: token no header
	if tokenHeader != "" && tokenHeader != userID {
		wa, err := waRepo.FindByToken(tokenHeader)
		if err == nil && wa != nil && wa.Id != "" {
			instID := strings.TrimSpace(wa.GetString("instancia_id"))
			g.Info("webhook/whatsmeow: resolve por token(header) → wa=%s instancia=%s", wa.Id, instID)
			return wa, instID
		}
	}

	return nil, ""
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
	tokenHeader := strings.TrimSpace(c.Request().Header.Get("token"))

	// Fallback para formato antigo (testes manuais)
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

	switch eventType {
	case "connected":
		waRepo := repositories.NewWhatsAppRepo(app.Dao())
		wa, instanciaID := resolveWARecord(waRepo, userID, tokenHeader)
		if wa == nil || wa.Id == "" {
			g.Warn("webhook/whatsmeow: Connected — nenhum registro encontrado userID=%q tokenHeader=%q", userID, maskWebhookPhone(tokenHeader))
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

		// Wuzapi limpa Events ao reconectar — re-sincronizar webhook + events
		if instanciaID != "" {
			go func(iid string) {
				cfg := config.NewWhatsMeowConfig()
				webhookURL := cfg.WebhookURL()
				if webhookURL == "" {
					return
				}
				wuzClient := wuzapi.NewClient(cfg)
				if err := wuzClient.UpdateAdminUser(iid, map[string]any{
					"webhook": webhookURL,
					"events":  "All",
				}); err != nil {
					g.Warn("webhook/whatsmeow: falha ao re-sync events após Connected instancia=%s: %v", iid, err)
				} else {
					g.Info("webhook/whatsmeow: events re-sync OK após Connected instancia=%s webhook=%s", iid, webhookURL)
				}
			}(instanciaID)
		}

	case "message":
		waRepo := repositories.NewWhatsAppRepo(app.Dao())
		wa, _ := resolveWARecord(waRepo, userID, tokenHeader)
		if wa == nil || wa.Id == "" {
			g.Warn("webhook/whatsmeow: Message — nenhum registro encontrado userID=%q tokenHeader=%q", userID, maskWebhookPhone(tokenHeader))
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

		// Suporte para MessageSource aninhado (algumas versões de wuzapi)
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
			return c.JSON(200, g.M("ok", true))
		}

		// Ignorar mensagens de grupo
		if cast.ToBool(info["IsGroup"]) {
			g.Debug("webhook/whatsmeow: ignorando mensagem de grupo")
			return c.JSON(200, g.M("ok", true))
		}

		// Ignorar mensagens de canais/newsletters (@newsletter)
		chatStr := strings.TrimSpace(cast.ToString(info["Chat"]))
		if chatStr == "" {
			if chatMap, ok := info["Chat"].(map[string]any); ok {
				chatStr = strings.TrimSpace(cast.ToString(chatMap["Server"]))
			}
		}
		if strings.Contains(chatStr, "@newsletter") || strings.Contains(chatStr, "newsletter") {
			g.Debug("webhook/whatsmeow: ignorando mensagem de newsletter/canal")
			return c.JSON(200, g.M("ok", true))
		}

		// Extrair telefone real do evento (prioriza JIDs @s.whatsapp.net e evita LID).
		telefone := extractBestPhoneFromInfo(info)

		if telefone == "" {
			g.Warn("webhook/whatsmeow: telefone vazio após extração de JID Sender=%v SenderAlt=%v Chat=%v", info["Sender"], info["SenderAlt"], info["Chat"])
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
			return c.JSON(200, g.M("ok", true))
		}

		g.Info("webhook/whatsmeow: Lead %s encontrado em campanha com IA ativa, processando...", maskWebhookPhone(telefone))

		go func(teamID, telefone, mensagem, nomeContato, msgID string) {
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
			if err := iaSvc.ProcessarMensagemRecebida(teamID, telefone, mensagem, nomeContato, msgID); err != nil {
				g.Error(err, "webhook/whatsmeow: erro ao processar mensagem telefone=%s", maskWebhookPhone(telefone))
				return
			}
			g.Info("webhook/whatsmeow: goroutine IA end telefone=%s", maskWebhookPhone(telefone))
		}(teamID, telefone, mensagem, nomeContato, msgIDExterno)
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

// pauseConversationalIAOnHumanMessage pausa conversa ativa para o telefone quando
// detectamos mensagem enviada pelo próprio usuário (intervenção manual).
func pauseConversationalIAOnHumanMessage(dao *daos.Dao, teamID string, info map[string]any, evt map[string]any) {
	if dao == nil || strings.TrimSpace(teamID) == "" {
		return
	}

	candidatos := resolvePhonesForManualIntervention(info)
	if len(candidatos) == 0 {
		g.Debug("webhook/whatsmeow: intervenção manual sem telefone elegível para pausar IA")
		return
	}

	conversaRepo := repositories.NewConversaRepo(dao)
	conversa := conversaRepo.FindByTelefoneCandidates(teamID, candidatos)
	if conversa == nil {
		g.Debug("webhook/whatsmeow: intervenção manual sem conversa encontrada team=%s candidatos=%v", teamID, candidatos)
		return
	}

	statusAtual := strings.ToUpper(strings.TrimSpace(conversa.GetString("status")))
	if statusAtual != "ATIVA" {
		return
	}

	// Evita pausar quando o webhook apenas ecoa uma mensagem que a própria IA acabou de enviar.
	msgEnviada := normalizeWAMessageForCompare(extractWAMessageText(evt))
	ultimaModel := normalizeWAMessageForCompare(lastModelMessageFromConversa(conversa))
	if msgEnviada != "" && ultimaModel != "" && msgEnviada == ultimaModel {
		return
	}

	conversa.Set("status", "PAUSADA")
	conversa.Set("ultima_mensagem_em", time.Now().UTC())
	if err := conversaRepo.Save(conversa); err != nil {
		g.Warn("webhook/whatsmeow: falha ao pausar conversa por intervenção manual conversa=%s err=%v", conversa.Id, err)
		return
	}

	g.Info(
		"webhook/whatsmeow: conversa pausada por intervenção manual conversa=%s team=%s telefone=%s",
		conversa.Id,
		teamID,
		maskWebhookPhone(candidatos[0]),
	)
}

func resolvePhonesForManualIntervention(info map[string]any) []string {
	candidates := make([]string, 0, 4)
	seen := map[string]struct{}{}

	add := func(raw string) {
		phone := normalizeWASender(raw)
		if phone == "" {
			return
		}
		if _, ok := seen[phone]; ok {
			return
		}
		seen[phone] = struct{}{}
		candidates = append(candidates, phone)
	}

	// Em mensagens enviadas manualmente (IsFromMe=true), o contato de destino costuma vir aqui.
	add(cast.ToString(info["RecipientAlt"]))
	add(cast.ToString(info["recipientAlt"]))

	// Fallback: em alguns payloads o número útil vem no SenderAlt quando há LID no Sender.
	add(cast.ToString(info["SenderAlt"]))
	add(cast.ToString(info["senderAlt"]))

	// Último fallback: tenta extração genérica dos campos Info.
	add(extractBestPhoneFromInfo(info))

	out := make([]string, 0, len(candidates)*2)
	outSeen := map[string]struct{}{}
	for _, phone := range candidates {
		for _, v := range buildTelefoneCandidates(phone) {
			if _, ok := outSeen[v]; ok {
				continue
			}
			outSeen[v] = struct{}{}
			out = append(out, v)
		}
	}

	return out
}

func lastModelMessageFromConversa(conversa *models.Record) string {
	if conversa == nil {
		return ""
	}

	raw := conversa.Get("mensagens")
	var mensagens []map[string]any

	switch v := raw.(type) {
	case []map[string]any:
		mensagens = v
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				mensagens = append(mensagens, m)
				continue
			}
			if m, ok := item.(map[string]interface{}); ok {
				tmp := map[string]any{}
				for k, val := range m {
					tmp[k] = val
				}
				mensagens = append(mensagens, tmp)
			}
		}
	}

	for i := len(mensagens) - 1; i >= 0; i-- {
		role := strings.TrimSpace(strings.ToLower(cast.ToString(mensagens[i]["role"])))
		if role != "model" && role != "assistant" {
			continue
		}
		content := strings.TrimSpace(cast.ToString(mensagens[i]["content"]))
		if content != "" {
			return content
		}
	}

	return ""
}

func normalizeWAMessageForCompare(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return ""
	}
	return strings.Join(strings.Fields(s), " ")
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

// extractWAInfoMessageID tenta obter o identificador da mensagem no evento WhatsApp (deduplicação).
func extractWAInfoMessageID(info map[string]any) string {
	if info == nil {
		return ""
	}
	for _, k := range []string{"ID", "Id", "id"} {
		if v := strings.TrimSpace(cast.ToString(info[k])); v != "" {
			return v
		}
	}
	return ""
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

// extractJIDPhone extrai o número de telefone de um JID WhatsApp,
// suportando tanto string ("5511999998888@s.whatsapp.net") quanto
// objeto/map ({"User": "5511999998888", "Server": "s.whatsapp.net"}).
func extractJIDPhone(raw any) string {
	if raw == nil {
		return ""
	}

	// Se for um JID objeto (map), extrair o campo "User"
	if m, ok := raw.(map[string]any); ok {
		user := strings.TrimSpace(cast.ToString(m["User"]))
		if user != "" {
			return normalizeWASender(user)
		}
		// Fallback: tentar campo "user" (minúsculo)
		user = strings.TrimSpace(cast.ToString(m["user"]))
		if user != "" {
			return normalizeWASender(user)
		}
		return ""
	}

	// Se for string, normalizar diretamente
	s := strings.TrimSpace(cast.ToString(raw))
	if s == "" || strings.HasPrefix(s, "map[") {
		// Proteção contra cast.ToString de um map não capturado acima
		return ""
	}
	return normalizeWASender(s)
}

// isLikelyLID verifica se o número parece ser um LID (Linked Identity) do WhatsApp
// ao invés de um número de telefone real. LIDs são IDs internos que não correspondem
// a números de telefone. Heurística: se NÃO começa com código de país conhecido
// e tem formato atípico, provavelmente é um LID.
func isLikelyLID(phone string) bool {
	if phone == "" {
		return false
	}
	// Telefones brasileiros com código de país começam com 55 e têm 12-13 dígitos
	// Telefones internacionais geralmente começam com 1-9 seguido de padrões reconhecíveis
	// LIDs são números longos que não seguem formatos de telefone (ex: 216192111399060)
	if len(phone) > 13 && !strings.HasPrefix(phone, "55") {
		return true
	}
	return false
}

func buildTelefoneCandidates(telefone string) []string {
	telefone = normalizeWASender(telefone)
	if telefone == "" {
		return nil
	}
	if isLikelyLID(telefone) {
		return []string{telefone}
	}

	seen := map[string]struct{}{}
	out := make([]string, 0, 6)

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

	// Com/sem código de país 55
	if strings.HasPrefix(telefone, "55") && len(telefone) > 10 {
		add(telefone[2:])
	}
	if !strings.HasPrefix(telefone, "55") && len(telefone) >= 10 {
		add("55" + telefone)
	}

	// Variantes do 9º dígito brasileiro:
	// Celulares no Brasil podem ter 8 ou 9 dígitos depois do DDD.
	// Exemplo: 5511999998888 (com 9) ↔ 551199998888 (sem 9)
	withCC := telefone
	if !strings.HasPrefix(withCC, "55") && len(withCC) >= 10 {
		withCC = "55" + withCC
	}

	if strings.HasPrefix(withCC, "55") && len(withCC) >= 12 {
		ddd := withCC[2:4]
		local := withCC[4:]

		if len(local) == 9 && local[0] == '9' {
			// Tem 9º dígito → tentar sem
			add("55" + ddd + local[1:])
			add(ddd + local[1:])
		} else if len(local) == 8 {
			// Não tem 9º dígito → tentar com
			add("55" + ddd + "9" + local)
			add(ddd + "9" + local)
		}
	}

	return out
}

func extractBestPhoneFromInfo(info map[string]any) string {
	if len(info) == 0 {
		return ""
	}

	orderedKeys := []string{
		"SenderAlt", "ChatAlt", "ParticipantAlt",
		"Sender", "Chat", "Participant",
		"From", "To", "Author",
	}

	// 1) Prioridade para campos conhecidos com @s.whatsapp.net
	for _, k := range orderedKeys {
		if raw, ok := info[k]; ok {
			if phone := extractPhoneFromAny(raw, true); phone != "" {
				return phone
			}
		}
	}

	// 2) Campos conhecidos sem exigir @s.whatsapp.net
	for _, k := range orderedKeys {
		if raw, ok := info[k]; ok {
			if phone := extractPhoneFromAny(raw, false); phone != "" {
				return phone
			}
		}
	}

	// 3) Fallback recursivo: vasculha todo payload do info
	if phone := extractPhoneFromAny(info, true); phone != "" {
		return phone
	}
	return extractPhoneFromAny(info, false)
}

func extractPhoneFromAny(v any, requireServerJID bool) string {
	switch t := v.(type) {
	case string:
		raw := strings.TrimSpace(t)
		if raw == "" {
			return ""
		}
		if requireServerJID && !strings.Contains(raw, "@s.whatsapp.net") {
			return ""
		}
		phone := extractJIDPhone(raw)
		if phone == "" {
			return ""
		}
		if requireServerJID && isLikelyLID(phone) {
			return ""
		}
		return phone

	case map[string]any:
		// Caso de JID objeto: {"User":"5511...", "Server":"s.whatsapp.net"}
		user := strings.TrimSpace(cast.ToString(t["User"]))
		server := strings.TrimSpace(cast.ToString(t["Server"]))
		if user != "" {
			if !requireServerJID || strings.Contains(server, "s.whatsapp.net") {
				phone := normalizeWASender(user)
				if phone != "" && (!requireServerJID || !isLikelyLID(phone)) {
					return phone
				}
			}
		}
		// fallback lowercase keys
		user = strings.TrimSpace(cast.ToString(t["user"]))
		server = strings.TrimSpace(cast.ToString(t["server"]))
		if user != "" {
			if !requireServerJID || strings.Contains(server, "s.whatsapp.net") {
				phone := normalizeWASender(user)
				if phone != "" && (!requireServerJID || !isLikelyLID(phone)) {
					return phone
				}
			}
		}
		for _, child := range t {
			if phone := extractPhoneFromAny(child, requireServerJID); phone != "" {
				return phone
			}
		}
		return ""

	case []any:
		for _, child := range t {
			if phone := extractPhoneFromAny(child, requireServerJID); phone != "" {
				return phone
			}
		}
		return ""

	default:
		return extractPhoneFromAny(cast.ToString(v), requireServerJID)
	}
}

func registrarRespostaWhatsAppSemIA(
	dao *daos.Dao,
	conversaRepo *repositories.ConversaRepo,
	conversa *models.Record,
	teamID, telefone, nomeContato, mensagem string,
	dest *models.Record,
	messageIDExterno string,
) error {
	if conversaRepo == nil {
		return nil
	}

	teamID = strings.TrimSpace(teamID)
	telefone = normalizeWASender(telefone)
	nomeContato = strings.TrimSpace(nomeContato)
	mensagem = strings.TrimSpace(mensagem)
	if teamID == "" || telefone == "" || mensagem == "" {
		return nil
	}

	campanhaID := ""
	destinatarioID := ""
	if dest != nil {
		campanhaID = strings.TrimSpace(dest.GetString("campanha_id"))
		destinatarioID = strings.TrimSpace(dest.Id)
	}
	if campanhaID == "" && conversa != nil {
		campanhaID = strings.TrimSpace(conversa.GetString("campanha_id"))
	}
	if destinatarioID == "" && conversa != nil {
		destinatarioID = strings.TrimSpace(conversa.GetString("destinatario_id"))
	}

	// Sem campanha vinculada não entra no relatório de campanhas.
	if campanhaID == "" {
		return nil
	}

	now := time.Now().UTC()

	if conversa == nil {
		payload := map[string]any{
			"team_id":            teamID,
			"campanha_id":        campanhaID,
			"telefone":           telefone,
			"nome_contato":       nomeContato,
			"mensagens":          []map[string]any{},
			"status":             "PAUSADA",
			"ultima_mensagem_em": now,
		}
		if destinatarioID != "" {
			payload["destinatario_id"] = destinatarioID
		}
		var err error
		conversa, err = conversaRepo.Create(payload)
		if err != nil {
			return err
		}
	}

	if strings.ToUpper(strings.TrimSpace(conversa.GetString("status"))) == "ATIVA" {
		conversa.Set("status", "PAUSADA")
	}
	if strings.TrimSpace(conversa.GetString("campanha_id")) == "" {
		conversa.Set("campanha_id", campanhaID)
	}
	if strings.TrimSpace(conversa.GetString("destinatario_id")) == "" && destinatarioID != "" {
		conversa.Set("destinatario_id", destinatarioID)
	}
	if strings.TrimSpace(conversa.GetString("nome_contato")) == "" && nomeContato != "" {
		conversa.Set("nome_contato", nomeContato)
	}

	mensagens := make([]map[string]any, 0)
	raw := conversa.Get("mensagens")
	switch v := raw.(type) {
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				mensagens = append(mensagens, map[string]any{
					"role":      cast.ToString(m["role"]),
					"content":   cast.ToString(m["content"]),
					"timestamp": cast.ToString(m["timestamp"]),
				})
				continue
			}
			if m, ok := item.(map[string]any); ok {
				mensagens = append(mensagens, map[string]any{
					"role":      cast.ToString(m["role"]),
					"content":   cast.ToString(m["content"]),
					"timestamp": cast.ToString(m["timestamp"]),
				})
			}
		}
	case []map[string]interface{}:
		for _, item := range v {
			mensagens = append(mensagens, map[string]any{
				"role":      cast.ToString(item["role"]),
				"content":   cast.ToString(item["content"]),
				"timestamp": cast.ToString(item["timestamp"]),
			})
		}
	}

	mensagens = append(mensagens, map[string]any{
		"role":      "user",
		"content":   mensagem,
		"timestamp": now.Format(time.RFC3339),
	})

	conversa.Set("mensagens", mensagens)
	conversa.Set("ultima_mensagem_em", now)

	if err := conversaRepo.Save(conversa); err != nil {
		return err
	}

	_ = repositories.SaveCampanhaLeadResposta(dao, repositories.CampanhaLeadRespostaInput{
		TeamID:           teamID,
		CampanhaID:       campanhaID,
		DestinatarioID:   destinatarioID,
		ConversaID:       conversa.Id,
		Canal:            "WHATSAPP",
		TelefoneE164:     telefone,
		NomeContato:      nomeContato,
		Corpo:            mensagem,
		MessageIDExterno: strings.TrimSpace(messageIDExterno),
		RecebidaEm:       now,
	})
	return nil
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

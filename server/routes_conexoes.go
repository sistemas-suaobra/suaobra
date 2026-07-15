package server

import (
	"strings"
	"time"

	"github.com/flarco/g"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
	"github.com/spf13/cast"

	"github.com/suaobra/suaobra-app/server/repositories"
	"github.com/suaobra/suaobra-app/server/services"
)

func newEmailService(req Request) *services.EmailService {
	conRepo := repositories.NewConexaoRepo(req.Dao())
	emailRepo := repositories.NewEmailRepo(req.Dao())

	return services.NewEmailService(conRepo, emailRepo)
}

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
// detectamos mensagem enviada pelo prÃ³prio usuÃ¡rio (intervenÃ§Ã£o manual).
func pauseConversationalIAOnHumanMessage(dao *daos.Dao, teamID string, info map[string]any, evt map[string]any) {
	if dao == nil || strings.TrimSpace(teamID) == "" {
		return
	}

	candidatos := resolvePhonesForManualIntervention(info)
	if len(candidatos) == 0 {
		g.Debug("webhook/whatsmeow: intervenÃ§Ã£o manual sem telefone elegÃ­vel para pausar IA")
		return
	}

	conversaRepo := repositories.NewConversaRepo(dao)
	conversa := conversaRepo.FindByTelefoneCandidates(teamID, candidatos)
	if conversa == nil {
		g.Debug("webhook/whatsmeow: intervenÃ§Ã£o manual sem conversa encontrada team=%s candidatos=%v", teamID, candidatos)
		return
	}

	statusAtual := strings.ToUpper(strings.TrimSpace(conversa.GetString("status")))
	if statusAtual != "ATIVA" {
		return
	}

	// Evita pausar quando o webhook apenas ecoa uma mensagem que a prÃ³pria IA acabou de enviar.
	msgEnviada := normalizeWAMessageForCompare(extractWAMessageText(evt))
	ultimaModel := normalizeWAMessageForCompare(lastModelMessageFromConversa(conversa))
	if msgEnviada != "" && ultimaModel != "" && msgEnviada == ultimaModel {
		return
	}

	conversa.Set("status", "PAUSADA")
	conversa.Set("ultima_mensagem_em", time.Now().UTC())
	if err := conversaRepo.Save(conversa); err != nil {
		g.Warn("webhook/whatsmeow: falha ao pausar conversa por intervenÃ§Ã£o manual conversa=%s err=%v", conversa.Id, err)
		return
	}

	g.Info(
		"webhook/whatsmeow: conversa pausada por intervenÃ§Ã£o manual conversa=%s team=%s telefone=%s",
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

	// Fallback: em alguns payloads o nÃºmero Ãºtil vem no SenderAlt quando hÃ¡ LID no Sender.
	add(cast.ToString(info["SenderAlt"]))
	add(cast.ToString(info["senderAlt"]))

	// Ãšltimo fallback: tenta extraÃ§Ã£o genÃ©rica dos campos Info.
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

// extractWAInfoMessageID tenta obter o identificador da mensagem no evento WhatsApp (deduplicaÃ§Ã£o).
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

// extractJIDPhone extrai o nÃºmero de telefone de um JID WhatsApp,
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
		// Fallback: tentar campo "user" (minÃºsculo)
		user = strings.TrimSpace(cast.ToString(m["user"]))
		if user != "" {
			return normalizeWASender(user)
		}
		return ""
	}

	// Se for string, normalizar diretamente
	s := strings.TrimSpace(cast.ToString(raw))
	if s == "" || strings.HasPrefix(s, "map[") {
		// ProteÃ§Ã£o contra cast.ToString de um map nÃ£o capturado acima
		return ""
	}
	return normalizeWASender(s)
}

// isLikelyLID verifica se o nÃºmero parece ser um LID (Linked Identity) do WhatsApp
// ao invÃ©s de um nÃºmero de telefone real. LIDs sÃ£o IDs internos que nÃ£o correspondem
// a nÃºmeros de telefone. HeurÃ­stica: se NÃƒO comeÃ§a com cÃ³digo de paÃ­s conhecido
// e tem formato atÃ­pico, provavelmente Ã© um LID.
func isLikelyLID(phone string) bool {
	if phone == "" {
		return false
	}
	// Telefones brasileiros com cÃ³digo de paÃ­s comeÃ§am com 55 e tÃªm 12-13 dÃ­gitos
	// Telefones internacionais geralmente comeÃ§am com 1-9 seguido de padrÃµes reconhecÃ­veis
	// LIDs sÃ£o nÃºmeros longos que nÃ£o seguem formatos de telefone (ex: 216192111399060)
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

	// Com/sem cÃ³digo de paÃ­s 55
	if strings.HasPrefix(telefone, "55") && len(telefone) > 10 {
		add(telefone[2:])
	}
	if !strings.HasPrefix(telefone, "55") && len(telefone) >= 10 {
		add("55" + telefone)
	}

	// Variantes do 9Âº dÃ­gito brasileiro:
	// Celulares no Brasil podem ter 8 ou 9 dÃ­gitos depois do DDD.
	// Exemplo: 5511999998888 (com 9) â†” 551199998888 (sem 9)
	withCC := telefone
	if !strings.HasPrefix(withCC, "55") && len(withCC) >= 10 {
		withCC = "55" + withCC
	}

	if strings.HasPrefix(withCC, "55") && len(withCC) >= 12 {
		ddd := withCC[2:4]
		local := withCC[4:]

		if len(local) == 9 && local[0] == '9' {
			// Tem 9Âº dÃ­gito â†’ tentar sem
			add("55" + ddd + local[1:])
			add(ddd + local[1:])
		} else if len(local) == 8 {
			// NÃ£o tem 9Âº dÃ­gito â†’ tentar com
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

	// Sem campanha vinculada nÃ£o entra no relatÃ³rio de campanhas.
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

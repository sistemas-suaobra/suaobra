package server

import (
	"os"
	"strings"

	"github.com/flarco/g"
	"github.com/flarco/g/net"
	"github.com/labstack/echo/v5"
	"github.com/samber/lo"
	"github.com/spf13/cast"

	"github.com/suaobra/suaobra-app/server/repositories"
	"github.com/suaobra/suaobra-app/server/services"
)

var (
	MessengerAuthorization = "App-Token " + os.Getenv("MESSENGER_TOKEN")
	MessengerBaseURL       = os.Getenv("MESSENGER_BASE_URL")
	MessengerHeaders       = map[string]string{
		echo.HeaderAuthorization: MessengerAuthorization,
		echo.HeaderContentType:   echo.MIMEApplicationJSON,
	}

	// Gemini API Configuration
	GeminiAPIKey  = os.Getenv("GEMINI_API_KEY")
	GeminiBaseURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent"
	GeminiHeaders = map[string]string{
		echo.HeaderContentType: echo.MIMEApplicationJSON,
	}
)

// called from messenger
func MessengerSync(c echo.Context) error {
	logMessengerRequestMeta(c, "MessengerSync: incoming request")

	req := NewRequest(c)
	if req.Error != nil {
		g.Error(req.Error, "MessengerSync: erro ao montar request")
		return ErrJSON(400, req.Error)
	}

	type PayloadType string

	var (
		PayloadTypeMessageQueue    PayloadType = "message_queue"
		PayloadTypeWhatsAppMessage PayloadType = "whats_app_message"
	)

	payloadType := PayloadType(req.Payload.String("type"))
	data := cast.ToStringMap(req.Payload["data"])

	g.Info(
		"MessengerSync: payload_type=%s payload=%s",
		payloadType,
		truncateMessengerLog(g.Marshal(req.Payload), 5000),
	)

	switch payloadType {
	case PayloadTypeMessageQueue:
		status := cast.ToString(data["status"])
		teamID := cast.ToString(data["user_id"])

		g.Info(
			"MessengerSync[message_queue]: status=%s team_id=%s data=%s",
			status,
			teamID,
			truncateMessengerLog(g.Marshal(data), 3000),
		)

		metaMap := cast.ToStringMap(data["meta"])
		if len(metaMap) == 0 {
			g.Warn("MessengerSync[message_queue]: meta ausente")
			return ErrJSON(404, g.Error("missing meta key"))
		}

		leadID := cast.ToString(metaMap["lead_id"])
		if leadID == "" {
			g.Warn("MessengerSync[message_queue]: meta.lead_id ausente meta=%s", truncateMessengerLog(g.Marshal(metaMap), 2000))
			return ErrJSON(404, g.Error("missing meta.lead_id key"))
		}

		isOwner := cast.ToString(metaMap["contact_role"]) == "owner"
		contactedAtCol := lo.Ternary(isOwner, "owner_contacted_at", "professional_contacted_at")

		messengerUpdateLead, _ := templates.ReadFile("templates/messenger/update_lead.sql")
		_, err := req.SqlQuery(g.Rm(string(messengerUpdateLead), g.M(
			"status",           status,
			"team_id",          teamID,
			"lead_id",          leadID,
			"contacted_at_col", contactedAtCol,
		)))
		if err != nil {
			g.Error(err, "MessengerSync[message_queue]: erro ao atualizar lead team=%s lead=%s", teamID, leadID)
			return ErrJSON(502, err, "error updating messenger contacted_at")
		}

		g.Info(
			"MessengerSync[message_queue]: lead atualizado com sucesso team=%s lead=%s contacted_col=%s",
			teamID,
			leadID,
			contactedAtCol,
		)

	case PayloadTypeWhatsAppMessage:
		g.Info(
			"MessengerSync[WA]: raw_data=%s",
			truncateMessengerLog(g.Marshal(data), 8000),
		)

		if shouldIgnoreWhatsAppWebhook(data) {
			g.Info("MessengerSync[WA]: ignorado por regra (grupo/from_me/sem conteúdo útil)")
			break
		}

		teamID := resolveWhatsAppTeamID(req, data)
		telefone := extractWhatsAppPhone(data)
		mensagem := extractWhatsAppText(data)
		nomeContato := extractWhatsAppContactName(data)

		g.Info(
			"MessengerSync[WA]: resolved team_id=%s telefone=%s nome=%s msg_len=%d mensagem=%s",
			teamID,
			maskMessengerPhone(telefone),
			nomeContato,
			len(strings.TrimSpace(mensagem)),
			truncateMessengerLog(mensagem, 1200),
		)

		if teamID == "" {
			g.Warn("MessengerSync[WA]: não foi possível resolver team_id data=%s", truncateMessengerLog(g.Marshal(data), 5000))
			break
		}

		if telefone == "" {
			g.Warn("MessengerSync[WA]: telefone não encontrado team=%s data=%s", teamID, truncateMessengerLog(g.Marshal(data), 5000))
			break
		}

		if strings.TrimSpace(mensagem) == "" {
			g.Warn("MessengerSync[WA]: mensagem vazia team=%s telefone=%s", teamID, maskMessengerPhone(telefone))
			break
		}

		intencaoRepo := repositories.NewIntencaoRepo(req.Dao())
		conversaRepo := repositories.NewConversaRepo(req.Dao())
		waSvc := newWhatsAppServiceFromRequest(req)
		geminiSvc := services.NewGeminiService()

		iaSvc := services.NewIAConversacionalService(
			req.Dao(),
			intencaoRepo,
			conversaRepo,
			waSvc,
			geminiSvc,
		)

		g.Info(
			"MessengerSync[WA]: chamando IA team=%s telefone=%s",
			teamID,
			maskMessengerPhone(telefone),
		)

		if err := iaSvc.ProcessarMensagemRecebida(teamID, telefone, mensagem, nomeContato); err != nil {
			g.Error(
				err,
				"MessengerSync[WA]: erro ao processar mensagem team=%s telefone=%s mensagem=%s",
				teamID,
				maskMessengerPhone(telefone),
				truncateMessengerLog(mensagem, 800),
			)
			break
		}

		g.Info(
			"MessengerSync[WA]: IA processou com sucesso team=%s telefone=%s",
			teamID,
			maskMessengerPhone(telefone),
		)

	default:
		g.Warn(
			"MessengerSync: payload_type desconhecido=%s payload=%s",
			payloadType,
			truncateMessengerLog(g.Marshal(req.Payload), 4000),
		)
	}

	return c.String(200, "OK")
}

// called from suaobra frontend
func MessengerQueueSubmit(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(400, req.Error)
	}

	// get the numbers for the names
	if req.Payload["records"] == nil {
		return ErrJSON(400, g.Error("missing records"))
	}

	var records []map[string]any
	if err := g.JSONConvert(req.Payload["records"], &records); err != nil {
		return ErrJSON(400, g.Error(err, "could not cast records to slice of map"))
	}

	// get names, and then name to number map
	names := make([]string, len(records))
	for i, rec := range records {
		names[i] = g.F("'%s'", rec["contact_name"]) // for sql injection
	}

	namePhoneSQL, _ := templates.ReadFile("templates/messenger/name_phones.sql")
	recordsRecipients, err := req.SqlQueryRecords(g.Rm(string(namePhoneSQL), g.M("names", strings.Join(names, ","))))
	if err != nil {
		return ErrJSON(502, err, "error querying phone data")
	}

	// get name to number map
	nameNumberMap := map[string]string{}
	for _, rec := range recordsRecipients {
		name := cast.ToString(rec["nome"])
		phone := formatPhoneNumber(cast.ToString(rec["telephone"]))
		if phone == "" {
			g.Warn("could not format phone number", rec)
			continue
		}
		nameNumberMap[name] = phone
	}

	// submit each record
	for _, record := range records {
		recipient := nameNumberMap[cast.ToString(record["contact_name"])]
		if recipient == "" {
			continue
		}

		// set user
		req.Payload["user"] = g.M("id", req.TeamID())
		req.Payload["message"] = g.M(
			"app_id", "suaobra",
			"user_id", req.TeamID(),
			"platform", "wa",
			"recipient", recipient,
			"body", record["text"],
			"meta", g.M(
				"obra_id", record["obra_id"],
				"contact_name", record["contact_name"],
				"contact_role", record["contact_role"],
			),
			"status", "pending",
		)

		payload := g.Marshal(req.Payload)
		url := g.F("%s/message/queue/add", MessengerBaseURL)
		resp, respBytes, err := net.ClientDo("POST", url, strings.NewReader(payload), MessengerHeaders)
		if strings.Contains(string(respBytes), "recipient already") {
			goto setPending
		}

		if err != nil {
			return ErrJSON(400, err)
		} else if resp.StatusCode >= 400 {
			return ErrJSON(resp.StatusCode, g.Error("could not submit to messenger"))
		}

	setPending:
		req.Payload["team_id"] = req.TeamID()
		req.Payload["obra_id"] = record["obra_id"]
		req.Payload["toggle_col"] = "owner_contact_pending_at"
		if record["contact_role"] == "Profissional" {
			req.Payload["toggle_col"] = "professional_contact_pending_at"
		}
		if _, err = LeadToggle(&req); err != nil {
			return err
		}
	}

	return c.JSON(200, g.M())
}

// called from suaobra frontend
func MessengerGenerateTemplates(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(400, req.Error)
	}

	teamRecord, err := req.TeamRecord()
	if err != nil {
		return ErrJSON(400, g.Error(err, "could not find team record"))
	}

	teamProps, err := g.UnmarshalMap(teamRecord.GetString("properties"))
	if err != nil {
		return ErrJSON(400, g.Error(err, "could not get team properties"))
	}

	storeName := cast.ToString(teamProps["name"])
	storeDescription := cast.ToString(teamProps["description"])
	templatesMap, _ := g.UnmarshalMap(cast.ToString(teamProps["templates"]))
	promptContext := strings.TrimSpace(cast.ToString(templatesMap["context"]))

	templateMap := g.M()

	req.Context.Wg.Read.Add()
	go func() bool {
		defer req.Context.Wg.Read.Done()
		payload := g.Marshal(g.M(
			"user", g.M("id", req.TeamID()),
			"prompt_key", "personalized_owner_message",
			"prompts", []string{
				g.F("Here is the store's name: %s\n\n", storeName),
				g.F("Here is the store's description:\n%s\n\n", storeDescription),
				lo.Ternary(promptContext == "", "", g.F("Emphasize on this:\n%s\n\n", promptContext)),
			},
		))

		url := g.F("%s/user/templates", MessengerBaseURL)
		resp, respBytes, err := net.ClientDo("POST", url, strings.NewReader(payload), MessengerHeaders)
		if err != nil {
			return req.Context.CaptureErr(err)
		} else if resp.StatusCode >= 400 {
			return req.Context.CaptureErr(g.Error("could not submit to messenger"))
		}

		respMap, _ := g.UnmarshalMap(string(respBytes))
		templateMap["owner"] = respMap["templates"]
		return false
	}()
	req.Context.Wg.Read.Wait()

	req.Context.Wg.Read.Add()
	go func() bool {
		defer req.Context.Wg.Read.Done()
		payload := g.Marshal(g.M(
			"user", g.M("id", req.TeamID()),
			"prompt_key", "personalized_professional_message",
		))

		url := g.F("%s/user/templates", MessengerBaseURL)
		resp, respBytes, err := net.ClientDo("POST", url, strings.NewReader(payload), MessengerHeaders)
		if err != nil {
			return req.Context.CaptureErr(err)
		} else if resp.StatusCode >= 400 {
			return req.Context.CaptureErr(g.Error("could not submit to messenger"))
		}

		respMap, _ := g.UnmarshalMap(string(respBytes))
		templateMap["professional"] = respMap["templates"]

		return false
	}()

	req.Context.Wg.Read.Wait()

	if err := req.Context.Err(); err != nil {
		return ErrJSON(400, g.Error(err, "error getting generating templates"))
	}

	return c.JSON(200, templateMap)
}

// called from suaobra frontend
func MessengerGenerateLeadIntroduction(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(400, req.Error)
	}

	if GeminiAPIKey == "" {
		g.LogError(g.Error("GEMINI_API_KEY not configured"))
		return ErrJSON(500, g.Error("AI service not configured"), "service configuration missing")
	}

	teamRecord, err := req.TeamRecord()
	if err != nil {
		return ErrJSON(400, g.Error(err, "could not find team record"))
	}

	teamProps, err := g.UnmarshalMap(teamRecord.GetString("properties"))
	if err != nil {
		return ErrJSON(400, g.Error(err, "could not get team properties"))
	}

	originalLeadIntroductionText := cast.ToString(teamProps["lead_introduction_text"])

	storeName := cast.ToString(teamProps["name"])
	storeDescription := cast.ToString(teamProps["description"])
	storeIndustry := cast.ToString(teamProps["industry"])
	storeKeywords := cast.ToString(teamProps["keywords"])
	foundedDate := cast.ToString(teamProps["founded_date"])

	if strings.TrimSpace(storeName) == "" {
		return ErrJSON(400, g.Error("store name is required to generate lead introduction"))
	}

	promptParts := []string{
		"Você deve criar um texto de apresentação profissional e informal para prospecção de leads.",
		g.F("Nome da empresa: %s", storeName),
	}

	if storeDescription != "" {
		promptParts = append(promptParts, g.F("Descrição da empresa: %s", storeDescription))
	}

	if storeIndustry != "" {
		promptParts = append(promptParts, g.F("Setor de atuação: %s", storeIndustry))
	}

	if storeKeywords != "" {
		promptParts = append(promptParts, g.F("Palavras-chave importantes: %s", storeKeywords))
	}

	if foundedDate != "" {
		promptParts = append(promptParts, g.F("Data de fundação: %s", foundedDate))
	}

	promptParts = append(promptParts,
		"",
		"Instruções:",
		"1. Use um tom informal-profissional, pessoal e acolhedor",
		"2. Destaque a empresa como parceira estratégica",
		"3. Crie conexão com o leitor",
		"4. Use entre 3 a 5 frases",
		"5. Seja conciso e impactante",
		"6. Ideal para primeiro contato com leads",
		"7. Não use aspas ou formatação especial",
		"8. Retorne apenas o texto de apresentação, sem explicações adicionais",
	)

	fullPrompt := strings.Join(promptParts, "\n")

	geminiPayload := g.M(
		"contents", []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"text": fullPrompt,
					},
				},
			},
		},
		"generationConfig", g.M(
			"temperature",     0.7,
			"topK",            40,
			"topP",            0.95,
			"maxOutputTokens", 200,
			"stopSequences",   []string{},
		),
	)

	url := g.F("%s?key=%s", GeminiBaseURL, GeminiAPIKey)
	resp, respBytes, err := net.ClientDo("POST", url, strings.NewReader(g.Marshal(geminiPayload)), GeminiHeaders)
	if err != nil {
		g.LogError(g.Error(err, "failed to call Gemini API for lead introduction"))
		return ErrJSON(500, g.Error("could not connect to AI service"), "service temporarily unavailable")
	}

	if resp.StatusCode >= 400 {
		g.LogError(g.Error("Gemini API returned error status: %d, response: %s", resp.StatusCode, string(respBytes)))
		return ErrJSON(500, g.Error("AI service error"), "could not generate lead introduction text")
	}

	respMap, err := g.UnmarshalMap(string(respBytes))
	if err != nil {
		g.LogError(g.Error(err, "failed to parse response from Gemini API"))
		return ErrJSON(500, g.Error("invalid response from AI service"))
	}

	candidates, ok := respMap["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		g.LogError(g.Error("no candidates in Gemini response"))
		return ErrJSON(500, g.Error("AI service did not generate any content"))
	}

	candidate, ok := candidates[0].(map[string]interface{})
	if !ok {
		g.LogError(g.Error("invalid candidate structure in Gemini response"))
		return ErrJSON(500, g.Error("AI service returned invalid response"))
	}

	content, ok := candidate["content"].(map[string]interface{})
	if !ok {
		g.LogError(g.Error("no content in Gemini candidate"))
		return ErrJSON(500, g.Error("AI service returned empty content"))
	}

	parts, ok := content["parts"].([]interface{})
	if !ok || len(parts) == 0 {
		g.LogError(g.Error("no parts in Gemini content"))
		return ErrJSON(500, g.Error("AI service returned no text parts"))
	}

	part, ok := parts[0].(map[string]interface{})
	if !ok {
		g.LogError(g.Error("invalid part structure in Gemini response"))
		return ErrJSON(500, g.Error("AI service returned invalid text structure"))
	}

	leadIntroductionText, ok := part["text"].(string)
	if !ok {
		g.LogError(g.Error("no text field in Gemini part"))
		return ErrJSON(500, g.Error("AI service returned no text content"))
	}

	leadIntroductionText = strings.TrimSpace(leadIntroductionText)
	if leadIntroductionText == "" {
		g.LogError(g.Error("empty text generated by Gemini API"))
		return ErrJSON(500, g.Error("AI service generated empty content"))
	}

	teamProps["lead_introduction_text"] = leadIntroductionText
	teamRecord.Set("properties", g.Marshal(teamProps))

	if err := req.Dao().SaveRecord(teamRecord); err != nil {
		g.LogError(g.Error(err, "failed to save lead introduction text to database"))
		return c.JSON(200, g.M(
			"lead_introduction_text", leadIntroductionText,
			"warning", "Text generated but not saved to database",
		))
	}

	return c.JSON(200, g.M(
		"lead_introduction_text", leadIntroductionText,
		"previous_text", originalLeadIntroductionText,
	))
}

// called from suaobra frontend
func MessengerQueueGet(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(400, req.Error)
	}

	req.Payload["user"] = g.M("id", req.TeamID())

	payload := g.Marshal(req.Payload)
	url := g.F("%s/message/queue/get", MessengerBaseURL)
	resp, respBytes, err := net.ClientDo("POST", url, strings.NewReader(payload), MessengerHeaders)
	if err != nil {
		return ErrJSON(400, err)
	} else if resp.StatusCode >= 400 {
		return ErrJSON(resp.StatusCode, g.Error("could not submit to messenger"))
	}

	return c.String(200, string(respBytes))
}

// called from suaobra frontend
func MessengerGetOrCreateUser(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(400, req.Error)
	}

	req.Payload["user"] = g.M("id", req.TeamID())

	payload := g.Marshal(req.Payload)
	url := g.F("%s/user/get", MessengerBaseURL)
	resp, respBytes, err := net.ClientDo("POST", url, strings.NewReader(payload), MessengerHeaders)
	if err != nil {
		if strings.Contains(err.Error(), "could not find user") {
			url = g.F("%s/user/set", MessengerBaseURL)
			req.Payload["user"] = g.M(
				"id", req.TeamID(),
				"app_id", "suaobra",
				"token", g.RandString(g.AlphaRunes, 32),
			)

			resp, respBytes, err = net.ClientDo("POST", url, strings.NewReader(g.Marshal(req.Payload)), MessengerHeaders)
			if err != nil {
				return ErrJSON(400, err)
			} else if resp.StatusCode >= 400 {
				return ErrJSON(resp.StatusCode, g.Error("could not submit to messenger to create user"))
			}
		} else {
			return ErrJSON(400, err)
		}
	} else if resp.StatusCode >= 400 {
		return ErrJSON(resp.StatusCode, g.Error("could not submit to messenger to get user"))
	}

	return c.String(200, string(respBytes))
}

// allows to query a number, to get status, whether a conversation was found or not
func MessengerExisting(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(400, req.Error)
	}

	if req.Payload["names"] == nil {
		return ErrJSON(400, g.Error("missing names"))
	}

	names := cast.ToStringSlice(req.Payload["names"])
	for i, name := range names {
		names[i] = g.F("'%s'", name) // for sql injection
	}

	namePhoneSQL, _ := templates.ReadFile("templates/messenger/name_phones.sql")
	recordsRecipients, err := req.SqlQueryRecords(g.Rm(string(namePhoneSQL), g.M("names", strings.Join(names, ","))))
	if err != nil {
		return ErrJSON(502, err, "error querying phone data")
	}

	numberNameMap := map[string]string{}
	recipients := []string{}
	for _, rec := range recordsRecipients {
		name := cast.ToString(rec["nome"])
		phone := formatPhoneNumber(cast.ToString(rec["telephone"]))
		if phone == "" {
			g.Warn("could not format phone number", rec)
			continue
		}

		numberNameMap[phone] = name
		recipients = append(recipients, phone)
	}

	req.Payload["user"] = g.M("id", req.TeamID())
	req.Payload["recipients"] = recipients
	delete(req.Payload, "names")

	payload := g.Marshal(req.Payload)

	nameContacted := map[string]bool{}

	req.Context.Wg.Read.Add()
	go func() bool {
		defer req.Context.Wg.Read.Done()

		url := g.F("%s/chat/existing", MessengerBaseURL)
		resp, respBytes, err := net.ClientDo("POST", url, strings.NewReader(payload), MessengerHeaders)
		if err != nil {
			return req.Context.CaptureErr(err)
		} else if resp.StatusCode >= 400 {
			return req.Context.CaptureErr(g.Error("could not submit to messenger"))
		}

		respMap, _ := g.UnmarshalMap(string(respBytes))

		var numberChatMap map[string]int
		if err = g.JSONConvert(respMap["recipient_chat_count_map"], &numberChatMap); err != nil {
			return req.Context.CaptureErr(g.Error(err, "could not cast recipient_chat_count_map"))
		}

		for number, count := range numberChatMap {
			if name, ok := numberNameMap[number]; ok && count > 0 {
				nameContacted[name] = true
			}
		}

		return false
	}()

	req.Context.Wg.Read.Add()
	go func() bool {
		defer req.Context.Wg.Read.Done()

		url := g.F("%s/message/existing", MessengerBaseURL)
		resp, respBytes, err := net.ClientDo("POST", url, strings.NewReader(payload), MessengerHeaders)
		if err != nil {
			return req.Context.CaptureErr(err)
		} else if resp.StatusCode >= 400 {
			return req.Context.CaptureErr(g.Error("could not submit to messenger"))
		}

		respMap, _ := g.UnmarshalMap(string(respBytes))

		var numberMessageMap map[string]int
		if err = g.JSONConvert(respMap["recipient_message_count_map"], &numberMessageMap); err != nil {
			return req.Context.CaptureErr(g.Error(err, "could not cast recipient_message_count_map"))
		}

		for number, count := range numberMessageMap {
			if name, ok := numberNameMap[number]; ok && count > 0 {
				nameContacted[name] = true
			}
		}

		return false
	}()

	req.Context.Wg.Read.Wait()

	if err := req.Context.Err(); err != nil {
		return ErrJSON(400, g.Error(err, "error getting existing details"))
	}

	return c.JSON(200, g.M("name_contacted_map", nameContacted))
}

func formatPhoneNumber(phone string) string {
	if !g.In(len(phone), 10, 11) {
		return ""
	}
	phone = "55" + phone
	return phone
}

func shouldIgnoreWhatsAppWebhook(data map[string]any) bool {
	if len(data) == 0 {
		return true
	}

	if cast.ToBool(data["from_me"]) ||
		cast.ToBool(data["fromMe"]) ||
		cast.ToBool(data["is_from_me"]) ||
		cast.ToBool(data["isFromMe"]) ||
		cast.ToBool(data["self"]) {
		return true
	}

	if cast.ToBool(data["is_group"]) ||
		cast.ToBool(data["isGroup"]) ||
		cast.ToBool(data["from_group"]) {
		return true
	}

	rawFrom := firstNonEmpty(
		cast.ToString(data["from"]),
		cast.ToString(data["chat_id"]),
		cast.ToString(cast.ToStringMap(data["sender"])["jid"]),
		cast.ToString(cast.ToStringMap(data["sender"])["id"]),
	)
	rawFrom = strings.TrimSpace(rawFrom)

	if strings.Contains(rawFrom, "@g.us") {
		return true
	}

	return false
}

func resolveWhatsAppTeamID(req Request, data map[string]any) string {
	meta := cast.ToStringMap(data["meta"])
	user := cast.ToStringMap(data["user"])

	teamID := strings.TrimSpace(firstNonEmpty(
		cast.ToString(data["team_id"]),
		cast.ToString(meta["team_id"]),
		cast.ToString(user["team_id"]),
	))
	if teamID != "" {
		g.Info("MessengerSync[WA]: team_id resolvido direto=%s", teamID)
		return teamID
	}

	userID := strings.TrimSpace(firstNonEmpty(
		cast.ToString(data["user_id"]),
		cast.ToString(meta["user_id"]),
		cast.ToString(user["id"]),
	))
	if strings.HasPrefix(userID, "team_") {
		g.Info("MessengerSync[WA]: team_id resolvido via user_id=%s", userID)
		return userID
	}

	tokenCandidate := strings.TrimSpace(firstNonEmpty(
		cast.ToString(data["token"]),
		cast.ToString(data["user_token"]),
		cast.ToString(meta["token"]),
	))
	if tokenCandidate != "" {
		g.Info("MessengerSync[WA]: tentando resolver team por token=%s", maskMessengerToken(tokenCandidate))
		if teamID := findTeamIDByWhatsAppToken(req, tokenCandidate); teamID != "" {
			g.Info("MessengerSync[WA]: team_id resolvido via token=%s", teamID)
			return teamID
		}
		g.Warn("MessengerSync[WA]: token não resolveu team_id token=%s", maskMessengerToken(tokenCandidate))
	}

	instanciaCandidate := strings.TrimSpace(firstNonEmpty(
		cast.ToString(data["instancia_id"]),
		cast.ToString(data["instance_id"]),
		cast.ToString(data["user_id"]),
		cast.ToString(meta["instancia_id"]),
		cast.ToString(meta["instance_id"]),
	))
	if instanciaCandidate != "" {
		g.Info("MessengerSync[WA]: tentando resolver team por instancia_id=%s", instanciaCandidate)
		if teamID := findTeamIDByWhatsAppInstance(req, instanciaCandidate); teamID != "" {
			g.Info("MessengerSync[WA]: team_id resolvido via instancia_id=%s => %s", instanciaCandidate, teamID)
			return teamID
		}
		g.Warn("MessengerSync[WA]: instancia_id não resolveu team_id instancia_id=%s", instanciaCandidate)
	}

	g.Warn("MessengerSync[WA]: falha total ao resolver team_id")
	return ""
}

func findTeamIDByWhatsAppToken(req Request, token string) string {
	waRepo := repositories.NewWhatsAppRepo(req.Dao())
	conRepo := repositories.NewConexaoRepo(req.Dao())

	wa, err := waRepo.FindByToken(strings.TrimSpace(token))
	if err != nil || wa == nil || wa.Id == "" {
		return ""
	}

	conIDs := wa.GetStringSlice("conexoes")
	if len(conIDs) == 0 {
		single := strings.TrimSpace(wa.GetString("conexoes"))
		if single != "" {
			conIDs = []string{single}
		}
	}

	for _, conID := range conIDs {
		conID = strings.TrimSpace(conID)
		if conID == "" {
			continue
		}

		con, err := conRepo.FindByID(conID)
		if err != nil || con == nil {
			continue
		}

		teamID := strings.TrimSpace(con.GetString("team_id"))
		if teamID != "" {
			return teamID
		}
	}

	return ""
}

func findTeamIDByWhatsAppInstance(req Request, instanciaID string) string {
	waRepo := repositories.NewWhatsAppRepo(req.Dao())
	conRepo := repositories.NewConexaoRepo(req.Dao())

	wa, err := waRepo.FindByInstanciaID(strings.TrimSpace(instanciaID))
	if err != nil || wa == nil || wa.Id == "" {
		return ""
	}

	conIDs := wa.GetStringSlice("conexoes")
	if len(conIDs) == 0 {
		single := strings.TrimSpace(wa.GetString("conexoes"))
		if single != "" {
			conIDs = []string{single}
		}
	}

	for _, conID := range conIDs {
		conID = strings.TrimSpace(conID)
		if conID == "" {
			continue
		}

		con, err := conRepo.FindByID(conID)
		if err != nil || con == nil {
			continue
		}

		teamID := strings.TrimSpace(con.GetString("team_id"))
		if teamID != "" {
			return teamID
		}
	}

	return ""
}

func extractWhatsAppPhone(data map[string]any) string {
	sender := cast.ToStringMap(data["sender"])
	from := cast.ToStringMap(data["from"])
	contact := cast.ToStringMap(data["contact"])

	raw := firstNonEmpty(
		cast.ToString(data["phone"]),
		cast.ToString(data["telefone"]),
		cast.ToString(data["from"]),
		cast.ToString(data["author"]),
		cast.ToString(data["chat_id"]),
		cast.ToString(sender["phone"]),
		cast.ToString(sender["id"]),
		cast.ToString(sender["jid"]),
		cast.ToString(from["phone"]),
		cast.ToString(from["id"]),
		cast.ToString(from["jid"]),
		cast.ToString(contact["phone"]),
		cast.ToString(contact["wa_id"]),
		cast.ToString(contact["id"]),
	)

	return normalizeWebhookPhone(raw)
}

func extractWhatsAppText(data map[string]any) string {
	message := cast.ToStringMap(data["message"])
	textObj := cast.ToStringMap(message["text"])
	extendedTextObj := cast.ToStringMap(message["extendedTextMessage"])

	text := firstNonEmpty(
		cast.ToString(data["body"]),
		cast.ToString(data["text"]),
		cast.ToString(data["message"]),
		cast.ToString(data["caption"]),
		cast.ToString(message["body"]),
		cast.ToString(message["text"]),
		cast.ToString(message["conversation"]),
		cast.ToString(message["caption"]),
		cast.ToString(textObj["body"]),
		cast.ToString(textObj["text"]),
		cast.ToString(extendedTextObj["text"]),
	)

	return strings.TrimSpace(text)
}

func extractWhatsAppContactName(data map[string]any) string {
	sender := cast.ToStringMap(data["sender"])
	contact := cast.ToStringMap(data["contact"])

	return strings.TrimSpace(firstNonEmpty(
		cast.ToString(data["name"]),
		cast.ToString(data["push_name"]),
		cast.ToString(data["pushName"]),
		cast.ToString(data["notify_name"]),
		cast.ToString(sender["name"]),
		cast.ToString(sender["push_name"]),
		cast.ToString(sender["pushName"]),
		cast.ToString(contact["name"]),
		cast.ToString(contact["push_name"]),
	))
}

func normalizeWebhookPhone(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	if idx := strings.Index(raw, "@"); idx > 0 {
		raw = raw[:idx]
	}

	var digits strings.Builder
	for _, c := range raw {
		if c >= '0' && c <= '9' {
			digits.WriteRune(c)
		}
	}

	phone := digits.String()
	if phone == "" {
		return ""
	}

	if len(phone) == 10 || len(phone) == 11 {
		phone = "55" + phone
	}

	return phone
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return ""
}

func logMessengerRequestMeta(c echo.Context, prefix string) {
	req := c.Request()
	g.Info(
		"%s method=%s path=%s remote_ip=%s user_agent=%s content_type=%s query=%s",
		prefix,
		req.Method,
		req.URL.Path,
		c.RealIP(),
		req.UserAgent(),
		req.Header.Get(echo.HeaderContentType),
		req.URL.RawQuery,
	)
}

func truncateMessengerLog(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "...(truncated)"
}

func maskMessengerPhone(phone string) string {
	phone = normalizeWebhookPhone(phone)
	if len(phone) <= 4 {
		return phone
	}
	return strings.Repeat("*", len(phone)-4) + phone[len(phone)-4:]
}

func maskMessengerToken(token string) string {
	token = strings.TrimSpace(token)
	if len(token) <= 6 {
		return token
	}
	return token[:3] + strings.Repeat("*", len(token)-6) + token[len(token)-3:]
}
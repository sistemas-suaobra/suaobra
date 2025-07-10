package server

import (
	"os"
	"strings"

	"github.com/flarco/g"
	"github.com/flarco/g/net"
	"github.com/labstack/echo/v5"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

var (
	MessengerAuthorization = "App-Token " + os.Getenv("MESSENGER_TOKEN")
	MessengerBaseURL       = os.Getenv("MESSENGER_BASE_URL")
	MessengerHeaders       = map[string]string{echo.HeaderAuthorization: MessengerAuthorization, echo.HeaderContentType: echo.MIMEApplicationJSON}
)

// called from messenger
func MessengerSync(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(400, req.Error)
	}

	type PayloadType string

	var (
		PayloadTypeMessageQueue    PayloadType = "message_queue"
		PayloadTypeWhatsAppMessage PayloadType = "whats_app_message"
	)

	payloadType := PayloadType(req.Payload.String("type"))
	data := cast.ToStringMap(req.Payload["data"])

	switch payloadType {
	case PayloadTypeMessageQueue:
		status := cast.ToString(data["status"])
		teamID := cast.ToString(data["user_id"]) // user_id corresponds to team_id here

		// need meta map
		metaMap := cast.ToStringMap(data["meta"])
		if len(metaMap) == 0 {
			return ErrJSON(404, g.Error("missing meta key"))
		}

		leadID := cast.ToString(metaMap["lead_id"])
		if leadID == "" {
			return ErrJSON(404, g.Error("missing meta.lead_id key"))
		}

		isOwner := cast.ToString(metaMap["contact_role"]) == "owner"
		contactedAtCol := lo.Ternary(isOwner, "owner_contacted_at", "professional_contacted_at")

		messengerUpdateLead, _ := templates.ReadFile("templates/messenger/update_lead.sql")
		_, err := req.SqlQuery(g.Rm(string(messengerUpdateLead), g.M("status", status, "team_id", teamID, "lead_id", leadID, "contacted_at_col", contactedAtCol)))
		if err != nil {
			return ErrJSON(502, err, "error updating messenger contacted_at")
		}

	case PayloadTypeWhatsAppMessage:
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

		// set as contact pending in sqlite
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
	req.Context.Wg.Read.Wait() // getting too many requests

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
func MessengerQueueGet(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(400, req.Error)
	}

	// set user
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

	// set user
	req.Payload["user"] = g.M("id", req.TeamID())

	payload := g.Marshal(req.Payload)
	url := g.F("%s/user/get", MessengerBaseURL)
	resp, respBytes, err := net.ClientDo("POST", url, strings.NewReader(payload), MessengerHeaders)
	if err != nil {
		if strings.Contains(err.Error(), "could not find user") {
			// create the user and token
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

// alows to query a number, to get status, whether a conversation was found or not
func MessengerExisting(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(400, req.Error)
	}

	// get the numbers for the names
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

		// add to map for reverse mapping
		numberNameMap[phone] = name
		recipients = append(recipients, phone)
	}

	// set payload
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
	// format phone
	// Landline: +55 (Country Code) - XX (DDD) - XXXX-XXXX
	// Mobile: +55 (Country Code) - XX (DDD) - 9XXXX-XXXX
	// can either by 10 or 11 digits (including DDD, without country code)
	if !g.In(len(phone), 10, 11) {
		return ""
	}
	// add country code
	phone = "55" + phone
	return phone
}

package server

import (
	"encoding/json"
	"strings"

	"github.com/flarco/g"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/models"
	"github.com/spf13/cast"
	"github.com/suaobra/suaobra-app/store"
)

var feedbackTypes = map[string]string{
	"bug":      "bug",
	"melhoria": "melhoria",
	"contato":  "contato",
}

func feedbackUserName(auth *models.Record) string {
	if auth == nil {
		return ""
	}
	props, err := propertiesMap(auth.Get("properties"))
	if err == nil {
		if name := strings.TrimSpace(cast.ToString(props["name"])); name != "" {
			return name
		}
	}
	return strings.TrimSpace(auth.Email())
}

func normalizeFeedbackAttachments(raw any) []map[string]string {
	out := make([]map[string]string, 0)
	appendEntry := func(name, kind, key, url, mime, size string) {
		name = strings.TrimSpace(name)
		if name == "" {
			return
		}
		kind = strings.TrimSpace(kind)
		if kind == "" {
			kind = "arquivo"
		}
		item := map[string]string{"name": name, "kind": kind}
		if key = strings.TrimSpace(key); key != "" {
			item["key"] = key
		}
		if url = strings.TrimSpace(url); url != "" {
			item["url"] = url
		}
		if mime = strings.TrimSpace(mime); mime != "" {
			item["mime"] = mime
		}
		if size = strings.TrimSpace(size); size != "" {
			item["size"] = size
		}
		out = append(out, item)
	}

	switch value := raw.(type) {
	case []any:
		for _, item := range value {
			entry, ok := item.(map[string]any)
			if !ok {
				continue
			}
			appendEntry(
				cast.ToString(entry["name"]),
				cast.ToString(entry["kind"]),
				cast.ToString(entry["key"]),
				cast.ToString(entry["url"]),
				cast.ToString(entry["mime"]),
				cast.ToString(entry["size"]),
			)
		}
	case []map[string]string:
		for _, entry := range value {
			appendEntry(entry["name"], entry["kind"], entry["key"], entry["url"], entry["mime"], entry["size"])
		}
	}
	if len(out) > 20 {
		out = out[:20]
	}
	return out
}

// CreateFeedback registra bug, melhoria ou contato enviado pelo usuário.
func CreateFeedback(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(401, req.Error, "error creating request")
	}
	auth := req.AuthRecord()
	if auth == nil {
		return ErrJSON(401, g.Error("unauthorized"))
	}
	if err := req.ValidatePayload("type", "subject", "message"); err != nil {
		return ErrJSON(400, err, "error validating payload")
	}

	kind, ok := feedbackTypes[strings.ToLower(strings.TrimSpace(req.Payload.String("type")))]
	if !ok {
		return ErrJSON(400, g.Error("type inválido"))
	}

	subject := strings.TrimSpace(req.Payload.String("subject"))
	message := strings.TrimSpace(req.Payload.String("message"))
	if subject == "" || message == "" {
		return ErrJSON(400, g.Error("assunto e mensagem são obrigatórios"))
	}
	if len(subject) > 200 {
		subject = subject[:200]
	}
	if len(message) > 8000 {
		message = message[:8000]
	}

	attachments := normalizeFeedbackAttachments(req.Payload["attachments"])
	encodedAttachments, err := json.Marshal(attachments)
	if err != nil {
		return ErrJSON(500, err, "error encoding attachments")
	}

	collection, err := req.Dao().FindCollectionByNameOrId("feedback")
	if err != nil {
		return ErrJSON(500, err, "coleção feedback não encontrada")
	}

	record := models.NewRecord(collection)
	record.SetId(store.ID(collection.Name, ""))
	record.Set("type", kind)
	record.Set("subject", subject)
	record.Set("message", message)
	record.Set("user_id", auth.Id)
	record.Set("team_id", cast.ToString(auth.Get("team_id")))
	record.Set("user_email", auth.Email())
	record.Set("user_name", feedbackUserName(auth))
	record.Set("attachments", string(encodedAttachments))
	record.Set("status", "novo")

	if err := req.Dao().SaveRecord(record); err != nil {
		return ErrJSON(500, err, "erro ao salvar feedback")
	}

	return c.JSON(201, g.M(
		"id", record.Id,
		"type", kind,
		"status", "novo",
	))
}

// PlatformFeedback lista bugs/melhorias/contatos para o backoffice.
func PlatformFeedback(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(401, req.Error, "error creating request")
	}
	if err := requirePlatformAdmin(&req); err != nil {
		return ErrJSON(403, err)
	}

	kindFilter := strings.ToLower(strings.TrimSpace(c.QueryParam("type")))
	sql := `
	select
		id,
		type,
		subject,
		message,
		user_id,
		team_id,
		user_email,
		user_name,
		attachments,
		status,
		created,
		updated
	from main.feedback
	`
	if kindFilter != "" {
		if _, ok := feedbackTypes[kindFilter]; !ok {
			return ErrJSON(400, g.Error("type inválido"))
		}
		sql += g.R(` where type = '{type}'`, "type", kindFilter)
	}
	sql += ` order by created desc limit 200`

	records, err := req.SqlQueryRecords(sql)
	if err != nil {
		return ErrJSON(500, err, "error listing feedback")
	}

	for _, record := range records {
		record["attachments"] = parseFeedbackAttachmentsField(record["attachments"])
	}

	return c.JSON(200, records)
}

func parseFeedbackAttachmentsField(raw any) []map[string]string {
	switch value := raw.(type) {
	case string:
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || trimmed == "null" {
			return []map[string]string{}
		}
		var parsed []map[string]string
		if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
			return normalizeFeedbackAttachments(toAnyList(parsed))
		}
		var asAny []any
		if err := json.Unmarshal([]byte(trimmed), &asAny); err == nil {
			return normalizeFeedbackAttachments(asAny)
		}
	case []byte:
		return parseFeedbackAttachmentsField(string(value))
	case []any:
		return normalizeFeedbackAttachments(value)
	}
	return []map[string]string{}
}

func toAnyList(items []map[string]string) []any {
	out := make([]any, 0, len(items))
	for _, item := range items {
		entry := map[string]any{}
		for key, value := range item {
			entry[key] = value
		}
		out = append(out, entry)
	}
	return out
}

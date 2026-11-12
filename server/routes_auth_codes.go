package server

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/flarco/g"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
	"github.com/spf13/cast"
	"github.com/suaobra/suaobra-app/store"
)

var authCodePurposes = map[string]string{
	"verify": "verify",
	"reset":  "reset",
}

func hashAuthCode(code string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(code)))
	return hex.EncodeToString(sum[:])
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func appFromEcho(c echo.Context) *pocketbase.PocketBase {
	app, _ := c.Get("app").(*pocketbase.PocketBase)
	return app
}

func bindJSONMap(c echo.Context) (map[string]any, error) {
	payload := map[string]any{}
	if err := c.Bind(&payload); err != nil {
		return nil, err
	}
	return payload, nil
}

// StoreAuthCode salva o hash de um código enviado pelo BFF (SendKit). Público para o BFF.
func StoreAuthCode(c echo.Context) error {
	app := appFromEcho(c)
	if app == nil {
		return ErrJSON(500, g.Error("app unavailable"))
	}
	payload, err := bindJSONMap(c)
	if err != nil {
		return ErrJSON(400, err, "payload inválido")
	}

	purpose, ok := authCodePurposes[strings.ToLower(strings.TrimSpace(cast.ToString(payload["purpose"])))]
	if !ok {
		return ErrJSON(400, g.Error("purpose inválido"))
	}

	email := normalizeEmail(cast.ToString(payload["email"]))
	code := strings.TrimSpace(cast.ToString(payload["code"]))
	if email == "" || len(code) < 4 {
		return ErrJSON(400, g.Error("e-mail e código são obrigatórios"))
	}

	user, err := app.Dao().FindAuthRecordByEmail("user", email)
	if err != nil || user == nil {
		return c.JSON(200, g.M("ok", true, "found", false))
	}

	collection, err := app.Dao().FindCollectionByNameOrId("auth_codes")
	if err != nil {
		return ErrJSON(500, err, "coleção auth_codes não encontrada")
	}

	_, _ = app.Dao().DB().NewQuery(`
		update main.auth_codes set used = 1
		where email = {:email} and purpose = {:purpose} and used = 0
	`).Bind(dbx.Params{"email": email, "purpose": purpose}).Execute()

	record := models.NewRecord(collection)
	record.SetId(store.ID(collection.Name, ""))
	record.Set("email", email)
	record.Set("user_id", user.Id)
	record.Set("purpose", purpose)
	record.Set("code_hash", hashAuthCode(code))
	record.Set("expires_at", time.Now().UTC().Add(15*time.Minute).Format(time.RFC3339))
	record.Set("used", false)

	if err := app.Dao().SaveRecord(record); err != nil {
		return ErrJSON(500, err, "erro ao salvar código")
	}

	return c.JSON(200, g.M(
		"ok", true,
		"found", true,
		"user_id", user.Id,
		"name", feedbackUserName(user),
		"email", email,
	))
}

// ConfirmAuthCode valida o código e aplica verify ou reset de senha. Público para o BFF.
func ConfirmAuthCode(c echo.Context) error {
	app := appFromEcho(c)
	if app == nil {
		return ErrJSON(500, g.Error("app unavailable"))
	}
	payload, err := bindJSONMap(c)
	if err != nil {
		return ErrJSON(400, err, "payload inválido")
	}

	purpose, ok := authCodePurposes[strings.ToLower(strings.TrimSpace(cast.ToString(payload["purpose"])))]
	if !ok {
		return ErrJSON(400, g.Error("purpose inválido"))
	}

	email := normalizeEmail(cast.ToString(payload["email"]))
	code := strings.TrimSpace(cast.ToString(payload["code"]))
	password := cast.ToString(payload["password"])
	if email == "" || code == "" {
		return ErrJSON(400, g.Error("e-mail e código são obrigatórios"))
	}
	if purpose == "reset" && strings.TrimSpace(password) == "" {
		return ErrJSON(400, g.Error("informe a nova senha"))
	}
	if purpose == "reset" && len(strings.TrimSpace(password)) < 8 {
		return ErrJSON(400, g.Error("a senha deve ter no mínimo 8 caracteres"))
	}

	user, err := app.Dao().FindAuthRecordByEmail("user", email)
	if err != nil || user == nil {
		return ErrJSON(400, g.Error("código inválido ou expirado"))
	}

	type codeRow struct {
		ID        string `db:"id"`
		CodeHash  string `db:"code_hash"`
		ExpiresAt string `db:"expires_at"`
	}
	rows := []codeRow{}
	err = app.Dao().DB().NewQuery(`
		select id, code_hash, expires_at
		from main.auth_codes
		where email = {:email} and purpose = {:purpose} and used = 0
		order by created desc
		limit 5
	`).Bind(dbx.Params{"email": email, "purpose": purpose}).All(&rows)
	if err != nil {
		return ErrJSON(500, err, "erro ao buscar código")
	}

	codeHash := hashAuthCode(code)
	var matchedID string
	now := time.Now().UTC()
	for _, row := range rows {
		expiresAt, parseErr := time.Parse(time.RFC3339, row.ExpiresAt)
		if parseErr != nil || now.After(expiresAt) {
			continue
		}
		if row.CodeHash == codeHash {
			matchedID = row.ID
			break
		}
	}
	if matchedID == "" {
		return ErrJSON(400, g.Error("código inválido ou expirado"))
	}

	if purpose == "verify" {
		user.SetVerified(true)
	} else {
		user.SetPassword(strings.TrimSpace(password))
		user.SetVerified(true)
	}
	if err := app.Dao().SaveRecord(user); err != nil {
		return ErrJSON(500, err, "erro ao atualizar usuário")
	}

	_, _ = app.Dao().DB().NewQuery(`
		update main.auth_codes set used = 1 where id = {:id}
	`).Bind(dbx.Params{"id": matchedID}).Execute()

	return c.JSON(200, g.M("ok", true, "purpose", purpose))
}

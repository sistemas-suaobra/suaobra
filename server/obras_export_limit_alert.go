package server

import (
	"net/mail"
	"strings"
	"time"

	"github.com/flarco/g"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/tools/mailer"
	"github.com/suaobra/suaobra-app/store"
)

const obrasExportLimitAlertTo = "lucas@suaobra.com.br"

func brtDayKey(now time.Time) string {
	return brtDayStart(now).Format("2006-01-02")
}

// claimExportLimitAlert marca o alerta do dia para a equipe.
// Retorna true se este processo "ganhou" o direito de enviar (primeira vez no dia).
func claimExportLimitAlert(teamID string, now time.Time) (bool, error) {
	if teamID == "" {
		return false, nil
	}

	dayKey := brtDayKey(now)
	_, err := store.MainDB.Exec(
		`INSERT INTO obras_export_limit_alert (id, team_id, day_key, created) VALUES (?, ?, ?, ?)`,
		store.ID("obras_export_limit_alert", ""),
		teamID,
		dayKey,
		now.UTC().Format(time.RFC3339),
	)
	if err != nil {
		msg := strings.ToLower(err.Error())
		if strings.Contains(msg, "unique") || strings.Contains(msg, "constraint") {
			return false, nil
		}
		if strings.Contains(msg, "no such table") {
			return false, nil
		}
		return false, g.Error(err, "erro ao registrar alerta de limite de exportação")
	}

	return true, nil
}

// maybeAlertExportDailyLimit avisa lucas@suaobra.com.br quando a equipe atinge 500/dia.
// Inclui e-mail do usuário e quantas obras ele exportou nesta operação.
func maybeAlertExportDailyLimit(
	app *pocketbase.PocketBase,
	teamID, teamName, exporterEmail string,
	obrasNesteExport int,
	now time.Time,
) {
	if app == nil || teamID == "" {
		return
	}

	used, err := dailyObrasExportCount(teamID, now)
	if err != nil {
		g.Warn("alerta exportação: falha ao contar uso diário: %v", err)
		return
	}
	if used < maxObrasExportLeadsPerDay {
		return
	}

	claimed, err := claimExportLimitAlert(teamID, now)
	if err != nil {
		g.Warn("alerta exportação: falha ao reservar alerta: %v", err)
		return
	}
	if !claimed {
		return
	}

	if teamName == "" {
		teamName = teamID
	}
	if exporterEmail == "" {
		exporterEmail = "(e-mail não disponível)"
	}

	dayLabel := brtDayStart(now).Format("02/01/2006")
	subject := g.F("SuaObra - Limite 500/dia atingido — %s", exporterEmail)
	html := g.Rm(`
		<p>Olá,</p>
		<p>Uma equipe atingiu o limite diário de exportação do Obras+ (500 leads/dia).</p>
		<p><strong>Usuário:</strong> <a href="mailto:{email}">{email}</a></p>
		<p><strong>Obras exportadas nesta operação:</strong> {obras}</p>
		<p><strong>Total exportado pela equipe hoje:</strong> {used}</p>
		<p><strong>Equipe:</strong> {team}</p>
		<p><strong>Data:</strong> {day}</p>
		<p><strong>Limite:</strong> {limit} leads por dia</p>
		<p>
			Obrigado,<br/>
			Equipe SuaObra
		</p>
	`, g.M(
		"email", exporterEmail,
		"obras", obrasNesteExport,
		"used", used,
		"team", teamName,
		"day", dayLabel,
		"limit", maxObrasExportLeadsPerDay,
	))

	settings := app.Settings()
	message := &mailer.Message{
		From: mail.Address{
			Address: settings.Meta.SenderAddress,
			Name:    settings.Meta.SenderName,
		},
		To: []mail.Address{
			{Address: obrasExportLimitAlertTo},
		},
		Subject: subject,
		HTML:    html,
	}

	if err := sendEmail(app, message); err != nil {
		g.Warn("alerta exportação: falha ao e-mail %s: %v", obrasExportLimitAlertTo, err)
		return
	}

	g.Info(
		"alerta exportação: e-mail enviado para %s (user=%s, obras=%d, team=%s, used=%d)",
		obrasExportLimitAlertTo, exporterEmail, obrasNesteExport, teamID, used,
	)
}

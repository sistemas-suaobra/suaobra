package server

import (
	"net/mail"
	"time"

	"github.com/flarco/g"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/tools/mailer"
	"github.com/spf13/cast"
	"github.com/suaobra/suaobra-app/store"
)

func NotifyReminders(app *pocketbase.PocketBase) {
	store.EnsureCoreReady()

	selectSQL, _ := templates.ReadFile("templates/crm/crm_reminders.sql")

	data, err := store.MainDB.Query(string(selectSQL))
	if g.LogError(err, "could not get list of crm_reminders") {
		return
	}

	records := data.RecordsCasted()
	g.Info("Processing %d reminder records", len(records))

	for i, rec := range records {
		// Rate limiting: delay between emails to respect provider limits (max 2/second)
		if i > 0 {
			time.Sleep(600 * time.Millisecond) // 0.6 seconds delay = ~1.5 emails/second
		}

		err = EmailReminder(app, rec)
		if g.LogError(err, "could not email reminder") {
			// Add exponential backoff on error
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		updateSQL := `
		update "main"."lead"
		set properties = json_patch(properties, '{"alerted": true}')
		where id = {:id}`
		updateSQL = store.BindSQL(updateSQL, g.M("id", rec["lead_id"]))
		_, err = store.MainDB.Exec(updateSQL)
		g.LogError(err, "could not set alerted for email reminder")
	}
}

// formatReminderAtBR formata alert_at (unix seconds da query) em horário de Brasília.
// alert_at em lead.properties é timestamp absoluto em ms (data personalizada no frontend).
func formatReminderAtBR(alertAtUnixSec any) string {
	sec := cast.ToInt64(alertAtUnixSec)
	if sec <= 0 {
		return ""
	}

	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.FixedZone("BRT", -3*3600)
	}

	return time.Unix(sec, 0).In(loc).Format("02/01/2006 15:04")
}

func EmailReminder(app *pocketbase.PocketBase, rec map[string]any) error {
	link := g.F("https://app.suaobra.com.br/venda-mais/?lead=%s", rec["lead_id"])
	title := g.F("%s - %s, %s", rec["owner"], rec["city"], rec["state"])
	scheduledAt := formatReminderAtBR(rec["alert_at"])

	g.Info("EmailReminder to '%s' for `%s` (Lead ID: %s, alert_at=%s)", rec["email"], title, rec["lead_id"], scheduledAt)

	scheduledLine := ""
	if scheduledAt != "" {
		scheduledLine = g.F("<p><strong>Data do lembrete</strong>: %s</p>", scheduledAt)
	}

	message := &mailer.Message{
		From: mail.Address{
			Address: app.Settings().Meta.SenderAddress,
			Name:    app.Settings().Meta.SenderName,
		},
		To: []mail.Address{
			{Address: cast.ToString(rec["email"])},
		},
		Subject: g.F("SuaObra - Retorno Agendado: %s", title),
		HTML: g.Rm(`
		<p>Olá,</p>
		<p>Chegou a hora do lembrete que você agendou para entrar em contato com o lead.</p>
		{scheduled_line}
		<p><strong>Proprietário</strong>: {owner}</p>
		<p><strong>Profissional</strong>: {professional}</p>
		<p><strong>Endereço</strong>: {address}</p>
		<p><strong>Link</strong>: <a href="{link}">Clique Aqui</a></p>

		<p>
			Obrigado,<br/>
			Equipe SuaObra
		</p>`,

			g.M(
				"scheduled_line", scheduledLine,
				"owner", rec["owner"],
				"professional", rec["professional"],
				"address", rec["address"],
				"link", link,
			),
		),
	}

	return sendEmail(app, message)
}

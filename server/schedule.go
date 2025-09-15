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
		_, err = store.MainDB.Query(updateSQL)
		g.LogError(err, "could not set alerted for email reminder")
	}
}

func EmailReminder(app *pocketbase.PocketBase, rec map[string]any) error {
	link := g.F("https://app.suaobra.com.br/venda-mais/?lead=%s", rec["lead_id"])
	title := g.F("%s - %s, %s", rec["owner"], rec["city"], rec["state"])

	g.Info("EmailReminder to '%s' for `%s` (Lead ID: %s)", rec["email"], title, rec["lead_id"])

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
		<p>Apenas para informar que você definiu um lembrete para entrar em contato com o lead.</p>
		<p><strong>Proprietário</strong>: {owner}</p>
		<p><strong>Profissional</strong>: {professional}</p>
		<p><strong>Enderco</strong>: {address}</p>
		<p><strong>Link</strong>: <a href="{link}">Clique Aqui</a></p>

		<p>
			Obrigado,<br/>
			Equipe SuaObra
		</p>`,

			g.M(
				"owner", rec["owner"],
				"professional", rec["professional"],
				"address", rec["address"],
				"link", link,
			),
		),
	}

	return sendEmail(app, message)
}

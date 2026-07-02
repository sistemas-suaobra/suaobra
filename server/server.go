package server

import (
	"os"

	"github.com/flarco/g"
	"github.com/flarco/g/net"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/tools/mailer"
)

func sendEmail(app *pocketbase.PocketBase, message *mailer.Message) (err error) {
	settings := app.Settings()
	if settings.Smtp.Host == "" || settings.Meta.SenderAddress == "" {
		return g.Error("SMTP do sistema não configurado no PocketBase")
	}

	if err := app.NewMailClient().Send(message); err != nil {
		NotifyDiscord(g.F("Could not send email:\n```\n%s\n```", g.Marshal(message)))
		return g.Error(err, "could not send email")
	}

	return nil
}

func NotifyDiscord(text string) {
	client := net.NewDiscordClient(os.Getenv("DISCORD_WEBHOOK_URL"))

	msg := net.DiscordMessage{
		Username: "Sua-Obra Backend: " + os.Getenv("ENV"),
		Content:  text,
	}

	g.LogError(client.Send(msg))
}

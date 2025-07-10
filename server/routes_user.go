package server

import (
	"net/mail"
	"sort"

	_ "embed"

	"github.com/flarco/g"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/mailer"
	"github.com/suaobra/suaobra-app/store"
)

func getUser(c echo.Context, userId string) (user store.User, err error) {

	sql, _ := templates.ReadFile("templates/user/user.sql")
	err = store.MainDbNewQuery(string(sql)).
		Bind(dbx.Params{"id": userId}).
		One(&user)
	if err != nil {
		err = g.Error(err, "error getting user")
	}

	// sorted cities
	sort.Strings(user.Team.Cities)

	return
}

func sendNotificationSignupToAdmins(app *pocketbase.PocketBase, rec *models.Record) error {
	message := &mailer.Message{
		From: mail.Address{
			Address: app.Settings().Meta.SenderAddress,
			Name:    app.Settings().Meta.SenderName,
		},
		// To:      []mail.Address{{Address: rec.Email()}},
		To: []mail.Address{
			{Address: "lucas@suaobra.com.br"},
			{Address: "flarco@gmail.com"},
		},
		Subject: g.F("SuaObra - Novo Usuário Cadastrado - %s", rec.Email()),
		HTML: g.F(`
		<p>Olá,</p>
		<p>Apenas para informar que um novo usuário se cadastrou.</p>
		<p>O e-mail do usuário: <a href="mailto:%s">%s</a></p>
		<p>
			Obrigado,<br/>
			Equipe SuaObra
		</p>
		`, rec.Email(), rec.Email()),
		// bcc, cc, attachments and custom headers are also supported...
	}

	return sendEmail(app, message)
}

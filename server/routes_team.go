package server

import (
	"net/mail"

	"github.com/flarco/g"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/mailer"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/spf13/cast"
	"github.com/suaobra/suaobra-app/store"
)

const teamInviteAppURL = "https://app.suaobra.com.br"

func sendTeamInviteEmail(app *pocketbase.PocketBase, email, password string, isNewUser bool) error {
	settings := app.Settings()
	message := &mailer.Message{
		From: mail.Address{
			Address: settings.Meta.SenderAddress,
			Name:    settings.Meta.SenderName,
		},
		To:      []mail.Address{{Address: email}},
		Subject: "Convite para Sua Obra",
	}

	if isNewUser {
		message.HTML = g.F(`
		<p>Olá,</p>
		<p>Você foi convidado para se juntar a uma equipe no Sua Obra.</p>
		<p>Email: <strong>%s</strong></p>
		<p>Senha temporária: <strong>%s</strong></p>
		<p>Por favor, acesse <a href="%s">%s</a> e altere sua senha após o primeiro login.</p>
		<p>
			Obrigado,<br/>
			Equipe SuaObra
		</p>
		`, email, password, teamInviteAppURL, teamInviteAppURL)
	} else {
		message.HTML = g.F(`
		<p>Olá,</p>
		<p>Você foi adicionado a uma equipe no Sua Obra.</p>
		<p>Acesse <a href="%s">%s</a> com seu email e senha para entrar.</p>
		<p>
			Obrigado,<br/>
			Equipe SuaObra
		</p>
		`, teamInviteAppURL, teamInviteAppURL)
	}

	return sendEmail(app, message)
}

// TeamMembers returns all team members for a team
func TeamMembers(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(401, req.Error, "error creating request")
	}

	teamID := req.TeamID()
	if teamID == "" {
		return ErrJSON(400, g.Error("team_id é necessário"))
	}

	// Get all users in the team
	sql := `
	SELECT
		u.id,
		u.email,
		u.created,
		u.updated,
		u.manager,
		u.properties
	FROM user u
	WHERE u.team_id = '{team_id}'
	`

	records, err := req.SqlQueryRecords(g.R(sql, "team_id", teamID))
	if err != nil {
		return ErrJSON(500, err, "error fetching team members")
	}

	return c.JSON(200, records)
}

// TeamInvite invites a new user to the team
func TeamInvite(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(401, req.Error, "error creating request")
	}

	// Validate parameters
	if err := req.ValidatePayload("team_id", "email"); err != nil {
		g.Debug("validate > " + err.Error())
		return ErrJSON(400, err, "error validating payload")
	}

	teamID := req.TeamID()
	email := req.Payload.String("email")

	// Check if email is valid
	_, err := mail.ParseAddress(email)
	if err != nil {
		g.Debug("ParseAddress > " + err.Error())
		return ErrJSON(400, err, "email inválido")
	}

	// Check if user already exists
	existingRecords, err := req.Dao().FindRecordsByExpr("user", dbx.HashExp{
		"email": email,
	})
	if err != nil {
		g.Debug("FindRecordsByExpr > " + err.Error())
		return ErrJSON(500, err, "erro ao verificar usuário existente")
	}

	var existingUser *models.Record
	if len(existingRecords) > 0 {
		existingUser = existingRecords[0]
	}

	if existingUser != nil {
		// If user exists, check if they're in their personal team or no team
		currentTeamID := cast.ToString(existingUser.Get("team_id"))
		existingUser.Set("manager", false)

		if currentTeamID == "" {
			// User has no team, we can set their team_id
			existingUser.Set("team_id", teamID)
			if err := req.Dao().SaveRecord(existingUser); err != nil {
				return ErrJSON(500, err, "erro ao adicionar usuário à equipe")
			}
			if err := sendTeamInviteEmail(req.App, email, "", false); err != nil {
				return ErrJSON(502, g.Error("Usuário adicionado à equipe, mas a notificação por e-mail falhou."))
			}
			return c.JSON(200, g.M("message", "Usuário adicionado à equipe"))
		}

		// If user already has a team, check if it's their personal team
		teamRecord, err := req.Dao().FindRecordById("team", currentTeamID)
		if err != nil {
			return ErrJSON(500, err, "erro ao verificar equipe atual do usuário")
		}

		teamName := cast.ToString(teamRecord.Get("name"))
		userEmail := cast.ToString(existingUser.Get("email"))

		// If the team name matches the user's email, it's their personal team
		// and we can move them to the new team
		if teamName == userEmail && currentTeamID != teamID {
			existingUser.Set("team_id", teamID)
			if err := req.Dao().SaveRecord(existingUser); err != nil {
				return ErrJSON(500, err, "erro ao adicionar usuário à equipe")
			}
			if err := sendTeamInviteEmail(req.App, email, "", false); err != nil {
				return ErrJSON(502, g.Error("Usuário adicionado à equipe, mas a notificação por e-mail falhou."))
			}
			return c.JSON(200, g.M("message", "Usuário adicionado à equipe"))
		}

		// User is already in a team (not their personal team) - can't invite them
		if currentTeamID == teamID {
			return ErrJSON(400, g.Error("usuário já está na equipe"))
		} else {
			return ErrJSON(400, g.Error("usuário já pertence a outra equipe"))
		}
	}

	// Create a new user and send invitation
	password := security.RandomString(10)

	// Create user userCollection
	userCollection, err := req.Dao().FindCollectionByNameOrId("user")
	if err != nil {
		return ErrJSON(500, err, "erro ao encontrar coleção de usuários")
	}

	// Create a new record
	newUser := models.NewRecord(userCollection)
	newUser.SetId(store.ID(newUser.TableName(), ""))
	newUser.Set("email", email)
	newUser.Set("username", g.F("user%d", g.RandInt(99999)+100000))
	newUser.SetPassword(password)
	newUser.Set("emailVisibility", true)
	newUser.Set("team_id", teamID)
	newUser.Set("manager", false)
	newUser.Set("properties", g.M())

	// Save the record
	if err := req.Dao().SaveRecord(newUser); err != nil {
		return ErrJSON(400, err, "erro ao criar usuário")
	}

	if err := sendTeamInviteEmail(req.App, email, password, true); err != nil {
		return ErrJSON(502, g.Error("Usuário criado, mas o e-mail de convite não foi enviado. Verifique a configuração SMTP."))
	}

	return c.JSON(200, g.M("message", "Convite enviado com sucesso"))
}

// TeamRemoveMember removes a user from the team
func TeamRemoveMember(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(401, req.Error, "error creating request")
	}

	// Validate parameters
	if err := req.ValidatePayload("team_id", "user_id"); err != nil {
		return ErrJSON(400, err, "error validating payload")
	}

	teamID := req.TeamID()
	userID := req.Payload.String("user_id")

	// Check if the user actually belongs to this team
	record, err := req.Dao().FindRecordById("user", userID)
	if err != nil {
		return ErrJSON(404, err, "usuário não encontrado")
	}

	recordTeamID := cast.ToString(record.Get("team_id"))
	if recordTeamID != teamID {
		return ErrJSON(403, g.Error("usuário não pertence a esta equipe"))
	}

	// Get user's email for finding their personal team
	userEmail := cast.ToString(record.Get("email"))

	// Find the user's personal team (with name matching their email)
	personalTeam, err := req.Dao().FindFirstRecordByData("team", "name", userEmail)
	if err != nil || personalTeam == nil {
		// If no personal team exists, create one
		teamCollection, err := req.Dao().FindCollectionByNameOrId("team")
		if err != nil {
			return ErrJSON(500, err, "erro ao encontrar coleção de equipe")
		}

		personalTeam = models.NewRecord(teamCollection)
		personalTeam.Set("name", userEmail)
		personalTeam.Set("owner_id", userID)
		personalTeam.Set("active", true)
		personalTeam.Set("blocked", false)
		personalTeam.Set("cities", "[]")
		personalTeam.Set("properties", g.Marshal(g.M()))
		personalTeam.Set("entitlements", g.Marshal(g.M()))

		if err := req.Dao().SaveRecord(personalTeam); err != nil {
			return ErrJSON(500, err, "erro ao criar equipe pessoal")
		}
	}

	// Update user to be on their personal team instead of setting to null
	record.Set("team_id", personalTeam.Id)
	if err := req.Dao().SaveRecord(record); err != nil {
		return ErrJSON(500, err, "erro ao remover usuário da equipe")
	}

	return c.JSON(200, g.M("message", "Usuário removido com sucesso"))
}

// TeamSetManager sets or unsets a user as a team manager
func TeamSetManager(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(401, req.Error, "error creating request")
	}

	// Validate parameters
	if err := req.ValidatePayload("team_id", "user_id", "is_manager"); err != nil {
		return ErrJSON(400, err, "error validating payload")
	}

	teamID := req.TeamID()
	userID := req.Payload.String("user_id")
	isManager := req.Payload.Bool("is_manager")

	// Check if the user actually belongs to this team
	record, err := req.Dao().FindRecordById("user", userID)
	if err != nil {
		return ErrJSON(404, err, "usuário não encontrado")
	}

	recordTeamID := cast.ToString(record.Get("team_id"))
	if recordTeamID != teamID {
		return ErrJSON(403, g.Error("usuário não pertence a esta equipe"))
	}

	// Update the record
	record.Set("manager", isManager)
	if err := req.Dao().SaveRecord(record); err != nil {
		return ErrJSON(500, err, "erro ao atualizar status de gerente")
	}

	message := "Usuário removido como gerente"
	if isManager {
		message = "Usuário definido como gerente"
	}
	return c.JSON(200, g.M("message", message))
}

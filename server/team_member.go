package server

import (
	"encoding/json"
	"strings"

	"github.com/flarco/g"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/tools/types"
	"github.com/spf13/cast"
)

const deactivatedAccountMessage = "Esta conta está desativada."

func propertiesMap(raw any) (map[string]any, error) {
	if raw == nil {
		return map[string]any{}, nil
	}

	props := map[string]any{}
	switch value := raw.(type) {
	case map[string]any:
		return value, nil
	case types.JsonRaw:
		if len(value) == 0 {
			return props, nil
		}
		if err := json.Unmarshal(value, &props); err != nil {
			return nil, err
		}
	case []byte:
		if len(value) == 0 {
			return props, nil
		}
		if err := json.Unmarshal(value, &props); err != nil {
			return nil, err
		}
	case string:
		if strings.TrimSpace(value) == "" || value == "null" {
			return props, nil
		}
		if err := json.Unmarshal([]byte(value), &props); err != nil {
			return nil, err
		}
	default:
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(encoded, &props); err != nil {
			return nil, err
		}
	}

	if props == nil {
		props = map[string]any{}
	}
	return props, nil
}

func activeFromProperties(raw any) bool {
	props, err := propertiesMap(raw)
	if err != nil {
		return true
	}
	active, ok := props["active"]
	if !ok || active == nil {
		return true
	}
	return cast.ToBool(active)
}

func UserIsActive(record *models.Record) bool {
	if record == nil {
		return true
	}
	return activeFromProperties(record.Get("properties"))
}

func userIsAdmin(record *models.Record) bool {
	if record == nil {
		return false
	}
	props, err := propertiesMap(record.Get("properties"))
	if err != nil {
		return false
	}
	return cast.ToBool(props["is_admin"])
}

func RejectDeactivatedRecord(record *models.Record) error {
	if record == nil || UserIsActive(record) {
		return nil
	}
	return apis.NewForbiddenError(deactivatedAccountMessage, nil)
}

func RejectDeactivatedAuth(c echo.Context) error {
	record, _ := c.Get(apis.ContextAuthRecordKey).(*models.Record)
	return RejectDeactivatedRecord(record)
}

// TeamSetActive desativa ou reativa alguém que continua na equipe.
func TeamSetActive(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(401, req.Error, "error creating request")
	}

	if err := req.ValidatePayload("user_id", "active"); err != nil {
		return ErrJSON(400, err, "error validating payload")
	}

	auth := req.AuthRecord()
	if auth == nil || (!auth.GetBool("manager") && !userIsAdmin(auth)) {
		return ErrJSON(403, g.Error("apenas gerentes podem desativar alguém da equipe"))
	}

	teamID := cast.ToString(auth.Get("team_id"))
	userID := req.Payload.String("user_id")
	active := req.Payload.Bool("active")

	if userID == auth.Id {
		return ErrJSON(400, g.Error("você não pode desativar a própria conta"))
	}

	record, err := req.Dao().FindRecordById("user", userID)
	if err != nil {
		return ErrJSON(404, err, "usuário não encontrado")
	}

	if cast.ToString(record.Get("team_id")) != teamID {
		return ErrJSON(403, g.Error("usuário não pertence a esta equipe"))
	}

	if userIsAdmin(record) && !userIsAdmin(auth) {
		return ErrJSON(403, g.Error("não é possível desativar um administrador"))
	}

	props, err := propertiesMap(record.Get("properties"))
	if err != nil {
		return ErrJSON(500, err, "erro ao ler os dados do usuário")
	}
	props["active"] = active
	record.Set("properties", props)
	if err := req.Dao().SaveRecord(record); err != nil {
		return ErrJSON(500, err, "erro ao atualizar o acesso do usuário")
	}

	message := "Usuário desativado"
	if active {
		message = "Usuário ativado"
	}
	return c.JSON(200, g.M("message", message))
}

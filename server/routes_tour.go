package server

import (
	"strings"

	"github.com/flarco/g"
	"github.com/labstack/echo/v5"
	"github.com/spf13/cast"
)

const tourCompletedKey = "tour_v2_completed"

func tourCompletedFromProps(props map[string]any) bool {
	return cast.ToBool(props[tourCompletedKey])
}

// TourStatus indica se o usuário já concluiu o tour de novidades.
func TourStatus(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(401, req.Error, "error creating request")
	}
	auth := req.AuthRecord()
	if auth == nil {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	props, err := propertiesMap(auth.Get("properties"))
	if err != nil {
		return ErrJSON(500, err, "erro ao ler propriedades")
	}

	return c.JSON(200, g.M(
		"completed", tourCompletedFromProps(props),
		"version", "v2",
	))
}

// TourComplete marca o tour de novidades como concluído (não volta a exibir).
func TourComplete(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(401, req.Error, "error creating request")
	}
	auth := req.AuthRecord()
	if auth == nil {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	user, err := req.Dao().FindRecordById("user", auth.Id)
	if err != nil || user == nil {
		return ErrJSON(404, g.Error("usuário não encontrado"))
	}

	props, err := propertiesMap(user.Get("properties"))
	if err != nil {
		return ErrJSON(500, err, "erro ao ler propriedades")
	}

	version := "v2"
	if payloadVersion := strings.TrimSpace(req.Payload.String("version")); payloadVersion != "" {
		version = payloadVersion
	}
	key := tourCompletedKey
	if version != "v2" {
		key = "tour_" + version + "_completed"
	}

	props[key] = true
	user.Set("properties", props)
	if err := req.Dao().SaveRecord(user); err != nil {
		return ErrJSON(500, err, "erro ao salvar conclusão do tour")
	}

	return c.JSON(200, g.M(
		"ok", true,
		"completed", true,
		"version", version,
	))
}

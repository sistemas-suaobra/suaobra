package server

import (
	"strings"

	"github.com/flarco/g"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/models"
	"github.com/spf13/cast"
	"github.com/suaobra/suaobra-app/store"
)

func createNotification(req *Request, input map[string]string) error {
	collection, err := req.Dao().FindCollectionByNameOrId("notifications")
	if err != nil {
		return err
	}
	record := models.NewRecord(collection)
	record.SetId(store.ID(collection.Name, ""))
	record.Set("user_id", strings.TrimSpace(input["user_id"]))
	record.Set("title", strings.TrimSpace(input["title"]))
	record.Set("body", strings.TrimSpace(input["body"]))
	record.Set("href", strings.TrimSpace(input["href"]))
	record.Set("kind", strings.TrimSpace(input["kind"]))
	record.Set("ref_id", strings.TrimSpace(input["ref_id"]))
	record.Set("read", false)
	return req.Dao().SaveRecord(record)
}

// ListNotifications lista notificações do usuário autenticado.
func ListNotifications(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(401, req.Error, "error creating request")
	}
	auth := req.AuthRecord()
	if auth == nil {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	records, err := req.SqlQueryRecords(g.R(`
		select id, user_id, title, body, href, kind, ref_id, read, created, updated
		from main.notifications
		where user_id = '{user_id}'
		order by created desc
		limit 100
	`, "user_id", auth.Id))
	if err != nil {
		return ErrJSON(500, err, "erro ao listar notificações")
	}
	return c.JSON(200, records)
}

// MarkNotificationsRead marca notificações como lidas.
func MarkNotificationsRead(c echo.Context) error {
	req := NewRequest(c)
	if req.Error != nil {
		return ErrJSON(401, req.Error, "error creating request")
	}
	auth := req.AuthRecord()
	if auth == nil {
		return ErrJSON(401, g.Error("unauthorized"))
	}

	idsRaw, _ := req.Payload["ids"].([]any)
	all := cast.ToBool(req.Payload["all"])
	if all {
		_, err := req.Dao().DB().NewQuery(`
			update main.notifications set read = 1 where user_id = {:user_id} and read = 0
		`).Bind(dbx.Params{"user_id": auth.Id}).Execute()
		if err != nil {
			return ErrJSON(500, err, "erro ao marcar notificações")
		}
		return c.JSON(200, g.M("ok", true))
	}

	ids := make([]string, 0, len(idsRaw))
	for _, item := range idsRaw {
		id := strings.TrimSpace(cast.ToString(item))
		if id != "" {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return ErrJSON(400, g.Error("informe ids ou all"))
	}

	for _, id := range ids {
		_, _ = req.Dao().DB().NewQuery(`
			update main.notifications
			set read = 1
			where id = {:id} and user_id = {:user_id}
		`).Bind(dbx.Params{"id": id, "user_id": auth.Id}).Execute()
	}
	return c.JSON(200, g.M("ok", true))
}

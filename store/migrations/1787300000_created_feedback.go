package migrations

import (
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		jsonData := `{
			"id": "col_feedback_suaobra01",
			"name": "feedback",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "fbtype00000001",
					"name": "type",
					"type": "text",
					"required": true,
					"presentable": true,
					"unique": false,
					"options": { "min": null, "max": 32, "pattern": "" }
				},
				{
					"system": false,
					"id": "fbsubject00001",
					"name": "subject",
					"type": "text",
					"required": true,
					"presentable": true,
					"unique": false,
					"options": { "min": null, "max": 200, "pattern": "" }
				},
				{
					"system": false,
					"id": "fbmessage00001",
					"name": "message",
					"type": "text",
					"required": true,
					"presentable": false,
					"unique": false,
					"options": { "min": null, "max": null, "pattern": "" }
				},
				{
					"system": false,
					"id": "fbuserid000001",
					"name": "user_id",
					"type": "text",
					"required": true,
					"presentable": false,
					"unique": false,
					"options": { "min": null, "max": 64, "pattern": "" }
				},
				{
					"system": false,
					"id": "fbteamid000001",
					"name": "team_id",
					"type": "text",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": { "min": null, "max": 64, "pattern": "" }
				},
				{
					"system": false,
					"id": "fbuseremail001",
					"name": "user_email",
					"type": "email",
					"required": false,
					"presentable": true,
					"unique": false,
					"options": { "exceptDomains": null, "onlyDomains": null }
				},
				{
					"system": false,
					"id": "fbusername0001",
					"name": "user_name",
					"type": "text",
					"required": false,
					"presentable": true,
					"unique": false,
					"options": { "min": null, "max": 200, "pattern": "" }
				},
				{
					"system": false,
					"id": "fbattach000001",
					"name": "attachments",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {}
				},
				{
					"system": false,
					"id": "fbstatus000001",
					"name": "status",
					"type": "text",
					"required": true,
					"presentable": true,
					"unique": false,
					"options": { "min": null, "max": 32, "pattern": "" }
				}
			],
			"indexes": [
				"CREATE INDEX ` + "`" + `idx_feedback_type_created` + "`" + ` ON ` + "`" + `feedback` + "`" + ` (` + "`" + `type` + "`" + `, ` + "`" + `created` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_feedback_user` + "`" + ` ON ` + "`" + `feedback` + "`" + ` (` + "`" + `user_id` + "`" + `)"
			],
			"listRule": null,
			"viewRule": null,
			"createRule": null,
			"updateRule": null,
			"deleteRule": null,
			"options": {}
		}`

		collection := &models.Collection{}
		if err := json.Unmarshal([]byte(jsonData), collection); err != nil {
			return err
		}
		return daos.New(db).SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db)
		collection, err := dao.FindCollectionByNameOrId("col_feedback_suaobra01")
		if err != nil {
			return nil
		}
		return dao.DeleteCollection(collection)
	})
}

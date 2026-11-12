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
			"id": "col_notifications_sua1",
			"name": "notifications",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "ntuserid000001",
					"name": "user_id",
					"type": "text",
					"required": true,
					"presentable": false,
					"unique": false,
					"options": { "min": null, "max": 64, "pattern": "" }
				},
				{
					"system": false,
					"id": "nttitle0000001",
					"name": "title",
					"type": "text",
					"required": true,
					"presentable": true,
					"unique": false,
					"options": { "min": null, "max": 200, "pattern": "" }
				},
				{
					"system": false,
					"id": "ntbody00000001",
					"name": "body",
					"type": "text",
					"required": true,
					"presentable": false,
					"unique": false,
					"options": { "min": null, "max": null, "pattern": "" }
				},
				{
					"system": false,
					"id": "nthref00000001",
					"name": "href",
					"type": "text",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": { "min": null, "max": 300, "pattern": "" }
				},
				{
					"system": false,
					"id": "ntkind00000001",
					"name": "kind",
					"type": "text",
					"required": true,
					"presentable": true,
					"unique": false,
					"options": { "min": null, "max": 64, "pattern": "" }
				},
				{
					"system": false,
					"id": "ntrefid0000001",
					"name": "ref_id",
					"type": "text",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": { "min": null, "max": 64, "pattern": "" }
				},
				{
					"system": false,
					"id": "ntread00000001",
					"name": "read",
					"type": "bool",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {}
				}
			],
			"indexes": [
				"CREATE INDEX ` + "`" + `idx_notifications_user_created` + "`" + ` ON ` + "`" + `notifications` + "`" + ` (` + "`" + `user_id` + "`" + `, ` + "`" + `created` + "`" + `)"
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
		collection, err := dao.FindCollectionByNameOrId("col_notifications_sua1")
		if err != nil {
			return nil
		}
		return dao.DeleteCollection(collection)
	})
}

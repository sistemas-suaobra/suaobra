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
			"id": "p980c5513p13sfl",
			"created": "2023-09-26 12:46:55.082Z",
			"updated": "2023-09-26 12:46:55.082Z",
			"name": "list_stage",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "igclyogo",
					"name": "list_id",
					"type": "relation",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"collectionId": "xe9w10fta5tu89v",
						"cascadeDelete": false,
						"minSelect": null,
						"maxSelect": 1,
						"displayFields": null
					}
				},
				{
					"system": false,
					"id": "dq4jyfk2",
					"name": "name",
					"type": "text",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"min": null,
						"max": null,
						"pattern": ""
					}
				},
				{
					"system": false,
					"id": "km8wl7qa",
					"name": "properties",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {}
				}
			],
			"indexes": [
				"CREATE INDEX ` + "`" + `idx_LubeNjx` + "`" + ` ON ` + "`" + `list_stage` + "`" + ` (` + "`" + `list_id` + "`" + `)"
			],
			"listRule": null,
			"viewRule": null,
			"createRule": null,
			"updateRule": null,
			"deleteRule": null,
			"options": {}
		}`

		collection := &models.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return daos.New(db).SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("p980c5513p13sfl")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}

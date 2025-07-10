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
			"id": "_collection_cydQmBK5ZeGpLfE",
			"created": "2023-11-06 12:31:24.500Z",
			"updated": "2023-11-06 12:31:24.500Z",
			"name": "obra_note",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "hhivoakm",
					"name": "obra_id",
					"type": "text",
					"required": true,
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
					"id": "zntkktub",
					"name": "note",
					"type": "editor",
					"required": true,
					"presentable": false,
					"unique": false,
					"options": {
						"convertUrls": false
					}
				},
				{
					"system": false,
					"id": "mupq4jb7",
					"name": "user_id",
					"type": "relation",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"collectionId": "_pb_users_auth_",
						"cascadeDelete": false,
						"minSelect": null,
						"maxSelect": 1,
						"displayFields": null
					}
				}
			],
			"indexes": [
				"CREATE INDEX ` + "`" + `idx_JX8RWC0` + "`" + ` ON ` + "`" + `obra_note` + "`" + ` (` + "`" + `obra_id` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_5iAymaC` + "`" + ` ON ` + "`" + `obra_note` + "`" + ` (` + "`" + `user_id` + "`" + `)"
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

		collection, err := dao.FindCollectionByNameOrId("_collection_cydQmBK5ZeGpLfE")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}

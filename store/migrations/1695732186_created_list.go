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
			"id": "xe9w10fta5tu89v",
			"created": "2023-09-26 12:43:06.798Z",
			"updated": "2023-09-26 12:43:06.798Z",
			"name": "list",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "dnwfvyw0",
					"name": "team_id",
					"type": "relation",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"collectionId": "r12uiife4c4e8zr",
						"cascadeDelete": false,
						"minSelect": null,
						"maxSelect": 1,
						"displayFields": null
					}
				},
				{
					"system": false,
					"id": "mxtgciz2",
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
					"id": "hs5smpht",
					"name": "properties",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {}
				},
				{
					"system": false,
					"id": "cmki8uze",
					"name": "deleted_at",
					"type": "date",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"min": "",
						"max": ""
					}
				}
			],
			"indexes": [
				"CREATE INDEX ` + "`" + `idx_cV2jd76` + "`" + ` ON ` + "`" + `list` + "`" + ` (` + "`" + `team_id` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_mU5cHL6` + "`" + ` ON ` + "`" + `list` + "`" + ` (` + "`" + `deleted_at` + "`" + `)"
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

		collection, err := dao.FindCollectionByNameOrId("xe9w10fta5tu89v")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}

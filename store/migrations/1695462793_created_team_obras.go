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
			"id": "qzdr0f0p4ddqfqe",
			"created": "2023-09-23 09:53:13.710Z",
			"updated": "2023-09-23 09:53:13.710Z",
			"name": "team_obras",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "5mvmmvku",
					"name": "team_id",
					"type": "relation",
					"required": false,
					"unique": false,
					"options": {
						"collectionId": "r12uiife4c4e8zr",
						"cascadeDelete": false,
						"minSelect": null,
						"maxSelect": 1,
						"displayFields": []
					}
				},
				{
					"system": false,
					"id": "esr9k2dm",
					"name": "obra_id",
					"type": "text",
					"required": false,
					"unique": false,
					"options": {
						"min": null,
						"max": null,
						"pattern": ""
					}
				},
				{
					"system": false,
					"id": "kxcdyvm4",
					"name": "visited_at",
					"type": "date",
					"required": false,
					"unique": false,
					"options": {
						"min": "",
						"max": ""
					}
				},
				{
					"system": false,
					"id": "wv156ip5",
					"name": "favorited_at",
					"type": "date",
					"required": false,
					"unique": false,
					"options": {
						"min": "",
						"max": ""
					}
				},
				{
					"system": false,
					"id": "3tmlcge1",
					"name": "excluded_at",
					"type": "date",
					"required": false,
					"unique": false,
					"options": {
						"min": "",
						"max": ""
					}
				}
			],
			"indexes": [
				"CREATE UNIQUE INDEX ` + "`" + `idx_al1JxBl` + "`" + ` ON ` + "`" + `team_obras` + "`" + ` (\n  ` + "`" + `team_id` + "`" + `,\n  ` + "`" + `obra_id` + "`" + `\n)",
				"CREATE INDEX ` + "`" + `idx_l70ogsl` + "`" + ` ON ` + "`" + `team_obras` + "`" + ` (` + "`" + `visited_at` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_RqdgzUy` + "`" + ` ON ` + "`" + `team_obras` + "`" + ` (` + "`" + `favorited_at` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_RvZyJvN` + "`" + ` ON ` + "`" + `team_obras` + "`" + ` (` + "`" + `excluded_at` + "`" + `)"
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

		collection, err := dao.FindCollectionByNameOrId("qzdr0f0p4ddqfqe")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}

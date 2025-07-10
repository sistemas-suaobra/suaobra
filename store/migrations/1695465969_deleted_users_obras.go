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
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("6r8ux9luhyejwyf")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	}, func(db dbx.Builder) error {
		jsonData := `{
			"id": "6r8ux9luhyejwyf",
			"created": "2023-09-23 03:11:42.422Z",
			"updated": "2023-09-23 09:33:58.838Z",
			"name": "users_obras",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "svk0wuem",
					"name": "user_id",
					"type": "relation",
					"required": true,
					"unique": false,
					"options": {
						"collectionId": "_pb_users_auth_",
						"cascadeDelete": false,
						"minSelect": null,
						"maxSelect": 1,
						"displayFields": []
					}
				},
				{
					"system": false,
					"id": "86ro1qoo",
					"name": "obra_id",
					"type": "text",
					"required": true,
					"unique": false,
					"options": {
						"min": null,
						"max": null,
						"pattern": ""
					}
				},
				{
					"system": false,
					"id": "bhknf4q8",
					"name": "status",
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
					"id": "4nzhw6xp",
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
					"id": "s9dk4jgf",
					"name": "excluded_at",
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
					"id": "fpcuw0cz",
					"name": "visited_at",
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
				"CREATE INDEX ` + "`" + `idx_8czuntv` + "`" + ` ON ` + "`" + `users_obras` + "`" + ` (` + "`" + `favorited_at` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_4bZz28i` + "`" + ` ON ` + "`" + `users_obras` + "`" + ` (` + "`" + `excluded_at` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_ob5OKBz` + "`" + ` ON ` + "`" + `users_obras` + "`" + ` (` + "`" + `visited_at` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_10Ig2Kd` + "`" + ` ON ` + "`" + `users_obras` + "`" + ` (` + "`" + `status` + "`" + `)",
				"CREATE UNIQUE INDEX ` + "`" + `idx_2eWg4zG` + "`" + ` ON ` + "`" + `users_obras` + "`" + ` (\n  ` + "`" + `user_id` + "`" + `,\n  ` + "`" + `obra_id` + "`" + `\n)"
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
	})
}

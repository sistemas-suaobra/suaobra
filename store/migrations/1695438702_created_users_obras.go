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
			"id": "6r8ux9luhyejwyf",
			"created": "2023-09-23 03:11:42.422Z",
			"updated": "2023-09-23 03:11:42.422Z",
			"name": "users_obras",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "z6rs3kp1",
					"name": "user_id",
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
					"id": "86ro1qoo",
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
				"CREATE UNIQUE INDEX ` + "`" + `idx_0ZGYJ4Q` + "`" + ` ON ` + "`" + `users_obras` + "`" + ` (\n  ` + "`" + `user_id` + "`" + `,\n  ` + "`" + `obra_id` + "`" + `\n)",
				"CREATE INDEX ` + "`" + `idx_8czuntv` + "`" + ` ON ` + "`" + `users_obras` + "`" + ` (` + "`" + `favorited_at` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_4bZz28i` + "`" + ` ON ` + "`" + `users_obras` + "`" + ` (` + "`" + `excluded_at` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_ob5OKBz` + "`" + ` ON ` + "`" + `users_obras` + "`" + ` (` + "`" + `visited_at` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_10Ig2Kd` + "`" + ` ON ` + "`" + `users_obras` + "`" + ` (` + "`" + `status` + "`" + `)"
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

		collection, err := dao.FindCollectionByNameOrId("6r8ux9luhyejwyf")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}

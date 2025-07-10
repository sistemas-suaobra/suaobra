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
			"id": "r12uiife4c4e8zr",
			"created": "2023-09-23 09:48:26.308Z",
			"updated": "2023-09-23 09:48:26.308Z",
			"name": "teams",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "qbtojp7p",
					"name": "owner",
					"type": "relation",
					"required": false,
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
					"id": "4d9feu6f",
					"name": "active",
					"type": "bool",
					"required": false,
					"unique": false,
					"options": {}
				},
				{
					"system": false,
					"id": "3lwx03l8",
					"name": "blocked",
					"type": "bool",
					"required": false,
					"unique": false,
					"options": {}
				}
			],
			"indexes": [
				"CREATE INDEX ` + "`" + `idx_LjSxmxc` + "`" + ` ON ` + "`" + `teams` + "`" + ` (` + "`" + `owner` + "`" + `)"
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

		collection, err := dao.FindCollectionByNameOrId("r12uiife4c4e8zr")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}

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
			"id": "hqvr9jtjmjjp1ps",
			"created": "2023-09-14 13:23:02.242Z",
			"updated": "2023-09-14 13:23:02.242Z",
			"name": "users_old",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "outxdcwn",
					"name": "email",
					"type": "email",
					"required": false,
					"unique": false,
					"options": {
						"exceptDomains": null,
						"onlyDomains": null
					}
				},
				{
					"system": false,
					"id": "zvd09klj",
					"name": "entitlements",
					"type": "json",
					"required": false,
					"unique": false,
					"options": {}
				}
			],
			"indexes": [
				"CREATE UNIQUE INDEX ` + "`" + `idx_fGzp6ue` + "`" + ` ON ` + "`" + `users_old` + "`" + ` (` + "`" + `email` + "`" + `)"
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

		collection, err := dao.FindCollectionByNameOrId("hqvr9jtjmjjp1ps")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}

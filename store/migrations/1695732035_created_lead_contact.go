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
			"id": "sonqjf6tyc0da5b",
			"created": "2023-09-26 12:40:35.289Z",
			"updated": "2023-09-26 12:40:35.289Z",
			"name": "lead_contact",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "smsjhbws",
					"name": "lead_id",
					"type": "relation",
					"required": true,
					"presentable": false,
					"unique": false,
					"options": {
						"collectionId": "qzdr0f0p4ddqfqe",
						"cascadeDelete": false,
						"minSelect": null,
						"maxSelect": 1,
						"displayFields": null
					}
				},
				{
					"system": false,
					"id": "dwhgzn2m",
					"name": "contact_id",
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
					"id": "afelderg",
					"name": "properties",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {}
				}
			],
			"indexes": [
				"CREATE UNIQUE INDEX ` + "`" + `idx_GZN5VcU` + "`" + ` ON ` + "`" + `lead_contact` + "`" + ` (\n  ` + "`" + `lead_id` + "`" + `,\n  ` + "`" + `contact_id` + "`" + `\n)"
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

		collection, err := dao.FindCollectionByNameOrId("sonqjf6tyc0da5b")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}

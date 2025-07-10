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
			"id": "m7bh3r7viqvvqrb",
			"created": "2023-09-26 12:44:50.977Z",
			"updated": "2023-09-26 12:44:50.977Z",
			"name": "list_lead",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "b6m9qmbx",
					"name": "lead_id",
					"type": "relation",
					"required": false,
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
					"id": "6yi8qpne",
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
					"id": "fsxnai18",
					"name": "properties",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {}
				}
			],
			"indexes": [
				"CREATE UNIQUE INDEX ` + "`" + `idx_krItb1x` + "`" + ` ON ` + "`" + `list_lead` + "`" + ` (\n  ` + "`" + `lead_id` + "`" + `,\n  ` + "`" + `list_id` + "`" + `\n)"
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

		collection, err := dao.FindCollectionByNameOrId("m7bh3r7viqvvqrb")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}

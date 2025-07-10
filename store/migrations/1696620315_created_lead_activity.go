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
			"id": "_collection_3jbxfRdyGCQsXpb",
			"created": "2023-10-06 19:25:15.187Z",
			"updated": "2023-10-06 19:25:15.187Z",
			"name": "lead_activity",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "s5x464jd",
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
					"id": "rywm2abm",
					"name": "actor_id",
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
				},
				{
					"system": false,
					"id": "mgsysajq",
					"name": "type",
					"type": "select",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"maxSelect": 1,
						"values": [
							"history",
							"note",
							"obra",
							"info-obras",
							"email",
							"whatsapp",
							"phone"
						]
					}
				},
				{
					"system": false,
					"id": "rc8dfk1o",
					"name": "body",
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
					"id": "zzron25n",
					"name": "properties",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {}
				}
			],
			"indexes": [
				"CREATE INDEX ` + "`" + `idx_FCVKM9h` + "`" + ` ON ` + "`" + `lead_activity` + "`" + ` (` + "`" + `lead_id` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_S8GRinC` + "`" + ` ON ` + "`" + `lead_activity` + "`" + ` (` + "`" + `type` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_ACZWHVQ` + "`" + ` ON ` + "`" + `lead_activity` + "`" + ` (` + "`" + `actor_id` + "`" + `)"
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

		collection, err := dao.FindCollectionByNameOrId("_collection_3jbxfRdyGCQsXpb")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}

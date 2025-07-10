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
			"id": "89ju1qkr9p4o35m",
			"created": "2023-09-12 20:34:03.740Z",
			"updated": "2023-09-12 20:34:03.740Z",
			"name": "events",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "cka8slig",
					"name": "messageId",
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
					"id": "itctb62d",
					"name": "type",
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
					"id": "gs3z95gc",
					"name": "event",
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
					"id": "qa0on8cg",
					"name": "channel",
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
					"id": "sra0deki",
					"name": "userId",
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
					"id": "qky9dfug",
					"name": "anonymousId",
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
					"id": "iw4nmlqh",
					"name": "originalTimestamp",
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
					"id": "7hwycvzm",
					"name": "sentAt",
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
					"id": "ifj43l3v",
					"name": "context",
					"type": "json",
					"required": false,
					"unique": false,
					"options": {}
				},
				{
					"system": false,
					"id": "tx3gshq9",
					"name": "properties",
					"type": "json",
					"required": false,
					"unique": false,
					"options": {}
				},
				{
					"system": false,
					"id": "avrtvvhm",
					"name": "integrations",
					"type": "json",
					"required": false,
					"unique": false,
					"options": {}
				}
			],
			"indexes": [],
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

		collection, err := dao.FindCollectionByNameOrId("89ju1qkr9p4o35m")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}

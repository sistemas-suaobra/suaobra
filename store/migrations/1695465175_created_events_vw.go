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
			"id": "8q056kyr077k5eu",
			"created": "2023-09-23 10:32:55.602Z",
			"updated": "2023-09-23 10:32:55.602Z",
			"name": "events_vw",
			"type": "view",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "n9yng1ga",
					"name": "originaltimestamp",
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
					"id": "cj0ehbcv",
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
					"id": "vaukksip",
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
					"id": "5ytpt9zg",
					"name": "properties",
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
			"options": {
				"query": "select id, originaltimestamp, type, event, properties from events"
			}
		}`

		collection := &models.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return daos.New(db).SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("8q056kyr077k5eu")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}

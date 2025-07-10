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
			"id": "obcn6aebh75cl8p",
			"created": "2023-09-09 18:38:37.605Z",
			"updated": "2023-09-09 18:38:37.605Z",
			"name": "bairro",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "ajabmk1q",
					"name": "uf",
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
					"id": "soucitgr",
					"name": "cidade",
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
					"id": "jhcaefih",
					"name": "bairro",
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
					"id": "f8r1ncu3",
					"name": "latitude",
					"type": "number",
					"required": false,
					"unique": false,
					"options": {
						"min": null,
						"max": null
					}
				},
				{
					"system": false,
					"id": "ydbfutef",
					"name": "longitude",
					"type": "number",
					"required": false,
					"unique": false,
					"options": {
						"min": null,
						"max": null
					}
				},
				{
					"system": false,
					"id": "kigl0ge9",
					"name": "payload",
					"type": "json",
					"required": false,
					"unique": false,
					"options": {}
				}
			],
			"indexes": [
				"CREATE UNIQUE INDEX ` + "`" + `idx_riJ7UVr` + "`" + ` ON ` + "`" + `bairro` + "`" + ` (\n  ` + "`" + `uf` + "`" + `,\n  ` + "`" + `cidade` + "`" + `,\n  ` + "`" + `bairro` + "`" + `\n)"
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

		collection, err := dao.FindCollectionByNameOrId("obcn6aebh75cl8p")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}

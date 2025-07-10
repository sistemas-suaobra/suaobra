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
			"id": "_collection_tnUtpZr6ec9Yjxi",
			"created": "2024-01-05 22:40:20.974Z",
			"updated": "2024-01-05 22:40:20.974Z",
			"name": "vw_usuarios",
			"type": "view",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "kl2msfva",
					"name": "email",
					"type": "email",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"exceptDomains": null,
						"onlyDomains": null
					}
				},
				{
					"system": false,
					"id": "h30ne8tz",
					"name": "nome",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {}
				},
				{
					"system": false,
					"id": "bqbg34zz",
					"name": "telephone",
					"type": "json",
					"required": false,
					"presentable": false,
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
				"query": "select\n  user.id,\n  user.email,\n  coalesce(nullif(user.name, ''), json_extract(user.properties, '$.name'), '-') as nome,\n  coalesce(json_extract(user.properties, '$.phone'), '-') as telephone,\n  created,\n  updated\nfrom user\norder by created desc"
			}
		}`

		collection := &models.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return daos.New(db).SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("_collection_tnUtpZr6ec9Yjxi")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}

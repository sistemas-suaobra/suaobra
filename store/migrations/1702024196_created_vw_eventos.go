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
			"id": "_collection_fgFv1hfA4mZ947D",
			"created": "2023-12-08 08:29:56.829Z",
			"updated": "2023-12-08 08:29:56.829Z",
			"name": "vw_eventos",
			"type": "view",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "mr9w7ifr",
					"name": "tipo_evento",
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
					"id": "2poee4py",
					"name": "nome_evento",
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
					"id": "nfvau0pt",
					"name": "usario_email",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {}
				},
				{
					"system": false,
					"id": "opgden11",
					"name": "data_evento",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {}
				},
				{
					"system": false,
					"id": "t4fxqdeg",
					"name": "uf",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {}
				},
				{
					"system": false,
					"id": "g7scxsag",
					"name": "cidade",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {}
				},
				{
					"system": false,
					"id": "ttb2axli",
					"name": "pagina",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {}
				},
				{
					"system": false,
					"id": "uvu0zeym",
					"name": "tamanho",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {}
				},
				{
					"system": false,
					"id": "lxbqs1kp",
					"name": "dados",
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
				"query": "select\n  id,\n  type as tipo_evento,\n  event as nome_evento,\n  json_extract(properties, '$.user.email') as usario_email,\n  datetime(julianday(originaltimestamp) - 0.125) as data_evento, -- BRT\n  json_extract(properties, '$.state') as uf,\n  json_extract(properties, '$.city') as cidade,\n  json_extract(properties, '$.page_number') as pagina,\n  json_extract(properties, '$.size') as tamanho,\n  properties as dados\nfrom rudderstack\norder by data_evento desc"
			}
		}`

		collection := &models.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return daos.New(db).SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("_collection_fgFv1hfA4mZ947D")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}

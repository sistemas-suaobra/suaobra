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
			"id": "_collection_1BPiQsdzfoeHAOH",
			"created": "2023-12-08 09:05:22.252Z",
			"updated": "2023-12-08 09:05:22.252Z",
			"name": "vw_eventos_diarios",
			"type": "view",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "5mtfmb5a",
					"name": "date",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {}
				},
				{
					"system": false,
					"id": "knnbdpig",
					"name": "email",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {}
				},
				{
					"system": false,
					"id": "xsqso8rk",
					"name": "eventos",
					"type": "number",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"min": null,
						"max": null,
						"noDecimal": false
					}
				},
				{
					"system": false,
					"id": "3aj8vfyb",
					"name": "login_eventos",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {}
				},
				{
					"system": false,
					"id": "ozdlh6wr",
					"name": "venda_mais_eventos",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {}
				},
				{
					"system": false,
					"id": "mf0y9lyc",
					"name": "obras_plus_eventos",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {}
				},
				{
					"system": false,
					"id": "ebrv8bbw",
					"name": "track_eventos",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {}
				},
				{
					"system": false,
					"id": "lzkgvz7a",
					"name": "page_eventos",
					"type": "json",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {}
				},
				{
					"system": false,
					"id": "oqw9zg3o",
					"name": "identity_eventos",
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
				"query": "select\n  (ROW_NUMBER() OVER()) as id,\n  date(data_evento) as date,\n  nullif(usario_email, '') email,\n  count(1) eventos,\n  sum(case when nome_evento like 'login-%' then 1 else 0 end) login_eventos,\n  sum(case when nome_evento like 'venda-mais-%' then 1 else 0 end) venda_mais_eventos,\n  sum(case when nome_evento like 'obras-plus-%' then 1 else 0 end) obras_plus_eventos,\n  sum(case when tipo_evento = 'track' then 1 else 0 end) track_eventos,\n  sum(case when tipo_evento = 'page' then 1 else 0 end) page_eventos,\n  sum(case when tipo_evento = 'identify' then 1 else 0 end) identity_eventos\nfrom vw_eventos\ngroup by 1, 2\norder by 1 desc"
			}
		}`

		collection := &models.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return daos.New(db).SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("_collection_1BPiQsdzfoeHAOH")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}

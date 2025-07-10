package migrations

import (
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models/schema"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("r12uiife4c4e8zr")
		if err != nil {
			return err
		}

		// update
		edit_cities := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "fmc0oxjv",
			"name": "cities",
			"type": "relation",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "2nsfuh4jhipsus4",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": null,
				"displayFields": null
			}
		}`), edit_cities)
		collection.Schema.AddField(edit_cities)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("r12uiife4c4e8zr")
		if err != nil {
			return err
		}

		// update
		edit_cities := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "fmc0oxjv",
			"name": "cidades",
			"type": "relation",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "2nsfuh4jhipsus4",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": null,
				"displayFields": null
			}
		}`), edit_cities)
		collection.Schema.AddField(edit_cities)

		return dao.SaveCollection(collection)
	})
}

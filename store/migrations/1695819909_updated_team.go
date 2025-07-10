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

		// add
		new_cidades := &schema.SchemaField{}
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
		}`), new_cidades)
		collection.Schema.AddField(new_cidades)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("r12uiife4c4e8zr")
		if err != nil {
			return err
		}

		// remove
		collection.Schema.RemoveField("fmc0oxjv")

		return dao.SaveCollection(collection)
	})
}

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
		new_name := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "djgum8mv",
			"name": "name",
			"type": "text",
			"required": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), new_name)
		collection.Schema.AddField(new_name)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("r12uiife4c4e8zr")
		if err != nil {
			return err
		}

		// remove
		collection.Schema.RemoveField("djgum8mv")

		return dao.SaveCollection(collection)
	})
}

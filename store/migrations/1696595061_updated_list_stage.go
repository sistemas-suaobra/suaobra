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

		collection, err := dao.FindCollectionByNameOrId("p980c5513p13sfl")
		if err != nil {
			return err
		}

		// add
		new_order := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "lugwt2hb",
			"name": "order",
			"type": "number",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"noDecimal": false
			}
		}`), new_order)
		collection.Schema.AddField(new_order)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("p980c5513p13sfl")
		if err != nil {
			return err
		}

		// remove
		collection.Schema.RemoveField("lugwt2hb")

		return dao.SaveCollection(collection)
	})
}

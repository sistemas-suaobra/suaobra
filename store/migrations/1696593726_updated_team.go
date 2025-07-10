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
		new_state := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "al8whxfv",
			"name": "state",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_state)
		collection.Schema.AddField(new_state)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("r12uiife4c4e8zr")
		if err != nil {
			return err
		}

		// remove
		collection.Schema.RemoveField("al8whxfv")

		return dao.SaveCollection(collection)
	})
}

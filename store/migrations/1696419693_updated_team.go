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
		new_entitlements := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "jiwykqw0",
			"name": "entitlements",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_entitlements)
		collection.Schema.AddField(new_entitlements)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("r12uiife4c4e8zr")
		if err != nil {
			return err
		}

		// remove
		collection.Schema.RemoveField("jiwykqw0")

		return dao.SaveCollection(collection)
	})
}

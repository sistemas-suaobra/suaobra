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

		collection, err := dao.FindCollectionByNameOrId("m7bh3r7viqvvqrb")
		if err != nil {
			return err
		}

		// remove
		collection.Schema.RemoveField("amcba1z8")

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("m7bh3r7viqvvqrb")
		if err != nil {
			return err
		}

		// add
		del_stage_rank := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "amcba1z8",
			"name": "stage_rank",
			"type": "number",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"noDecimal": true
			}
		}`), del_stage_rank)
		collection.Schema.AddField(del_stage_rank)

		return dao.SaveCollection(collection)
	})
}

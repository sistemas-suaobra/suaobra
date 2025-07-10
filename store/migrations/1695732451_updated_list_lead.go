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

		json.Unmarshal([]byte(`[
			"CREATE UNIQUE INDEX ` + "`" + `idx_krItb1x` + "`" + ` ON ` + "`" + `list_lead` + "`" + ` (\n  ` + "`" + `lead_id` + "`" + `,\n  ` + "`" + `list_id` + "`" + `\n)",
			"CREATE INDEX ` + "`" + `idx_4L0OZAl` + "`" + ` ON ` + "`" + `list_lead` + "`" + ` (` + "`" + `stage_id` + "`" + `)"
		]`), &collection.Indexes)

		// add
		new_stage_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "cjexxcbt",
			"name": "stage_id",
			"type": "relation",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "p980c5513p13sfl",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), new_stage_id)
		collection.Schema.AddField(new_stage_id)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("m7bh3r7viqvvqrb")
		if err != nil {
			return err
		}

		json.Unmarshal([]byte(`[
			"CREATE UNIQUE INDEX ` + "`" + `idx_krItb1x` + "`" + ` ON ` + "`" + `list_lead` + "`" + ` (\n  ` + "`" + `lead_id` + "`" + `,\n  ` + "`" + `list_id` + "`" + `\n)"
		]`), &collection.Indexes)

		// remove
		collection.Schema.RemoveField("cjexxcbt")

		return dao.SaveCollection(collection)
	})
}

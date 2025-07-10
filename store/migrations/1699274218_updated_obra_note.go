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

		collection, err := dao.FindCollectionByNameOrId("_collection_cydQmBK5ZeGpLfE")
		if err != nil {
			return err
		}

		json.Unmarshal([]byte(`[
			"CREATE INDEX ` + "`" + `idx_JX8RWC0` + "`" + ` ON ` + "`" + `obra_note` + "`" + ` (` + "`" + `obra_id` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_5iAymaC` + "`" + ` ON ` + "`" + `obra_note` + "`" + ` (` + "`" + `user_id` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_7IE7GnX` + "`" + ` ON ` + "`" + `obra_note` + "`" + ` (` + "`" + `info_obra_id` + "`" + `)"
		]`), &collection.Indexes)

		// add
		new_info_obra_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "6rwxy9tx",
			"name": "info_obra_id",
			"type": "number",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"noDecimal": false
			}
		}`), new_info_obra_id)
		collection.Schema.AddField(new_info_obra_id)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("_collection_cydQmBK5ZeGpLfE")
		if err != nil {
			return err
		}

		json.Unmarshal([]byte(`[
			"CREATE INDEX ` + "`" + `idx_JX8RWC0` + "`" + ` ON ` + "`" + `obra_note` + "`" + ` (` + "`" + `obra_id` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_5iAymaC` + "`" + ` ON ` + "`" + `obra_note` + "`" + ` (` + "`" + `user_id` + "`" + `)"
		]`), &collection.Indexes)

		// remove
		collection.Schema.RemoveField("6rwxy9tx")

		return dao.SaveCollection(collection)
	})
}

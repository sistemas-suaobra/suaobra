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

		collection, err := dao.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		json.Unmarshal([]byte(`[
			"CREATE UNIQUE INDEX ` + "`" + `idx_jOLoFsU` + "`" + ` ON ` + "`" + `users` + "`" + ` (` + "`" + `legacy_id` + "`" + `)"
		]`), &collection.Indexes)

		// add
		new_legacy_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "98vujz64",
			"name": "legacy_id",
			"type": "text",
			"required": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), new_legacy_id)
		collection.Schema.AddField(new_legacy_id)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		json.Unmarshal([]byte(`[]`), &collection.Indexes)

		// remove
		collection.Schema.RemoveField("98vujz64")

		return dao.SaveCollection(collection)
	})
}

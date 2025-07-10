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

		collection, err := dao.FindCollectionByNameOrId("_collection_3jbxfRdyGCQsXpb")
		if err != nil {
			return err
		}

		json.Unmarshal([]byte(`[
			"CREATE INDEX ` + "`" + `idx_FCVKM9h` + "`" + ` ON ` + "`" + `lead_activity` + "`" + ` (` + "`" + `lead_id` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_S8GRinC` + "`" + ` ON ` + "`" + `lead_activity` + "`" + ` (` + "`" + `type` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_5aHR3oQ` + "`" + ` ON ` + "`" + `lead_activity` + "`" + ` (` + "`" + `team_id` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_ieldg7X` + "`" + ` ON ` + "`" + `lead_activity` + "`" + ` (` + "`" + `actor_email` + "`" + `)"
		]`), &collection.Indexes)

		// remove
		collection.Schema.RemoveField("rywm2abm")

		// add
		new_actor_email := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "b31n7xbm",
			"name": "actor_email",
			"type": "email",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"exceptDomains": [],
				"onlyDomains": []
			}
		}`), new_actor_email)
		collection.Schema.AddField(new_actor_email)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("_collection_3jbxfRdyGCQsXpb")
		if err != nil {
			return err
		}

		json.Unmarshal([]byte(`[
			"CREATE INDEX ` + "`" + `idx_FCVKM9h` + "`" + ` ON ` + "`" + `lead_activity` + "`" + ` (` + "`" + `lead_id` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_S8GRinC` + "`" + ` ON ` + "`" + `lead_activity` + "`" + ` (` + "`" + `type` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_ACZWHVQ` + "`" + ` ON ` + "`" + `lead_activity` + "`" + ` (` + "`" + `actor_id` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_5aHR3oQ` + "`" + ` ON ` + "`" + `lead_activity` + "`" + ` (` + "`" + `team_id` + "`" + `)"
		]`), &collection.Indexes)

		// add
		del_actor_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "rywm2abm",
			"name": "actor_id",
			"type": "relation",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "_pb_users_auth_",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), del_actor_id)
		collection.Schema.AddField(del_actor_id)

		// remove
		collection.Schema.RemoveField("b31n7xbm")

		return dao.SaveCollection(collection)
	})
}

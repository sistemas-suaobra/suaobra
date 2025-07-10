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
			"CREATE INDEX ` + "`" + `idx_ACZWHVQ` + "`" + ` ON ` + "`" + `lead_activity` + "`" + ` (` + "`" + `actor_id` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_5aHR3oQ` + "`" + ` ON ` + "`" + `lead_activity` + "`" + ` (` + "`" + `team_id` + "`" + `)"
		]`), &collection.Indexes)

		// add
		new_team_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "1swcgosy",
			"name": "team_id",
			"type": "relation",
			"required": true,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "r12uiife4c4e8zr",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), new_team_id)
		collection.Schema.AddField(new_team_id)

		// update
		edit_lead_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "s5x464jd",
			"name": "lead_id",
			"type": "relation",
			"required": true,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "qzdr0f0p4ddqfqe",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), edit_lead_id)
		collection.Schema.AddField(edit_lead_id)

		// update
		edit_type := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "mgsysajq",
			"name": "type",
			"type": "select",
			"required": true,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSelect": 1,
				"values": [
					"history",
					"note",
					"obra",
					"info-obras",
					"email",
					"whatsapp",
					"phone"
				]
			}
		}`), edit_type)
		collection.Schema.AddField(edit_type)

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
			"CREATE INDEX ` + "`" + `idx_ACZWHVQ` + "`" + ` ON ` + "`" + `lead_activity` + "`" + ` (` + "`" + `actor_id` + "`" + `)"
		]`), &collection.Indexes)

		// remove
		collection.Schema.RemoveField("1swcgosy")

		// update
		edit_lead_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "s5x464jd",
			"name": "lead_id",
			"type": "relation",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "qzdr0f0p4ddqfqe",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), edit_lead_id)
		collection.Schema.AddField(edit_lead_id)

		// update
		edit_type := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "mgsysajq",
			"name": "type",
			"type": "select",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"maxSelect": 1,
				"values": [
					"history",
					"note",
					"obra",
					"info-obras",
					"email",
					"whatsapp",
					"phone"
				]
			}
		}`), edit_type)
		collection.Schema.AddField(edit_type)

		return dao.SaveCollection(collection)
	})
}

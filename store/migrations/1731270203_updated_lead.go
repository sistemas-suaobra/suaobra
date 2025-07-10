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

		collection, err := dao.FindCollectionByNameOrId("qzdr0f0p4ddqfqe")
		if err != nil {
			return err
		}

		json.Unmarshal([]byte(`[
			"CREATE UNIQUE INDEX ` + "`" + `idx_al1JxBl` + "`" + ` ON ` + "`" + `lead` + "`" + ` (\n  ` + "`" + `team_id` + "`" + `,\n  ` + "`" + `obra_id` + "`" + `\n)",
			"CREATE INDEX ` + "`" + `idx_l70ogsl` + "`" + ` ON ` + "`" + `lead` + "`" + ` (` + "`" + `visited_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_RqdgzUy` + "`" + ` ON ` + "`" + `lead` + "`" + ` (` + "`" + `favorited_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_Yeq3ZSJ` + "`" + ` ON ` + "`" + `lead` + "`" + ` (` + "`" + `owner_id` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_5nvenYI` + "`" + ` ON ` + "`" + `lead` + "`" + ` (` + "`" + `excluded_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_p3U3oaz` + "`" + ` ON ` + "`" + `lead` + "`" + ` (` + "`" + `owner_contacted_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_V2TZ0Mj` + "`" + ` ON ` + "`" + `lead` + "`" + ` (` + "`" + `professional_contacted_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_sZ5oFIb` + "`" + ` ON ` + "`" + `lead` + "`" + ` (` + "`" + `professional_contact_pending_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_N6mA4kk` + "`" + ` ON ` + "`" + `lead` + "`" + ` (` + "`" + `owner_contact_pending_at` + "`" + `)"
		]`), &collection.Indexes)

		// add
		new_professional_contact_pending_at := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "titkt578",
			"name": "professional_contact_pending_at",
			"type": "date",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": "",
				"max": ""
			}
		}`), new_professional_contact_pending_at)
		collection.Schema.AddField(new_professional_contact_pending_at)

		// update
		edit_owner_contact_pending_at := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "vcljkso9",
			"name": "owner_contact_pending_at",
			"type": "date",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": "",
				"max": ""
			}
		}`), edit_owner_contact_pending_at)
		collection.Schema.AddField(edit_owner_contact_pending_at)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("qzdr0f0p4ddqfqe")
		if err != nil {
			return err
		}

		json.Unmarshal([]byte(`[
			"CREATE UNIQUE INDEX ` + "`" + `idx_al1JxBl` + "`" + ` ON ` + "`" + `lead` + "`" + ` (\n  ` + "`" + `team_id` + "`" + `,\n  ` + "`" + `obra_id` + "`" + `\n)",
			"CREATE INDEX ` + "`" + `idx_l70ogsl` + "`" + ` ON ` + "`" + `lead` + "`" + ` (` + "`" + `visited_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_RqdgzUy` + "`" + ` ON ` + "`" + `lead` + "`" + ` (` + "`" + `favorited_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_Yeq3ZSJ` + "`" + ` ON ` + "`" + `lead` + "`" + ` (` + "`" + `owner_id` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_5nvenYI` + "`" + ` ON ` + "`" + `lead` + "`" + ` (` + "`" + `excluded_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_p3U3oaz` + "`" + ` ON ` + "`" + `lead` + "`" + ` (` + "`" + `owner_contacted_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_V2TZ0Mj` + "`" + ` ON ` + "`" + `lead` + "`" + ` (` + "`" + `professional_contacted_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_sZ5oFIb` + "`" + ` ON ` + "`" + `lead` + "`" + ` (` + "`" + `contact_pending_at` + "`" + `)"
		]`), &collection.Indexes)

		// remove
		collection.Schema.RemoveField("titkt578")

		// update
		edit_owner_contact_pending_at := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "vcljkso9",
			"name": "contact_pending_at",
			"type": "date",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": "",
				"max": ""
			}
		}`), edit_owner_contact_pending_at)
		collection.Schema.AddField(edit_owner_contact_pending_at)

		return dao.SaveCollection(collection)
	})
}

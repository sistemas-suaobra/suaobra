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
			"CREATE INDEX ` + "`" + `idx_5nvenYI` + "`" + ` ON ` + "`" + `lead` + "`" + ` (` + "`" + `excluded_at` + "`" + `)"
		]`), &collection.Indexes)

		// add
		new_excluded_at := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "4ylqgrke",
			"name": "excluded_at",
			"type": "date",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": "",
				"max": ""
			}
		}`), new_excluded_at)
		collection.Schema.AddField(new_excluded_at)

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
			"CREATE INDEX ` + "`" + `idx_Yeq3ZSJ` + "`" + ` ON ` + "`" + `lead` + "`" + ` (` + "`" + `owner_id` + "`" + `)"
		]`), &collection.Indexes)

		// remove
		collection.Schema.RemoveField("4ylqgrke")

		return dao.SaveCollection(collection)
	})
}

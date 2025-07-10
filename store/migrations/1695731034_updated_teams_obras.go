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

		collection.Name = "lead"

		json.Unmarshal([]byte(`[
			"CREATE UNIQUE INDEX ` + "`" + `idx_al1JxBl` + "`" + ` ON ` + "`" + `lead` + "`" + ` (\n  ` + "`" + `team_id` + "`" + `,\n  ` + "`" + `obra_id` + "`" + `\n)",
			"CREATE INDEX ` + "`" + `idx_l70ogsl` + "`" + ` ON ` + "`" + `lead` + "`" + ` (` + "`" + `visited_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_RqdgzUy` + "`" + ` ON ` + "`" + `lead` + "`" + ` (` + "`" + `favorited_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_Yeq3ZSJ` + "`" + ` ON ` + "`" + `lead` + "`" + ` (` + "`" + `owner_id` + "`" + `)"
		]`), &collection.Indexes)

		// remove
		collection.Schema.RemoveField("3tmlcge1")

		// add
		new_owner_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "myvc8buo",
			"name": "owner_id",
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
		}`), new_owner_id)
		collection.Schema.AddField(new_owner_id)

		// add
		new_properties := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "wm0uplhe",
			"name": "properties",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_properties)
		collection.Schema.AddField(new_properties)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("qzdr0f0p4ddqfqe")
		if err != nil {
			return err
		}

		collection.Name = "teams_obras"

		json.Unmarshal([]byte(`[
			"CREATE UNIQUE INDEX ` + "`" + `idx_al1JxBl` + "`" + ` ON ` + "`" + `teams_obras` + "`" + ` (\n  ` + "`" + `team_id` + "`" + `,\n  ` + "`" + `obra_id` + "`" + `\n)",
			"CREATE INDEX ` + "`" + `idx_l70ogsl` + "`" + ` ON ` + "`" + `teams_obras` + "`" + ` (` + "`" + `visited_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_RqdgzUy` + "`" + ` ON ` + "`" + `teams_obras` + "`" + ` (` + "`" + `favorited_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_RvZyJvN` + "`" + ` ON ` + "`" + `teams_obras` + "`" + ` (` + "`" + `excluded_at` + "`" + `)"
		]`), &collection.Indexes)

		// add
		del_excluded_at := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "3tmlcge1",
			"name": "excluded_at",
			"type": "date",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": "",
				"max": ""
			}
		}`), del_excluded_at)
		collection.Schema.AddField(del_excluded_at)

		// remove
		collection.Schema.RemoveField("myvc8buo")

		// remove
		collection.Schema.RemoveField("wm0uplhe")

		return dao.SaveCollection(collection)
	})
}

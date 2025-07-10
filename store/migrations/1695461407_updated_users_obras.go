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

		collection, err := dao.FindCollectionByNameOrId("6r8ux9luhyejwyf")
		if err != nil {
			return err
		}

		json.Unmarshal([]byte(`[
			"CREATE INDEX ` + "`" + `idx_8czuntv` + "`" + ` ON ` + "`" + `users_obras` + "`" + ` (` + "`" + `favorited_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_4bZz28i` + "`" + ` ON ` + "`" + `users_obras` + "`" + ` (` + "`" + `excluded_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_ob5OKBz` + "`" + ` ON ` + "`" + `users_obras` + "`" + ` (` + "`" + `visited_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_10Ig2Kd` + "`" + ` ON ` + "`" + `users_obras` + "`" + ` (` + "`" + `status` + "`" + `)",
			"CREATE UNIQUE INDEX ` + "`" + `idx_2eWg4zG` + "`" + ` ON ` + "`" + `users_obras` + "`" + ` (\n  ` + "`" + `user_id` + "`" + `,\n  ` + "`" + `obra_id` + "`" + `\n)"
		]`), &collection.Indexes)

		// add
		new_user_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "svk0wuem",
			"name": "user_id",
			"type": "relation",
			"required": true,
			"unique": false,
			"options": {
				"collectionId": "_pb_users_auth_",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": null,
				"displayFields": []
			}
		}`), new_user_id)
		collection.Schema.AddField(new_user_id)

		// update
		edit_obra_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "86ro1qoo",
			"name": "obra_id",
			"type": "text",
			"required": true,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), edit_obra_id)
		collection.Schema.AddField(edit_obra_id)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("6r8ux9luhyejwyf")
		if err != nil {
			return err
		}

		json.Unmarshal([]byte(`[
			"CREATE INDEX ` + "`" + `idx_8czuntv` + "`" + ` ON ` + "`" + `users_obras` + "`" + ` (` + "`" + `favorited_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_4bZz28i` + "`" + ` ON ` + "`" + `users_obras` + "`" + ` (` + "`" + `excluded_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_ob5OKBz` + "`" + ` ON ` + "`" + `users_obras` + "`" + ` (` + "`" + `visited_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_10Ig2Kd` + "`" + ` ON ` + "`" + `users_obras` + "`" + ` (` + "`" + `status` + "`" + `)"
		]`), &collection.Indexes)

		// remove
		collection.Schema.RemoveField("svk0wuem")

		// update
		edit_obra_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "86ro1qoo",
			"name": "obra_id",
			"type": "text",
			"required": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), edit_obra_id)
		collection.Schema.AddField(edit_obra_id)

		return dao.SaveCollection(collection)
	})
}

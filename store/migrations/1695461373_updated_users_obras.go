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
			"CREATE INDEX ` + "`" + `idx_10Ig2Kd` + "`" + ` ON ` + "`" + `users_obras` + "`" + ` (` + "`" + `status` + "`" + `)"
		]`), &collection.Indexes)

		// remove
		collection.Schema.RemoveField("z6rs3kp1")

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("6r8ux9luhyejwyf")
		if err != nil {
			return err
		}

		json.Unmarshal([]byte(`[
			"CREATE UNIQUE INDEX ` + "`" + `idx_0ZGYJ4Q` + "`" + ` ON ` + "`" + `users_obras` + "`" + ` (\n  ` + "`" + `user_id` + "`" + `,\n  ` + "`" + `obra_id` + "`" + `\n)",
			"CREATE INDEX ` + "`" + `idx_8czuntv` + "`" + ` ON ` + "`" + `users_obras` + "`" + ` (` + "`" + `favorited_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_4bZz28i` + "`" + ` ON ` + "`" + `users_obras` + "`" + ` (` + "`" + `excluded_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_ob5OKBz` + "`" + ` ON ` + "`" + `users_obras` + "`" + ` (` + "`" + `visited_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_10Ig2Kd` + "`" + ` ON ` + "`" + `users_obras` + "`" + ` (` + "`" + `status` + "`" + `)"
		]`), &collection.Indexes)

		// add
		del_user_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "z6rs3kp1",
			"name": "user_id",
			"type": "text",
			"required": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), del_user_id)
		collection.Schema.AddField(del_user_id)

		return dao.SaveCollection(collection)
	})
}

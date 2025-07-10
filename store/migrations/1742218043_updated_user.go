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
			"CREATE INDEX ` + "`" + `idx_jOLoFsU` + "`" + ` ON ` + "`" + `user` + "`" + ` (` + "`" + `legacy_id` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_FVIuOFp` + "`" + ` ON ` + "`" + `user` + "`" + ` (` + "`" + `team_id` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_T2SXCNf` + "`" + ` ON ` + "`" + `user` + "`" + ` (` + "`" + `manager` + "`" + `)"
		]`), &collection.Indexes)

		// add
		new_manager := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "bd5mq0ck",
			"name": "manager",
			"type": "bool",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_manager)
		collection.Schema.AddField(new_manager)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		json.Unmarshal([]byte(`[
			"CREATE INDEX ` + "`" + `idx_jOLoFsU` + "`" + ` ON ` + "`" + `user` + "`" + ` (` + "`" + `legacy_id` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_FVIuOFp` + "`" + ` ON ` + "`" + `user` + "`" + ` (` + "`" + `team_id` + "`" + `)"
		]`), &collection.Indexes)

		// remove
		collection.Schema.RemoveField("bd5mq0ck")

		return dao.SaveCollection(collection)
	})
}

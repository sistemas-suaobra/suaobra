package migrations

import (
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		collection.Name = "user"

		json.Unmarshal([]byte(`[
			"CREATE UNIQUE INDEX ` + "`" + `idx_jOLoFsU` + "`" + ` ON ` + "`" + `user` + "`" + ` (` + "`" + `legacy_id` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_beDGb9Y` + "`" + ` ON ` + "`" + `user` + "`" + ` (` + "`" + `team_id` + "`" + `)"
		]`), &collection.Indexes)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		collection.Name = "users"

		json.Unmarshal([]byte(`[
			"CREATE UNIQUE INDEX ` + "`" + `idx_jOLoFsU` + "`" + ` ON ` + "`" + `users` + "`" + ` (` + "`" + `legacy_id` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_beDGb9Y` + "`" + ` ON ` + "`" + `users` + "`" + ` (` + "`" + `team_id` + "`" + `)"
		]`), &collection.Indexes)

		return dao.SaveCollection(collection)
	})
}

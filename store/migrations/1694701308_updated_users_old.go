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

		collection, err := dao.FindCollectionByNameOrId("hqvr9jtjmjjp1ps")
		if err != nil {
			return err
		}

		collection.Name = "users_legacy"

		json.Unmarshal([]byte(`[
			"CREATE UNIQUE INDEX ` + "`" + `idx_fGzp6ue` + "`" + ` ON ` + "`" + `users_legacy` + "`" + ` (` + "`" + `email` + "`" + `)"
		]`), &collection.Indexes)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("hqvr9jtjmjjp1ps")
		if err != nil {
			return err
		}

		collection.Name = "users_old"

		json.Unmarshal([]byte(`[
			"CREATE UNIQUE INDEX ` + "`" + `idx_fGzp6ue` + "`" + ` ON ` + "`" + `users_old` + "`" + ` (` + "`" + `email` + "`" + `)"
		]`), &collection.Indexes)

		return dao.SaveCollection(collection)
	})
}

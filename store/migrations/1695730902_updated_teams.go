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

		collection, err := dao.FindCollectionByNameOrId("r12uiife4c4e8zr")
		if err != nil {
			return err
		}

		collection.Name = "team"

		json.Unmarshal([]byte(`[
			"CREATE INDEX ` + "`" + `idx_LjSxmxc` + "`" + ` ON ` + "`" + `team` + "`" + ` (` + "`" + `owner` + "`" + `)"
		]`), &collection.Indexes)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("r12uiife4c4e8zr")
		if err != nil {
			return err
		}

		collection.Name = "teams"

		json.Unmarshal([]byte(`[
			"CREATE INDEX ` + "`" + `idx_LjSxmxc` + "`" + ` ON ` + "`" + `teams` + "`" + ` (` + "`" + `owner` + "`" + `)"
		]`), &collection.Indexes)

		return dao.SaveCollection(collection)
	})
}

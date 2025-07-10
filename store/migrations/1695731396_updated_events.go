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

		collection, err := dao.FindCollectionByNameOrId("89ju1qkr9p4o35m")
		if err != nil {
			return err
		}

		collection.Name = "rudderstack"

		json.Unmarshal([]byte(`[
			"CREATE INDEX ` + "`" + `idx_5Mpur7t` + "`" + ` ON ` + "`" + `rudderstack` + "`" + ` (` + "`" + `anonymousId` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_WTl97PJ` + "`" + ` ON ` + "`" + `rudderstack` + "`" + ` (\n  ` + "`" + `type` + "`" + `,\n  ` + "`" + `event` + "`" + `\n)"
		]`), &collection.Indexes)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("89ju1qkr9p4o35m")
		if err != nil {
			return err
		}

		collection.Name = "events"

		json.Unmarshal([]byte(`[
			"CREATE INDEX ` + "`" + `idx_5Mpur7t` + "`" + ` ON ` + "`" + `events` + "`" + ` (` + "`" + `anonymousId` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_WTl97PJ` + "`" + ` ON ` + "`" + `events` + "`" + ` (\n  ` + "`" + `type` + "`" + `,\n  ` + "`" + `event` + "`" + `\n)"
		]`), &collection.Indexes)

		return dao.SaveCollection(collection)
	})
}

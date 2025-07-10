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

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("qzdr0f0p4ddqfqe")
		if err != nil {
			return err
		}

		collection.Name = "team_obras"

		json.Unmarshal([]byte(`[
			"CREATE UNIQUE INDEX ` + "`" + `idx_al1JxBl` + "`" + ` ON ` + "`" + `team_obras` + "`" + ` (\n  ` + "`" + `team_id` + "`" + `,\n  ` + "`" + `obra_id` + "`" + `\n)",
			"CREATE INDEX ` + "`" + `idx_l70ogsl` + "`" + ` ON ` + "`" + `team_obras` + "`" + ` (` + "`" + `visited_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_RqdgzUy` + "`" + ` ON ` + "`" + `team_obras` + "`" + ` (` + "`" + `favorited_at` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_RvZyJvN` + "`" + ` ON ` + "`" + `team_obras` + "`" + ` (` + "`" + `excluded_at` + "`" + `)"
		]`), &collection.Indexes)

		return dao.SaveCollection(collection)
	})
}

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

		collection, err := dao.FindCollectionByNameOrId("_collection_cydQmBK5ZeGpLfE")
		if err != nil {
			return err
		}

		json.Unmarshal([]byte(`[
			"CREATE UNIQUE INDEX ` + "`" + `idx_JX8RWC0` + "`" + ` ON ` + "`" + `obra_note` + "`" + ` (` + "`" + `obra_id` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_5iAymaC` + "`" + ` ON ` + "`" + `obra_note` + "`" + ` (` + "`" + `user_id` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_7IE7GnX` + "`" + ` ON ` + "`" + `obra_note` + "`" + ` (` + "`" + `info_obra_id` + "`" + `)"
		]`), &collection.Indexes)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("_collection_cydQmBK5ZeGpLfE")
		if err != nil {
			return err
		}

		json.Unmarshal([]byte(`[
			"CREATE INDEX ` + "`" + `idx_JX8RWC0` + "`" + ` ON ` + "`" + `obra_note` + "`" + ` (` + "`" + `obra_id` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_5iAymaC` + "`" + ` ON ` + "`" + `obra_note` + "`" + ` (` + "`" + `user_id` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_7IE7GnX` + "`" + ` ON ` + "`" + `obra_note` + "`" + ` (` + "`" + `info_obra_id` + "`" + `)"
		]`), &collection.Indexes)

		return dao.SaveCollection(collection)
	})
}

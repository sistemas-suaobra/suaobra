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

		collection, err := dao.FindCollectionByNameOrId("hqvr9jtjmjjp1ps")
		if err != nil {
			return err
		}

		json.Unmarshal([]byte(`[
			"CREATE UNIQUE INDEX ` + "`" + `idx_fGzp6ue` + "`" + ` ON ` + "`" + `users_legacy` + "`" + ` (` + "`" + `email` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_vLulWsq` + "`" + ` ON ` + "`" + `users_legacy` + "`" + ` (` + "`" + `team_id` + "`" + `)"
		]`), &collection.Indexes)

		// add
		new_team_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "wwlzc0ja",
			"name": "team_id",
			"type": "relation",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "r12uiife4c4e8zr",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), new_team_id)
		collection.Schema.AddField(new_team_id)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("hqvr9jtjmjjp1ps")
		if err != nil {
			return err
		}

		json.Unmarshal([]byte(`[
			"CREATE UNIQUE INDEX ` + "`" + `idx_fGzp6ue` + "`" + ` ON ` + "`" + `users_legacy` + "`" + ` (` + "`" + `email` + "`" + `)"
		]`), &collection.Indexes)

		// remove
		collection.Schema.RemoveField("wwlzc0ja")

		return dao.SaveCollection(collection)
	})
}

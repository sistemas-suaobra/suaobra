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
			"CREATE UNIQUE INDEX ` + "`" + `idx_jOLoFsU` + "`" + ` ON ` + "`" + `user` + "`" + ` (` + "`" + `legacy_id` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_FVIuOFp` + "`" + ` ON ` + "`" + `user` + "`" + ` (` + "`" + `team_id` + "`" + `)"
		]`), &collection.Indexes)

		// remove
		collection.Schema.RemoveField("bhv2wgie")

		// add
		new_team_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "0sugomuk",
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

		collection, err := dao.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}

		json.Unmarshal([]byte(`[
			"CREATE UNIQUE INDEX ` + "`" + `idx_jOLoFsU` + "`" + ` ON ` + "`" + `user` + "`" + ` (` + "`" + `legacy_id` + "`" + `)",
			"CREATE INDEX ` + "`" + `idx_beDGb9Y` + "`" + ` ON ` + "`" + `user` + "`" + ` (` + "`" + `team_id` + "`" + `)"
		]`), &collection.Indexes)

		// add
		del_team_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "bhv2wgie",
			"name": "team_id",
			"type": "text",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), del_team_id)
		collection.Schema.AddField(del_team_id)

		// remove
		collection.Schema.RemoveField("0sugomuk")

		return dao.SaveCollection(collection)
	})
}

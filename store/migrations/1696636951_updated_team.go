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

		collection, err := dao.FindCollectionByNameOrId("r12uiife4c4e8zr")
		if err != nil {
			return err
		}

		json.Unmarshal([]byte(`[
			"CREATE INDEX ` + "`" + `idx_LjSxmxc` + "`" + ` ON ` + "`" + `team` + "`" + ` (` + "`" + `owner_id` + "`" + `)"
		]`), &collection.Indexes)

		// update
		edit_owner_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "qbtojp7p",
			"name": "owner_id",
			"type": "relation",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "_pb_users_auth_",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": []
			}
		}`), edit_owner_id)
		collection.Schema.AddField(edit_owner_id)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("r12uiife4c4e8zr")
		if err != nil {
			return err
		}

		json.Unmarshal([]byte(`[
			"CREATE INDEX ` + "`" + `idx_LjSxmxc` + "`" + ` ON ` + "`" + `team` + "`" + ` (` + "`" + `owner` + "`" + `)"
		]`), &collection.Indexes)

		// update
		edit_owner_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "qbtojp7p",
			"name": "owner",
			"type": "relation",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "_pb_users_auth_",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": []
			}
		}`), edit_owner_id)
		collection.Schema.AddField(edit_owner_id)

		return dao.SaveCollection(collection)
	})
}

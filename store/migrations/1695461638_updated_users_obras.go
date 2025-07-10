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

		collection, err := dao.FindCollectionByNameOrId("6r8ux9luhyejwyf")
		if err != nil {
			return err
		}

		// update
		edit_user_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "svk0wuem",
			"name": "user_id",
			"type": "relation",
			"required": true,
			"unique": false,
			"options": {
				"collectionId": "_pb_users_auth_",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": []
			}
		}`), edit_user_id)
		collection.Schema.AddField(edit_user_id)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("6r8ux9luhyejwyf")
		if err != nil {
			return err
		}

		// update
		edit_user_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "svk0wuem",
			"name": "user_id",
			"type": "relation",
			"required": true,
			"unique": false,
			"options": {
				"collectionId": "_pb_users_auth_",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": null,
				"displayFields": []
			}
		}`), edit_user_id)
		collection.Schema.AddField(edit_user_id)

		return dao.SaveCollection(collection)
	})
}

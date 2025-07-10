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

		collection, err := dao.FindCollectionByNameOrId("xe9w10fta5tu89v")
		if err != nil {
			return err
		}

		// update
		edit_team_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "dnwfvyw0",
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
		}`), edit_team_id)
		collection.Schema.AddField(edit_team_id)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("xe9w10fta5tu89v")
		if err != nil {
			return err
		}

		// update
		edit_team_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "dnwfvyw0",
			"name": "team_id",
			"type": "relation",
			"required": false,
			"presentable": true,
			"unique": false,
			"options": {
				"collectionId": "r12uiife4c4e8zr",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), edit_team_id)
		collection.Schema.AddField(edit_team_id)

		return dao.SaveCollection(collection)
	})
}

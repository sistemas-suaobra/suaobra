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

		collection, err := dao.FindCollectionByNameOrId("qzdr0f0p4ddqfqe")
		if err != nil {
			return err
		}

		// update
		edit_team_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "5mvmmvku",
			"name": "team_id",
			"type": "relation",
			"required": true,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "r12uiife4c4e8zr",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": []
			}
		}`), edit_team_id)
		collection.Schema.AddField(edit_team_id)

		// update
		edit_obra_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "esr9k2dm",
			"name": "obra_id",
			"type": "text",
			"required": true,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), edit_obra_id)
		collection.Schema.AddField(edit_obra_id)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("qzdr0f0p4ddqfqe")
		if err != nil {
			return err
		}

		// update
		edit_team_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "5mvmmvku",
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
				"displayFields": []
			}
		}`), edit_team_id)
		collection.Schema.AddField(edit_team_id)

		// update
		edit_obra_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "esr9k2dm",
			"name": "obra_id",
			"type": "text",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), edit_obra_id)
		collection.Schema.AddField(edit_obra_id)

		return dao.SaveCollection(collection)
	})
}

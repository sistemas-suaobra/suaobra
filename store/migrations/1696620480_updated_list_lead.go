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

		collection, err := dao.FindCollectionByNameOrId("m7bh3r7viqvvqrb")
		if err != nil {
			return err
		}

		// update
		edit_list_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "6yi8qpne",
			"name": "list_id",
			"type": "relation",
			"required": true,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "xe9w10fta5tu89v",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), edit_list_id)
		collection.Schema.AddField(edit_list_id)

		// update
		edit_lead_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "b6m9qmbx",
			"name": "lead_id",
			"type": "relation",
			"required": true,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "qzdr0f0p4ddqfqe",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), edit_lead_id)
		collection.Schema.AddField(edit_lead_id)

		// update
		edit_stage_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "cjexxcbt",
			"name": "stage_id",
			"type": "relation",
			"required": true,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "p980c5513p13sfl",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), edit_stage_id)
		collection.Schema.AddField(edit_stage_id)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("m7bh3r7viqvvqrb")
		if err != nil {
			return err
		}

		// update
		edit_list_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "6yi8qpne",
			"name": "list_id",
			"type": "relation",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "xe9w10fta5tu89v",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), edit_list_id)
		collection.Schema.AddField(edit_list_id)

		// update
		edit_lead_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "b6m9qmbx",
			"name": "lead_id",
			"type": "relation",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "qzdr0f0p4ddqfqe",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), edit_lead_id)
		collection.Schema.AddField(edit_lead_id)

		// update
		edit_stage_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "cjexxcbt",
			"name": "stage_id",
			"type": "relation",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"collectionId": "p980c5513p13sfl",
				"cascadeDelete": false,
				"minSelect": null,
				"maxSelect": 1,
				"displayFields": null
			}
		}`), edit_stage_id)
		collection.Schema.AddField(edit_stage_id)

		return dao.SaveCollection(collection)
	})
}

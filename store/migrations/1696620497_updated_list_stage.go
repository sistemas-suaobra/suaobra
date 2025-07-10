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

		collection, err := dao.FindCollectionByNameOrId("p980c5513p13sfl")
		if err != nil {
			return err
		}

		// update
		edit_list_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "igclyogo",
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
		edit_name := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "dq4jyfk2",
			"name": "name",
			"type": "text",
			"required": true,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), edit_name)
		collection.Schema.AddField(edit_name)

		// update
		edit_order := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "lugwt2hb",
			"name": "order",
			"type": "number",
			"required": true,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"noDecimal": true
			}
		}`), edit_order)
		collection.Schema.AddField(edit_order)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("p980c5513p13sfl")
		if err != nil {
			return err
		}

		// update
		edit_list_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "igclyogo",
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
		edit_name := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "dq4jyfk2",
			"name": "name",
			"type": "text",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), edit_name)
		collection.Schema.AddField(edit_name)

		// update
		edit_order := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "lugwt2hb",
			"name": "order",
			"type": "number",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"noDecimal": false
			}
		}`), edit_order)
		collection.Schema.AddField(edit_order)

		return dao.SaveCollection(collection)
	})
}

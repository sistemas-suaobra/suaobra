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

		collection, err := dao.FindCollectionByNameOrId("8q056kyr077k5eu")
		if err != nil {
			return err
		}

		collection.Name = "rudderstack_vw"

		options := map[string]any{}
		json.Unmarshal([]byte(`{
			"query": "select id, originaltimestamp, type, event, json_extract(properties, '$.user.email') as userEmail, properties from rudderstack"
		}`), &options)
		collection.SetOptions(options)

		// remove
		collection.Schema.RemoveField("t01lehvs")

		// remove
		collection.Schema.RemoveField("h1wtqpdn")

		// remove
		collection.Schema.RemoveField("8lbq47ze")

		// remove
		collection.Schema.RemoveField("lkgyxwl4")

		// remove
		collection.Schema.RemoveField("beafbuit")

		// add
		new_originaltimestamp := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "3hpnz8ur",
			"name": "originaltimestamp",
			"type": "date",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": "",
				"max": ""
			}
		}`), new_originaltimestamp)
		collection.Schema.AddField(new_originaltimestamp)

		// add
		new_type := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "i120wkj2",
			"name": "type",
			"type": "text",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), new_type)
		collection.Schema.AddField(new_type)

		// add
		new_event := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "zvhycyjl",
			"name": "event",
			"type": "text",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), new_event)
		collection.Schema.AddField(new_event)

		// add
		new_userEmail := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "js8nikbv",
			"name": "userEmail",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_userEmail)
		collection.Schema.AddField(new_userEmail)

		// add
		new_properties := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "5ydce1rt",
			"name": "properties",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_properties)
		collection.Schema.AddField(new_properties)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("8q056kyr077k5eu")
		if err != nil {
			return err
		}

		collection.Name = "events_vw"

		options := map[string]any{}
		json.Unmarshal([]byte(`{
			"query": "select id, originaltimestamp, type, event, json_extract(properties, '$.user.email') as userEmail, properties from events"
		}`), &options)
		collection.SetOptions(options)

		// add
		del_originaltimestamp := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "t01lehvs",
			"name": "originaltimestamp",
			"type": "date",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": "",
				"max": ""
			}
		}`), del_originaltimestamp)
		collection.Schema.AddField(del_originaltimestamp)

		// add
		del_type := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "h1wtqpdn",
			"name": "type",
			"type": "text",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), del_type)
		collection.Schema.AddField(del_type)

		// add
		del_event := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "8lbq47ze",
			"name": "event",
			"type": "text",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), del_event)
		collection.Schema.AddField(del_event)

		// add
		del_userEmail := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "lkgyxwl4",
			"name": "userEmail",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), del_userEmail)
		collection.Schema.AddField(del_userEmail)

		// add
		del_properties := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "beafbuit",
			"name": "properties",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), del_properties)
		collection.Schema.AddField(del_properties)

		// remove
		collection.Schema.RemoveField("3hpnz8ur")

		// remove
		collection.Schema.RemoveField("i120wkj2")

		// remove
		collection.Schema.RemoveField("zvhycyjl")

		// remove
		collection.Schema.RemoveField("js8nikbv")

		// remove
		collection.Schema.RemoveField("5ydce1rt")

		return dao.SaveCollection(collection)
	})
}

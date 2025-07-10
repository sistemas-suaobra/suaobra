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

		options := map[string]any{}
		json.Unmarshal([]byte(`{
			"query": "select id, originaltimestamp, type, event, json_extract(properties, '$.user.email') as userEmail, properties from events"
		}`), &options)
		collection.SetOptions(options)

		// remove
		collection.Schema.RemoveField("hbfeqvk9")

		// remove
		collection.Schema.RemoveField("fsqlvuqe")

		// remove
		collection.Schema.RemoveField("ininu5ar")

		// remove
		collection.Schema.RemoveField("6u6rt2is")

		// add
		new_originaltimestamp := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "t01lehvs",
			"name": "originaltimestamp",
			"type": "date",
			"required": false,
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
			"id": "h1wtqpdn",
			"name": "type",
			"type": "text",
			"required": false,
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
			"id": "8lbq47ze",
			"name": "event",
			"type": "text",
			"required": false,
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
			"id": "lkgyxwl4",
			"name": "userEmail",
			"type": "json",
			"required": false,
			"unique": false,
			"options": {}
		}`), new_userEmail)
		collection.Schema.AddField(new_userEmail)

		// add
		new_properties := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "beafbuit",
			"name": "properties",
			"type": "json",
			"required": false,
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

		options := map[string]any{}
		json.Unmarshal([]byte(`{
			"query": "select id, originaltimestamp , type, event, properties from events"
		}`), &options)
		collection.SetOptions(options)

		// add
		del_originaltimestamp := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "hbfeqvk9",
			"name": "originaltimestamp",
			"type": "json",
			"required": false,
			"unique": false,
			"options": {}
		}`), del_originaltimestamp)
		collection.Schema.AddField(del_originaltimestamp)

		// add
		del_type := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "fsqlvuqe",
			"name": "type",
			"type": "date",
			"required": false,
			"unique": false,
			"options": {
				"min": "",
				"max": ""
			}
		}`), del_type)
		collection.Schema.AddField(del_type)

		// add
		del_event := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "ininu5ar",
			"name": "event",
			"type": "text",
			"required": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), del_event)
		collection.Schema.AddField(del_event)

		// add
		del_properties := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "6u6rt2is",
			"name": "properties",
			"type": "json",
			"required": false,
			"unique": false,
			"options": {}
		}`), del_properties)
		collection.Schema.AddField(del_properties)

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

		return dao.SaveCollection(collection)
	})
}

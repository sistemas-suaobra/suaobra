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

		collection, err := dao.FindCollectionByNameOrId("_collection_tnUtpZr6ec9Yjxi")
		if err != nil {
			return err
		}

		options := map[string]any{}
		json.Unmarshal([]byte(`{
			"query": "select\n  user.id,\n  user.email,\n  coalesce(nullif(user.name, ''), json_extract(user.properties, '$.name'), '-') as nome,\n  coalesce(json_extract(user.properties, '$.phone'), '-') as telephone,\n  created,\n  updated\nfrom user\nwhere verified\norder by created desc"
		}`), &options)
		collection.SetOptions(options)

		// remove
		collection.Schema.RemoveField("kl2msfva")

		// remove
		collection.Schema.RemoveField("h30ne8tz")

		// remove
		collection.Schema.RemoveField("bqbg34zz")

		// add
		new_email := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "a59uxyq3",
			"name": "email",
			"type": "email",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"exceptDomains": null,
				"onlyDomains": null
			}
		}`), new_email)
		collection.Schema.AddField(new_email)

		// add
		new_nome := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "pat3tgm4",
			"name": "nome",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_nome)
		collection.Schema.AddField(new_nome)

		// add
		new_telephone := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "kubbevpj",
			"name": "telephone",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_telephone)
		collection.Schema.AddField(new_telephone)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("_collection_tnUtpZr6ec9Yjxi")
		if err != nil {
			return err
		}

		options := map[string]any{}
		json.Unmarshal([]byte(`{
			"query": "select\n  user.id,\n  user.email,\n  coalesce(nullif(user.name, ''), json_extract(user.properties, '$.name'), '-') as nome,\n  coalesce(json_extract(user.properties, '$.phone'), '-') as telephone,\n  created,\n  updated\nfrom user\norder by created desc"
		}`), &options)
		collection.SetOptions(options)

		// add
		del_email := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "kl2msfva",
			"name": "email",
			"type": "email",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"exceptDomains": null,
				"onlyDomains": null
			}
		}`), del_email)
		collection.Schema.AddField(del_email)

		// add
		del_nome := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "h30ne8tz",
			"name": "nome",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), del_nome)
		collection.Schema.AddField(del_nome)

		// add
		del_telephone := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "bqbg34zz",
			"name": "telephone",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), del_telephone)
		collection.Schema.AddField(del_telephone)

		// remove
		collection.Schema.RemoveField("a59uxyq3")

		// remove
		collection.Schema.RemoveField("pat3tgm4")

		// remove
		collection.Schema.RemoveField("kubbevpj")

		return dao.SaveCollection(collection)
	})
}

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

		collection, err := dao.FindCollectionByNameOrId("_collection_1BPiQsdzfoeHAOH")
		if err != nil {
			return err
		}

		options := map[string]any{}
		json.Unmarshal([]byte(`{
			"query": "select\n  date(data_evento) as date,\n  nullif(usario_email, '') email,\n  count(1) eventos,\n  sum(case when nome_evento like 'login-%' then 1 else 0 end) login_eventos,\n  sum(case when nome_evento like 'venda-mais-%' then 1 else 0 end) venda_mais_eventos,\n  sum(case when nome_evento like 'obras-plus-%' then 1 else 0 end) obras_plus_eventos,\n  sum(case when tipo_evento = 'track' then 1 else 0 end) track_eventos,\n  sum(case when tipo_evento = 'page' then 1 else 0 end) page_eventos,\n  sum(case when tipo_evento = 'identify' then 1 else 0 end) identity_eventos,\n  substr(lower(hex(randomblob(10))), 1, 15) as id\nfrom vw_eventos\ngroup by 1, 2\norder by 1 desc"
		}`), &options)
		collection.SetOptions(options)

		// remove
		collection.Schema.RemoveField("wfb8nobz")

		// remove
		collection.Schema.RemoveField("xsowve8l")

		// remove
		collection.Schema.RemoveField("f8r76kxd")

		// remove
		collection.Schema.RemoveField("smuombl0")

		// remove
		collection.Schema.RemoveField("ethzlqvx")

		// remove
		collection.Schema.RemoveField("knqi4zxe")

		// remove
		collection.Schema.RemoveField("xltxbayg")

		// remove
		collection.Schema.RemoveField("0uwpufwb")

		// remove
		collection.Schema.RemoveField("zzak0e8m")

		// add
		new_date := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "xnsvgxkc",
			"name": "date",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_date)
		collection.Schema.AddField(new_date)

		// add
		new_email := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "36lycmjm",
			"name": "email",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_email)
		collection.Schema.AddField(new_email)

		// add
		new_eventos := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "jz4zk5p2",
			"name": "eventos",
			"type": "number",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"noDecimal": false
			}
		}`), new_eventos)
		collection.Schema.AddField(new_eventos)

		// add
		new_login_eventos := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "tb3pdu3h",
			"name": "login_eventos",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_login_eventos)
		collection.Schema.AddField(new_login_eventos)

		// add
		new_venda_mais_eventos := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "cptluuav",
			"name": "venda_mais_eventos",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_venda_mais_eventos)
		collection.Schema.AddField(new_venda_mais_eventos)

		// add
		new_obras_plus_eventos := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "5isyrey7",
			"name": "obras_plus_eventos",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_obras_plus_eventos)
		collection.Schema.AddField(new_obras_plus_eventos)

		// add
		new_track_eventos := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "0zbb2ndm",
			"name": "track_eventos",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_track_eventos)
		collection.Schema.AddField(new_track_eventos)

		// add
		new_page_eventos := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "xnght2pc",
			"name": "page_eventos",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_page_eventos)
		collection.Schema.AddField(new_page_eventos)

		// add
		new_identity_eventos := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "pfh547mm",
			"name": "identity_eventos",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_identity_eventos)
		collection.Schema.AddField(new_identity_eventos)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("_collection_1BPiQsdzfoeHAOH")
		if err != nil {
			return err
		}

		options := map[string]any{}
		json.Unmarshal([]byte(`{
			"query": "select\n  date(data_evento) as date,\n  nullif(usario_email, '') email,\n  count(1) eventos,\n  sum(case when nome_evento like 'login-%' then 1 else 0 end) login_eventos,\n  sum(case when nome_evento like 'venda-mais-%' then 1 else 0 end) venda_mais_eventos,\n  sum(case when nome_evento like 'obras-plus-%' then 1 else 0 end) obras_plus_eventos,\n  sum(case when tipo_evento = 'track' then 1 else 0 end) track_eventos,\n  sum(case when tipo_evento = 'page' then 1 else 0 end) page_eventos,\n  sum(case when tipo_evento = 'identify' then 1 else 0 end) identity_eventos,\n  '-' as id\nfrom vw_eventos\ngroup by 1, 2\norder by 1 desc"
		}`), &options)
		collection.SetOptions(options)

		// add
		del_date := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "wfb8nobz",
			"name": "date",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), del_date)
		collection.Schema.AddField(del_date)

		// add
		del_email := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "xsowve8l",
			"name": "email",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), del_email)
		collection.Schema.AddField(del_email)

		// add
		del_eventos := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "f8r76kxd",
			"name": "eventos",
			"type": "number",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"noDecimal": false
			}
		}`), del_eventos)
		collection.Schema.AddField(del_eventos)

		// add
		del_login_eventos := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "smuombl0",
			"name": "login_eventos",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), del_login_eventos)
		collection.Schema.AddField(del_login_eventos)

		// add
		del_venda_mais_eventos := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "ethzlqvx",
			"name": "venda_mais_eventos",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), del_venda_mais_eventos)
		collection.Schema.AddField(del_venda_mais_eventos)

		// add
		del_obras_plus_eventos := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "knqi4zxe",
			"name": "obras_plus_eventos",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), del_obras_plus_eventos)
		collection.Schema.AddField(del_obras_plus_eventos)

		// add
		del_track_eventos := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "xltxbayg",
			"name": "track_eventos",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), del_track_eventos)
		collection.Schema.AddField(del_track_eventos)

		// add
		del_page_eventos := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "0uwpufwb",
			"name": "page_eventos",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), del_page_eventos)
		collection.Schema.AddField(del_page_eventos)

		// add
		del_identity_eventos := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "zzak0e8m",
			"name": "identity_eventos",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), del_identity_eventos)
		collection.Schema.AddField(del_identity_eventos)

		// remove
		collection.Schema.RemoveField("xnsvgxkc")

		// remove
		collection.Schema.RemoveField("36lycmjm")

		// remove
		collection.Schema.RemoveField("jz4zk5p2")

		// remove
		collection.Schema.RemoveField("tb3pdu3h")

		// remove
		collection.Schema.RemoveField("cptluuav")

		// remove
		collection.Schema.RemoveField("5isyrey7")

		// remove
		collection.Schema.RemoveField("0zbb2ndm")

		// remove
		collection.Schema.RemoveField("xnght2pc")

		// remove
		collection.Schema.RemoveField("pfh547mm")

		return dao.SaveCollection(collection)
	})
}

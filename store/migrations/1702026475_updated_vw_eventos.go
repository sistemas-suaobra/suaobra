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

		collection, err := dao.FindCollectionByNameOrId("_collection_fgFv1hfA4mZ947D")
		if err != nil {
			return err
		}

		options := map[string]any{}
		json.Unmarshal([]byte(`{
			"query": "select\n  rudderstack.id,\n  rudderstack.type as tipo_evento,\n  coalesce(nullif(rudderstack.event, ''), rudderstack.type) as nome_evento,\n  coalesce(json_extract(rudderstack.properties, '$.user.email'), json_extract(rudderstack.context, '$.traits.user.email'), json_extract(rudderstack.properties, '$.email'), user.email) as usario_email,\n  coalesce(json_extract(rudderstack.properties, '$.userId'), json_extract(rudderstack.properties, '$.user.id')) as usario_id,\n  datetime(julianday(rudderstack.originaltimestamp) - 0.125) as data_evento, -- BRT\n  json_extract(rudderstack.properties, '$.state') as uf,\n  json_extract(rudderstack.properties, '$.city') as cidade,\n  json_extract(rudderstack.properties, '$.page_number') as pagina,\n  json_extract(rudderstack.properties, '$.size') as tamanho,\n  json_extract(rudderstack.properties, '$.url') as link,\n  json_extract(rudderstack.properties, '$.title') as titulo,\n  rudderstack.properties as dados\nfrom rudderstack\nleft join user on json_extract(rudderstack.properties, '$.userId') = user.id\norder by data_evento desc"
		}`), &options)
		collection.SetOptions(options)

		// remove
		collection.Schema.RemoveField("wcz0okfo")

		// remove
		collection.Schema.RemoveField("vxn3vybg")

		// remove
		collection.Schema.RemoveField("aphjfrou")

		// remove
		collection.Schema.RemoveField("bwsvnish")

		// remove
		collection.Schema.RemoveField("ezejigh1")

		// remove
		collection.Schema.RemoveField("8rli2whk")

		// remove
		collection.Schema.RemoveField("4h3jg5fx")

		// remove
		collection.Schema.RemoveField("ivusadta")

		// remove
		collection.Schema.RemoveField("17163ehc")

		// remove
		collection.Schema.RemoveField("ol1xjaqy")

		// remove
		collection.Schema.RemoveField("vmibhzr0")

		// remove
		collection.Schema.RemoveField("rvibbrzg")

		// add
		new_tipo_evento := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "gkxpw2sd",
			"name": "tipo_evento",
			"type": "text",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), new_tipo_evento)
		collection.Schema.AddField(new_tipo_evento)

		// add
		new_nome_evento := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "hw5wouwe",
			"name": "nome_evento",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_nome_evento)
		collection.Schema.AddField(new_nome_evento)

		// add
		new_usario_email := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "vc63opzx",
			"name": "usario_email",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_usario_email)
		collection.Schema.AddField(new_usario_email)

		// add
		new_usario_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "l7mhmx4v",
			"name": "usario_id",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_usario_id)
		collection.Schema.AddField(new_usario_id)

		// add
		new_data_evento := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "lpsfputh",
			"name": "data_evento",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_data_evento)
		collection.Schema.AddField(new_data_evento)

		// add
		new_uf := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "acywbctf",
			"name": "uf",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_uf)
		collection.Schema.AddField(new_uf)

		// add
		new_cidade := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "mo9m5hyz",
			"name": "cidade",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_cidade)
		collection.Schema.AddField(new_cidade)

		// add
		new_pagina := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "4ohn2ycj",
			"name": "pagina",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_pagina)
		collection.Schema.AddField(new_pagina)

		// add
		new_tamanho := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "cvv5hdgm",
			"name": "tamanho",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_tamanho)
		collection.Schema.AddField(new_tamanho)

		// add
		new_link := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "4anr0dhp",
			"name": "link",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_link)
		collection.Schema.AddField(new_link)

		// add
		new_titulo := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "2h1pumpm",
			"name": "titulo",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_titulo)
		collection.Schema.AddField(new_titulo)

		// add
		new_dados := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "qruxthyt",
			"name": "dados",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), new_dados)
		collection.Schema.AddField(new_dados)

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("_collection_fgFv1hfA4mZ947D")
		if err != nil {
			return err
		}

		options := map[string]any{}
		json.Unmarshal([]byte(`{
			"query": "select\n  rudderstack.id,\n  rudderstack.type as tipo_evento,\n  coalesce(nullif(rudderstack.event, ''), rudderstack.type) as nome_evento,\n  coalesce(json_extract(rudderstack.properties, '$.user .email'), json_extract(rudderstack.context, '$.traits.user .email'), json_extract(rudderstack.properties, '$.email'), user.email) as usario_email,\n  coalesce(json_extract(rudderstack.properties, '$.userId'), json_extract(rudderstack.properties, '$.user .id')) as usario_id,\n  datetime(julianday(rudderstack.originaltimestamp) - 0.125) as data_evento, -- BRT\n  json_extract(rudderstack.properties, '$.state') as uf,\n  json_extract(rudderstack.properties, '$.city') as cidade,\n  json_extract(rudderstack.properties, '$.page_number') as pagina,\n  json_extract(rudderstack.properties, '$.size') as tamanho,\n  json_extract(rudderstack.properties, '$.url') as link,\n  json_extract(rudderstack.properties, '$.title') as titulo,\n  rudderstack.properties as dados\nfrom rudderstack\nleft join user on json_extract(rudderstack.properties, '$.userId') = user.id\norder by data_evento desc"
		}`), &options)
		collection.SetOptions(options)

		// add
		del_tipo_evento := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "wcz0okfo",
			"name": "tipo_evento",
			"type": "text",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {
				"min": null,
				"max": null,
				"pattern": ""
			}
		}`), del_tipo_evento)
		collection.Schema.AddField(del_tipo_evento)

		// add
		del_nome_evento := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "vxn3vybg",
			"name": "nome_evento",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), del_nome_evento)
		collection.Schema.AddField(del_nome_evento)

		// add
		del_usario_email := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "aphjfrou",
			"name": "usario_email",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), del_usario_email)
		collection.Schema.AddField(del_usario_email)

		// add
		del_usario_id := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "bwsvnish",
			"name": "usario_id",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), del_usario_id)
		collection.Schema.AddField(del_usario_id)

		// add
		del_data_evento := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "ezejigh1",
			"name": "data_evento",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), del_data_evento)
		collection.Schema.AddField(del_data_evento)

		// add
		del_uf := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "8rli2whk",
			"name": "uf",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), del_uf)
		collection.Schema.AddField(del_uf)

		// add
		del_cidade := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "4h3jg5fx",
			"name": "cidade",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), del_cidade)
		collection.Schema.AddField(del_cidade)

		// add
		del_pagina := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "ivusadta",
			"name": "pagina",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), del_pagina)
		collection.Schema.AddField(del_pagina)

		// add
		del_tamanho := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "17163ehc",
			"name": "tamanho",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), del_tamanho)
		collection.Schema.AddField(del_tamanho)

		// add
		del_link := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "ol1xjaqy",
			"name": "link",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), del_link)
		collection.Schema.AddField(del_link)

		// add
		del_titulo := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "vmibhzr0",
			"name": "titulo",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), del_titulo)
		collection.Schema.AddField(del_titulo)

		// add
		del_dados := &schema.SchemaField{}
		json.Unmarshal([]byte(`{
			"system": false,
			"id": "rvibbrzg",
			"name": "dados",
			"type": "json",
			"required": false,
			"presentable": false,
			"unique": false,
			"options": {}
		}`), del_dados)
		collection.Schema.AddField(del_dados)

		// remove
		collection.Schema.RemoveField("gkxpw2sd")

		// remove
		collection.Schema.RemoveField("hw5wouwe")

		// remove
		collection.Schema.RemoveField("vc63opzx")

		// remove
		collection.Schema.RemoveField("l7mhmx4v")

		// remove
		collection.Schema.RemoveField("lpsfputh")

		// remove
		collection.Schema.RemoveField("acywbctf")

		// remove
		collection.Schema.RemoveField("mo9m5hyz")

		// remove
		collection.Schema.RemoveField("4ohn2ycj")

		// remove
		collection.Schema.RemoveField("cvv5hdgm")

		// remove
		collection.Schema.RemoveField("4anr0dhp")

		// remove
		collection.Schema.RemoveField("2h1pumpm")

		// remove
		collection.Schema.RemoveField("qruxthyt")

		return dao.SaveCollection(collection)
	})
}

package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models/schema"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db)

		collection, err := dao.FindCollectionByNameOrId("89ju1qkr9p4o35m")
		if err != nil {
			return err
		}

		// Adicionar campo receivedAt
		collection.Schema.AddField(&schema.SchemaField{
			Name:     "receivedAt",
			Type:     schema.FieldTypeDate,
			Required: false,
		})

		// Adicionar campo request_ip
		collection.Schema.AddField(&schema.SchemaField{
			Name:     "request_ip",
			Type:     schema.FieldTypeText,
			Required: false,
		})

		// Adicionar campo rudderId
		collection.Schema.AddField(&schema.SchemaField{
			Name:     "rudderId",
			Type:     schema.FieldTypeText,
			Required: false,
		})

		// Adicionar campo timestamp
		collection.Schema.AddField(&schema.SchemaField{
			Name:     "timestamp",
			Type:     schema.FieldTypeDate,
			Required: false,
		})

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db)

		collection, err := dao.FindCollectionByNameOrId("89ju1qkr9p4o35m")
		if err != nil {
			return err
		}

		// Remover campos na reversão
		collection.Schema.RemoveField("receivedAt")
		collection.Schema.RemoveField("request_ip")
		collection.Schema.RemoveField("rudderId")
		collection.Schema.RemoveField("timestamp")

		return dao.SaveCollection(collection)
	})
}

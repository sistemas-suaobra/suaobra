package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db)

		collection, err := dao.FindCollectionByNameOrId("rudderstack")
		if err != nil {
			return err
		}

		// Adicionar índices para melhorar performance
		collection.Indexes = []string{
			"CREATE INDEX IF NOT EXISTS idx_rudderstack_timestamp ON rudderstack (originaltimestamp)",
			"CREATE INDEX IF NOT EXISTS idx_rudderstack_type_event ON rudderstack (type, event)",
			"CREATE INDEX IF NOT EXISTS idx_rudderstack_anonymousId ON rudderstack (anonymousId)",
			"CREATE INDEX IF NOT EXISTS idx_rudderstack_created ON rudderstack (created)",
		}

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		// Rollback - remover índices
		dao := daos.New(db)

		collection, err := dao.FindCollectionByNameOrId("rudderstack")
		if err != nil {
			return err
		}

		collection.Indexes = []string{}

		return dao.SaveCollection(collection)
	})
}

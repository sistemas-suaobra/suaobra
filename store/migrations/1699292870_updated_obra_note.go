package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("_collection_cydQmBK5ZeGpLfE")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("@request.auth.id != \"\"")

		collection.CreateRule = types.Pointer("@request.auth.properties.is_admin = \"true\"")

		collection.UpdateRule = types.Pointer("@request.auth.properties.is_admin = \"true\"")

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("_collection_cydQmBK5ZeGpLfE")
		if err != nil {
			return err
		}

		collection.ListRule = nil

		collection.CreateRule = nil

		collection.UpdateRule = nil

		return dao.SaveCollection(collection)
	})
}

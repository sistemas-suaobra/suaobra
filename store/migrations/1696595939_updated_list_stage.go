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

		collection, err := dao.FindCollectionByNameOrId("p980c5513p13sfl")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("@request.auth.team_id = @collection.list.team_id")

		collection.ViewRule = types.Pointer("@request.auth.team_id = @collection.list.team_id")

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("p980c5513p13sfl")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("@request.auth.team_id.id.id = @collection.list.team_id.id")

		collection.ViewRule = types.Pointer("@request.auth.team_id.id.id = @collection.list.team_id.id")

		return dao.SaveCollection(collection)
	})
}

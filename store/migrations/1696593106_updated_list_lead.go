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

		collection, err := dao.FindCollectionByNameOrId("m7bh3r7viqvvqrb")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("@request.auth.team_id = @collection.list.team_id")

		collection.ViewRule = types.Pointer("@request.auth.team_id = @collection.list.team_id")

		return dao.SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("m7bh3r7viqvvqrb")
		if err != nil {
			return err
		}

		collection.ListRule = types.Pointer("@request.auth.team_id = @collection.team.id")

		collection.ViewRule = types.Pointer("@request.auth.team_id = @collection.team.id")

		return dao.SaveCollection(collection)
	})
}

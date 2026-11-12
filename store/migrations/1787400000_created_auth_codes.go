package migrations

import (
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		jsonData := `{
			"id": "col_auth_codes_suaobra1",
			"name": "auth_codes",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "acemail0000001",
					"name": "email",
					"type": "email",
					"required": true,
					"presentable": true,
					"unique": false,
					"options": { "exceptDomains": null, "onlyDomains": null }
				},
				{
					"system": false,
					"id": "acuserid000001",
					"name": "user_id",
					"type": "text",
					"required": true,
					"presentable": false,
					"unique": false,
					"options": { "min": null, "max": 64, "pattern": "" }
				},
				{
					"system": false,
					"id": "acpurpose00001",
					"name": "purpose",
					"type": "text",
					"required": true,
					"presentable": true,
					"unique": false,
					"options": { "min": null, "max": 32, "pattern": "" }
				},
				{
					"system": false,
					"id": "achash00000001",
					"name": "code_hash",
					"type": "text",
					"required": true,
					"presentable": false,
					"unique": false,
					"options": { "min": null, "max": 128, "pattern": "" }
				},
				{
					"system": false,
					"id": "acexpires00001",
					"name": "expires_at",
					"type": "text",
					"required": true,
					"presentable": false,
					"unique": false,
					"options": { "min": null, "max": 64, "pattern": "" }
				},
				{
					"system": false,
					"id": "acused00000001",
					"name": "used",
					"type": "bool",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {}
				}
			],
			"indexes": [
				"CREATE INDEX ` + "`" + `idx_auth_codes_email_purpose` + "`" + ` ON ` + "`" + `auth_codes` + "`" + ` (` + "`" + `email` + "`" + `, ` + "`" + `purpose` + "`" + `, ` + "`" + `created` + "`" + `)"
			],
			"listRule": null,
			"viewRule": null,
			"createRule": null,
			"updateRule": null,
			"deleteRule": null,
			"options": {}
		}`

		collection := &models.Collection{}
		if err := json.Unmarshal([]byte(jsonData), collection); err != nil {
			return err
		}
		return daos.New(db).SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db)
		collection, err := dao.FindCollectionByNameOrId("col_auth_codes_suaobra1")
		if err != nil {
			return nil
		}
		return dao.DeleteCollection(collection)
	})
}

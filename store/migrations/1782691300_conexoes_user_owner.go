package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models/schema"
)

// Conexão de WhatsApp passa a ser por USUÁRIO (dono), não mais só por equipe.
//   - conexoes.user_id: dono da conexão (vazio = conexão legada/compartilhada do time).
//   - campanhas.criado_por: usuário que criou a campanha (usado para enviar pelo
//     número do dono). O frontend já envia esse valor; aqui garantimos o campo.
// Campos opcionais e sem backfill destrutivo: registros antigos ficam com valor
// vazio e continuam funcionando via fallback compartilhado.
func init() {
	m.Register(func(db dbx.Builder) error {
		dao := daos.New(db)

		if col, err := dao.FindCollectionByNameOrId("conexoes"); err == nil {
			col.Schema.AddField(&schema.SchemaField{
				Name:     "user_id",
				Type:     schema.FieldTypeText,
				Required: false,
			})
			if err := dao.SaveCollection(col); err != nil {
				return err
			}
		} else {
			return err
		}

		if col, err := dao.FindCollectionByNameOrId("campanhas"); err == nil {
			// Só adiciona se ainda não existir (pode já ter sido criado via Admin).
			if col.Schema.GetFieldByName("criado_por") == nil {
				col.Schema.AddField(&schema.SchemaField{
					Name:     "criado_por",
					Type:     schema.FieldTypeText,
					Required: false,
				})
				if err := dao.SaveCollection(col); err != nil {
					return err
				}
			}
		} else {
			return err
		}

		return nil
	}, func(db dbx.Builder) error {
		dao := daos.New(db)

		if col, err := dao.FindCollectionByNameOrId("conexoes"); err == nil {
			col.Schema.RemoveField("user_id")
			if err := dao.SaveCollection(col); err != nil {
				return err
			}
		}

		// criado_por não é removido no down por poder ser usado por outras telas.
		return nil
	})
}

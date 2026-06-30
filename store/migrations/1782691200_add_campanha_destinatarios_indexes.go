package migrations

import (
	"github.com/pocketbase/dbx"
	m "github.com/pocketbase/pocketbase/migrations"
)

// Índices para acelerar as consultas de campanha:
//   - leads-plus usa duas subqueries correlacionadas por linha em
//     campanha_destinatarios (team_id, obra_id, contato_tipo, status);
//   - o envio busca pendentes/duplicados por (campanha_id, status) e
//     (campanha_id, telefone_e164).
// Sem esses índices, cada linha de Obras+ faz um scan da tabela, deixando
// o carregamento da tela de campanhas lento.
func init() {
	upQueries := []string{
		"CREATE INDEX IF NOT EXISTS idx_camp_dest_team_obra_tipo_status ON campanha_destinatarios (team_id, obra_id, contato_tipo, status)",
		"CREATE INDEX IF NOT EXISTS idx_camp_dest_campanha_status ON campanha_destinatarios (campanha_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_camp_dest_campanha_telefone ON campanha_destinatarios (campanha_id, telefone_e164)",
	}

	downQueries := []string{
		"DROP INDEX IF EXISTS idx_camp_dest_team_obra_tipo_status",
		"DROP INDEX IF EXISTS idx_camp_dest_campanha_status",
		"DROP INDEX IF EXISTS idx_camp_dest_campanha_telefone",
	}

	m.Register(func(db dbx.Builder) error {
		for _, q := range upQueries {
			if _, err := db.NewQuery(q).Execute(); err != nil {
				return err
			}
		}
		return nil
	}, func(db dbx.Builder) error {
		for _, q := range downQueries {
			if _, err := db.NewQuery(q).Execute(); err != nil {
				return err
			}
		}
		return nil
	})
}

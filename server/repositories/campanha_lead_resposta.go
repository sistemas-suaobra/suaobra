package repositories

import (
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
)

// CampanhaLeadRespostaInput dados de uma mensagem recebida do lead (webhook).
type CampanhaLeadRespostaInput struct {
	TeamID           string
	CampanhaID       string
	DestinatarioID   string
	ConversaID       string
	Canal            string
	TelefoneE164     string
	Email            string
	NomeContato      string
	Corpo            string
	RecebidaEm       time.Time
	MessageIDExterno string
	Payload          map[string]any
}

// SaveCampanhaLeadResposta grava um evento em campanha_lead_respostas (métricas do dashboard).
// Ignora gravação se corpo ou team_id estiverem vazios. Deduplica por team_id + message_id_externo quando informado.
func SaveCampanhaLeadResposta(dao *daos.Dao, in CampanhaLeadRespostaInput) error {
	if dao == nil {
		return nil
	}
	in.TeamID = strings.TrimSpace(in.TeamID)
	in.Corpo = strings.TrimSpace(in.Corpo)
	if in.TeamID == "" || in.Corpo == "" {
		return nil
	}
	if strings.TrimSpace(in.Canal) == "" {
		in.Canal = "WHATSAPP"
	}
	if in.RecebidaEm.IsZero() {
		in.RecebidaEm = time.Now().UTC()
	}

	msgID := strings.TrimSpace(in.MessageIDExterno)
	if msgID != "" {
		existing, _ := dao.FindFirstRecordByFilter(
			"campanha_lead_respostas",
			"team_id = {:tid} && message_id_externo = {:mid}",
			dbx.Params{"tid": in.TeamID, "mid": msgID},
		)
		if existing != nil && existing.Id != "" {
			return nil
		}
	}

	collection, err := dao.FindCollectionByNameOrId("campanha_lead_respostas")
	if err != nil {
		return err
	}

	rec := models.NewRecord(collection)
	rec.Set("team_id", in.TeamID)
	rec.Set("recebida_em", in.RecebidaEm)
	rec.Set("corpo", in.Corpo)
	rec.Set("canal", strings.TrimSpace(in.Canal))
	if strings.TrimSpace(in.TelefoneE164) != "" {
		rec.Set("telefone_e164", strings.TrimSpace(in.TelefoneE164))
	}
	if strings.TrimSpace(in.Email) != "" {
		rec.Set("email", strings.TrimSpace(in.Email))
	}
	if strings.TrimSpace(in.NomeContato) != "" {
		rec.Set("nome_contato", strings.TrimSpace(in.NomeContato))
	}
	if strings.TrimSpace(in.CampanhaID) != "" {
		rec.Set("campanha_id", strings.TrimSpace(in.CampanhaID))
	}
	if strings.TrimSpace(in.DestinatarioID) != "" {
		rec.Set("destinatario_id", strings.TrimSpace(in.DestinatarioID))
	}
	if strings.TrimSpace(in.ConversaID) != "" {
		rec.Set("conversa_id", strings.TrimSpace(in.ConversaID))
	}
	if msgID != "" {
		rec.Set("message_id_externo", msgID)
	}
	if len(in.Payload) > 0 {
		rec.Set("payload", in.Payload)
	}

	return dao.SaveRecord(rec)
}

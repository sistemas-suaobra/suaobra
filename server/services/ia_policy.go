package services

import (
	"strings"

	"github.com/flarco/g"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
	"github.com/suaobra/suaobra-app/server/repositories"
)

func normalizePhoneDigits(telefone string) string {
	var b strings.Builder
	for _, r := range telefone {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// BuildTelefoneCandidates gera variações do número para matching no banco.
func BuildTelefoneCandidates(telefone string) []string {
	telefone = normalizePhoneDigits(telefone)
	if telefone == "" {
		return nil
	}

	seen := map[string]struct{}{}
	out := make([]string, 0, 6)

	add := func(v string) {
		v = normalizePhoneDigits(v)
		if v == "" {
			return
		}
		if _, ok := seen[v]; ok {
			return
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}

	add(telefone)

	if strings.HasPrefix(telefone, "55") && len(telefone) > 10 {
		add(telefone[2:])
	}
	if !strings.HasPrefix(telefone, "55") && len(telefone) >= 10 {
		add("55" + telefone)
	}

	withCC := telefone
	if !strings.HasPrefix(withCC, "55") && len(withCC) >= 10 {
		withCC = "55" + withCC
	}
	if strings.HasPrefix(withCC, "55") && len(withCC) >= 12 {
		ddd := withCC[2:4]
		local := withCC[4:]
		if len(local) == 9 && local[0] == '9' {
			add("55" + ddd + local[1:])
			add(ddd + local[1:])
		} else if len(local) == 8 {
			add("55" + ddd + "9" + local)
			add(ddd + "9" + local)
		}
	}

	return out
}

// FindDestinatarioEnviadoRecente retorna o destinatário ENVIADO mais recente para o telefone.
func FindDestinatarioEnviadoRecente(dao *daos.Dao, teamID string, candidatos []string) *models.Record {
	teamID = strings.TrimSpace(teamID)
	if dao == nil || teamID == "" {
		return nil
	}

	for _, candidato := range candidatos {
		candidato = strings.TrimSpace(candidato)
		if candidato == "" {
			continue
		}

		recs, err := dao.FindRecordsByFilter(
			"campanha_destinatarios",
			"team_id = {:teamId} && telefone_e164 = {:telefone} && status = 'ENVIADO'",
			"-enviado_em,-updated",
			1,
			0,
			dbx.Params{
				"teamId":   teamID,
				"telefone": candidato,
			},
		)
		if err != nil {
			g.Warn("ia_policy: erro ao buscar destinatário team=%s telefone=%s err=%v", teamID, candidato, err)
			continue
		}
		if len(recs) > 0 {
			return recs[0]
		}
	}

	return nil
}

// CampanhaPermiteIA verifica se a campanha permite resposta automática da IA.
func CampanhaPermiteIA(dao *daos.Dao, campanhaID string) bool {
	campanhaID = strings.TrimSpace(campanhaID)
	if dao == nil || campanhaID == "" {
		return false
	}

	campanha, err := dao.FindRecordById("campanhas", campanhaID)
	if err != nil || campanha == nil {
		return false
	}

	return campanha.GetBool("manter_ia")
}

// IAAtivaParaTelefone decide se a IA pode responder com base apenas na campanha
// mais recente que enviou mensagem para o telefone (sem fallback em conversas antigas).
func IAAtivaParaTelefone(dao *daos.Dao, teamID string, candidatos []string) bool {
	dest := FindDestinatarioEnviadoRecente(dao, teamID, candidatos)
	if dest == nil {
		return false
	}

	campanhaID := strings.TrimSpace(dest.GetString("campanha_id"))
	permitido := CampanhaPermiteIA(dao, campanhaID)
	g.Info(
		"ia_policy: telefone candidatos=%v dest_id=%s campanha=%s manter_ia=%v",
		candidatos,
		dest.Id,
		campanhaID,
		permitido,
	)
	return permitido
}

// PausarConversaIA marca conversa ativa como PAUSADA para o telefone.
func PausarConversaIA(conversaRepo *repositories.ConversaRepo, teamID string, candidatos []string) {
	if conversaRepo == nil {
		return
	}

	conversa := conversaRepo.FindByTelefoneCandidates(teamID, candidatos)
	if conversa == nil {
		return
	}

	status := strings.ToUpper(strings.TrimSpace(conversa.GetString("status")))
	if status != "ATIVA" {
		return
	}

	conversa.Set("status", "PAUSADA")
	if err := conversaRepo.Save(conversa); err != nil {
		g.Warn("ia_policy: falha ao pausar conversa %s err=%v", conversa.Id, err)
		return
	}

	g.Info("ia_policy: conversa pausada id=%s team=%s", conversa.Id, teamID)
}

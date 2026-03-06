package services

import (
	"strings"
	"time"

	"github.com/flarco/g"
	"github.com/pocketbase/pocketbase/models"
	"github.com/spf13/cast"

	"github.com/suaobra/suaobra-app/server/repositories"
	"github.com/suaobra/suaobra-app/store"
)

const (
	maxEnviosPorHora = 5
	maxEnviosPorDia  = 100
	delayEntreEnvios = 10 * time.Second
)

type CampanhaService struct {
	repo        *repositories.CampanhaRepo
	waSvc       *WhatsAppService
	emailSvc    *EmailService
	conversaRepo *repositories.ConversaRepo
}

func NewCampanhaService(
	repo *repositories.CampanhaRepo,
	waSvc *WhatsAppService,
	emailSvc *EmailService,
	conversaRepo *repositories.ConversaRepo,
) *CampanhaService {
	return &CampanhaService{repo: repo, waSvc: waSvc, emailSvc: emailSvc, conversaRepo: conversaRepo}
}

// IniciarCampanha inicia uma campanha e dispara o processamento em background
func (s *CampanhaService) IniciarCampanha(teamID, campanhaID string) error {
	campanha, err := s.repo.FindByID(campanhaID)
	if err != nil {
		return g.Error(err, "campanha não encontrada")
	}

	if campanha.GetString("team_id") != teamID {
		return g.Error("não autorizado")
	}

	status := campanha.GetString("status")
	if status == repositories.CampanhaStatusEmAndamento || status == repositories.CampanhaStatusConcluida {
		return g.Error("campanha já iniciada ou concluída")
	}

	// Verifica canais da campanha
	canais := campanha.GetStringSlice("canal")
	if len(canais) == 0 {
		canalStr := campanha.GetString("canal")
		if canalStr != "" {
			canais = []string{canalStr}
		}
	}

	usaWhatsApp := false
	usaEmail := false
	for _, c := range canais {
		if c == "WHATSAPP" {
			usaWhatsApp = true
		}
		if c == "EMAIL" {
			usaEmail = true
		}
	}

	// Valida que WhatsApp está conectado
	if usaWhatsApp {
		_, _, wa, err := s.waSvc.GetByTeam(teamID)
		if err != nil || wa == nil {
			return g.Error("Conexão WhatsApp não encontrada. Configure o WhatsApp em Conexões.")
		}
		if wa.GetDateTime("conectado_em").Time().IsZero() {
			return g.Error("WhatsApp não está conectado. Conecte o WhatsApp antes de iniciar a campanha.")
		}
	}

	// Valida que e-mail está configurado
	if usaEmail && s.emailSvc != nil {
		exists, _, _, err := s.emailSvc.GetConfig(teamID)
		if err != nil || !exists {
			return g.Error("Configuração de e-mail não encontrada. Configure o E-mail em Conexões.")
		}
	}

	// Verifica limite diário
	enviadosHoje, _ := s.repo.CountEnviadosHojeByTeam(teamID)
	if enviadosHoje >= maxEnviosPorDia {
		return g.Error("Limite diário de %d mensagens já atingido. Tente novamente amanhã.", maxEnviosPorDia)
	}

	// Auto-enriquece destinatários com dados de contato do core.db
	s.EnriquecerDestinatarios(teamID, campanhaID)

	// Atualiza status
	if err := s.repo.UpdateStatus(campanha, repositories.CampanhaStatusEmAndamento); err != nil {
		return g.Error(err, "erro ao atualizar status")
	}

	// Dispara processamento em background
	go s.ProcessarCampanhaAsync(campanhaID)

	return nil
}

// ProcessarCampanhaAsync processa os envios de uma campanha em background
func (s *CampanhaService) ProcessarCampanhaAsync(campanhaID string) {
	campanha, err := s.repo.FindByID(campanhaID)
	if err != nil {
		g.Error(err, "erro ao buscar campanha %s", campanhaID)
		return
	}

	teamID := campanha.GetString("team_id")
	manterIA := campanha.GetBool("manter_ia")
	mensagemTemplate := campanha.GetString("mensagem_template")
	assuntoEmail := campanha.GetString("assunto_email")
	if assuntoEmail == "" {
		assuntoEmail = campanha.GetString("nome")
	}

	canais := campanha.GetStringSlice("canal")
	if len(canais) == 0 {
		canalStr := campanha.GetString("canal")
		if canalStr != "" {
			canais = []string{canalStr}
		} else {
			canais = []string{"WHATSAPP"}
		}
	}

	usaWhatsApp := false
	usaEmail := false
	for _, c := range canais {
		if c == "WHATSAPP" {
			usaWhatsApp = true
		}
		if c == "EMAIL" {
			usaEmail = true
		}
	}

	g.Info("Processando campanha %s para team %s canais %v", campanhaID, teamID, canais)

	destinatarios, err := s.repo.FindDestinatariosPendentes(campanhaID)
	if err != nil {
		g.Error(err, "erro ao buscar destinatários")
		s.repo.UpdateStatus(campanha, repositories.CampanhaStatusConcluida)
		return
	}

	g.Info("Encontrados %d destinatários pendentes", len(destinatarios))

	enviadosHoje, _ := s.repo.CountEnviadosHojeByTeam(teamID)
	enviosNestaHora := 0
	inicioHora := time.Now()

	enviados := 0
	falhas := 0

	for _, dest := range destinatarios {
		campanha, _ = s.repo.FindByID(campanhaID)
		status := campanha.GetString("status")
		if status == repositories.CampanhaStatusPausada || status == repositories.CampanhaStatusCancelada {
			g.Info("Campanha %s foi %s, interrompendo", campanhaID, status)
			break
		}

		if enviadosHoje+enviados >= maxEnviosPorDia {
			g.Info("Limite diário de %d mensagens atingido para team %s", maxEnviosPorDia, teamID)
			break
		}

		if enviosNestaHora >= maxEnviosPorHora {
			elapsed := time.Since(inicioHora)
			remaining := time.Hour - elapsed
			if remaining > 0 {
				g.Info("Rate limit: %d/hora atingido, aguardando %v", maxEnviosPorHora, remaining)
				if s.waitWithCancelCheck(campanhaID, remaining) {
					break
				}
			}
			inicioHora = time.Now()
			enviosNestaHora = 0
		}

		nomeContato := dest.GetString("nome_contato")
		telefone := dest.GetString("telefone_e164")
		email := dest.GetString("email")

		cidade := ""
		bairro := ""
		uf := ""

		leadID := dest.GetString("lead_id")
		leadRecord, _ := s.repo.FindLeadByID(leadID)

		if leadRecord != nil {
			propsAny := leadRecord.Get("properties")
			props := cast.ToStringMap(propsAny)

			if nomeContato == "" {
				nomeContato = cast.ToString(props["nome_contato"])
				if nomeContato == "" {
					nomeContato = cast.ToString(props["contato_nome"])
				}
			}

			if telefone == "" {
				telefone = cast.ToString(props["telefone_e164"])
				if telefone == "" {
					telefone = cast.ToString(props["contato_telefone"])
				}
				if telefone != "" {
					telefone = s.formatPhone(telefone)
				}
			}

			if email == "" {
				email = cast.ToString(props["email"])
				if email == "" {
					email = cast.ToString(props["contato_email"])
				}
			}

			cidade = cast.ToString(props["city"])
			bairro = cast.ToString(props["bairro"])
			uf = cast.ToString(props["state"])

			if (nomeContato == "" && telefone == "" && email == "") || (cidade == "" && uf == "") {
				obraID := leadRecord.GetString("obra_id")
				if contato, err := s.BuscarContatoObra(obraID); err == nil && contato != nil {
					if nomeContato == "" {
						nomeContato = contato["nome"]
					}
					if telefone == "" {
						telefone = s.formatPhone(contato["telefone"])
					}
					if email == "" {
						email = contato["email"]
					}
					if cidade == "" {
						cidade = contato["cidade"]
					}
					if uf == "" {
						uf = contato["uf"]
					}
				}
			}

			changed := false
			if dest.GetString("nome_contato") == "" && nomeContato != "" {
				dest.Set("nome_contato", nomeContato)
				changed = true
			}
			if dest.GetString("telefone_e164") == "" && telefone != "" {
				dest.Set("telefone_e164", telefone)
				changed = true
			}
			if dest.GetString("email") == "" && email != "" {
				dest.Set("email", email)
				changed = true
			}
			if changed {
				_ = s.repo.SaveDestinatario(dest)
			}
		}

		if nomeContato == "" && telefone == "" && email == "" {
			s.repo.UpdateDestinatarioStatus(dest, repositories.DestStatusIgnorado, "Sem contato disponível")
			falhas++
			continue
		}

		mensagem := s.PersonalizarMensagem(mensagemTemplate, map[string]string{
			"nome":   nomeContato,
			"cidade": cidade,
			"bairro": bairro,
			"uf":     uf,
		})

		var sendErr error
		enviouAlgum := false

		if usaWhatsApp && telefone != "" {
			sendErr = s.enviarWhatsApp(teamID, telefone, mensagem)
			if sendErr == nil {
				enviouAlgum = true
				g.Info("WhatsApp enviado para %s", telefone)

				if manterIA && s.conversaRepo != nil {
					existente, _ := s.conversaRepo.FindByTelefone(teamID, telefone)
					if existente == nil {
						_, cErr := s.conversaRepo.Create(map[string]interface{}{
							"team_id":           teamID,
							"campanha_id":       campanhaID,
							"telefone":          telefone,
							"nome_contato":      nomeContato,
							"mensagens":         []map[string]interface{}{{"role": "assistant", "content": mensagem, "timestamp": time.Now().UTC().Format(time.RFC3339)}},
							"status":            "ATIVA",
							"ultima_mensagem_em": time.Now().UTC(),
						})
						if cErr != nil {
							g.Warn("manter_ia: erro ao criar conversa para %s: %v", telefone, cErr)
						}
					}
				}
			} else {
				g.Error(sendErr, "Falha WhatsApp para %s", telefone)
			}
		}

		if usaEmail && email != "" {
			emailErr := s.enviarEmail(teamID, email, assuntoEmail, mensagem)
			if emailErr == nil {
				enviouAlgum = true
				g.Info("Email enviado para %s", email)
			} else {
				g.Error(emailErr, "Falha Email para %s", email)
				if sendErr == nil {
					sendErr = emailErr
				}
			}
		}

		if !enviouAlgum {
			errMsg := "Sem contato para os canais selecionados"
			if sendErr != nil {
				errMsg = sendErr.Error()
			}
			s.repo.UpdateDestinatarioStatus(dest, repositories.DestStatusFalhou, errMsg)
			falhas++
			continue
		}

		s.repo.UpdateDestinatarioStatus(dest, repositories.DestStatusEnviado, "")
		enviados++
		enviosNestaHora++
		g.Info("Enviado %d/%d lead=%s cidade=%s bairro=%s uf=%s", enviados, len(destinatarios), dest.GetString("lead_id"), cidade, bairro, uf)

		time.Sleep(delayEntreEnvios)
	}

	campanha, _ = s.repo.FindByID(campanhaID)
	if campanha.GetString("status") == repositories.CampanhaStatusEmAndamento {
		s.repo.UpdateStatus(campanha, repositories.CampanhaStatusConcluida)
	}
	g.Info("Campanha %s finalizada: %d enviados, %d falhas", campanhaID, enviados, falhas)
}

// BuscarContatoObra busca o contato (telefone/email) de uma obra pelo ID
// Usa a mesma estratégia do Obras Plus: busca owner/professional da obra,
// depois busca telefone e email separadamente nas tabelas específicas
func (s *CampanhaService) BuscarContatoObra(obraID string) (map[string]string, error) {
	// 1. Busca dados da obra (owner, professional, cidade, uf)
	sqlObra := `
		SELECT 
			owner, 
			professional,
			city,
			state
		FROM core.core_obras_plus
		WHERE id = {:obraID}
		LIMIT 1
	`
	
	boundSQL := store.BindSQL(sqlObra, map[string]any{"obraID": obraID})
	dataObra, err := store.MainDB.Query(boundSQL)
	if err != nil {
		return nil, g.Error(err, "erro ao buscar obra %s", obraID)
	}

	recsObra := dataObra.RecordsCasted(true)
	if len(recsObra) == 0 {
		return nil, nil
	}

	obra := recsObra[0]
	owner := cast.ToString(obra["owner"])
	professional := cast.ToString(obra["professional"])
	cidade := cast.ToString(obra["city"])
	uf := cast.ToString(obra["state"])

	if owner == "" && professional == "" {
		return nil, nil
	}

	// 2. Busca telefone do owner primeiro, depois professional
	telefone := ""
	nome := owner
	if owner != "" {
		telefone = s.buscarTelefone(owner, cidade, uf)
	}
	if telefone == "" && professional != "" {
		telefone = s.buscarTelefone(professional, cidade, uf)
		if telefone != "" {
			nome = professional
		}
	}

	// 3. Busca email do owner primeiro, depois professional
	email := ""
	if owner != "" {
		email = s.buscarEmail(owner, cidade, uf)
	}
	if email == "" && professional != "" {
		email = s.buscarEmail(professional, cidade, uf)
	}

	// Se não encontrou nenhum contato, retorna nil
	if telefone == "" && email == "" {
		return nil, nil
	}

	return map[string]string{
		"nome":     nome,
		"telefone": telefone,
		"email":    email,
		"cidade":   cidade,
		"uf":       uf,
	}, nil
}

// buscarTelefone busca telefone de uma pessoa na tabela core_obras_plus_phone
func (s *CampanhaService) buscarTelefone(nome, cidade, uf string) string {
	sql := `
		WITH prep AS (
			SELECT
				telefone,
				cidade,
				uf,
				row_number() OVER(PARTITION BY telefone ORDER BY COALESCE(poder_aquisitivo, 1) DESC) as row_num
			FROM core.core_obras_plus_phone
			WHERE nome = {:nome}
		)
		SELECT telefone
		FROM prep
		WHERE row_num = 1
		ORDER BY
			CASE WHEN uf = {:uf} AND cidade = {:cidade} THEN 1 ELSE 2 END ASC,
			telefone DESC
		LIMIT 1
	`
	
	boundSQL := store.BindSQL(sql, map[string]any{
		"nome":   nome,
		"cidade": cidade,
		"uf":     uf,
	})
	
	data, err := store.MainDB.Query(boundSQL)
	if err != nil {
		return ""
	}
	
	recs := data.RecordsCasted(true)
	if len(recs) == 0 {
		return ""
	}
	
	return cast.ToString(recs[0]["telefone"])
}

// buscarEmail busca email de uma pessoa na tabela core_obras_plus_email
func (s *CampanhaService) buscarEmail(nome, cidade, uf string) string {
	sql := `
		WITH prep AS (
			SELECT
				email,
				cidade,
				uf,
				poder_aquisitivo,
				row_number() OVER(PARTITION BY email ORDER BY COALESCE(poder_aquisitivo, 1) DESC) as row_num
			FROM core.core_obras_plus_email
			WHERE nome = {:nome}
		)
		SELECT email
		FROM prep
		WHERE row_num = 1
		ORDER BY
			CASE WHEN uf = {:uf} AND cidade = {:cidade} THEN 1 ELSE 2 END ASC,
			poder_aquisitivo DESC
		LIMIT 1
	`
	
	boundSQL := store.BindSQL(sql, map[string]any{
		"nome":   nome,
		"cidade": cidade,
		"uf":     uf,
	})
	
	g.Debug("Query email para nome=%s: %s", nome, boundSQL)
	
	data, err := store.MainDB.Query(boundSQL)
	if err != nil {
		g.Warn("Erro ao buscar email para %s: %v", nome, err)
		return ""
	}
	
	recs := data.RecordsCasted(true)
	g.Debug("Email query retornou %d registros para nome=%s", len(recs), nome)
	
	if len(recs) == 0 {
		return ""
	}
	
	email := cast.ToString(recs[0]["email"])
	g.Info("Email encontrado para %s: %s", nome, email)
	return email
}

// EnriquecerDestinatarios preenche nome_contato, telefone_e164 e email
// de todos os destinatários pendentes de uma campanha usando dados do core.db
func (s *CampanhaService) EnriquecerDestinatarios(teamID, campanhaID string) (int, int, error) {
	campanha, err := s.repo.FindByID(campanhaID)
	if err != nil {
		return 0, 0, g.Error(err, "campanha não encontrada")
	}

	if campanha.GetString("team_id") != teamID {
		return 0, 0, g.Error("não autorizado")
	}

	destinatarios, err := s.repo.FindDestinatariosPendentes(campanhaID)
	if err != nil {
		return 0, 0, g.Error(err, "erro ao buscar destinatários")
	}

	enriquecidos := 0
	semContato := 0

	for _, dest := range destinatarios {
		leadID := dest.GetString("lead_id")
		leadRecord, err := s.repo.FindLeadByID(leadID)
		if err != nil || leadRecord == nil {
			semContato++
			continue
		}

		nome := strings.TrimSpace(dest.GetString("nome_contato"))
		telefone := strings.TrimSpace(dest.GetString("telefone_e164"))
		email := strings.TrimSpace(dest.GetString("email"))

		propsAny := leadRecord.Get("properties")
		props := cast.ToStringMap(propsAny)

		if nome == "" {
			nome = cast.ToString(props["nome_contato"])
			if nome == "" {
				nome = cast.ToString(props["contato_nome"])
			}
		}

		if telefone == "" {
			telefone = cast.ToString(props["telefone_e164"])
			if telefone == "" {
				telefone = cast.ToString(props["contato_telefone"])
			}
			if telefone != "" {
				telefone = s.formatPhone(telefone)
			}
		}

		if email == "" {
			email = cast.ToString(props["email"])
			if email == "" {
				email = cast.ToString(props["contato_email"])
			}
		}

		if nome == "" && telefone == "" && email == "" {
			obraID := leadRecord.GetString("obra_id")
			contato, err := s.BuscarContatoObra(obraID)
			if err != nil || contato == nil {
				semContato++
				continue
			}

			nome = strings.TrimSpace(contato["nome"])
			telefone = strings.TrimSpace(contato["telefone"])
			email = strings.TrimSpace(contato["email"])

			if telefone != "" {
				telefone = s.formatPhone(telefone)
			}
		}

		if nome == "" && telefone == "" && email == "" {
			semContato++
			continue
		}

		changed := false

		if strings.TrimSpace(dest.GetString("nome_contato")) == "" && nome != "" {
			dest.Set("nome_contato", nome)
			changed = true
		}

		if strings.TrimSpace(dest.GetString("telefone_e164")) == "" && telefone != "" {
			dest.Set("telefone_e164", telefone)
			changed = true
		}

		if strings.TrimSpace(dest.GetString("email")) == "" && email != "" {
			dest.Set("email", email)
			changed = true
		}

		if !changed {
			enriquecidos++
			continue
		}

		if err := s.repo.SaveDestinatario(dest); err != nil {
			g.Error(err, "erro ao salvar destinatário %s", dest.Id)
			continue
		}

		enriquecidos++
	}

	g.Info("Enriquecimento campanha %s: %d enriquecidos, %d sem contato", campanhaID, enriquecidos, semContato)
	return enriquecidos, semContato, nil
}

// PersonalizarMensagem substitui os placeholders na mensagem
func (s *CampanhaService) PersonalizarMensagem(template string, contato map[string]string) string {
	nome := contato["nome"]
	if nome == "" {
		nome = "Cliente"
	}

	h := time.Now().Hour()
	saudacao := "boa noite"
	if h < 12 {
		saudacao = "bom dia"
	} else if h < 18 {
		saudacao = "boa tarde"
	}

	msg := template

	replace := func(key, val string) {
		variants := []string{
			"{{" + strings.ToLower(key) + "}}",
			"{{" + strings.ToUpper(key) + "}}",
			"{{ " + strings.ToLower(key) + " }}",
			"{{ " + strings.ToUpper(key) + " }}",
		}
		for _, v := range variants {
			msg = strings.ReplaceAll(msg, v, val)
		}
	}

	replace("NOME", nome)
	replace("SAUDACAO", saudacao)
	replace("CIDADE", contato["cidade"])
	replace("BAIRRO", contato["bairro"])

	return msg
}

// PausarCampanha pausa uma campanha em andamento
func (s *CampanhaService) PausarCampanha(teamID, campanhaID string) error {
	campanha, err := s.repo.FindByID(campanhaID)
	if err != nil {
		return g.Error(err, "campanha não encontrada")
	}

	if campanha.GetString("team_id") != teamID {
		return g.Error("não autorizado")
	}

	if campanha.GetString("status") != repositories.CampanhaStatusEmAndamento {
		return g.Error("campanha não está em andamento")
	}

	return s.repo.UpdateStatus(campanha, repositories.CampanhaStatusPausada)
}

// CancelarCampanha cancela uma campanha
func (s *CampanhaService) CancelarCampanha(teamID, campanhaID string) error {
	campanha, err := s.repo.FindByID(campanhaID)
	if err != nil {
		return g.Error(err, "campanha não encontrada")
	}

	if campanha.GetString("team_id") != teamID {
		return g.Error("não autorizado")
	}

	return s.repo.UpdateStatus(campanha, repositories.CampanhaStatusCancelada)
}

// GetStatus retorna o status detalhado de uma campanha
func (s *CampanhaService) GetStatus(teamID, campanhaID string) (*models.Record, map[string]any, error) {
	campanha, err := s.repo.FindByID(campanhaID)
	if err != nil {
		return nil, nil, g.Error(err, "campanha não encontrada")
	}

	if campanha.GetString("team_id") != teamID {
		return nil, nil, g.Error("não autorizado")
	}

	destStats, err := s.repo.CountDestinatariosByStatus(campanhaID)
	if err != nil {
		return campanha, nil, err
	}
	
	stats := make(map[string]any)
	for k, v := range destStats {
		stats[k] = v
	}

	// Adiciona estatísticas de respostas
	respostas, err := s.repo.CountRespostas(campanhaID)
	if err != nil {
		g.Warn("Erro ao contar respostas para campanha %s: %v", campanhaID, err)
		stats["respostas"] = 0
		stats["taxa_resposta"] = 0.0
	} else {
		stats["respostas"] = respostas
		enviados := destStats["enviado"]
		if enviados > 0 {
			stats["taxa_resposta"] = float64(respostas) / float64(enviados)
		} else {
			stats["taxa_resposta"] = 0.0
		}
	}

	return campanha, stats, nil
}

// formatPhone formata telefone para E164
func (s *CampanhaService) formatPhone(phone string) string {
	result := ""
	for _, c := range phone {
		if c >= '0' && c <= '9' {
			result += string(c)
		}
	}
	// Adiciona código do país se não tiver
	if len(result) == 10 || len(result) == 11 {
		result = "55" + result
	}
	return result
}

// enviarWhatsApp envia mensagem via WhatsApp
func (s *CampanhaService) enviarWhatsApp(teamID, telefone, mensagem string) error {
	phone := s.formatPhone(telefone)
	_, err := s.waSvc.SendTestMessage(teamID, phone, mensagem)
	return err
}

// waitWithCancelCheck aguarda uma duração, verificando pause/cancel a cada 30s.
// Retorna true se a campanha foi pausada/cancelada durante a espera.
func (s *CampanhaService) waitWithCancelCheck(campanhaID string, duration time.Duration) bool {
	remaining := duration
	for remaining > 0 {
		sleep := 30 * time.Second
		if remaining < sleep {
			sleep = remaining
		}
		time.Sleep(sleep)
		remaining -= sleep

		campanha, _ := s.repo.FindByID(campanhaID)
		if campanha != nil {
			st := campanha.GetString("status")
			if st == repositories.CampanhaStatusPausada || st == repositories.CampanhaStatusCancelada {
				return true
			}
		}
	}
	return false
}

// enviarEmail envia mensagem via Email usando o serviço de email configurado
func (s *CampanhaService) enviarEmail(teamID, email, assunto, mensagem string) error {
	if s.emailSvc == nil {
		return g.Error("serviço de email não configurado")
	}
	return s.emailSvc.SendTestEmail(teamID, email, assunto, mensagem)
}

// GetDashboardStats retorna estatísticas agregadas para o dashboard
func (s *CampanhaService) GetDashboardStats(teamID string) (map[string]any, error) {
	stats, err := s.repo.GetDashboardStats(teamID)
	if err != nil {
		return nil, g.Error(err, "erro ao obter estatísticas")
	}

	enviadas, _ := stats["enviadas"].(int)
	respostas, _ := stats["respostas"].(int)
	
	taxa := 0.0
	if enviadas > 0 {
		taxa = float64(respostas) / float64(enviadas)
	}
	stats["taxa"] = taxa

	return stats, nil
}

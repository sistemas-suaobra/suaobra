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
	// Intervalo de 2 minutos entre envios de WhatsApp (por campanha).
	delayEntreEnviosWhatsApp = 2 * time.Minute

	ContatoTipoOwner        = "OWNER"
	ContatoTipoProfessional = "PROFISSIONAL"

	maxTelefonesPorContato = 10
	maxEmailsPorContato    = 10
)

type CampanhaDestinatarioInput struct {
	ObraID      string `json:"obra_id"`
	ContatoTipo string `json:"contato_tipo"` // OWNER | PROFISSIONAL
}

type ObraContatoCampanha struct {
	ObraID        string
	ContatoTipo   string
	NomeContato   string
	TelefonesE164 []string
	Emails        []string
	Cidade        string
	Bairro        string
	UF            string
	Endereco      string
}

type obraBase struct {
	Owner        string
	Professional string
	Address      string
	Bairro       string
	Cidade       string
	UF           string
}

type CampanhaService struct {
	repo         *repositories.CampanhaRepo
	waSvc        *WhatsAppService
	emailSvc     *EmailService
	conversaRepo *repositories.ConversaRepo
}

func NewCampanhaService(
	repo *repositories.CampanhaRepo,
	waSvc *WhatsAppService,
	emailSvc *EmailService,
	conversaRepo *repositories.ConversaRepo,
) *CampanhaService {
	return &CampanhaService{
		repo:         repo,
		waSvc:        waSvc,
		emailSvc:     emailSvc,
		conversaRepo: conversaRepo,
	}
}

func normalizarContatoTipo(v string) string {
	val := strings.ToUpper(strings.TrimSpace(v))

	switch val {
	case "OWNER", "PROPRIETARIO", "PROPRIETÁRIO":
		return ContatoTipoOwner
	case "PROFESSIONAL", "PROFISSIONAL":
		return ContatoTipoProfessional
	default:
		return ""
	}
}

func uniqueNonEmpty(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))

	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, exists := seen[v]; exists {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}

	return out
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (s *CampanhaService) getCampaignChannels(campanha *models.Record) (usaWhatsApp bool, usaEmail bool) {
	if campanha == nil {
		return false, false
	}

	canais := campanha.GetStringSlice("canal")
	if len(canais) == 0 {
		canalStr := strings.TrimSpace(campanha.GetString("canal"))
		if canalStr != "" {
			canais = []string{canalStr}
		}
	}

	for _, c := range canais {
		switch strings.ToUpper(strings.TrimSpace(c)) {
		case "WHATSAPP":
			usaWhatsApp = true
		case "EMAIL":
			usaEmail = true
		}
	}

	return usaWhatsApp, usaEmail
}

func (s *CampanhaService) getTargetChannelsForDest(
	dest *models.Record,
	usaWhatsApp, usaEmail bool,
) (targetWhatsApp bool, targetEmail bool) {
	if dest == nil {
		return usaWhatsApp, usaEmail
	}

	hasPhone := strings.TrimSpace(dest.GetString("telefone_e164")) != ""
	hasEmail := strings.TrimSpace(dest.GetString("email")) != ""

	switch {
	case hasPhone && !hasEmail:
		return usaWhatsApp, false
	case hasEmail && !hasPhone:
		return false, usaEmail
	default:
		return usaWhatsApp, usaEmail
	}
}

func (s *CampanhaService) jaFoiContactado(teamID, obraID, contatoTipo string) (bool, error) {
	tipo := normalizarContatoTipo(contatoTipo)
	if teamID == "" || obraID == "" || tipo == "" {
		return false, nil
	}

	return s.repo.ExistsContatoEnviado(teamID, obraID, tipo)
}

// AdicionarDestinatariosObrasPlus cria destinatários direto da base Obras Plus.
// Agora também respeita a flag "ocultar já contactados".
func (s *CampanhaService) AdicionarDestinatariosObrasPlus(
	teamID, campanhaID string,
	inputs []CampanhaDestinatarioInput,
	ocultarJaContactados bool,
) (int, int, error) {
	campanha, err := s.repo.FindByID(campanhaID)
	if err != nil {
		return 0, 0, g.Error(err, "campanha não encontrada")
	}

	if campanha.GetString("team_id") != teamID {
		return 0, 0, g.Error("não autorizado")
	}

	usaWhatsApp, usaEmail := s.getCampaignChannels(campanha)
	if !usaWhatsApp && !usaEmail {
		return 0, 0, g.Error("a campanha não possui canais configurados")
	}

	criados := 0
	ignorados := 0

	for _, item := range inputs {
		obraID := strings.TrimSpace(item.ObraID)
		contatoTipo := normalizarContatoTipo(item.ContatoTipo)

		if obraID == "" || contatoTipo == "" {
			ignorados++
			continue
		}

		if ocultarJaContactados {
			jaFoi, err := s.jaFoiContactado(teamID, obraID, contatoTipo)
			if err != nil {
				return criados, ignorados, g.Error(err, "erro ao verificar histórico de contato")
			}
			if jaFoi {
				ignorados++
				continue
			}
		}

		contato, err := s.BuscarContatoObraPorTipo(obraID, contatoTipo)
		if err != nil {
			g.Warn("erro ao resolver contato da obra %s (%s): %v", obraID, contatoTipo, err)
			ignorados++
			continue
		}
		if contato == nil {
			ignorados++
			continue
		}

		telefones := make([]string, 0)
		emails := make([]string, 0)

		if usaWhatsApp {
			telefones = contato.TelefonesE164
		}
		if usaEmail {
			emails = contato.Emails
		}

		criadosItem := 0

		for _, telefone := range telefones {
			telefone = s.formatPhone(telefone)
			if telefone == "" {
				ignorados++
				continue
			}

			existenteTelefone, err := s.repo.FindDestinatarioByCampanhaTelefone(campanhaID, telefone)
			if err != nil {
				return criados, ignorados, g.Error(err, "erro ao verificar destinatário duplicado por telefone")
			}
			if existenteTelefone != nil {
				ignorados++
				continue
			}

			existente, err := s.repo.FindDestinatarioByCampanhaContatoValor(
				campanhaID,
				contato.ObraID,
				contato.ContatoTipo,
				telefone,
				"",
			)
			if err != nil {
				return criados, ignorados, g.Error(err, "erro ao verificar destinatário existente (telefone)")
			}
			if existente != nil {
				ignorados++
				continue
			}

			_, err = s.repo.CreateDestinatario(map[string]any{
				"team_id":       teamID,
				"campanha_id":   campanhaID,
				"obra_id":       contato.ObraID,
				"contato_tipo":  contato.ContatoTipo,
				"nome_contato":  contato.NomeContato,
				"telefone_e164": telefone,
				"email":         "",
				"cidade":        contato.Cidade,
				"bairro":        contato.Bairro,
				"uf":            contato.UF,
				"address":       contato.Endereco,
				"status":        repositories.DestStatusPendente,
				"erro":          "",
				"tentativas":    0,
			})
			if err != nil {
				return criados, ignorados, g.Error(err, "erro ao criar destinatário de telefone")
			}

			criados++
			criadosItem++
		}

		for _, email := range emails {
			existente, err := s.repo.FindDestinatarioByCampanhaContatoValor(
				campanhaID,
				contato.ObraID,
				contato.ContatoTipo,
				"",
				email,
			)
			if err != nil {
				return criados, ignorados, g.Error(err, "erro ao verificar destinatário existente (email)")
			}
			if existente != nil {
				ignorados++
				continue
			}

			_, err = s.repo.CreateDestinatario(map[string]any{
				"team_id":       teamID,
				"campanha_id":   campanhaID,
				"obra_id":       contato.ObraID,
				"contato_tipo":  contato.ContatoTipo,
				"nome_contato":  contato.NomeContato,
				"telefone_e164": "",
				"email":         email,
				"cidade":        contato.Cidade,
				"bairro":        contato.Bairro,
				"uf":            contato.UF,
				"address":       contato.Endereco,
				"status":        repositories.DestStatusPendente,
				"erro":          "",
				"tentativas":    0,
			})
			if err != nil {
				return criados, ignorados, g.Error(err, "erro ao criar destinatário de email")
			}

			criados++
			criadosItem++
		}

		if criadosItem == 0 {
			semContatoExistente, err := s.repo.FindDestinatarioByCampanhaContatoValor(
				campanhaID,
				contato.ObraID,
				contato.ContatoTipo,
				"",
				"",
			)
			if err != nil {
				return criados, ignorados, g.Error(err, "erro ao verificar destinatário ignorado")
			}

			if semContatoExistente == nil {
				_, err = s.repo.CreateDestinatario(map[string]any{
					"team_id":       teamID,
					"campanha_id":   campanhaID,
					"obra_id":       contato.ObraID,
					"contato_tipo":  contato.ContatoTipo,
					"nome_contato":  contato.NomeContato,
					"telefone_e164": "",
					"email":         "",
					"cidade":        contato.Cidade,
					"bairro":        contato.Bairro,
					"uf":            contato.UF,
					"address":       contato.Endereco,
					"status":        repositories.DestStatusIgnorado,
					"erro":          "Sem contato disponível para os canais selecionados",
					"tentativas":    0,
				})
				if err != nil {
					return criados, ignorados, g.Error(err, "erro ao criar destinatário ignorado")
				}
			}

			ignorados++
		}
	}

	return criados, ignorados, nil
}

func appendMensagemNaConversa(
	conversa *models.Record,
	role string,
	content string,
	timestamp time.Time,
) {
	if conversa == nil {
		return
	}

	mensagens := make([]map[string]interface{}, 0)

	raw := conversa.Get("mensagens")
	switch v := raw.(type) {
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				mensagens = append(mensagens, map[string]interface{}{
					"role":      cast.ToString(m["role"]),
					"content":   cast.ToString(m["content"]),
					"timestamp": cast.ToString(m["timestamp"]),
				})
				continue
			}

			if m, ok := item.(map[string]any); ok {
				mensagens = append(mensagens, map[string]interface{}{
					"role":      cast.ToString(m["role"]),
					"content":   cast.ToString(m["content"]),
					"timestamp": cast.ToString(m["timestamp"]),
				})
			}
		}

	case []map[string]interface{}:
		for _, item := range v {
			mensagens = append(mensagens, map[string]interface{}{
				"role":      cast.ToString(item["role"]),
				"content":   cast.ToString(item["content"]),
				"timestamp": cast.ToString(item["timestamp"]),
			})
		}
	}

	mensagens = append(mensagens, map[string]interface{}{
		"role":      strings.TrimSpace(role),
		"content":   strings.TrimSpace(content),
		"timestamp": timestamp.UTC().Format(time.RFC3339),
	})

	conversa.Set("mensagens", mensagens)
	conversa.Set("ultima_mensagem_em", timestamp.UTC())
}

func (s *CampanhaService) garantirConversaIA(
	teamID, campanhaID string,
	dest *models.Record,
	telefone, nomeContato, mensagem string,
) error {
	if s.conversaRepo == nil {
		return nil
	}

	telefone = s.formatPhone(telefone)
	nomeContato = strings.TrimSpace(nomeContato)
	mensagem = strings.TrimSpace(mensagem)

	if teamID == "" || telefone == "" || mensagem == "" {
		return nil
	}

	now := time.Now().UTC()

	conversa, err := s.conversaRepo.FindByTelefone(teamID, telefone)
	if err != nil {
		return err
	}

	if conversa == nil {
		payload := map[string]interface{}{
			"team_id":            teamID,
			"campanha_id":        campanhaID,
			"telefone":           telefone,
			"nome_contato":       nomeContato,
			"mensagens":          []map[string]interface{}{},
			"status":             "ATIVA",
			"ultima_mensagem_em": now,
		}

		if dest != nil && dest.Id != "" {
			payload["destinatario_id"] = dest.Id

			if payload["nome_contato"] == "" {
				payload["nome_contato"] = strings.TrimSpace(dest.GetString("nome_contato"))
			}
		}

		conversa, err = s.conversaRepo.Create(payload)
		if err != nil {
			return err
		}
	} else {
		if campanhaID != "" {
			conversa.Set("campanha_id", campanhaID)
		}

		if dest != nil && dest.Id != "" {
			conversa.Set("destinatario_id", dest.Id)
		}

		if conversa.GetString("nome_contato") == "" && nomeContato != "" {
			conversa.Set("nome_contato", nomeContato)
		}

		if strings.ToUpper(strings.TrimSpace(conversa.GetString("status"))) != "ATIVA" {
			conversa.Set("status", "ATIVA")
		}
	}

	appendMensagemNaConversa(conversa, "assistant", mensagem, now)
	return s.conversaRepo.Save(conversa)
}

// ProcessarCampanhaAsync processa os envios de uma campanha em background.
// O loop busca destinatários PENDENTE continuamente (um a um) até que não
// sobre nenhum, garantindo que todos sejam processados mesmo em caso de erro.
// O intervalo de 2 minutos entre WhatsApp é POR CAMPANHA – cada campanha
// roda a sua própria goroutine, então os timers são independentes.
func (s *CampanhaService) ProcessarCampanhaAsync(campanhaID string) {
	// Garantir que a campanha seja finalizada mesmo em caso de panic
	defer func() {
		if r := recover(); r != nil {
			g.Error(g.Error("panic na goroutine da campanha %s: %v", campanhaID, r))
		}
		// Sempre tenta finalizar a campanha ao sair da goroutine
		s.finalizarCampanhaSeNecessario(campanhaID)
	}()

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

	usaWhatsApp, usaEmail := s.getCampaignChannels(campanha)

	g.Info("Processando campanha %s para team %s (WHATSAPP=%v EMAIL=%v delay_wa=%v)",
		campanhaID, teamID, usaWhatsApp, usaEmail, delayEntreEnviosWhatsApp)

	enviados := 0
	falhas := 0
	jaEnviadosPorTelefone := map[string]struct{}{}

	// Loop contínuo: busca PENDENTE → processa → espera → repete até acabar.
	for {
		// 1. Verificar se campanha foi pausada ou cancelada
		campanha, _ = s.repo.FindByID(campanhaID)
		if campanha == nil {
			g.Warn("Campanha %s não encontrada, encerrando", campanhaID)
			break
		}
		status := campanha.GetString("status")
		if status == repositories.CampanhaStatusPausada || status == repositories.CampanhaStatusCancelada {
			g.Info("Campanha %s foi %s, interrompendo (enviados=%d falhas=%d)", campanhaID, status, enviados, falhas)
			break
		}

		// 2. Buscar próximo destinatário PENDENTE
		pendentes, err := s.repo.FindDestinatariosPendentes(campanhaID)
		if err != nil {
			g.Error(err, "erro ao buscar destinatários pendentes da campanha %s", campanhaID)
			break
		}
		if len(pendentes) == 0 {
			g.Info("Campanha %s: nenhum destinatário PENDENTE restante", campanhaID)
			break
		}

		dest := pendentes[0] // processa um de cada vez

		g.Info("Campanha %s: processando destinatário %s (%d pendentes restantes)",
			campanhaID, dest.Id, len(pendentes))

		// 3. Resolver canais e dados do destinatário
		targetWhatsApp, targetEmail := s.getTargetChannelsForDest(dest, usaWhatsApp, usaEmail)

		nomeContato := strings.TrimSpace(dest.GetString("nome_contato"))
		telefone := strings.TrimSpace(dest.GetString("telefone_e164"))
		email := strings.TrimSpace(dest.GetString("email"))
		cidade := strings.TrimSpace(dest.GetString("cidade"))
		bairro := strings.TrimSpace(dest.GetString("bairro"))
		uf := strings.TrimSpace(dest.GetString("uf"))

		needsEnrichment := nomeContato == "" || cidade == "" || uf == ""
		if targetWhatsApp && telefone == "" {
			needsEnrichment = true
		}
		if targetEmail && email == "" {
			needsEnrichment = true
		}

		if needsEnrichment {
			changed, enrichErr := s.enriquecerDestinatarioComObra(dest, targetWhatsApp, targetEmail)
			if enrichErr != nil {
				g.Warn("erro ao enriquecer destinatário %s: %v", dest.Id, enrichErr)
			} else if changed {
				if err := s.repo.SaveDestinatario(dest); err != nil {
					g.Warn("erro ao salvar destinatário enriquecido %s: %v", dest.Id, err)
				}
			}

			nomeContato = strings.TrimSpace(dest.GetString("nome_contato"))
			telefone = strings.TrimSpace(dest.GetString("telefone_e164"))
			email = strings.TrimSpace(dest.GetString("email"))
			cidade = strings.TrimSpace(dest.GetString("cidade"))
			bairro = strings.TrimSpace(dest.GetString("bairro"))
			uf = strings.TrimSpace(dest.GetString("uf"))
		}

		if nomeContato == "" {
			nomeContato = "Cliente"
		}

		if targetWhatsApp && telefone != "" {
			telefone = s.formatPhone(telefone)
			if telefone == "" {
				targetWhatsApp = false
			}
		}

		// 4. Sem contato disponível → IGNORADO (pula sem delay)
		if !((!targetWhatsApp || telefone != "") && (!targetEmail || email != "")) {
			g.Warn("Destinatário %s ignorado: sem contato para os canais selecionados", dest.Id)
			if err := s.repo.UpdateDestinatarioStatus(dest, repositories.DestStatusIgnorado, "Sem contato para os canais selecionados"); err != nil {
				g.Error(err, "Erro ao atualizar status para IGNORADO do destinatário %s", dest.Id)
			}
			falhas++
			continue // não aplica delay – segue pro próximo
		}

		// 5. Personalizar mensagem
		mensagem := s.PersonalizarMensagem(mensagemTemplate, map[string]string{
			"nome":   nomeContato,
			"cidade": cidade,
			"bairro": bairro,
			"uf":     uf,
		})

		// 6. Enviar
		var sendErr error
		enviouAlgum := false
		enviouWhatsApp := false

		if targetWhatsApp && telefone != "" {
			if _, jaEnviadoNoLoop := jaEnviadosPorTelefone[telefone]; jaEnviadoNoLoop {
				g.Info("Destinatário %s ignorado: telefone %s já enviado nesta execução", dest.Id, telefone)
				if err := s.repo.UpdateDestinatarioStatus(dest, repositories.DestStatusIgnorado, "Número duplicado na campanha"); err != nil {
					g.Error(err, "Erro ao atualizar status para IGNORADO do destinatário %s", dest.Id)
				}
				continue
			}

			jaEnviadoAntes, err := s.repo.ExistsEnviadoByCampanhaTelefone(campanhaID, telefone, dest.Id)
			if err != nil {
				g.Warn("erro ao verificar telefone já enviado campanha=%s telefone=%s: %v", campanhaID, telefone, err)
			}
			if jaEnviadoAntes {
				g.Info("Destinatário %s ignorado: telefone %s já recebeu envio nesta campanha", dest.Id, telefone)
				if err := s.repo.UpdateDestinatarioStatus(dest, repositories.DestStatusIgnorado, "Número já enviado nesta campanha"); err != nil {
					g.Error(err, "Erro ao atualizar status para IGNORADO do destinatário %s", dest.Id)
				}
				continue
			}

			sendErr = s.enviarWhatsApp(teamID, telefone, mensagem)
			if sendErr == nil {
				enviouAlgum = true
				enviouWhatsApp = true
				jaEnviadosPorTelefone[telefone] = struct{}{}
				g.Info("WhatsApp enviado para %s", telefone)

				if manterIA && s.conversaRepo != nil {
					if err := s.garantirConversaIA(teamID, campanhaID, dest, telefone, nomeContato, mensagem); err != nil {
						g.Warn("manter_ia: erro ao garantir conversa para %s: %v", telefone, err)
					}
				}
			} else {
				g.Error(sendErr, "Falha WhatsApp para %s (destinatário %s)", telefone, dest.Id)
			}
		}

		if targetEmail && email != "" {
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

		// 7. Atualizar status do destinatário
		if !enviouAlgum {
			errMsg := "Sem contato para os canais selecionados"
			if sendErr != nil {
				errMsg = sendErr.Error()
			}
			g.Warn("Destinatário %s marcado como FALHOU: %s", dest.Id, errMsg)
			if err := s.repo.UpdateDestinatarioStatus(dest, repositories.DestStatusFalhou, errMsg); err != nil {
				g.Error(err, "Erro ao atualizar status para FALHOU do destinatário %s", dest.Id)
			}
			falhas++
			// Não aplica delay de 2 min em caso de falha – segue para o próximo
			continue
		}

		if err := s.repo.UpdateDestinatarioStatus(dest, repositories.DestStatusEnviado, ""); err != nil {
			g.Error(err, "Erro ao atualizar status para ENVIADO do destinatário %s", dest.Id)
		}
		enviados++

		g.Info(
			"Campanha %s: enviado %d (falhas=%d) obra=%s tipo=%s cidade=%s bairro=%s uf=%s telefone=%s email=%s",
			campanhaID,
			enviados,
			falhas,
			dest.GetString("obra_id"),
			dest.GetString("contato_tipo"),
			cidade,
			bairro,
			uf,
			telefone,
			email,
		)

		// 8. Aguardar 2 minutos entre envios de WhatsApp (por campanha).
		//    Se o envio foi apenas email, continua sem delay.
		//    Verifica se ainda há pendentes ANTES de esperar para não aguardar
		//    2 minutos desnecessários após o último envio.
		if enviouWhatsApp {
			proximosPendentes, _ := s.repo.FindDestinatariosPendentes(campanhaID)
			if len(proximosPendentes) == 0 {
				g.Info("Campanha %s: último envio concluído, sem mais pendentes", campanhaID)
				break
			}
			g.Info("Campanha %s: aguardando %v antes do próximo envio WhatsApp (%d pendentes)...",
				campanhaID, delayEntreEnviosWhatsApp, len(proximosPendentes))
			if s.waitWithCancelCheck(campanhaID, delayEntreEnviosWhatsApp) {
				g.Info("Campanha %s foi pausada/cancelada durante espera", campanhaID)
				break
			}
		}
	}

	g.Info("Campanha %s: loop encerrado — enviados=%d falhas=%d", campanhaID, enviados, falhas)
	// A finalização real é feita no defer (finalizarCampanhaSeNecessario)
}

// restante do arquivo continua igual...
func (s *CampanhaService) BuscarContatoObraPorTipo(obraID, contatoTipo string) (*ObraContatoCampanha, error) {
	tipo := normalizarContatoTipo(contatoTipo)
	if tipo == "" {
		return nil, g.Error("contato_tipo inválido")
	}

	obra, err := s.buscarObraBase(obraID)
	if err != nil {
		return nil, err
	}
	if obra == nil {
		return nil, nil
	}

	nome := ""
	switch tipo {
	case ContatoTipoOwner:
		nome = strings.TrimSpace(obra.Owner)
	case ContatoTipoProfessional:
		nome = strings.TrimSpace(obra.Professional)
	}

	if nome == "" {
		return nil, nil
	}

	telefones := s.buscarTelefones(nome, obra.Cidade, obra.UF)
	emails := s.buscarEmails(nome, obra.Cidade, obra.UF)

	return &ObraContatoCampanha{
		ObraID:        obraID,
		ContatoTipo:   tipo,
		NomeContato:   nome,
		TelefonesE164: telefones,
		Emails:        emails,
		Cidade:        obra.Cidade,
		Bairro:        obra.Bairro,
		UF:            obra.UF,
		Endereco:      obra.Address,
	}, nil
}

func (s *CampanhaService) BuscarContatoObra(obraID string) (map[string]string, error) {
	obra, err := s.buscarObraBase(obraID)
	if err != nil {
		return nil, err
	}
	if obra == nil {
		return nil, nil
	}

	if obra.Owner == "" && obra.Professional == "" {
		return nil, nil
	}

	nome := strings.TrimSpace(obra.Owner)
	contatoTipo := ContatoTipoOwner

	telefones := make([]string, 0)
	emails := make([]string, 0)

	if nome != "" {
		telefones = s.buscarTelefones(nome, obra.Cidade, obra.UF)
		emails = s.buscarEmails(nome, obra.Cidade, obra.UF)
	}

	if len(telefones) == 0 && len(emails) == 0 && strings.TrimSpace(obra.Professional) != "" {
		nome = strings.TrimSpace(obra.Professional)
		contatoTipo = ContatoTipoProfessional
		telefones = s.buscarTelefones(nome, obra.Cidade, obra.UF)
		emails = s.buscarEmails(nome, obra.Cidade, obra.UF)
	}

	if len(telefones) == 0 && len(emails) == 0 {
		return nil, nil
	}

	telefone := ""
	email := ""

	if len(telefones) > 0 {
		telefone = telefones[0]
	}
	if len(emails) > 0 {
		email = emails[0]
	}

	return map[string]string{
		"nome":         nome,
		"telefone":     telefone,
		"email":        email,
		"cidade":       obra.Cidade,
		"bairro":       obra.Bairro,
		"uf":           obra.UF,
		"address":      obra.Address,
		"contato_tipo": contatoTipo,
	}, nil
}

func (s *CampanhaService) buscarObraBase(obraID string) (*obraBase, error) {
	sqlObra := `
		SELECT
			owner,
			professional,
			address,
			bairro,
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

	return &obraBase{
		Owner:        cast.ToString(obra["owner"]),
		Professional: cast.ToString(obra["professional"]),
		Address:      cast.ToString(obra["address"]),
		Bairro:       cast.ToString(obra["bairro"]),
		Cidade:       cast.ToString(obra["city"]),
		UF:           cast.ToString(obra["state"]),
	}, nil
}

func (s *CampanhaService) EnriquecerDestinatarios(teamID, campanhaID string) (int, int, error) {
	campanha, err := s.repo.FindByID(campanhaID)
	if err != nil {
		return 0, 0, g.Error(err, "campanha não encontrada")
	}

	if campanha.GetString("team_id") != teamID {
		return 0, 0, g.Error("não autorizado")
	}

	usaWhatsApp, usaEmail := s.getCampaignChannels(campanha)

	destinatarios, err := s.repo.FindDestinatariosPendentes(campanhaID)
	if err != nil {
		return 0, 0, g.Error(err, "erro ao buscar destinatários")
	}

	enriquecidos := 0
	semContato := 0

	for _, dest := range destinatarios {
		targetWhatsApp, targetEmail := s.getTargetChannelsForDest(dest, usaWhatsApp, usaEmail)

		nome := strings.TrimSpace(dest.GetString("nome_contato"))
		telefone := strings.TrimSpace(dest.GetString("telefone_e164"))
		email := strings.TrimSpace(dest.GetString("email"))
		cidade := strings.TrimSpace(dest.GetString("cidade"))
		uf := strings.TrimSpace(dest.GetString("uf"))

		needsEnrichment := nome == "" || cidade == "" || uf == ""
		if targetWhatsApp && telefone == "" {
			needsEnrichment = true
		}
		if targetEmail && email == "" {
			needsEnrichment = true
		}

		if !needsEnrichment {
			enriquecidos++
			continue
		}

		changed, err := s.enriquecerDestinatarioComObra(dest, targetWhatsApp, targetEmail)
		if err != nil {
			semContato++
			continue
		}

		if changed {
			if err := s.repo.SaveDestinatario(dest); err != nil {
				g.Error(err, "erro ao salvar destinatário %s", dest.Id)
				continue
			}
		}

		telefone = strings.TrimSpace(dest.GetString("telefone_e164"))
		email = strings.TrimSpace(dest.GetString("email"))

		if (!targetWhatsApp || telefone != "") && (!targetEmail || email != "") {
			enriquecidos++
		} else {
			semContato++
		}
	}

	g.Info("Enriquecimento campanha %s: %d enriquecidos, %d sem contato", campanhaID, enriquecidos, semContato)
	return enriquecidos, semContato, nil
}

func (s *CampanhaService) enriquecerDestinatarioComObra(
	dest *models.Record,
	fillPhone bool,
	fillEmail bool,
) (bool, error) {
	if dest == nil {
		return false, nil
	}

	obraID := strings.TrimSpace(dest.GetString("obra_id"))
	if obraID == "" {
		return false, nil
	}

	contatoTipo := normalizarContatoTipo(dest.GetString("contato_tipo"))

	var contato *ObraContatoCampanha
	var err error

	if contatoTipo != "" {
		contato, err = s.BuscarContatoObraPorTipo(obraID, contatoTipo)
		if err != nil {
			return false, err
		}
	} else {
		m, err := s.BuscarContatoObra(obraID)
		if err != nil {
			return false, err
		}
		if m != nil {
			telefones := make([]string, 0, 1)
			emails := make([]string, 0, 1)

			if m["telefone"] != "" {
				telefones = append(telefones, m["telefone"])
			}
			if m["email"] != "" {
				emails = append(emails, m["email"])
			}

			contato = &ObraContatoCampanha{
				ObraID:        obraID,
				ContatoTipo:   m["contato_tipo"],
				NomeContato:   m["nome"],
				TelefonesE164: telefones,
				Emails:        emails,
				Cidade:        m["cidade"],
				Bairro:        m["bairro"],
				UF:            m["uf"],
				Endereco:      m["address"],
			}
		}
	}

	if contato == nil {
		return false, nil
	}

	changed := false

	if dest.GetString("contato_tipo") == "" && contato.ContatoTipo != "" {
		dest.Set("contato_tipo", contato.ContatoTipo)
		changed = true
	}
	if dest.GetString("nome_contato") == "" && contato.NomeContato != "" {
		dest.Set("nome_contato", contato.NomeContato)
		changed = true
	}
	if fillPhone && dest.GetString("telefone_e164") == "" && len(contato.TelefonesE164) > 0 {
		dest.Set("telefone_e164", contato.TelefonesE164[0])
		changed = true
	}
	if fillEmail && dest.GetString("email") == "" && len(contato.Emails) > 0 {
		dest.Set("email", contato.Emails[0])
		changed = true
	}
	if dest.GetString("cidade") == "" && contato.Cidade != "" {
		dest.Set("cidade", contato.Cidade)
		changed = true
	}
	if dest.GetString("bairro") == "" && contato.Bairro != "" {
		dest.Set("bairro", contato.Bairro)
		changed = true
	}
	if dest.GetString("uf") == "" && contato.UF != "" {
		dest.Set("uf", contato.UF)
		changed = true
	}
	if dest.GetString("address") == "" && contato.Endereco != "" {
		dest.Set("address", contato.Endereco)
		changed = true
	}

	return changed, nil
}

func (s *CampanhaService) buscarTelefones(nome, cidade, uf string) []string {
	sql := `
		WITH prep AS (
			SELECT
				telefone,
				cidade,
				uf,
				row_number() OVER(PARTITION BY telefone ORDER BY COALESCE(poder_aquisitivo, 1) DESC) AS row_num
			FROM core.core_obras_plus_phone
			WHERE nome = {:nome}
		)
		SELECT telefone
		FROM prep
		WHERE row_num = 1
		ORDER BY
			CASE WHEN uf = {:uf} AND cidade = {:cidade} THEN 1 ELSE 2 END ASC,
			telefone DESC
		LIMIT {:limite}
	`

	boundSQL := store.BindSQL(sql, map[string]any{
		"nome":   nome,
		"cidade": cidade,
		"uf":     uf,
		"limite": maxTelefonesPorContato,
	})

	data, err := store.MainDB.Query(boundSQL)
	if err != nil {
		return []string{}
	}

	recs := data.RecordsCasted(true)
	if len(recs) == 0 {
		return []string{}
	}

	telefones := make([]string, 0, len(recs))
	for _, rec := range recs {
		telefone := s.formatPhone(cast.ToString(rec["telefone"]))
		if telefone != "" {
			telefones = append(telefones, telefone)
		}
	}

	return uniqueNonEmpty(telefones)
}

func (s *CampanhaService) buscarEmails(nome, cidade, uf string) []string {
	sql := `
		WITH prep AS (
			SELECT
				email,
				cidade,
				uf,
				poder_aquisitivo,
				row_number() OVER(PARTITION BY email ORDER BY COALESCE(poder_aquisitivo, 1) DESC) AS row_num
			FROM core.core_obras_plus_email
			WHERE nome = {:nome}
		)
		SELECT email
		FROM prep
		WHERE row_num = 1
		ORDER BY
			CASE WHEN uf = {:uf} AND cidade = {:cidade} THEN 1 ELSE 2 END ASC,
			poder_aquisitivo DESC
		LIMIT {:limite}
	`

	boundSQL := store.BindSQL(sql, map[string]any{
		"nome":   nome,
		"cidade": cidade,
		"uf":     uf,
		"limite": maxEmailsPorContato,
	})

	data, err := store.MainDB.Query(boundSQL)
	if err != nil {
		g.Warn("Erro ao buscar emails para %s: %v", nome, err)
		return []string{}
	}

	recs := data.RecordsCasted(true)
	if len(recs) == 0 {
		return []string{}
	}

	emails := make([]string, 0, len(recs))
	for _, rec := range recs {
		email := normalizeEmail(cast.ToString(rec["email"]))
		if email != "" {
			emails = append(emails, email)
		}
	}

	return uniqueNonEmpty(emails)
}

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
			"{{" + key + "}}",
			"{{ " + key + " }}",
		}
		for _, v := range variants {
			msg = strings.ReplaceAll(msg, v, val)
		}
	}

	primeiroNome := strings.Split(strings.TrimSpace(nome), " ")[0]

	replace("NOME", nome)
	replace("primeiroNome", primeiroNome)
	replace("SAUDACAO", saudacao)
	replace("CIDADE", contato["cidade"])
	replace("BAIRRO", contato["bairro"])
	replace("UF", contato["uf"])

	return msg
}

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

func (s *CampanhaService) formatPhone(phone string) string {
	result := ""
	for _, c := range phone {
		if c >= '0' && c <= '9' {
			result += string(c)
		}
	}
	if len(result) == 10 || len(result) == 11 {
		result = "55" + result
	}
	return result
}

func (s *CampanhaService) enviarWhatsApp(teamID, telefone, mensagem string) error {
	phone := s.formatPhone(telefone)
	_, err := s.waSvc.SendTestMessage(teamID, phone, mensagem)
	return err
}

// finalizarCampanhaSeNecessario verifica se a campanha ainda está EM_ANDAMENTO
// e, se não houver mais destinatários PENDENTE, atualiza o status para CONCLUIDA.
// Chamada sempre no defer de ProcessarCampanhaAsync para garantir que a campanha
// não fique presa como EM_ANDAMENTO mesmo em caso de panic ou erro inesperado.
func (s *CampanhaService) finalizarCampanhaSeNecessario(campanhaID string) {
	campanha, err := s.repo.FindByID(campanhaID)
	if err != nil {
		g.Error(err, "finalizarCampanha: erro ao buscar campanha %s", campanhaID)
		return
	}
	if campanha == nil {
		g.Warn("finalizarCampanha: campanha %s não encontrada", campanhaID)
		return
	}

	status := campanha.GetString("status")
	if status != repositories.CampanhaStatusEmAndamento {
		g.Info("finalizarCampanha: campanha %s já está como %s, nenhuma ação necessária", campanhaID, status)
		return
	}

	// Verificar se realmente não há mais pendentes
	pendentes, _ := s.repo.FindDestinatariosPendentes(campanhaID)
	if len(pendentes) > 0 {
		g.Warn("finalizarCampanha: campanha %s ainda tem %d pendentes — mantendo EM_ANDAMENTO", campanhaID, len(pendentes))
		return
	}

	if err := s.repo.UpdateStatus(campanha, repositories.CampanhaStatusConcluida); err != nil {
		g.Error(err, "finalizarCampanha: ERRO ao atualizar campanha %s para CONCLUIDA", campanhaID)
		return
	}
	g.Info("finalizarCampanha: campanha %s atualizada para CONCLUIDA com sucesso ✓", campanhaID)
}

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

func (s *CampanhaService) enviarEmail(teamID, email, assunto, mensagem string) error {
	if s.emailSvc == nil {
		return g.Error("serviço de email não configurado")
	}
	return s.emailSvc.SendTestEmail(teamID, email, assunto, mensagem)
}

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

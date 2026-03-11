package services

import (
	"strings"
	"time"

	"github.com/flarco/g"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"
	"github.com/spf13/cast"
	"github.com/suaobra/suaobra-app/server/repositories"
)

// IAConversacionalService gerencia conversas de IA com leads
type IAConversacionalService struct {
	dao          *daos.Dao
	intencaoRepo *repositories.IntencaoRepo
	conversaRepo *repositories.ConversaRepo
	geminiSvc    *GeminiService
	whatsappSvc  *WhatsAppService
}

// NewIAConversacionalService cria uma nova instância do serviço
func NewIAConversacionalService(
	dao *daos.Dao,
	intencaoRepo *repositories.IntencaoRepo,
	conversaRepo *repositories.ConversaRepo,
	geminiSvc *GeminiService,
	whatsappSvc *WhatsAppService,
) *IAConversacionalService {
	return &IAConversacionalService{
		dao:          dao,
		intencaoRepo: intencaoRepo,
		conversaRepo: conversaRepo,
		geminiSvc:    geminiSvc,
		whatsappSvc:  whatsappSvc,
	}
}

// ProcessarMensagemRecebida processa uma mensagem recebida do lead
func (s *IAConversacionalService) ProcessarMensagemRecebida(
	teamID string,
	telefone string,
	mensagem string,
	nomeContato string,
) error {
	g.Info("IA: Processando mensagem de %s (%s): %s", nomeContato, telefone, mensagem)

	// Busca ou cria conversa
	conversa, err := s.conversaRepo.FindByTelefone(teamID, telefone)
	if err != nil {
		// Conversa não existe, cria nova e tenta associar com campanha
		conversa, err = s.criarNovaConversa(teamID, telefone, nomeContato, mensagem)
		if err != nil {
			return g.Error(err, "erro ao criar nova conversa")
		}
	} else {
		// Adiciona mensagem ao histórico
		if err := s.adicionarMensagemAoHistorico(conversa, "user", mensagem); err != nil {
			return g.Error(err, "erro ao adicionar mensagem ao histórico")
		}
	}

	// Busca intenções ativas da team
	intencoes, err := s.intencaoRepo.FindAtivasByTeamID(teamID)
	if err != nil {
		g.Warn("Erro ao buscar intenções para team %s: %v", teamID, err)
		intencoes = []*models.Record{}
	}

	// Tenta fazer match com intenções
	intencaoMatch, resposta := s.buscarIntencaoMatch(mensagem, intencoes)

	var respostaIA string
	var intencaoDetectadaID string

	if intencaoMatch != nil {
		// Intenção encontrada - usa resposta configurada
		respostaIA = resposta
		intencaoDetectadaID = intencaoMatch.Id
		g.Info("IA: Intenção detectada '%s' para telefone %s", intencaoMatch.GetString("nome"), telefone)
	} else {
		// Sem match de intenção - usa Gemini para resposta conversacional
		g.Info("IA: Nenhuma intenção detectada, usando Gemini para telefone %s", telefone)

		// Passa as intenções para construir contexto mais rico
		contexto := s.construirContextoNegocio(conversa, intencoes)
		historico := s.extrairHistoricoGemini(conversa)

		g.Debug("IA: Contexto para Gemini: %s", contexto)
		g.Debug("IA: Histórico tem %d mensagens", len(historico))
		// Se o histórico está vazio, tenta buscar interações salvas (mensagens da campanha)
		if len(historico) == 0 {
			// Busca destinatário vinculado à conversa
			destID := conversa.GetString("destinatario_id")
			var interacoes []string
			if destID != "" {
				// Busca registro do destinatário
				dest, err := s.dao.FindRecordById("campanha_destinatarios", destID)
				if err == nil && dest != nil {
					// Busca campo de interações salvas (ex: mensagens_enviadas)
					if msgsRaw := dest.Get("mensagens_enviadas"); msgsRaw != nil {
						if arr, ok := msgsRaw.([]interface{}); ok {
							for _, m := range arr {
								if str, ok := m.(string); ok && str != "" {
									interacoes = append(interacoes, str)
								}
							}
						}
					}
				}
			}
			// Se encontrou interações, adiciona ao histórico
			for _, msg := range interacoes {
				historico = append(historico, GeminiMessage{Role: "user", Content: msg})
			}
			// Sempre adiciona a mensagem atual do usuário
			historico = append(historico, GeminiMessage{Role: "user", Content: mensagem})
			g.Debug("IA: Histórico reconstruído, agora tem %d mensagens", len(historico))
		}

		respostaIA, err = s.geminiSvc.GenerateConversationalResponse(historico, contexto, 0.8)
		if err != nil {
			g.Warn("IA: ERRO Gemini para telefone %s: %v", telefone, err)
			respostaIA = "Desculpe, estou com dificuldades técnicas no momento. Em breve retornarei o contato! 😊"
		} else {
			g.Debug("IA: Gemini retornou resposta: %s", respostaIA)
		}
	}

	// Adiciona resposta da IA ao histórico
	if err := s.adicionarMensagemAoHistorico(conversa, "model", respostaIA); err != nil {
		return g.Error(err, "erro ao adicionar resposta ao histórico")
	}

	// Atualiza conversa
	updates := map[string]interface{}{
		"ultima_mensagem_em": time.Now().UTC(),
	}
	if intencaoDetectadaID != "" {
		updates["intencao_detectada"] = intencaoDetectadaID
	}
	if err := s.conversaRepo.Update(conversa, updates); err != nil {
		g.Error(err, "Erro ao atualizar conversa %s", conversa.Id)
	}

	// Envia resposta via WhatsApp
	if err := s.enviarRespostaWhatsApp(teamID, telefone, respostaIA); err != nil {
		return g.Error(err, "erro ao enviar resposta WhatsApp")
	}

	g.Info("IA: Resposta enviada para %s (%s)", nomeContato, telefone)
	return nil
}

// criarNovaConversa cria uma nova conversa de IA
func (s *IAConversacionalService) criarNovaConversa(
	teamID, telefone, nomeContato, primeiraMensagem string,
) (*models.Record, error) {
	mensagens := []map[string]interface{}{
		{
			"role":      "user",
			"content":   primeiraMensagem,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	}

	data := map[string]interface{}{
		"team_id":            teamID,
		"telefone":           telefone,
		"nome_contato":       nomeContato,
		"mensagens":          mensagens,
		"status":             "ATIVA",
		"ultima_mensagem_em": time.Now().UTC(),
	}

	// Tenta encontrar um destinatário recente com esse telefone para associar campanha
	dest, err := s.dao.FindFirstRecordByFilter(
		"campanha_destinatarios",
		"team_id = {:teamId} && telefone_e164 = {:telefone} && status = 'ENVIADO'",
		dbx.Params{
			"teamId":   teamID,
			"telefone": telefone,
		},
	)
	if err == nil && dest != nil {
		data["campanha_id"] = dest.GetString("campanha_id")
		data["destinatario_id"] = dest.Id
		data["lead_id"] = dest.GetString("lead_id")

		// Usa o nome do destinatário se não tiver
		if nomeContato == "" {
			data["nome_contato"] = dest.GetString("nome_contato")
		}
		g.Info("IA: Nova conversa associada à campanha %s", dest.GetString("campanha_id"))
	}

	return s.conversaRepo.Create(data)
}

// adicionarMensagemAoHistorico adiciona uma mensagem ao histórico da conversa
func (s *IAConversacionalService) adicionarMensagemAoHistorico(
	conversa *models.Record,
	role string,
	content string,
) error {
	mensagensRaw := conversa.Get("mensagens")
	mensagens := []map[string]interface{}{}

	if mensagensRaw != nil {
		if arr, ok := mensagensRaw.([]interface{}); ok {
			for _, item := range arr {
				if m, ok := item.(map[string]interface{}); ok {
					mensagens = append(mensagens, m)
				}
			}
		}
	}

	mensagens = append(mensagens, map[string]interface{}{
		"role":      role,
		"content":   content,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})

	conversa.Set("mensagens", mensagens)
	return s.conversaRepo.Save(conversa)
}

// buscarIntencaoMatch busca a melhor intenção que combina com a mensagem
func (s *IAConversacionalService) buscarIntencaoMatch(
	mensagem string,
	intencoes []*models.Record,
) (*models.Record, string) {
	var melhorIntencao *models.Record
	var melhorScore float64
	var melhorResposta string

	for _, intencao := range intencoes {
		palavrasChaveRaw := intencao.Get("palavras_chave")
		palavrasChave := ExtrairPalavrasChave(palavrasChaveRaw)

		match, score := MatchIntencao(mensagem, palavrasChave)
		if match && score > melhorScore {
			melhorScore = score
			melhorIntencao = intencao
			melhorResposta = intencao.GetString("resposta")
		}
	}

	// Só retorna se o score for >= 0.5 (pelo menos metade das palavras-chave)
	if melhorScore >= 0.5 {
		return melhorIntencao, melhorResposta
	}

	return nil, ""
}

// extrairHistoricoGemini extrai histórico de mensagens no formato Gemini
func (s *IAConversacionalService) extrairHistoricoGemini(conversa *models.Record) []GeminiMessage {
	mensagensRaw := conversa.Get("mensagens")
	historico := []GeminiMessage{}

	if mensagensRaw == nil {
		return historico
	}

	if arr, ok := mensagensRaw.([]interface{}); ok {
		// Limita a últimas 20 mensagens para não estourar token
		start := 0
		if len(arr) > 20 {
			start = len(arr) - 20
		}

		for i := start; i < len(arr); i++ {
			if m, ok := arr[i].(map[string]interface{}); ok {
				role := cast.ToString(m["role"])
				content := cast.ToString(m["content"])
				if role != "" && content != "" {
					historico = append(historico, GeminiMessage{
						Role:    role,
						Content: content,
					})
				}
			}
		}
	}

	return historico
}

// construirContextoNegocio constrói contexto do negócio para a IA
func (s *IAConversacionalService) construirContextoNegocio(conversa *models.Record, intencoes []*models.Record) string {
	var partes []string

	// Nome do contato
	nomeContato := conversa.GetString("nome_contato")
	if nomeContato != "" {
		partes = append(partes, g.F("O nome do cliente é: %s", nomeContato))
	}

	// Busca informações da campanha se existir
	campanhaID := conversa.GetString("campanha_id")
	if campanhaID != "" {
		campanha, err := s.dao.FindRecordById("campanhas", campanhaID)
		if err == nil && campanha != nil {
			mensagemOriginal := campanha.GetString("mensagem_template")
			if mensagemOriginal != "" {
				partes = append(partes, g.F("A mensagem original enviada para o cliente foi:\n---\n%s\n---", mensagemOriginal))
			}
		}
	}

	// Busca informações do destinatário se existir
	destinatarioID := conversa.GetString("destinatario_id")
	if destinatarioID != "" {
		dest, err := s.dao.FindRecordById("campanha_destinatarios", destinatarioID)
		if err == nil && dest != nil {
			// Pode buscar mais contexto do lead se necessário
			leadID := dest.GetString("lead_id")
			if leadID != "" {
				lead, err := s.dao.FindRecordById("lead", leadID)
				if err == nil && lead != nil {
					obraID := lead.GetString("obra_id")
					if obraID != "" {
						partes = append(partes, g.F("Este é um proprietário de obra (ID: %s). Você pode mencionar que viu a obra dele e está oferecendo serviços relacionados.", obraID))
					}
				}
			}
		}
	}

	// Adiciona intenções configuradas como guia
	if len(intencoes) > 0 {
		var intencoesDesc []string
		for _, intencao := range intencoes {
			nome := intencao.GetString("nome")
			descricao := intencao.GetString("descricao")
			resposta := intencao.GetString("resposta")
			if nome != "" {
				info := g.F("- %s", nome)
				if descricao != "" {
					info += g.F(": %s", descricao)
				}
				if resposta != "" {
					info += g.F(" (responda algo como: '%s')", resposta)
				}
				intencoesDesc = append(intencoesDesc, info)
			}
		}
		if len(intencoesDesc) > 0 {
			partes = append(partes, g.F("Tópicos que você pode abordar baseado nas perguntas do cliente:\n%s", strings.Join(intencoesDesc, "\n")))
		}
	}

	// Contexto padrão se não tiver nada
	if len(partes) == 0 {
		return "Você está realizando atendimento e qualificação de leads. Seja cordial e profissional."
	}

	return strings.Join(partes, "\n\n")
}

// enviarRespostaWhatsApp envia a resposta via WhatsApp
func (s *IAConversacionalService) enviarRespostaWhatsApp(teamID, telefone, mensagem string) error {
	_, err := s.whatsappSvc.SendTestMessage(teamID, telefone, mensagem)
	if err != nil {
		return g.Error(err, "falha ao enviar mensagem WhatsApp")
	}
	return nil
}

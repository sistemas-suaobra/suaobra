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

type IAConversacionalService struct {
	dao          *daos.Dao
	intencaoRepo *repositories.IntencaoRepo
	conversaRepo *repositories.ConversaRepo
	whatsappSvc  *WhatsAppService
	geminiSvc    *GeminiService
}

func NewIAConversacionalService(
	dao *daos.Dao,
	intencaoRepo *repositories.IntencaoRepo,
	conversaRepo *repositories.ConversaRepo,
	whatsappSvc *WhatsAppService,
	geminiSvc *GeminiService,
) *IAConversacionalService {
	return &IAConversacionalService{
		dao:          dao,
		intencaoRepo: intencaoRepo,
		conversaRepo: conversaRepo,
		whatsappSvc:  whatsappSvc,
		geminiSvc:    geminiSvc,
	}
}

// ProcessarMensagemRecebida processa uma mensagem recebida do lead
func (s *IAConversacionalService) ProcessarMensagemRecebida(
	teamID string,
	telefone string,
	mensagem string,
	nomeContato string,
) error {
	startedAt := time.Now()

	teamID = strings.TrimSpace(teamID)
	telefone = s.normalizarTelefoneE164(telefone)
	mensagem = strings.TrimSpace(mensagem)
	nomeContato = strings.TrimSpace(nomeContato)

	g.Info(
		"IA: START team=%s telefone=%s nome=%s msg_len=%d mensagem=%s",
		teamID,
		maskIAPhone(telefone),
		nomeContato,
		len(mensagem),
		truncateIAContent(mensagem, 1200),
	)

	if teamID == "" || telefone == "" || mensagem == "" {
		g.Warn(
			"IA: abortando por dados obrigatórios ausentes team=%s telefone=%s msg_len=%d",
			teamID,
			maskIAPhone(telefone),
			len(mensagem),
		)
		return nil
	}

	conversa, err := s.conversaRepo.FindByTelefone(teamID, telefone)
	if err != nil {
		g.Error(err, "IA: erro ao buscar conversa team=%s telefone=%s", teamID, maskIAPhone(telefone))
		return g.Error(err, "erro ao buscar conversa")
	}

	if conversa == nil {
		g.Info("IA: nenhuma conversa encontrada, criando nova team=%s telefone=%s", teamID, maskIAPhone(telefone))
		conversa, err = s.criarNovaConversa(teamID, telefone, nomeContato, mensagem)
		if err != nil {
			g.Error(err, "IA: erro ao criar nova conversa team=%s telefone=%s", teamID, maskIAPhone(telefone))
			return g.Error(err, "erro ao criar nova conversa")
		}
		g.Info("IA: conversa criada id=%s status=%s", conversa.Id, conversa.GetString("status"))
	} else {
		g.Info(
			"IA: conversa encontrada id=%s status=%s historico_itens=%d",
			conversa.Id,
			conversa.GetString("status"),
			len(s.obterMensagens(conversa)),
		)

		if strings.ToUpper(strings.TrimSpace(conversa.GetString("status"))) != "ATIVA" {
			g.Warn(
				"IA: conversa %s não está ativa (status=%s), ignorando resposta automática",
				conversa.Id,
				conversa.GetString("status"),
			)
			return nil
		}

		if conversa.GetString("nome_contato") == "" && nomeContato != "" {
			conversa.Set("nome_contato", nomeContato)
		}

		if err := s.adicionarMensagemAoHistorico(conversa, "user", mensagem); err != nil {
			g.Error(err, "IA: erro ao adicionar mensagem do usuário no histórico conversa=%s", conversa.Id)
			return g.Error(err, "erro ao adicionar mensagem ao histórico")
		}

		g.Info("IA: mensagem do usuário adicionada ao histórico conversa=%s", conversa.Id)
	}

	intencoes, err := s.intencaoRepo.FindAtivasByTeamID(teamID)
	if err != nil {
		g.Warn("IA: erro ao buscar intenções para team %s: %v", teamID, err)
		intencoes = []*models.Record{}
	}
	g.Info("IA: total de intenções ativas team=%s count=%d", teamID, len(intencoes))

	var respostaIA string
	var intencaoDetectadaID string

	intencaoMatch, resposta := s.buscarIntencaoMatch(conversa, mensagem, intencoes)
	if intencaoMatch != nil {
		respostaIA = resposta
		intencaoDetectadaID = intencaoMatch.Id
		g.Info(
			"IA: intenção detectada id=%s nome=%s resposta=%s",
			intencaoMatch.Id,
			intencaoMatch.GetString("nome"),
			truncateIAContent(respostaIA, 1000),
		)
	} else if s.geminiSvc != nil {
		g.Info("IA: nenhuma intenção detectada, usando Gemini telefone=%s", maskIAPhone(telefone))

		contexto := s.construirContextoNegocio(conversa, intencoes)
		historico := s.extrairHistoricoGemini(conversa)

		g.Info(
			"IA: Gemini contexto_len=%d historico_msgs=%d",
			len(contexto),
			len(historico),
		)

		if len(historico) == 0 {
			historico = append(historico, GeminiMessage{
				Role:    "user",
				Content: mensagem,
			})
		}

		respostaIA, err = s.geminiSvc.GenerateConversationalResponse(historico, contexto, 0.8)
		if err != nil {
			g.Error(err, "IA: erro Gemini telefone=%s", maskIAPhone(telefone))
			respostaIA = "Desculpe, estou com dificuldades técnicas no momento. Em breve retornarei o contato."
		} else {
			g.Info("IA: Gemini respondeu telefone=%s resposta=%s", maskIAPhone(telefone), truncateIAContent(respostaIA, 1000))
		}
	} else {
		g.Warn("IA: geminiSvc nil, usando fallback telefone=%s", maskIAPhone(telefone))
		respostaIA = "Recebi sua mensagem. Vou te responder em instantes."
	}

	respostaIA = strings.TrimSpace(respostaIA)
	if respostaIA == "" {
		g.Warn("IA: resposta vazia telefone=%s", maskIAPhone(telefone))
		return nil
	}

	g.Info("IA: enviando resposta WhatsApp telefone=%s resposta=%s", maskIAPhone(telefone), truncateIAContent(respostaIA, 1000))
	if err := s.enviarRespostaWhatsApp(teamID, telefone, respostaIA); err != nil {
		g.Error(err, "IA: falha ao enviar resposta WhatsApp telefone=%s", maskIAPhone(telefone))
		return g.Error(err, "erro ao enviar resposta WhatsApp")
	}
	g.Info("IA: resposta enviada no WhatsApp telefone=%s", maskIAPhone(telefone))

	if err := s.adicionarMensagemAoHistorico(conversa, "model", respostaIA); err != nil {
		g.Error(err, "IA: erro ao adicionar resposta ao histórico conversa=%s", conversa.Id)
		return g.Error(err, "erro ao adicionar resposta ao histórico")
	}

	conversa.Set("ultima_mensagem_em", time.Now().UTC())
	if intencaoDetectadaID != "" {
		conversa.Set("intencao_detectada", intencaoDetectadaID)
	}

	if err := s.conversaRepo.Save(conversa); err != nil {
		g.Error(err, "IA: erro ao salvar conversa %s", conversa.Id)
	}

	g.Info(
		"IA: END conversa=%s telefone=%s duration=%s",
		conversa.Id,
		maskIAPhone(telefone),
		time.Since(startedAt),
	)
	return nil
}

// criarNovaConversa cria uma nova conversa de IA
func (s *IAConversacionalService) criarNovaConversa(
	teamID, telefone, nomeContato, primeiraMensagem string,
) (*models.Record, error) {
	now := time.Now().UTC()
	telefone = s.normalizarTelefoneE164(telefone)

	mensagens := []map[string]interface{}{
		{
			"role":      "user",
			"content":   strings.TrimSpace(primeiraMensagem),
			"timestamp": now.Format(time.RFC3339),
		},
	}

	data := map[string]interface{}{
		"team_id":            teamID,
		"telefone":           telefone,
		"nome_contato":       strings.TrimSpace(nomeContato),
		"mensagens":          mensagens,
		"status":             "ATIVA",
		"ultima_mensagem_em": now,
	}

	destinatarios, err := s.dao.FindRecordsByFilter(
		"campanha_destinatarios",
		"team_id = {:teamId} && telefone_e164 = {:telefone} && status = 'ENVIADO'",
		"-enviado_em,-updated",
		1,
		0,
		dbx.Params{
			"teamId":   teamID,
			"telefone": telefone,
		},
	)
	if err == nil && len(destinatarios) > 0 {
		dest := destinatarios[0]

		data["campanha_id"] = dest.GetString("campanha_id")
		data["destinatario_id"] = dest.Id

		if cast.ToString(data["nome_contato"]) == "" {
			data["nome_contato"] = strings.TrimSpace(dest.GetString("nome_contato"))
		}

		g.Info(
			"IA: nova conversa associada à campanha %s e destinatário %s",
			dest.GetString("campanha_id"),
			dest.Id,
		)
	}

	return s.conversaRepo.Create(data)
}

// adicionarMensagemAoHistorico adiciona uma mensagem ao histórico da conversa
func (s *IAConversacionalService) adicionarMensagemAoHistorico(
	conversa *models.Record,
	role string,
	content string,
) error {
	mensagens := s.obterMensagens(conversa)

	mensagens = append(mensagens, map[string]interface{}{
		"role":      strings.TrimSpace(role),
		"content":   strings.TrimSpace(content),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})

	conversa.Set("mensagens", mensagens)
	conversa.Set("ultima_mensagem_em", time.Now().UTC())

	return s.conversaRepo.Save(conversa)
}

// buscarIntencaoMatch busca a melhor intenção que combina com a mensagem
func (s *IAConversacionalService) buscarIntencaoMatch(
	conversa *models.Record,
	mensagem string,
	intencoes []*models.Record,
) (*models.Record, string) {
	var melhorIntencao *models.Record
	var melhorScore float64
	var melhorResposta string
	var melhorPrioridade int

	vars := s.carregarVariaveisConversa(conversa)

	for _, intencao := range intencoes {
		palavrasChaveRaw := intencao.Get("palavras_chave")
		palavrasChave := ExtrairPalavrasChave(palavrasChaveRaw)

		match, score := MatchIntencao(mensagem, palavrasChave)
		if !match {
			continue
		}

		prioridade := intencao.GetInt("prioridade")
		resposta := s.personalizarRespostaIntencao(intencao.GetString("resposta"), vars)

		if score > melhorScore || (score == melhorScore && prioridade > melhorPrioridade) {
			melhorScore = score
			melhorIntencao = intencao
			melhorResposta = resposta
			melhorPrioridade = prioridade
		}
	}

	if melhorScore >= 0.5 {
		return melhorIntencao, melhorResposta
	}

	return nil, ""
}

// extrairHistoricoGemini extrai histórico de mensagens no formato Gemini
func (s *IAConversacionalService) extrairHistoricoGemini(conversa *models.Record) []GeminiMessage {
	mensagens := s.obterMensagens(conversa)
	historico := []GeminiMessage{}

	if len(mensagens) == 0 {
		return historico
	}

	start := 0
	if len(mensagens) > 20 {
		start = len(mensagens) - 20
	}

	for i := start; i < len(mensagens); i++ {
		role := cast.ToString(mensagens[i]["role"])
		content := cast.ToString(mensagens[i]["content"])

		if role == "" || content == "" {
			continue
		}

		geminiRole := "user"
		if role == "model" || role == "assistant" {
			geminiRole = "model"
		}

		historico = append(historico, GeminiMessage{
			Role:    geminiRole,
			Content: content,
		})
	}

	return historico
}

// construirContextoNegocio constrói contexto do negócio para a IA
func (s *IAConversacionalService) construirContextoNegocio(conversa *models.Record, intencoes []*models.Record) string {
	var partes []string

	nomeContato := ""
	if conversa != nil {
		nomeContato = strings.TrimSpace(conversa.GetString("nome_contato"))
	}
	if nomeContato != "" {
		partes = append(partes, g.F("Nome do contato: %s", nomeContato))
	}

	if conversa != nil {
		campanhaID := strings.TrimSpace(conversa.GetString("campanha_id"))
		if campanhaID != "" {
			campanha, err := s.dao.FindRecordById("campanhas", campanhaID)
			if err == nil && campanha != nil {
				mensagemOriginal := strings.TrimSpace(campanha.GetString("mensagem_template"))
				if mensagemOriginal != "" {
					partes = append(partes, g.F("Mensagem inicial enviada na campanha:\n%s", mensagemOriginal))
				}
			}
		}

		destinatarioID := strings.TrimSpace(conversa.GetString("destinatario_id"))
		if destinatarioID != "" {
			dest, err := s.dao.FindRecordById("campanha_destinatarios", destinatarioID)
			if err == nil && dest != nil {
				if v := strings.TrimSpace(dest.GetString("nome_contato")); v != "" {
					partes = append(partes, g.F("Nome salvo do contato: %s", v))
				}
				if v := strings.TrimSpace(dest.GetString("cidade")); v != "" {
					partes = append(partes, g.F("Cidade: %s", v))
				}
				if v := strings.TrimSpace(dest.GetString("bairro")); v != "" {
					partes = append(partes, g.F("Bairro: %s", v))
				}
				if v := strings.TrimSpace(dest.GetString("uf")); v != "" {
					partes = append(partes, g.F("UF: %s", v))
				}
				if v := strings.TrimSpace(dest.GetString("address")); v != "" {
					partes = append(partes, g.F("Endereço: %s", v))
				}
				if v := strings.TrimSpace(dest.GetString("obra_id")); v != "" {
					partes = append(partes, g.F("Obra ID interno: %s", v))
				}
				if v := strings.TrimSpace(dest.GetString("contato_tipo")); v != "" {
					partes = append(partes, g.F("Tipo de contato: %s", v))
				}
			}
		}
	}

	if len(intencoes) > 0 {
		var intencoesDesc []string
		for _, intencao := range intencoes {
			nome := strings.TrimSpace(intencao.GetString("nome"))
			descricao := strings.TrimSpace(intencao.GetString("descricao"))
			resposta := strings.TrimSpace(intencao.GetString("resposta"))

			if nome == "" {
				continue
			}

			info := g.F("- %s", nome)
			if descricao != "" {
				info += g.F(": %s", descricao)
			}
			if resposta != "" {
				info += g.F(" (responda algo próximo de: '%s')", resposta)
			}

			intencoesDesc = append(intencoesDesc, info)
		}

		if len(intencoesDesc) > 0 {
			partes = append(partes, g.F("Intenções configuradas:\n%s", strings.Join(intencoesDesc, "\n")))
		}
	}

	if len(partes) == 0 {
		return "Você está realizando atendimento e qualificação de leads. Seja cordial e profissional."
	}

	return strings.Join(partes, "\n\n")
}

// enviarRespostaWhatsApp envia a resposta via WhatsApp
func (s *IAConversacionalService) enviarRespostaWhatsApp(teamID, telefone, mensagem string) error {
	_, err := s.whatsappSvc.SendTestMessage(teamID, s.normalizarTelefoneE164(telefone), mensagem)
	if err != nil {
		return g.Error(err, "falha ao enviar mensagem WhatsApp")
	}
	return nil
}

func (s *IAConversacionalService) obterMensagens(conversa *models.Record) []map[string]interface{} {
	mensagens := []map[string]interface{}{}
	if conversa == nil {
		return mensagens
	}

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
		return mensagens

	case []map[string]interface{}:
		for _, item := range v {
			mensagens = append(mensagens, map[string]interface{}{
				"role":      cast.ToString(item["role"]),
				"content":   cast.ToString(item["content"]),
				"timestamp": cast.ToString(item["timestamp"]),
			})
		}
		return mensagens
	}

	return mensagens
}

func (s *IAConversacionalService) carregarVariaveisConversa(conversa *models.Record) map[string]string {
	vars := map[string]string{
		"nome":    "",
		"cidade":  "",
		"bairro":  "",
		"uf":      "",
		"address": "",
	}

	if conversa == nil {
		return vars
	}

	vars["nome"] = strings.TrimSpace(conversa.GetString("nome_contato"))

	destinatarioID := strings.TrimSpace(conversa.GetString("destinatario_id"))
	if destinatarioID == "" {
		return vars
	}

	dest, err := s.dao.FindRecordById("campanha_destinatarios", destinatarioID)
	if err != nil || dest == nil {
		return vars
	}

	if vars["nome"] == "" {
		vars["nome"] = strings.TrimSpace(dest.GetString("nome_contato"))
	}
	vars["cidade"] = strings.TrimSpace(dest.GetString("cidade"))
	vars["bairro"] = strings.TrimSpace(dest.GetString("bairro"))
	vars["uf"] = strings.TrimSpace(dest.GetString("uf"))
	vars["address"] = strings.TrimSpace(dest.GetString("address"))

	return vars
}

func (s *IAConversacionalService) personalizarRespostaIntencao(template string, vars map[string]string) string {
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

	nome := strings.TrimSpace(vars["nome"])
	primeiroNome := strings.Split(nome, " ")[0]

	replace("NOME", nome)
	replace("primeiroNome", primeiroNome)
	replace("CIDADE", vars["cidade"])
	replace("BAIRRO", vars["bairro"])
	replace("UF", vars["uf"])
	replace("ADDRESS", vars["address"])

	return strings.TrimSpace(msg)
}

func (s *IAConversacionalService) normalizarTelefoneE164(phone string) string {
	var digits strings.Builder
	for _, c := range phone {
		if c >= '0' && c <= '9' {
			digits.WriteRune(c)
		}
	}

	result := digits.String()
	if len(result) == 10 || len(result) == 11 {
		result = "55" + result
	}

	return result
}

func truncateIAContent(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "...(truncated)"
}

func maskIAPhone(phone string) string {
	var digits strings.Builder
	for _, c := range phone {
		if c >= '0' && c <= '9' {
			digits.WriteRune(c)
		}
	}
	v := digits.String()
	if len(v) <= 4 {
		return v
	}
	return strings.Repeat("*", len(v)-4) + v[len(v)-4:]
}

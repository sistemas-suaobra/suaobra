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

// ultimasRespostasModel retorna as últimas N mensagens enviadas pela IA (role "model").
func (s *IAConversacionalService) ultimasRespostasModel(conversa *models.Record, n int) []string {
	mensagens := s.obterMensagens(conversa)
	var respostas []string
	for i := len(mensagens) - 1; i >= 0 && len(respostas) < n; i-- {
		if cast.ToString(mensagens[i]["role"]) == "model" {
			c := strings.TrimSpace(cast.ToString(mensagens[i]["content"]))
			if c != "" {
				respostas = append(respostas, c)
			}
		}
	}
	return respostas
}

// respostaSimilar verifica se `nova` é muito parecida com alguma resposta anterior.
func respostaSimilar(nova string, anteriores []string) bool {
	novaNorm := normalizarTextoMatch(nova)
	if novaNorm == "" {
		return false
	}
	for _, ant := range anteriores {
		antNorm := normalizarTextoMatch(ant)
		if antNorm == "" {
			continue
		}
		if novaNorm == antNorm {
			return true
		}
		// Se >70% das palavras são iguais, considerar similar
		novaPalavras := strings.Fields(novaNorm)
		antPalavras := strings.Fields(antNorm)
		if len(novaPalavras) == 0 || len(antPalavras) == 0 {
			continue
		}
		antSet := map[string]struct{}{}
		for _, p := range antPalavras {
			antSet[p] = struct{}{}
		}
		common := 0
		for _, p := range novaPalavras {
			if _, ok := antSet[p]; ok {
				common++
			}
		}
		ratio := float64(common) / float64(len(novaPalavras))
		if ratio > 0.7 {
			return true
		}
	}
	return false
}

// ProcessarMensagemRecebida processa uma mensagem recebida do lead.
// messageIDExterno é o ID da mensagem no provedor (ex.: WhatsApp), para deduplicação em campanha_lead_respostas.
func (s *IAConversacionalService) ProcessarMensagemRecebida(
	teamID string,
	telefone string,
	mensagem string,
	nomeContato string,
	messageIDExterno string,
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

	candidatosTelefone := s.buildPhoneCandidates(telefone)
	if !IAAtivaParaTelefone(s.dao, teamID, candidatosTelefone) {
		g.Info(
			"IA: campanha sem manter_ia para telefone=%s team=%s — ignorando resposta automática",
			maskIAPhone(telefone),
			teamID,
		)
		return nil
	}

	// Gerar candidatos de telefone para busca mais robusta
	g.Info("IA: candidatos telefone=%v", candidatosTelefone)

	conversa := s.conversaRepo.FindByTelefoneCandidates(teamID, candidatosTelefone)

	var err error
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

	_ = repositories.SaveCampanhaLeadResposta(s.dao, repositories.CampanhaLeadRespostaInput{
		TeamID:           teamID,
		CampanhaID:       strings.TrimSpace(conversa.GetString("campanha_id")),
		DestinatarioID:   strings.TrimSpace(conversa.GetString("destinatario_id")),
		ConversaID:       conversa.Id,
		Canal:            "WHATSAPP",
		TelefoneE164:     telefone,
		NomeContato:      nomeContato,
		Corpo:            mensagem,
		MessageIDExterno: strings.TrimSpace(messageIDExterno),
		RecebidaEm:       time.Now().UTC(),
	})

	intencoes, err := s.intencaoRepo.FindAtivasByTeamID(teamID)
	if err != nil {
		g.Warn("IA: erro ao buscar intenções para team %s: %v", teamID, err)
		intencoes = []*models.Record{}
	}
	g.Info("IA: total de intenções ativas team=%s count=%d", teamID, len(intencoes))

	var respostaIA string
	var intencaoDetectadaID string

	anteriores := s.ultimasRespostasModel(conversa, 3)

	intencaoMatch, resposta := s.buscarIntencaoMatch(conversa, mensagem, intencoes)
	if intencaoMatch != nil {
		intencaoDetectadaID = intencaoMatch.Id
		g.Info(
			"IA: intenção detectada id=%s nome=%s resposta=%s",
			intencaoMatch.Id,
			intencaoMatch.GetString("nome"),
			truncateIAContent(resposta, 1000),
		)

		if respostaSimilar(resposta, anteriores) && s.geminiSvc != nil {
			g.Info("IA: resposta da intenção é repetida, delegando ao Gemini para variar")
			contexto := s.construirContextoNegocio(conversa, intencoes)
			contexto += "\n\nIMPORTANTE: A intenção detectada é '" + intencaoMatch.GetString("nome") +
				"'. A resposta-base é: '" + resposta +
				"'. Reformule de forma diferente e NÃO repita frases anteriores."
			historico := s.extrairHistoricoGemini(conversa)
			if len(historico) == 0 {
				historico = append(historico, GeminiMessage{Role: "user", Content: mensagem})
			}
			reformulada, err := s.geminiSvc.GenerateConversationalResponse(historico, contexto, 0.9)
			if err == nil && strings.TrimSpace(reformulada) != "" {
				respostaIA = reformulada
			} else {
				respostaIA = resposta
			}
		} else {
			respostaIA = resposta
		}
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

		if respostaSimilar(respostaIA, anteriores) {
			g.Info("IA: Gemini gerou resposta repetida, tentando novamente com temperatura maior")
			contexto += "\n\nNÃO repita nenhuma das frases anteriores. Avance a conversa."
			retry, retryErr := s.geminiSvc.GenerateConversationalResponse(historico, contexto, 1.0)
			if retryErr == nil && strings.TrimSpace(retry) != "" && !respostaSimilar(retry, anteriores) {
				respostaIA = retry
			}
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

	if !s.conversaPermiteRespostaAutomatica(conversa.Id) {
		g.Info(
			"IA: resposta automática abortada por intervenção manual conversa=%s telefone=%s",
			conversa.Id,
			maskIAPhone(telefone),
		)
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

	// Buscar destinatário usando múltiplos candidatos de telefone
	candidatos := s.buildPhoneCandidates(telefone)
	var destMatch *models.Record
	for _, candidato := range candidatos {
		recs, findErr := s.dao.FindRecordsByFilter(
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
		if findErr == nil && len(recs) > 0 {
			destMatch = recs[0]
			break
		}
	}

	if destMatch != nil {
		data["campanha_id"] = destMatch.GetString("campanha_id")
		data["destinatario_id"] = destMatch.Id

		if cast.ToString(data["nome_contato"]) == "" {
			data["nome_contato"] = strings.TrimSpace(destMatch.GetString("nome_contato"))
		}

		g.Info(
			"IA: nova conversa associada à campanha %s e destinatário %s",
			destMatch.GetString("campanha_id"),
			destMatch.Id,
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

// buscarIntencaoMatch busca a melhor intenção que combina com a mensagem.
// Se houver qualquer match de palavra-chave, retorna a melhor intenção (sem threshold mínimo).
// Prioridade desempata scores iguais.
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

	if melhorIntencao != nil {
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
		return "Você está realizando atendimento e qualificação de leads. Seja cordial e profissional. Não comece com saudação se a conversa já tem histórico."
	}

	return strings.Join(partes, "\n\n") + "\n\nLEMBRETE: Nunca repita suas respostas anteriores. Avance a conversa."
}

// enviarRespostaWhatsApp envia a resposta via WhatsApp
func (s *IAConversacionalService) enviarRespostaWhatsApp(teamID, telefone, mensagem string) error {
	_, err := s.whatsappSvc.SendTestMessage(teamID, s.normalizarTelefoneE164(telefone), mensagem)
	if err != nil {
		return g.Error(err, "falha ao enviar mensagem WhatsApp")
	}
	return nil
}

func (s *IAConversacionalService) conversaPermiteRespostaAutomatica(conversaID string) bool {
	conversaID = strings.TrimSpace(conversaID)
	if conversaID == "" {
		return false
	}

	rec, err := s.dao.FindRecordById("conversas_ia", conversaID)
	if err != nil || rec == nil {
		return false
	}

	status := strings.ToUpper(strings.TrimSpace(rec.GetString("status")))
	return status == "ATIVA"
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

// buildPhoneCandidates gera variações de número de telefone para matching
// robusto: com/sem código de país 55 e com/sem o 9º dígito brasileiro.
func (s *IAConversacionalService) buildPhoneCandidates(phone string) []string {
	phone = s.normalizarTelefoneE164(phone)
	if phone == "" {
		return nil
	}

	seen := map[string]struct{}{}
	out := make([]string, 0, 6)

	add := func(v string) {
		v = s.normalizarTelefoneE164(v)
		if v == "" {
			return
		}
		if _, ok := seen[v]; ok {
			return
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}

	add(phone)

	// Com/sem código de país 55
	if strings.HasPrefix(phone, "55") && len(phone) > 10 {
		add(phone[2:])
	}
	if !strings.HasPrefix(phone, "55") && len(phone) >= 10 {
		add("55" + phone)
	}

	// Variantes do 9º dígito brasileiro
	withCC := phone
	if !strings.HasPrefix(withCC, "55") && len(withCC) >= 10 {
		withCC = "55" + withCC
	}
	if strings.HasPrefix(withCC, "55") && len(withCC) >= 12 {
		ddd := withCC[2:4]
		local := withCC[4:]

		if len(local) == 9 && local[0] == '9' {
			add("55" + ddd + local[1:])
		} else if len(local) == 8 {
			add("55" + ddd + "9" + local)
		}
	}

	return out
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

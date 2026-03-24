package services

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/flarco/g"
	"github.com/flarco/g/net"
	"github.com/labstack/echo/v5"
	"github.com/spf13/cast"
	"github.com/suaobra/suaobra-app/server/config"
)

var keywordSplitRegex = regexp.MustCompile(`[,\n;|]+`)

// GeminiService integra com a API do Google Gemini para geração de respostas conversacionais
type GeminiService struct {
	apiKey  string
	apiBase string
	model   string
	headers map[string]string
}

func NewGeminiService() *GeminiService {
	apiKey := strings.TrimSpace(config.EnvOr("GEMINI_API_KEY", ""))
	if apiKey == "" {
		g.Warn("GEMINI_API_KEY não configurada - serviço de IA desabilitado")
	}

	model := strings.TrimSpace(config.EnvOr("GEMINI_MODEL", "gemini-2.5-flash-lite"))

	return &GeminiService{
		apiKey:  apiKey,
		apiBase: "https://generativelanguage.googleapis.com/v1beta",
		model:   model,
		headers: map[string]string{
			echo.HeaderContentType: "application/json",
			"x-goog-api-key":       apiKey,
		},
	}
}

type GeminiMessage struct {
	Role    string `json:"role"`    // "user" ou "model"
	Content string `json:"content"` // texto da mensagem
}

func (s *GeminiService) generateContentURL() string {
	return fmt.Sprintf("%s/models/%s:generateContent", strings.TrimRight(s.apiBase, "/"), s.model)
}

func (s *GeminiService) GenerateConversationalResponse(
	historico []GeminiMessage,
	contextoNegocio string,
	temperatura float64,
) (string, error) {
	if strings.TrimSpace(s.apiKey) == "" {
		return "", g.Error("GEMINI_API_KEY não configurada")
	}

	if temperatura <= 0 {
		temperatura = 0.8
	}

	hasHistory := len(historico) > 1

	systemPrompt := `Você é um atendente/vendedor por WhatsApp.

Regras ABSOLUTAS (violar = erro):
1. NUNCA comece com saudação se já houver histórico. Proibido: "Olá", "Oi", "Tudo bem?", "Bom dia/tarde/noite", "Como vai?" e variações. Vá direto ao ponto.
2. NUNCA repita uma resposta que você já deu no histórico. Leia o que já disse e avance a conversa.
3. Máximo 2 frases + 1 pergunta. Sem listas, sem markdown, sem asteriscos.
4. Máximo 1 emoji (opcional).
5. Nunca diga: "vi que", "notei que", "pelo que vi", "você é proprietário", "sua obra".
6. Não explique contexto do lead. Use dados (nome/cidade/bairro) só se necessário.
7. Foque em qualificar o lead: entender interesse, necessidade, orçamento, prazo.
8. Se o lead já demonstrou interesse, avance para próximo passo (agendar visita, enviar proposta, etc.).
`

	if !hasHistory {
		systemPrompt += "\nEstado: PRIMEIRA mensagem da conversa. Pode cumprimentar UMA vez de forma breve e ir direto ao assunto.\n"
	} else {
		systemPrompt += "\nEstado: CONVERSA JÁ EM ANDAMENTO. Proibido qualquer cumprimento. Vá direto ao assunto.\n"
	}

	// Injetar últimas respostas do model para o Gemini saber o que NÃO repetir
	var respostasAnteriores []string
	for _, h := range historico {
		if h.Role == "model" && strings.TrimSpace(h.Content) != "" {
			respostasAnteriores = append(respostasAnteriores, h.Content)
		}
	}
	if len(respostasAnteriores) > 0 {
		lastN := respostasAnteriores
		if len(lastN) > 3 {
			lastN = lastN[len(lastN)-3:]
		}
		systemPrompt += "\nSuas últimas respostas (NÃO repita nenhuma delas, reformule ou avance):\n"
		for i, r := range lastN {
			systemPrompt += fmt.Sprintf("  %d. \"%s\"\n", i+1, r)
		}
	}

	if strings.TrimSpace(contextoNegocio) != "" {
		systemPrompt += "\nContexto interno (não revele):\n" + strings.TrimSpace(contextoNegocio) + "\n"
	}

	contents := make([]map[string]interface{}, 0, len(historico))
	for _, msg := range historico {
		role := msg.Role
		if role != "user" && role != "model" {
			role = "user"
		}

		txt := strings.TrimSpace(msg.Content)
		if txt == "" {
			continue
		}

		contents = append(contents, map[string]interface{}{
			"role": role,
			"parts": []map[string]interface{}{
				{"text": txt},
			},
		})
	}

	if len(contents) == 0 {
		return "", g.Error("histórico vazio")
	}

	payload := g.M(
		"system_instruction", g.M(
			"parts", []map[string]interface{}{
				{"text": systemPrompt},
			},
		),
		"contents", contents,
		"generationConfig", g.M(
			"temperature", temperatura,
			"topK", 35,
			"topP", 0.89,
			"maxOutputTokens", 220,
		),
	)

	url := s.generateContentURL()
	g.Debug("Gemini: chamando API url=%s model=%s", url, s.model)

	resp, respBytes, err := net.ClientDo("POST", url, strings.NewReader(g.Marshal(payload)), s.headers)
	if err != nil {
		g.Warn("Gemini: erro na requisição HTTP: %v", err)
		return "", g.Error(err, "falha ao chamar Gemini API")
	}

	g.Debug("Gemini: resposta status=%d len=%d", resp.StatusCode, len(respBytes))

	if resp.StatusCode >= 400 {
		body := string(respBytes)
		g.Warn("Gemini API erro status=%d body=%s", resp.StatusCode, body)

		if resp.StatusCode == 404 {
			return "", g.Error("Gemini API 404: model inválido ou não suporta generateContent. Confira GEMINI_MODEL e/ou use /v1beta/models para listar.")
		}
		if resp.StatusCode == 429 {
			return "", g.Error("Gemini API rate limit atingido (429) - aguarde alguns segundos")
		}
		return "", g.Error("Gemini API retornou erro: status %d body=%s", resp.StatusCode, body)
	}

	respMap, err := g.UnmarshalMap(string(respBytes))
	if err != nil {
		return "", g.Error(err, "falha ao parsear resposta Gemini")
	}

	candidates, ok := respMap["candidates"].([]interface{})
	if !ok || len(candidates) == 0 {
		return "", g.Error("Gemini não retornou candidatos")
	}

	candidate, ok := candidates[0].(map[string]interface{})
	if !ok {
		return "", g.Error("estrutura de candidato inválida")
	}

	content, ok := candidate["content"].(map[string]interface{})
	if !ok {
		return "", g.Error("sem content no candidato")
	}

	parts, ok := content["parts"].([]interface{})
	if !ok || len(parts) == 0 {
		return "", g.Error("sem parts no content")
	}

	part, ok := parts[0].(map[string]interface{})
	if !ok {
		return "", g.Error("estrutura de part inválida")
	}

	text, ok := part["text"].(string)
	if !ok {
		return "", g.Error("sem texto na resposta")
	}

	return strings.TrimSpace(text), nil
}

func (s *GeminiService) GenerateSimpleResponse(
	mensagemUsuario string,
	contexto string,
	temperatura float64,
) (string, error) {
	historico := []GeminiMessage{
		{Role: "user", Content: mensagemUsuario},
	}
	return s.GenerateConversationalResponse(historico, contexto, temperatura)
}

func (s *GeminiService) GenerateCampaignMessage(
	objetivo string,
	temperatura float64,
) (string, error) {
	contexto := strings.TrimSpace(config.EnvOr("GEMINI_BUSINESS_CONTEXT", ""))

	if temperatura <= 0 {
		temperatura = 0.6
	}

	// ✅ Aqui está o pulo do gato: objetivo vira instrução de template final
	userPrompt := fmt.Sprintf(
		`Crie UMA única mensagem pronta para enviar no WhatsApp para um lead.

Objetivo: %s

Regras:
- Sem "que ótimo", sem "prazer", sem enrolação.
- 1 a 2 frases curtas, máximo 220 caracteres.
- No máximo 1 emoji (opcional).
- Use variáveis: {{nome}} {{primeiroNome}} {{cidade}} {{bairro}} (se fizer sentido).
- Termine com 1 pergunta curta de qualificação (prazo/orçamento/quantidade).
- Retorne APENAS o texto da mensagem (sem título, sem aspas).`,
		strings.TrimSpace(objetivo),
	)

	return s.GenerateSimpleResponse(userPrompt, contexto, temperatura)
}

func (s *GeminiService) ListModels() (string, error) {
	if strings.TrimSpace(s.apiKey) == "" {
		return "", g.Error("GEMINI_API_KEY não configurada")
	}

	url := fmt.Sprintf("%s/models", strings.TrimRight(s.apiBase, "/"))
	resp, b, err := net.ClientDo("GET", url, nil, s.headers)
	if err != nil {
		return "", g.Error(err, "falha ao listar modelos")
	}
	if resp.StatusCode >= 400 {
		return "", g.Error("erro ao listar modelos: status %d body=%s", resp.StatusCode, string(b))
	}
	return string(b), nil
}

func MatchIntencao(mensagem string, palavrasChave []string) (bool, float64) {
	if len(palavrasChave) == 0 {
		return false, 0
	}

	mensagemNorm := normalizarTextoMatch(mensagem)
	if mensagemNorm == "" {
		return false, 0
	}

	msgTokens := strings.Fields(mensagemNorm)
	matches := 0

	for _, palavra := range palavrasChave {
		kwNorm := normalizarTextoMatch(palavra)
		if kwNorm == "" {
			continue
		}

		// 1) substring direta (ex: "orcamento" dentro de "quero um orcamento")
		if strings.Contains(mensagemNorm, kwNorm) {
			matches++
			continue
		}

		// 2) qualquer token da mensagem está contido na keyword ou vice-versa
		for _, tok := range msgTokens {
			if len(tok) < 3 {
				continue
			}
			if strings.Contains(kwNorm, tok) || strings.Contains(tok, kwNorm) {
				matches++
				break
			}
		}
	}

	if matches == 0 {
		return false, 0
	}

	score := float64(matches) / float64(len(palavrasChave))

	// Bonus agressivo: 1 match já deve ser suficiente para ativar a intenção.
	if matches >= 1 && score < 0.5 {
		score = 0.5 + score*0.5
	}
	if score > 1 {
		score = 1
	}

	return true, score
}

func ExtrairPalavrasChave(jsonData interface{}) []string {
	palavras := []string{}

	switch v := jsonData.(type) {
	case []interface{}:
		for _, item := range v {
			if str := cast.ToString(item); str != "" {
				palavras = append(palavras, str)
			}
		}
	case []string:
		palavras = v
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return palavras
		}

		var arr []interface{}
		if err := g.JSONUnmarshal([]byte(raw), &arr); err == nil {
			for _, item := range arr {
				if str := cast.ToString(item); str != "" {
					palavras = append(palavras, str)
				}
			}
			return limparPalavrasChave(palavras)
		}

		// Fallback para texto simples: "orcamento, preço, valor"
		parts := keywordSplitRegex.Split(raw, -1)
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				palavras = append(palavras, p)
			}
		}
	}

	return limparPalavrasChave(palavras)
}

func limparPalavrasChave(in []string) []string {
	if len(in) == 0 {
		return in
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, p := range in {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		key := normalizarTextoMatch(p)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, p)
	}
	return out
}

func normalizarTextoMatch(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return ""
	}

	replacer := strings.NewReplacer(
		"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"í", "i", "ì", "i", "î", "i", "ï", "i",
		"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
		"ú", "u", "ù", "u", "û", "u", "ü", "u",
		"ç", "c",
	)
	s = replacer.Replace(s)

	// Normaliza pontuação comum para espaço.
	for _, ch := range []string{".", ",", "!", "?", ":", ";", "(", ")", "\"", "'", "\t", "\r", "\n"} {
		s = strings.ReplaceAll(s, ch, " ")
	}
	return strings.Join(strings.Fields(s), " ")
}

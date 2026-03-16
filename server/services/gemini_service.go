package services

import (
	"fmt"
	"strings"

	"github.com/flarco/g"
	"github.com/flarco/g/net"
	"github.com/labstack/echo/v5"
	"github.com/spf13/cast"
	"github.com/suaobra/suaobra-app/server/config"
)

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

	Regras obrigatórias:
	- Não recomece conversa. Assuma que já estamos falando.
	- Se a conversa já tiver histórico, NÃO cumprimente (sem "Olá", "Oi", "bom dia/tarde/noite").
	- Nunca diga: "vi que", "notei que", "pelo que vi", "você é proprietário", "sua obra".
	- Não explique contexto do lead. Use dados (nome/cidade/bairro) só se ajudarem e de forma natural.
	- Resposta curta: máximo 2 frases + 1 pergunta.
	- Máximo 1 emoji (opcional).
	- Sem listas e sem markdown.
	`

	if !hasHistory {
		// primeira mensagem: pode cumprimentar 1 vez (sem "vi que...")
		systemPrompt += "\nEstado: primeira mensagem. Pode cumprimentar UMA vez, mas direto.\n"
	} else {
		systemPrompt += "\nEstado: conversa em andamento. Proibido cumprimentar.\n"
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

	mensagemLower := strings.ToLower(strings.TrimSpace(mensagem))
	matches := 0

	for _, palavra := range palavrasChave {
		palavraLower := strings.ToLower(strings.TrimSpace(palavra))
		if palavraLower == "" {
			continue
		}
		if strings.Contains(mensagemLower, palavraLower) {
			matches++
		}
	}

	if matches == 0 {
		return false, 0
	}

	score := float64(matches) / float64(len(palavrasChave))
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
		var arr []interface{}
		if err := g.JSONUnmarshal([]byte(v), &arr); err == nil {
			for _, item := range arr {
				if str := cast.ToString(item); str != "" {
					palavras = append(palavras, str)
				}
			}
		}
	}

	return palavras
}

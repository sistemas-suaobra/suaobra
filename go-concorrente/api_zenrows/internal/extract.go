package api_zenrows

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// ObraData representa os dados extraídos para cadastro
type ObraData struct {
	Id               string
	ObraNumber       string
	Professional     string
	Owner            string
	Address          string
	Bairro           string
	City             string
	State            string
	StartDate        string
	EndDate          string
	FirstListingDate string
	Activity         string
	Type             string
	Size             float64
	Unidade          string
}

// Extrai dados estruturados do texto HTML
func ExtractObraData(html string) (*ObraData, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(html)))
	if err != nil {
		logExtraction("", "error", err.Error(), nil)
		return nil, err
	}

	// Helper to find the text in <b> within a <span> with an exact label match.
	getTextFromBold := func(sel string) string {
		var result string
		doc.Find("span").EachWithBreak(func(i int, s *goquery.Selection) bool {
			// Extract the text of the span itself, ignoring children like <b>.
			// This prevents partial matches (e.g., "Logradouro:" matching "Tipo de Logradouro:").
			labelText := s.Contents().Not("b").Text()
			if strings.TrimSpace(labelText) == sel {
				result = s.Find("b").Text()
				return false // Found it, stop searching.
			}
			return true // Keep searching.
		})
		return strings.TrimSpace(result)
	}

	formatDate := func(dateStr string) string {
		parts := strings.Split(dateStr, "/")
		if len(parts) == 3 {
			return parts[2] + "-" + parts[1] + "-" + parts[0]
		}
		return ""
	}

	obraNumber := getTextFromBold("Número do RRT:")
	if obraNumber == "" {
		err := errors.New("não foi possível extrair o número da obra")
		logExtraction("", "error", err.Error(), nil)
		return nil, err
	}

	id := "obra_" + md5Hex(obraNumber)
	professional := getTextFromBold("Arquiteto(a) e Urbanista:")
	owner := getTextFromBold("Nome/Razão Social:")
	bairro := getTextFromBold("Bairro:")
	startDate := formatDate(getTextFromBold("Data de Início:"))
	endDate := formatDate(getTextFromBold("Previsão de Término:"))
	firstListingDate := formatDate(getTextFromBold("Data de Registro:"))

	// Tratamento especial para Cidade/UF que pode ter múltiplas tags <b>
	cityStateText := ""
	doc.Find("span:contains('Cidade/UF:')").Find("b").Each(func(i int, s *goquery.Selection) {
		if i > 0 {
			cityStateText += "/"
		}
		cityStateText += strings.TrimSpace(s.Text())
	})
	city, state := "", ""
	parts := strings.Split(cityStateText, "/")
	if len(parts) == 2 {
		city = strings.TrimSpace(parts[0])
		state = strings.TrimSpace(parts[1])
	}

	tipoLogradouro := getTextFromBold("Tipo de Logradouro:")
	logradouro := getTextFromBold("Logradouro:")
	numero := getTextFromBold("Número/Ano:")
	complemento := getTextFromBold("Complemento:")
	// Monta o endereço corretamente, sem duplicar o logradouro
	addressParts := []string{}
	if tipoLogradouro != "" {
		addressParts = append(addressParts, tipoLogradouro)
	}
	if logradouro != "" {
		addressParts = append(addressParts, logradouro)
	}
	if numero != "" {
		addressParts = append(addressParts, numero)
	}
	if complemento != "" {
		addressParts = append(addressParts, complemento)
	}
	address := strings.Join(addressParts, ", ") + ", " + bairro + ", " + city + " - " + state

	// Extração da tabela de Atividades usando as funções regex
	activity := ""
	typeVal := ""
	size := 0.0
	unidade := ""

	// Encontra a primeira linha da tabela de atividades e extrai os dados
	doc.Find("span:contains('Atividades')").Next().Find("tbody tr").First().Each(func(i int, row *goquery.Selection) {
		activityText := row.Find("td").Eq(0).Text()
		sizeUnitText := row.Find("td").Eq(1).Text()

		activity = extractActivity(activityText)
		typeVal = extractType(activityText) // A função extractType parece não corresponder ao formato, ajustando
		if typeVal == "" {
			// Fallback simples se a regex de tipo falhar
			re := regexp.MustCompile(`\s*-\s*([\w\sÇÃÁÉÍÓÚ]+)`)
			match := re.FindStringSubmatch(activity)
			if len(match) > 1 {
				typeVal = strings.TrimSpace(match[1])
			}
		}
		size, unidade = extractSizeUnit(sizeUnitText)
	})

	result := &ObraData{
		Id:               id,
		ObraNumber:       obraNumber,
		Professional:     professional,
		Owner:            owner,
		Address:          address,
		Bairro:           bairro,
		City:             city,
		State:            state,
		StartDate:        startDate,
		EndDate:          endDate,
		FirstListingDate: firstListingDate,
		Activity:         activity,
		Type:             typeVal,
		Size:             size,
		Unidade:          unidade,
	}
	logExtraction(obraNumber, "success", "ok", result)
	return result, nil
}

// Registra o resultado da extração no arquivo extract_result.log
func logExtraction(rrt string, status string, msg string, data *ObraData) {
	f, err := os.OpenFile("../extract_result.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Error opening log file:", err)
		return
	}
	defer f.Close()

	var jsonData string
	if data != nil {
		b, err := json.Marshal(data)
		if err != nil {
			jsonData = "{\"error\":\"json marshal failed\"}"
		} else {
			jsonData = string(b)
		}
	}

	line := rrt + "," + status + "," + msg + "," + time.Now().Format("2006-01-02 15:04:05") + "," + jsonData + "\n"
	if _, err := f.WriteString(line); err != nil {
		log.Println("Error writing to log file:", err)
	}
}

func md5Hex(s string) string {
	h := md5.Sum([]byte(s))
	return hex.EncodeToString(h[:])
}

// As funções regex fornecidas parecem adequadas para os dados na tabela de atividades.
func extractActivity(section string) string {
	re := regexp.MustCompile(`(\d+\s*-\s*[^\r\n]+)`)
	match := re.FindStringSubmatch(section)
	if len(match) > 1 {
		return strings.TrimSpace(match[1])
	}
	return strings.TrimSpace(section)
}

func extractType(section string) string {
	// A regex original para 'Type' não correspondia bem.
	// Esta é uma tentativa de extrair o texto principal da atividade.
	re := regexp.MustCompile(`-\s*([\s\S]*?)`)
	match := re.FindStringSubmatch(section)
	if len(match) > 1 {
		return strings.TrimSpace(match[1])
	}
	return ""
}

func extractSizeUnit(section string) (float64, string) {
	re := regexp.MustCompile(`(\d+\.?\d*)\s*/\s*([\w\s²]+)`)
	match := re.FindStringSubmatch(section)
	if len(match) > 2 {
		sizeVal := match[1]
		// Remove pontos de milhar antes de converter
		sizeVal = strings.ReplaceAll(sizeVal, ".", "")
		size, _ := strconv.ParseFloat(sizeVal, 64)
		return size, strings.TrimSpace(match[2])
	}
	return 0.0, ""
}
package api_zenrows

import (
	"io"
	"net/http"
	"os"
	"log"
	"fmt"
	"strings"
	"github.com/PuerkitoBio/goquery"
	"time"
)

func FetchHTMLFromZenRows(rrt, apiKey string) (string, error) {
	url := "https://api.zenrows.com/v1/?apikey=" + apiKey + "&url=https%3A%2F%2Facesso.caubr.gov.br%2Fautenticidade%2Frrt%3Fnumero%3D" + rrt + "%26retificador%3D&js_render=true&wait_for=.title-acction&premium_proxy=true&proxy_country=br"
	client := &http.Client{Timeout: 200 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	html := string(body)
	// Detect Captcha block (redirect to main page)
	if strings.Contains(html, "<title>CAU/BR - Conselho de Arquitetura e Urbanismo do Brasil</title>") {
		cleanedHTML := removeSVGElements(html)
		cleanedHTML = removeIMGElements(cleanedHTML)
		logFailedHTML(rrt, cleanedHTML)
		return "", fmt.Errorf("captcha_block")
	}
	// Detect absence of data (customize this check as needed)
	if strings.Contains(html, "Nenhum dado encontrado") || strings.Contains(html, "Não foi possível localizar") {
		cleanedHTML := removeSVGElements(html)
		cleanedHTML = removeIMGElements(cleanedHTML)
		logFailedHTML(rrt, cleanedHTML)
		return "", fmt.Errorf("no_data")
	}
	if resp.StatusCode != http.StatusOK {
		cleanedHTML := removeSVGElements(html)
		cleanedHTML = removeIMGElements(cleanedHTML)
		logFailedHTML(rrt, cleanedHTML)
		return "", fmt.Errorf("requisição ZenRows falhou: status %d", resp.StatusCode)
	}
	cleanedHTML := removeSVGElements(html)
	cleanedHTML = removeIMGElements(cleanedHTML)
	return cleanedHTML, nil
}

// removeSVGElements removes all <svg>...</svg> elements from HTML string
func removeSVGElements(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return html // fallback: return original if parsing fails
	}
	doc.Find("svg").Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})
	cleaned, err := doc.Html()
	if err != nil {
		return html
	}
	return cleaned
}

// removeIMGElements removes all <img>...</img> elements from HTML string
func removeIMGElements(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return html // fallback: return original if parsing fails
	}
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})
	cleaned, err := doc.Html()
	if err != nil {
		return html
	}
	return cleaned
}

// logFailedHTML logs the HTML of failed requests to html_rrt.log
func logFailedHTML(rrt, html string) {
	f, err := os.OpenFile("html_rrt.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Erro ao abrir html_rrt.log: %v", err)
		return
	}
	defer f.Close()
	logEntry := fmt.Sprintf("RRT: %s\n%s\n---\n", rrt, html)
	if _, err := f.WriteString(logEntry); err != nil {
		log.Printf("Erro ao escrever html falho: %v", err)
	}
}

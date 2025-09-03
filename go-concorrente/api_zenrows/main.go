package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
	"strings"

	"api_zenrows/internal"
)

const (
	ZENROWS_API_KEY = "b67313ec4485fd294ff26be3f995989e4b7ab61b"
	DB_PATH         = "core.db"
	MAX_WORKERS     = 10
	MAX_RETRIES     = 4
	RETRY_DELAY     = 10 * time.Second
	BACKOFF_FACTOR  = 1.2
	MAX_CONSECUTIVE_TIMEOUTS = 3
)

var totalTimeouts = 0
const FAILED_RRT_LOG = "failed_rrts_zenrows.log"

func logFailedRRT(rrt string) {
	f, err := os.OpenFile(FAILED_RRT_LOG, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Erro ao abrir arquivo de log de falhas: %v", err)
		return
	}
	defer f.Close()
	if _, err := f.WriteString(rrt + "\n"); err != nil {
		log.Printf("Erro ao escrever no arquivo de log de falhas: %v", err)
	}
}

func main() {
	log.Printf("🚀 Iniciando processamento concorrente de RRTs via ZenRows API")
	log.Printf("🔑 ZenRows API Key: %s", ZENROWS_API_KEY)
	log.Printf("👷 Workers: %d", MAX_WORKERS)
	log.Printf("💾 Database: %s", DB_PATH)

	fmt.Println("Escolha o modo de busca de RRTs:")
	fmt.Println("1 - Intervalo (número inicial e final)")
	fmt.Println("2 - Lista personalizada (separada por vírgula ou arquivo)")
	fmt.Print("Digite 1 ou 2: ")
	var mode int
	fmt.Scanln(&mode)

	var rrts []string
	if mode == 1 {
		var startRRT, endRRT int
		fmt.Print("Digite o número inicial da RRT: ")
		fmt.Scanln(&startRRT)
		fmt.Print("Digite o número final da RRT: ")
		fmt.Scanln(&endRRT)
		if startRRT > endRRT {
			log.Fatal("Número inicial deve ser menor ou igual ao final")
		}
		for i := startRRT; i <= endRRT; i++ {
			rrts = append(rrts, strconv.Itoa(i))
		}
	} else if mode == 2 {
		fmt.Print("Digite a lista de RRTs separada por vírgula ou o caminho do arquivo: ")
		var input string
		fmt.Scanln(&input)
		if strings.HasSuffix(input, ".txt") {
			// Read RRTs from file
			file, err := os.ReadFile(input)
			if err != nil {
				log.Fatalf("Erro ao ler arquivo: %v", err)
			}
			lines := strings.Split(string(file), "\n")
			for _, line := range lines {
				rrt := strings.TrimSpace(line)
				if rrt != "" {
					rrts = append(rrts, rrt)
				}
			}
		} else {
			// Parse comma-separated list
			items := strings.Split(input, ",")
			for _, item := range items {
				rrt := strings.TrimSpace(item)
				if rrt != "" {
					rrts = append(rrts, rrt)
				}
			}
		}
	} else {
		log.Fatal("Modo inválido. Digite 1 ou 2.")
	}

	results := processRRTsConcurrentlyZenRows(rrts)

	successful := 0
	errors := 0
	ignored := 0

	for _, result := range results {
		if result.Success {
			successful++
		} else if result.Error != nil {
			errors++
		} else {
			ignored++
		}
	}

	log.Printf("📊 Resultado final:")
	log.Printf("✅ Sucessos: %d", successful)
	log.Printf("❌ Erros: %d", errors)
	log.Printf("⚠️  Ignorados: %d", ignored)
	log.Printf("📈 Taxa de sucesso: %.1f%%", float64(successful)/float64(len(rrts))*100)
}

type ProcessResult struct {
	RRT     string
	Success bool
	Error   error
	Data    *api_zenrows.ObraData
}

func processRRTsConcurrentlyZenRows(rrts []string) []ProcessResult {
	jobs := make(chan string, len(rrts))
	results := make(chan ProcessResult, len(rrts))

	var wg sync.WaitGroup
	for i := 0; i < MAX_WORKERS; i++ {
		wg.Add(1)
		go workerZenRows(&wg, jobs, results)
	}

	for _, rrt := range rrts {
		jobs <- rrt
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var allResults []ProcessResult
	for result := range results {
		allResults = append(allResults, result)
	}

	return allResults
}

func workerZenRows(wg *sync.WaitGroup, jobs <-chan string, results chan<- ProcessResult) {
	defer wg.Done()

	consecutiveErrors := 0
	consecutiveTimeouts := 0

	for rrt := range jobs {
		result := processRRTZenRows(rrt)
		results <- result

		if result.Error != nil {
			consecutiveErrors++
			consecutiveTimeouts = 0

			if isTimeoutError(result.Error) {
				consecutiveTimeouts++
				delay := time.Duration(10+consecutiveTimeouts*5) * time.Second
				log.Printf("⏰ Timeout detectado (%d consecutivos), aguardando %v...", consecutiveTimeouts, delay)
				time.Sleep(delay)
			} else {
				delay := time.Duration(5+consecutiveErrors*2) * time.Second
				log.Printf("❌ Erro detectado (%d consecutivos), aguardando %v...", consecutiveErrors, delay)
				log.Printf("Erro: %v", result.Error)
				time.Sleep(delay)
			}
		} else {
			consecutiveErrors = 0
			consecutiveTimeouts = 0
			time.Sleep(3 * time.Second)
		}
	}
}

func processRRTZenRows(rrt string) ProcessResult {
	log.Printf("🔄 Processando RRT %s via ZenRows", rrt)

	var html string
	var err error
	for attempt := 1; attempt <= MAX_RETRIES; attempt++ {
		html, err = api_zenrows.FetchHTMLFromZenRows(rrt, ZENROWS_API_KEY)
		if err == nil {
			break
		}
		errMsg := err.Error()
		if errMsg == "captcha_block" {
			// Retry only on captcha block
			delay := time.Duration(float64(RETRY_DELAY) * float64(attempt) * BACKOFF_FACTOR)
			log.Printf("🔒 Bloqueio de Captcha detectado. Tentativa %d/%d. Aguardando %v...", attempt, MAX_RETRIES, delay)
			time.Sleep(delay)
			continue
		} else if errMsg == "no_data" {
			// Absence of data: ignore, do not retry
			log.Printf("⚠️  Ausência de dados para RRT %s. Ignorando.", rrt)
			return ProcessResult{RRT: rrt, Success: false, Error: nil}
		} else if isTimeoutError(err) {
			totalTimeouts++
			if totalTimeouts >= MAX_CONSECUTIVE_TIMEOUTS {
				logFailedRRT(rrt)
				return ProcessResult{RRT: rrt, Success: false, Error: fmt.Errorf("circuit breaker: %d timeouts consecutivos", totalTimeouts)}
			}
			delay := time.Duration(float64(RETRY_DELAY) * float64(attempt) * BACKOFF_FACTOR)
			log.Printf("⏰ Timeout detectado. Tentativa %d/%d. Aguardando %v...", attempt, MAX_RETRIES, delay)
			time.Sleep(delay)
			continue
		} else {
			// Other errors: do not retry
			logFailedRRT(rrt)
			return ProcessResult{RRT: rrt, Success: false, Error: err}
		}
	}
	if err != nil {
		logFailedRRT(rrt)
		return ProcessResult{RRT: rrt, Success: false, Error: err}
	}

	// Grava o HTML completo em um arquivo de log
	htmlLogFile := "html_rrt.log"
	f, errLog := os.OpenFile(htmlLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if errLog == nil {
		defer f.Close()
		f.WriteString(fmt.Sprintf("[HTML RRT %s]\n%s\n\n", rrt, html))
	} else {
		log.Printf("Erro ao gravar HTML no arquivo de log: %v", errLog)
	}

	obra, err := api_zenrows.ExtractObraData(html)
	if err != nil {
		logFailedRRT(rrt)
		return ProcessResult{RRT: rrt, Success: false, Error: err}
	}

	err = api_zenrows.SaveObraToDB(*obra, DB_PATH)
	if err != nil {
		return ProcessResult{RRT: rrt, Success: false, Error: err}
	}

	return ProcessResult{RRT: rrt, Success: true, Data: obra}
}

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "timeout") || strings.Contains(msg, "deadline exceeded")
}

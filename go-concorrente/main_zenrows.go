package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
	_ "github.com/mattn/go-sqlite3"
)

const (
	ZENROWS_API_URL = "https://api.zenrows.com/v1/"
	ZENROWS_API_KEY = "b67313ec4485fd294ff26be3f995989e4b7ab61b"
	DB_PATH         = "core.db"
	MAX_WORKERS     = 10
	REQUEST_TIMEOUT = 120 * time.Second
	MAX_RETRIES     = 4
	RETRY_DELAY     = 10 * time.Second
	BACKOFF_FACTOR  = 1.2
	MAX_CONSECUTIVE_TIMEOUTS = 3
)

var (
	totalTimeouts = 0
)

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
	log.Printf("⏱️  Timeout: %v", REQUEST_TIMEOUT)
	log.Printf("🔄 Max retries: %d", MAX_RETRIES)
	log.Printf("👷 Workers: %d", MAX_WORKERS)
	log.Printf("💾 Database: %s", DB_PATH)
	log.Printf("⚡ Estratégia: Timeout-aware com backoff inteligente")

	var startRRT, endRRT int
	fmt.Print("Digite o número inicial da RRT: ")
	fmt.Scanln(&startRRT)
	fmt.Print("Digite o número final da RRT: ")
	fmt.Scanln(&endRRT)

	if startRRT > endRRT {
		log.Fatal("Número inicial deve ser menor ou igual ao final")
	}

	var rrts []string
	for i := startRRT; i <= endRRT; i++ {
		rrts = append(rrts, strconv.Itoa(i))
	}

	log.Printf("📋 Lista gerada: %d RRTs (%d a %d)", len(rrts), startRRT, endRRT)

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

// Estrutura para resultado do processamento
type ProcessResult struct {
	RRT     string
	Success bool
	Error   error
	Data    *ObraData
}

// Estrutura para dados da obra (igual main.go)
type ObraData struct {
	Id                string  `json:"id"`
	ObraNumber        string  `json:"obra_number"`
	Professional      string  `json:"professional"`
	Owner             string  `json:"owner"`
	Address           string  `json:"address"`
	Bairro            string  `json:"bairro"`
	City              string  `json:"city"`
	State             string  `json:"state"`
	StartDate         string  `json:"start_date"`
	EndDate           string  `json:"end_date"`
	FirstListingDate  string  `json:"first_listing_date"`
	Activity          string  `json:"activity"`
	Type              string  `json:"type"`
	Size              float64 `json:"size"`
	Unidade           string  `json:"unidade"`
}

// Função para salvar dados no banco de dados
func saveToDB(data ObraData) error {
	dbPath := DB_PATH
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("erro ao abrir conexão com banco: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return fmt.Errorf("erro ao configurar WAL mode: %v", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return fmt.Errorf("erro ao habilitar foreign keys: %v", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout=30000"); err != nil {
		return fmt.Errorf("erro ao configurar busy timeout: %v", err)
	}

	var tx *sql.Tx
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		tx, err = db.Begin()
		if err == nil {
			break
		}
		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
		}
	}
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação após %d tentativas: %v", maxRetries, err)
	}
	defer tx.Rollback()

	now := time.Now()

	var existingID string
	err = tx.QueryRow("SELECT id FROM core_obras_plus WHERE obra_number = ?", data.ObraNumber).Scan(&existingID)

	if err == sql.ErrNoRows {
		insertStmt := `INSERT INTO core_obras_plus 
		(id, obra_number, owner, professional, address, bairro, city, state, 
		 start_date, end_date, activity, type, size, unidade, 
		 first_listing_date, last_listing_date, _sling_loaded_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

		_, err = tx.Exec(insertStmt,
			data.Id,
			data.ObraNumber,
			data.Owner,
			data.Professional,
			data.Address,
			data.Bairro,
			data.City,
			data.State,
			data.StartDate,
			data.EndDate,
			data.Activity,
			data.Type,
			int64(data.Size),
			data.Unidade,
			data.FirstListingDate,
			now.Format("2006-01-02"),
			now.Unix(),
		)

		if err != nil {
			return fmt.Errorf("erro ao inserir registro: %v", err)
		}
		log.Printf("📥 INSERT: RRT %s inserido com sucesso", data.ObraNumber)
	} else if err == nil {
		updateStmt := `UPDATE core_obras_plus SET id=?,
			owner=?, professional=?, address=?, bairro=?, city=?, state=?,
			start_date=?, end_date=?, activity=?, type=?, size=?, unidade=?,
			last_listing_date=?, _sling_loaded_at=?
		WHERE obra_number=?`

		_, err = tx.Exec(updateStmt,
			data.Id,
			data.Owner,
			data.Professional,
			data.Address,
			data.Bairro,
			data.City,
			data.State,
			data.StartDate,
			data.EndDate,
			data.Activity,
			data.Type,
			int64(data.Size),
			data.Unidade,
			now.Format("2006-01-02"),
			now.Unix(),
			data.ObraNumber,
		)

		if err != nil {
			return fmt.Errorf("erro ao atualizar registro: %v", err)
		}
		log.Printf("🔄 UPDATE: RRT %s atualizado com sucesso", data.ObraNumber)
	} else {
		return fmt.Errorf("erro ao verificar existência do registro: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("erro ao commitar transação: %v", err)
	}

	return nil
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

	obra, err := fetchRRTFromZenRows(rrt)
	if err != nil {
		return ProcessResult{RRT: rrt, Success: false, Error: err}
	}

	// Regras de negócio podem ser aplicadas aqui (exemplo: ignorar se obra for nula)
	if obra == nil {
		return ProcessResult{RRT: rrt, Success: false, Error: nil}
	}

	// Salvar no banco de dados
	err = saveToDB(*obra)
	if err != nil {
		return ProcessResult{RRT: rrt, Success: false, Error: err}
	}

	return ProcessResult{RRT: rrt, Success: true, Data: obra}
}

func fetchRRTFromZenRows(rrt string) (*ObraData, error) {
	if totalTimeouts >= MAX_CONSECUTIVE_TIMEOUTS {
		return nil, fmt.Errorf("circuit breaker ativado: %d timeouts consecutivos detectados", totalTimeouts)
	}

	url := fmt.Sprintf("%s?apikey=%s&url=https%%3A%%2F%%2Facesso.caubr.gov.br%%2Fautenticidade%%2Frrt%%3Fnumero%%3D%s%%26retificador%%3D&js_render=true&wait_for=.title&premium_proxy=true&proxy_country=br", ZENROWS_API_URL, ZENROWS_API_KEY, rrt)

	client := &http.Client{Timeout: REQUEST_TIMEOUT}

	var lastErr error
	consecutiveTimeouts := 0

	for attempt := 1; attempt <= MAX_RETRIES; attempt++ {
		log.Printf("🔄 Tentativa %d/%d para RRT %s: %s", attempt, MAX_RETRIES, rrt, url)

		resp, err := client.Get(url)
		if err != nil {
			lastErr = fmt.Errorf("erro na requisição HTTP: %v", err)
			log.Printf("❌ Tentativa %d falhou: %v", attempt, lastErr)

			if isTimeoutError(err) {
				consecutiveTimeouts++
				totalTimeouts++
				log.Printf("⏰ Timeout detectado (%d consecutivos, %d total)", consecutiveTimeouts, totalTimeouts)

				if totalTimeouts >= MAX_CONSECUTIVE_TIMEOUTS {
					log.Printf("🚫 Circuit breaker: muitos timeouts consecutivos, abortando")
					return nil, fmt.Errorf("circuit breaker: %d timeouts consecutivos", totalTimeouts)
				}

				if attempt < MAX_RETRIES {
					delay := time.Duration(float64(RETRY_DELAY) * float64(attempt) * BACKOFF_FACTOR)
					log.Printf("⏳ Aguardando %v antes da próxima tentativa...", delay)
					time.Sleep(delay)
				}
			} else {
				consecutiveTimeouts = 0
				if attempt < MAX_RETRIES {
					delay := time.Duration(float64(RETRY_DELAY) * float64(attempt-1) * BACKOFF_FACTOR)
					log.Printf("⏳ Aguardando %v antes da próxima tentativa...", delay)
					time.Sleep(delay)
				}
			}
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("status HTTP não esperado: %d", resp.StatusCode)
			log.Printf("❌ Status HTTP %d para RRT %s", resp.StatusCode, rrt)

			if attempt < MAX_RETRIES {
				delay := time.Duration(float64(RETRY_DELAY) * float64(attempt-1) * BACKOFF_FACTOR)
				time.Sleep(delay)
			}
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("erro ao ler resposta: %v", err)
			log.Printf("❌ Erro ao ler resposta para RRT %s: %v", rrt, lastErr)
			continue
		}

		bodyStr := string(body)
		if len(bodyStr) > 200 {
			log.Printf("📄 Resposta recebida para RRT %s: %s...", rrt, bodyStr[:200])
		} else {
			log.Printf("📄 Resposta recebida para RRT %s: %s", rrt, bodyStr)
		}

		// TODO: Fazer parsing do HTML para preencher ObraData
		// Exemplo: usar goquery ou regex para extrair campos
		// Aqui retorna nil para Data, pois depende do HTML
		// return &ObraData{ObraNumber: rrt, ...}, nil
		return nil, nil
	}

	logFailedRRT(rrt)
	return nil, fmt.Errorf("falhou após %d tentativas: %v", MAX_RETRIES, lastErr)
}

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	if err.Error() == "timeout" || err.Error() == "deadline exceeded" {
		return true
	}
	return false
}

// Opcional: salvar HTML no banco de dados
// func saveHTMLToDB(rrt string, html string) error {
// 	db, err := sql.Open("sqlite3", DB_PATH)
// 	if err != nil {
// 		return fmt.Errorf("erro ao abrir conexão com banco: %v", err)
// 	}
// 	defer db.Close()
// 	// ...
// 	return nil
// }

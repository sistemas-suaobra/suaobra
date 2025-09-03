package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Configurações
const (
	DEFAULT_API_URL = "https://caubr-api-rem6jzjgfa-uc.a.run.app"
	DB_PATH         = "core.db"
	MAX_WORKERS     = 10  // Reduzido para não sobrecarregar a API
	REQUEST_TIMEOUT = 120 * time.Second  // Aumentado para 2 minutos
	MAX_RETRIES     = 4  // Reduzido para não demorar tanto
	RETRY_DELAY     = 10 * time.Second   // Aumentado para dar mais tempo
	BACKOFF_FACTOR  = 1.2  // Backoff mais conservador
	MAX_CONSECUTIVE_TIMEOUTS = 3  // Máximo de timeouts consecutivos antes de parar
)

// Configurações globais
var (
	API_BASE_URL = getEnv("CAUBR_API_URL", DEFAULT_API_URL)
	totalTimeouts = 0
)

// Função helper para obter variáveis de ambiente com valor padrão
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Função para detectar se um erro é relacionado a timeout
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	timeoutKeywords := []string{
		"timeout",
		"deadline exceeded",
		"Client.Timeout exceeded",
		"context deadline exceeded",
		"i/o timeout",
		"TLS handshake timeout",
	}

	for _, keyword := range timeoutKeywords {
		if strings.Contains(strings.ToLower(errStr), keyword) {
			return true
		}
	}

	// Verificar se é um erro de rede relacionado a timeout
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return true
	}

	return false
}

// Estrutura para resposta completa da API
type APIResponse struct {
	Success bool      `json:"success"`
	Numero  string    `json:"numero"`
	Data    ObraData  `json:"data,omitempty"`
	Error   string    `json:"error,omitempty"`
	Message string    `json:"message,omitempty"`
}

// Estrutura para dados da obra vindos da API
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

// Estrutura para resultado do processamento
type ProcessResult struct {
	RRT     string
	Success bool
	Error   error
	Data    *ObraData
}

func main() {
	log.Printf("🚀 Iniciando processamento concorrente de RRTs")
	log.Printf("📡 API URL: %s", API_BASE_URL)
	log.Printf("⏱️  Timeout: %v", REQUEST_TIMEOUT)
	log.Printf("🔄 Max retries: %d", MAX_RETRIES)
	log.Printf("👷 Workers: %d", MAX_WORKERS)
	log.Printf("💾 Database: %s", DB_PATH)
	log.Printf("⚡ Estratégia: Timeout-aware com backoff inteligente", DB_PATH)

	// Testar conectividade com a API
	if err := testAPIConnectivity(); err != nil {
		log.Fatalf("❌ Falha na conectividade com a API: %v", err)
	}

	// Verificar saúde da API antes de processar
	if err := checkAPIHealth(); err != nil {
		log.Fatalf("❌ API não está saudável: %v", err)
	}

	// Solicita input do usuário para números inicial e final
	var startRRT, endRRT int
	fmt.Print("Digite o número inicial da RRT: ")
	fmt.Scanln(&startRRT)
	fmt.Print("Digite o número final da RRT: ")
	fmt.Scanln(&endRRT)

	if startRRT > endRRT {
		log.Fatal("Número inicial deve ser menor ou igual ao final")
	}

	// Gera lista de RRTs
	var rrts []string
	for i := startRRT; i <= endRRT; i++ {
		rrts = append(rrts, strconv.Itoa(i))
	}

	log.Printf("📋 Lista gerada: %d RRTs (%d a %d)", 
		len(rrts), startRRT, endRRT)

	results := processRRTsConcurrently(rrts)
	
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

// Função para verificar se a API está saudável antes de processar muitos RRTs
func checkAPIHealth() error {
	log.Printf("🏥 Verificando saúde da API...")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Testar endpoint de saúde
	resp, err := client.Get(API_BASE_URL + "/health")
	if err != nil {
		return fmt.Errorf("endpoint /health falhou: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("endpoint /health retornou status %d", resp.StatusCode)
	}

	return nil
}

// Função para testar conectividade com a API
func testAPIConnectivity() error {
	log.Printf("🔍 Testando conectividade com a API...")

	client := &http.Client{
		Timeout: 15 * time.Second, // Timeout menor para teste rápido
		Transport: &http.Transport{
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
		},
	}

	startTime := time.Now()
	resp, err := client.Get(API_BASE_URL + "/health")
	elapsed := time.Since(startTime)

	if err != nil {
		if isTimeoutError(err) {
			return fmt.Errorf("timeout ao conectar com a API (%.2fs): %v", elapsed.Seconds(), err)
		}
		return fmt.Errorf("erro ao conectar com a API (%.2fs): %v", elapsed.Seconds(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API retornou status %d (%.2fs)", resp.StatusCode, elapsed.Seconds())
	}

	log.Printf("✅ API está respondendo corretamente (latência: %.2fs)", elapsed.Seconds())

	return nil
}

func processRRTsConcurrently(rrts []string) []ProcessResult {
	jobs := make(chan string, len(rrts))
	results := make(chan ProcessResult, len(rrts))
	
	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < MAX_WORKERS; i++ {
		wg.Add(1)
		go worker(&wg, jobs, results)
	}

	// Send jobs
	for _, rrt := range rrts {
		jobs <- rrt
	}
	close(jobs)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var allResults []ProcessResult
	for result := range results {
		allResults = append(allResults, result)
	}

	return allResults
}

func worker(wg *sync.WaitGroup, jobs <-chan string, results chan<- ProcessResult) {
	defer wg.Done()

	consecutiveErrors := 0
	consecutiveTimeouts := 0

	for rrt := range jobs {
		result := processRRT(rrt)
		results <- result

		// Ajustar delay baseado no resultado e histórico
		if result.Error != nil {
			consecutiveErrors++
			consecutiveTimeouts = 0

			if isTimeoutError(result.Error) {
				consecutiveTimeouts++
				// Para timeouts, aumentar significativamente o delay
				delay := time.Duration(10+consecutiveTimeouts*5) * time.Second
				log.Printf("⏰ Timeout detectado (%d consecutivos), aguardando %v...", consecutiveTimeouts, delay)
				time.Sleep(delay)
			} else {
				// Para outros erros, delay normal
				delay := time.Duration(5+consecutiveErrors*2) * time.Second
				log.Printf("❌ Erro detectado (%d consecutivos), aguardando %v...", consecutiveErrors, delay)
				log.Printf("Erro: %v", result.Error)
				time.Sleep(delay)
			}
		} else {
			// Sucesso - resetar contadores e usar delay normal
			consecutiveErrors = 0
			consecutiveTimeouts = 0
			time.Sleep(3 * time.Second)
		}
	}
}

func processRRT(rrt string) ProcessResult {
	log.Printf("🔄 Processando RRT %s", rrt)

	// Buscar dados da API (já inclui retry logic)
	data, err := fetchRRTFromAPI(rrt)
	if err != nil {
		return ProcessResult{RRT: rrt, Success: false, Error: err}
	}

	// Aplicar regras de negócio
	if !isValidForProcessing(data) {
		return ProcessResult{RRT: rrt, Success: false, Error: nil} // Ignorado
	}

	// Salvar no banco de dados
	err = saveToDB(*data)
	if err != nil {
		return ProcessResult{RRT: rrt, Success: false, Error: err}
	}

	return ProcessResult{RRT: rrt, Success: true, Data: data}
}

func fetchRRTFromAPI(rrt string) (*ObraData, error) {
	// Circuit breaker: se muitos timeouts consecutivos, parar
	if totalTimeouts >= MAX_CONSECUTIVE_TIMEOUTS {
		return nil, fmt.Errorf("circuit breaker ativado: %d timeouts consecutivos detectados", totalTimeouts)
	}

	url := fmt.Sprintf("%s/caubr?numero=%s", API_BASE_URL, rrt)

	// Cliente HTTP com configurações otimizadas
	client := &http.Client{
		Timeout: REQUEST_TIMEOUT,
		Transport: &http.Transport{
			TLSHandshakeTimeout:   15 * time.Second,
			ResponseHeaderTimeout: 60 * time.Second,  // Aumentado para headers
			ExpectContinueTimeout: 5 * time.Second,
			IdleConnTimeout:       60 * time.Second,
			MaxIdleConns:          5,
			MaxIdleConnsPerHost:   1,
		},
	}

	var lastErr error
	consecutiveTimeouts := 0

	for attempt := 1; attempt <= MAX_RETRIES; attempt++ {
		log.Printf("🔄 Tentativa %d/%d para RRT %s: %s", attempt, MAX_RETRIES, rrt, url)

		startTime := time.Now()
		resp, err := client.Get(url)
		elapsed := time.Since(startTime)

		if err != nil {
			lastErr = fmt.Errorf("erro na requisição HTTP (%.2fs): %v", elapsed.Seconds(), err)
			log.Printf("❌ Tentativa %d falhou (%.2fs): %v", attempt, elapsed.Seconds(), lastErr)

			// Para timeouts, aguardar mais tempo
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
				// Para outros erros, usar delay normal
				consecutiveTimeouts = 0  // Resetar contador de timeouts
				if attempt < MAX_RETRIES {
					delay := time.Duration(float64(RETRY_DELAY) * float64(attempt-1) * BACKOFF_FACTOR)
					log.Printf("⏳ Aguardando %v antes da próxima tentativa...", delay)
					time.Sleep(delay)
				}
			}
			continue
		}

		// Resetar contadores em caso de sucesso
		consecutiveTimeouts = 0
		totalTimeouts = 0

		log.Printf("📡 Resposta recebida em %.2fs", elapsed.Seconds())
		defer resp.Body.Close()

		// Verificar status HTTP
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("status HTTP não esperado: %d", resp.StatusCode)
			log.Printf("❌ Status HTTP %d para RRT %s", resp.StatusCode, rrt)

			if attempt < MAX_RETRIES {
				delay := time.Duration(float64(RETRY_DELAY) * float64(attempt-1) * BACKOFF_FACTOR)
				time.Sleep(delay)
			}
			continue
		}

		// Ler resposta
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("erro ao ler resposta: %v", err)
			log.Printf("❌ Erro ao ler resposta para RRT %s: %v", rrt, lastErr)
			continue
		}

		// Log da resposta para debug (limitado para não poluir o log)
		bodyStr := string(body)
		if len(bodyStr) > 200 {
			log.Printf("📄 Resposta recebida para RRT %s: %s...", rrt, bodyStr[:200])
		} else {
			log.Printf("📄 Resposta recebida para RRT %s: %s", rrt, bodyStr)
		}

		// Decodificar JSON
		var apiResp APIResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			lastErr = fmt.Errorf("erro ao decodificar JSON: %v", err)
			log.Printf("❌ Erro JSON para RRT %s: %v", rrt, lastErr)
			continue
		}

		// Verificar se a API retornou sucesso
		if !apiResp.Success {
			errorMsg := apiResp.Error
			if errorMsg == "" {
				errorMsg = apiResp.Message
			}
			if errorMsg == "" {
				errorMsg = "erro desconhecido na API"
			}
			lastErr = fmt.Errorf("API retornou erro: %s", errorMsg)
			log.Printf("❌ API erro para RRT %s: %s", rrt, errorMsg)

			// Para erros da API, não faz sentido tentar novamente
			break
		}

		// Validar dados recebidos
		if err := validateObraData(&apiResp.Data); err != nil {
			lastErr = fmt.Errorf("dados inválidos recebidos: %v", err)
			log.Printf("❌ Dados inválidos para RRT %s: %v", rrt, lastErr)
			continue
		}

		log.Printf("✅ RRT %s processado com sucesso em %.2fs", rrt, elapsed.Seconds())
		return &apiResp.Data, nil
	}

	return nil, fmt.Errorf("falhou após %d tentativas: %v", MAX_RETRIES, lastErr)
}

// Função para validar dados da obra
func validateObraData(data *ObraData) error {
	if data == nil {
		return fmt.Errorf("dados são nulos")
	}

	if data.ObraNumber == "" {
		return fmt.Errorf("número da obra está vazio")
	}

	if data.Id == "" {
		return fmt.Errorf("ID está vazio")
	}

	// Validar datas se presentes
	if data.StartDate != "" {
		if _, err := time.Parse("2006-01-02", data.StartDate); err != nil {
			return fmt.Errorf("data de início inválida: %s", data.StartDate)
		}
	}

	if data.EndDate != "" {
		if _, err := time.Parse("2006-01-02", data.EndDate); err != nil {
			return fmt.Errorf("data de término inválida: %s", data.EndDate)
		}
	}

	if data.FirstListingDate != "" {
		if _, err := time.Parse("2006-01-02", data.FirstListingDate); err != nil {
			return fmt.Errorf("data de primeiro anúncio inválida: %s", data.FirstListingDate)
		}
	}

	return nil
}

func isValidForProcessing(data *ObraData) bool {
	if data == nil {
		log.Printf("⚠️  Dados nulos recebidos")
		return false
	}

	// Regra 1: Se type for "RRT MÍNIMO" deve ser ignorado
	if data.Type == "RRT MÍNIMO" {
		log.Printf("📋 RRT %s ignorado: tipo RRT MÍNIMO", data.ObraNumber)
		return false
	}

	// Regra 2: Data de término deve ser superior a dois anos de 01/09/2025
	if data.EndDate != "" {
		cutoffDate := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
		twoYearsAgo := cutoffDate.AddDate(-2, 0, 0) // 01/09/2023
		
		endDate, err := time.Parse("2006-01-02", data.EndDate)
		if err != nil {
			log.Printf("⚠️  RRT %s: erro ao parsear data de término %s: %v", data.ObraNumber, data.EndDate, err)
			return false
		}
		
		if endDate.Before(twoYearsAgo) {
			log.Printf("📅 RRT %s ignorado: data de término %s é anterior a %s", 
				data.ObraNumber, data.EndDate, twoYearsAgo.Format("2006-01-02"))
			return false
		}
	}

	// Regra 3: Validar campos obrigatórios
	if data.Owner == "" {
		log.Printf("⚠️  RRT %s: proprietário vazio", data.ObraNumber)
		return false
	}

	if data.Address == "" {
		log.Printf("⚠️  RRT %s: endereço vazio", data.ObraNumber)
		return false
	}

	if data.City == "" {
		log.Printf("⚠️  RRT %s: cidade vazia", data.ObraNumber)
		return false
	}

	return true
}

func saveToDB(data ObraData) error {
	dbPath := DB_PATH
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("erro ao abrir conexão com banco: %v", err)
	}
	defer db.Close()

	// Habilitar WAL mode e foreign keys
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return fmt.Errorf("erro ao configurar WAL mode: %v", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return fmt.Errorf("erro ao habilitar foreign keys: %v", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout=30000"); err != nil {
		return fmt.Errorf("erro ao configurar busy timeout: %v", err)
	}

	// Iniciar transação com retry
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

	// Verificar se registro já existe
	var existingID string
	err = tx.QueryRow("SELECT id FROM core_obras_plus WHERE obra_number = ?", data.ObraNumber).Scan(&existingID)

	if err == sql.ErrNoRows {
		// INSERT - novo registro
		insertStmt := `INSERT INTO core_obras_plus 
		(id, obra_number, owner, professional, address, bairro, city, state, 
		 start_date, end_date, activity, type, size, unidade, 
		 first_listing_date, last_listing_date, _sling_loaded_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		
		_, err = tx.Exec(insertStmt,
			data.Id,                           // id
			data.ObraNumber,                   // obra_number
			data.Owner,                        // owner
			data.Professional,                 // professional
			data.Address,                      // address
			data.Bairro,                      // bairro
			data.City,                        // city
			data.State,                       // state
			data.StartDate,                   // start_date
			data.EndDate,                     // end_date
			data.Activity,                    // activity
			data.Type,                        // type
			int64(data.Size),                 // size
			data.Unidade,                     // unidade
			data.FirstListingDate,            // first_listing_date (da API)
			now.Format("2006-01-02"),         // last_listing_date
			now.Unix(),                       // _sling_loaded_at
		)
		
		if err != nil {
			return fmt.Errorf("erro ao inserir registro: %v", err)
		}
		
		log.Printf("📥 INSERT: RRT %s inserido com sucesso", data.ObraNumber)
		
	} else if err == nil {
		// UPDATE - registro existente
		updateStmt := `UPDATE core_obras_plus SET id=?,
			owner=?, professional=?, address=?, bairro=?, city=?, state=?,
			start_date=?, end_date=?, activity=?, type=?, size=?, unidade=?,
			last_listing_date=?, _sling_loaded_at=?
		WHERE obra_number=?`
		
		_, err = tx.Exec(updateStmt,
			data.Id,                    // id
			data.Owner,                  // owner
			data.Professional,           // professional
			data.Address,                // address
			data.Bairro,                // bairro
			data.City,                  // city
			data.State,                 // state
			data.StartDate,             // start_date
			data.EndDate,               // end_date
			data.Activity,              // activity
			data.Type,                  // type
			int64(data.Size),           // size
			data.Unidade,               // unidade
			now.Format("2006-01-02"),   // last_listing_date
			now.Unix(),                 // _sling_loaded_at
			data.ObraNumber,            // WHERE obra_number=?
		)
		
		if err != nil {
			return fmt.Errorf("erro ao atualizar registro: %v", err)
		}
		
		log.Printf("🔄 UPDATE: RRT %s atualizado com sucesso", data.ObraNumber)
		
	} else {
		return fmt.Errorf("erro ao verificar existência do registro: %v", err)
	}

	// Commit da transação
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("erro ao commitar transação: %v", err)
	}

	return nil
}

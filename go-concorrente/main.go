package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Configurações
const (
	API_BASE_URL  = "http://localhost:5000"
	DB_PATH       = "./core_new.db"
	MAX_WORKERS   = 5
)

// Estrutura para dados da obra vindos da API
type ObraData struct {
	Id		   string  `json:"id"`
	ObraNumber   string  `json:"obra_number"`
	Professional string  `json:"professional"`
	Owner        string  `json:"owner"`
	Address      string  `json:"address"`
	Bairro       string  `json:"bairro"`
	City         string  `json:"city"`
	State        string  `json:"state"`
	StartDate    string  `json:"start_date"`
	EndDate      string  `json:"end_date"`
	Activity     string  `json:"activity"`
	Type         string  `json:"type"`
	Size         float64 `json:"size"`
	Unidade      string  `json:"unidade"`
}

// Estrutura para erro da API
type APIError struct {
	Error string `json:"error"`
}

// Estrutura para resultado do processamento
type ProcessResult struct {
	RRT     string
	Success bool
	Error   error
	Data    *ObraData
}

func main() {
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

	log.Printf("Iniciando processamento de %d RRTs (%d a %d) com %d workers", 
		len(rrts), startRRT, endRRT, MAX_WORKERS)

	results := processRRTsConcurrently(rrts)
	
	successful := 0
	errors := 0
	ignored := 0
	
	for _, result := range results {
		if result.Success {
			successful++
			log.Printf("✅ RRT %s processado com sucesso", result.RRT)
		} else if result.Error != nil {
			errors++
			log.Printf("❌ RRT %s falhou: %v", result.RRT, result.Error)
		} else {
			ignored++
			log.Printf("⚠️  RRT %s ignorado por regras de negócio", result.RRT)
		}
	}

	log.Printf("Processamento concluído: %d sucessos, %d erros, %d ignorados", successful, errors, ignored)
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
	
	for rrt := range jobs {
		result := processRRT(rrt)
		results <- result
	}
}

func processRRT(rrt string) ProcessResult {
	log.Printf("🔄 Processando RRT %s", rrt)
	
	// Fazer requisição para a API
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
	url := fmt.Sprintf("%s/rrt/%s", API_BASE_URL, rrt)
	
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer requisição: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiError APIError
		if err := json.Unmarshal(body, &apiError); err == nil {
			return nil, fmt.Errorf("erro da API: %s", apiError.Error)
		}
		return nil, fmt.Errorf("API retornou status %d", resp.StatusCode)
	}

	var data ObraData
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("erro ao decodificar JSON: %v", err)
	}

	return &data, nil
}

func isValidForProcessing(data *ObraData) bool {
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
			log.Printf("⚠️  RRT %s: erro ao parsear data de término %s", data.ObraNumber, data.EndDate)
			return false
		}
		
		if endDate.Before(twoYearsAgo) {
			log.Printf("📅 RRT %s ignorado: data de término %s é anterior a %s", 
				data.ObraNumber, data.EndDate, twoYearsAgo.Format("2006-01-02"))
			return false
		}
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

	// Iniciar transação
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %v", err)
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
			data.Id,            // id
			data.ObraNumber,            // obra_number
			data.Owner,                 // owner
			data.Professional,          // professional
			data.Address,               // address
			data.Bairro,               // bairro
			data.City,                 // city
			data.State,                // state
			data.StartDate,            // start_date
			data.EndDate,              // end_date
			data.Activity,             // activity
			data.Type,                 // type
			int64(data.Size),          // size
			data.Unidade,              // unidade
			now.Format("2006-01-02"),  // first_listing_date
			now.Format("2006-01-02"),  // last_listing_date
			now.Unix(),                // _sling_loaded_at
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

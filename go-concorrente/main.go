package main

import (
	"database/sql"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ObraData struct {
	ObraNumber       string
	Professional     string
	Owner            string
	Address          string
	Bairro           string
	City             string
	State            string
	StartDate        string
	EndDate          string
	Activity         string
	Type             string
	Size             int64
	Unidade          string
	FirstListingDate string
}

func main() {
	rrtToTest := "9999969"
	log.Printf("Iniciando teste com parser final para RRT %s", rrtToTest)

	cmd := exec.Command("python3", "automacao-caubr/automacao_caubr.py", rrtToTest)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Erro ao executar script Python: %v\nSaída: %s", err, string(output))
	}

	if len(output) < 100 {
		log.Fatalf("Dados não encontrados para RRT %s.", rrtToTest)
	}

	obra, err := parseData(string(output))
	if err != nil {
		log.Fatalf("Erro ao analisar dados: %v", err)
	}

	err = saveToDB(obra)
	if err != nil {
		log.Fatalf("Erro ao salvar no banco de dados: %v", err)
	}

	log.Printf("Processo para RRT %s concluído com sucesso.", rrtToTest)
}

func parseData(text string) (ObraData, error) {
    data := ObraData{}

    // extract é uma função aninhada que previne panics
    extract := func(re *regexp.Regexp, content string) string {
        match := re.FindStringSubmatch(content)
        if len(match) > 1 {
            return strings.TrimSpace(match[1])
        }
        return ""
    }

    // formatDate é uma função aninhada para formatação de data
    formatDate := func(dateStr string) string {
        t, err := time.Parse("02/01/2006", dateStr)
        if err != nil {
            return ""
        }
        return t.Format("2006-01-02")
    }

    // Extração de dados simples
    data.ObraNumber = extract(regexp.MustCompile(`(?i)Número do RRT:\s*(\d+)`), text)
    data.Professional = extract(regexp.MustCompile(`(?i)Arquiteto\(a\) e Urbanista:\s*(.*)`), text)

    // Lógica robusta para o Proprietário
    ownerValue := extract(regexp.MustCompile(`(?i)Nome/Razão Social:\s*(.*)`), text)
    if strings.HasPrefix(ownerValue, "CPF / CNPJ") {
        data.Owner = ""
    } else {
        data.Owner = ownerValue
    }

    data.Bairro = extract(regexp.MustCompile(`(?i)Bairro:\s*(.*)`), text)
    data.StartDate = formatDate(extract(regexp.MustCompile(`(?i)Data de Início:\s*(.*)`), text))
    data.EndDate = formatDate(extract(regexp.MustCompile(`(?i)Previsão de Término:\s*(.*)`), text))
    data.FirstListingDate = formatDate(extract(regexp.MustCompile(`(?i)Data de Registro:\s*(.*)`), text))

    // Lógica robusta para Cidade e Estado
    if cityState := extract(regexp.MustCompile(`(?i)Cidade/UF:\s*(.*)`), text); cityState != "" {
        parts := strings.Split(cityState, "/")
        if len(parts) == 2 {
            data.City = strings.TrimSpace(parts[0])
            data.State = strings.TrimSpace(parts[1])
        }
    }

    // Lógica para Atividade e Tipo com verificação de segurança
    activityStr := extract(regexp.MustCompile(`(?i)Atividade Subordinada([\s\S]*?)Pagamento`), text)
    data.Activity = extract(regexp.MustCompile(`(?i)(\d+\.\d+\.\d+\s*-\s*[^\d\s-][\s\S]*?)[\s]*[\d]`), activityStr)
    
    reType := regexp.MustCompile(`(?i)([\d\.]+)\s*([\wÇÃÁÉÍÓÚ]+)\s*>\s*(\d+)\s*>\s*([\d\.]+\s*-\s*[\w\sÇÃÁÉÍÓÚ]+)`)
    matchType := reType.FindStringSubmatch(activityStr)
    if len(matchType) > 3 {
        data.Type = fmt.Sprintf("%s - %s", matchType[3], matchType[2])
    }

    // Lógica para Tamanho e Unidade
    reSizeUnit := regexp.MustCompile(`(?i)(\d+\.?\d*)\s*/\s*([\w\s²]+)`)
    matchSizeUnit := reSizeUnit.FindStringSubmatch(activityStr)
    if len(matchSizeUnit) > 2 {
        sizeValue := strings.Split(matchSizeUnit[1], ".")[0]
        if floatSize, err := strconv.ParseFloat(sizeValue, 64); err == nil {
            data.Size = int64(floatSize)
        }
        data.Unidade = matchSizeUnit[2]
    }
    
    // Lógica para Endereço completo
    tipoLogradouro := extract(regexp.MustCompile(`(?i)Tipo de Logradouro:\s*(.*)`), text)
    logradouro := extract(regexp.MustCompile(`(?i)Logradouro:\s*(.*)`), text)
    numero := extract(regexp.MustCompile(`(?i)Número/Ano:\s*(.*)`), text)
    complemento := extract(regexp.MustCompile(`(?i)Complemento:\s*(.*)`), text)
    data.Address = fmt.Sprintf("%s %s, %s, %s, %s - %s, %s", tipoLogradouro, logradouro, numero, complemento, data.Bairro, data.City, data.State)

    return data, nil
}

func saveToDB(data ObraData) error {
	dbPath := "./core_new.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil { return err }
	defer db.Close()

	tx, err := db.Begin()
	if err != nil { return err }

	var existingID string
	err = tx.QueryRow("SELECT id FROM core_obras_plus WHERE obra_number = ?", data.ObraNumber).Scan(&existingID)

	updateStmt := `UPDATE core_obras_plus SET owner=?, professional=?, address=?, bairro=?, city=?, state=?, start_date=?, end_date=?, activity=?, type=?, size=?, unidade=?, last_listing_date=?, _sling_loaded_at=? WHERE obra_number=?`
	insertStmt := `INSERT INTO core_obras_plus (id, obra_number, owner, professional, address, bairro, city, state, start_date, end_date, activity, type, size, unidade, first_listing_date, last_listing_date, _sling_loaded_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	if err == sql.ErrNoRows {
		log.Println("INSERT: Registro não encontrado, inserindo...")
		_, err = tx.Exec(insertStmt, data.ObraNumber, data.ObraNumber, data.Owner, data.Professional, data.Address, data.Bairro, data.City, data.State, data.StartDate, data.EndDate, data.Activity, data.Type, data.Size, data.Unidade, data.FirstListingDate, time.Now().Format("2006-01-02"), time.Now().Unix())
	} else if err == nil {
		log.Println("UPDATE: Registro encontrado, atualizando...")
		_, err = tx.Exec(updateStmt, data.Owner, data.Professional, data.Address, data.Bairro, data.City, data.State, data.StartDate, data.EndDate, data.Activity, data.Type, data.Size, data.Unidade, time.Now().Format("2006-01-02"), time.Now().Unix(), data.ObraNumber)
	} else {
		tx.Rollback()
		return err
	}

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

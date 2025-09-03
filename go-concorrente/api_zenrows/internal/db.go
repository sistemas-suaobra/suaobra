package api_zenrows

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	_ "github.com/mattn/go-sqlite3"
)

func SaveObraToDB(data ObraData, dbPath string) error {
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

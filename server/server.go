package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Cotacao struct {
	USDBRL struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func main() {
	db, err := sql.Open("sqlite3", "./cotacoes.db")
	if err != nil {
		log.Fatal("Erro ao abrir conexão com o banco de dados:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Erro ao conectar ao banco de dados:", err)
	}

	if err := checkAndCreateTable(db); err != nil {
		log.Fatalf("Erro ao verificar ou criar tabela: %v", err)
	}

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 2000*time.Millisecond)
		defer cancel()

		cotacao, err := getCotacao(ctx)
		if err != nil {
			http.Error(w, "Erro ao obter cotação: "+err.Error(), http.StatusInternalServerError)
			return
		}

		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		err = saveCotacao(ctx, db, cotacao)
		if err != nil {
			http.Error(w, "Erro ao salvar cotação: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cotacao.USDBRL.Bid)
	})

	log.Println("Servidor rodando na porta 8080...")
	http.ListenAndServe(":8080", nil)
}

func getCotacao(ctx context.Context) (Cotacao, error) {
	req, err := http.NewRequest("GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return Cotacao{}, err
	}

	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Cotacao{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Cotacao{}, fmt.Errorf("erro ao obter dados da API: status %d", resp.StatusCode)
	}

	var cotacao Cotacao
	err = json.NewDecoder(resp.Body).Decode(&cotacao)
	if err != nil {
		return Cotacao{}, err
	}

	return cotacao, nil
}

func saveCotacao(ctx context.Context, db *sql.DB, cotacao Cotacao) error {
	stmt, err := db.Prepare("INSERT INTO usdbrl (bid, created_at) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, cotacao.USDBRL.Bid, time.Now())
	return err
}

func checkAndCreateTable(db *sql.DB) error {
	createTableSQL := `CREATE TABLE IF NOT EXISTS usdbrl (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		bid TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("erro ao criar tabela: %v", err)
	}
	return nil
}

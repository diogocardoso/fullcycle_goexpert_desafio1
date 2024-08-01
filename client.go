package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Cotacao struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	cotacao, err := getCotacao(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = saveCotacao(cotacao)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Success: Bid (%s) registered in %s", cotacao.Bid, time.Now().Format("02/02/2002 15:04:05"))
}

func getCotacao(ctx context.Context) (Cotacao, error) {
	req, err := http.NewRequest("GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		return Cotacao{}, err
	}

	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Cotacao{}, err
	}
	defer resp.Body.Close()

	var cotacao Cotacao

	err = json.NewDecoder(resp.Body).Decode(&cotacao)
	if err != nil {
		return Cotacao{}, err
	}

	return cotacao, nil
}

func saveCotacao(cotacao Cotacao) error {
	f, err := os.Create("cotacao.txt")
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("Dólar: %s\n", cotacao.Bid))
	return err
}

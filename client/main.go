package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type CurrencyQuote struct {
	Bid string `json:"bid"`
}

func main() {
	quote, err := getDolarQuote()
	if err != nil {
		panic(err)
	}

	err = storeQuoteFile(quote)
	if err != nil {
		panic(err)
	}
}

func getDolarQuote() (*CurrencyQuote, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()

	log.Println("[INFO] Requesting currency quote...")

	var quote CurrencyQuote

	chn := make(chan CurrencyQuote)

	go getDolarQuoteRequest(ctx, chn, &quote)

	select {
	case chnQuote := <-chn:
		if chnQuote == quote {
			log.Println("[INFO] Request completed successfully")
			return &quote, nil
		}
	case <-ctx.Done():
		log.Printf("[ERROR] The request has done: %v\n", ctx.Err())
	}

	return nil, ctx.Err()
}

func getDolarQuoteRequest(ctx context.Context, chn chan CurrencyQuote, quote *CurrencyQuote) {
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
	}

	err = json.Unmarshal(body, &quote)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
	}

	chn <- *quote
}

func storeQuoteFile(quote *CurrencyQuote) error {
	log.Println("[INFO] Storing currency quote in file...")

	file, err := os.Create("cotacao.txt")
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return err
	}
	defer file.Close()

	_, err = file.Write([]byte(fmt.Sprintf("Dólar: %v", quote.Bid)))
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return err
	}

	log.Println("[INFO] Currency quote storage completed")

	return nil
}

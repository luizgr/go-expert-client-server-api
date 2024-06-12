package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type CurrencyQuote struct {
	USDBRL QuoteDetail `json:"USDBRL"`
}

type QuoteDetail struct {
	Code       string `json:"code"`
	CodeIn     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/cotacao", QuoteHandler)

	http.ListenAndServe(":8080", mux)
}

// QuoteHandler is a handler that returns the currency quote

func QuoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	quote, err := getCurrencyQuote()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = storeCurrencyQuote(quote)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(quote.USDBRL)
}

// getCurrencyQuote gets the currency quote from the API

func getCurrencyQuote() (*CurrencyQuote, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()

	log.Println("[INFO] Requesting currency quote...")

	var quote CurrencyQuote

	chn := make(chan CurrencyQuote)

	go getCurrencyQuoteFromRequest(ctx, chn, &quote)

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

func getCurrencyQuoteFromRequest(ctx context.Context, chn chan CurrencyQuote, quote *CurrencyQuote) {
	req, err := http.NewRequestWithContext(
		ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return
	}

	err = json.Unmarshal(body, &quote)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return
	}

	chn <- *quote
}

// storeCurrencyQuote stores the currency quote in the database

func storeCurrencyQuote(quote *CurrencyQuote) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()

	log.Println("[INFO] Storing currency quote in database...")

	chn := make(chan CurrencyQuote)

	go storeCurrencyQuoteDatabase(ctx, chn, quote)

	select {
	case chnQuote := <-chn:
		if chnQuote == *quote {
			log.Println("[INFO] Currency quote stored successfully")
			return nil
		}
	case <-ctx.Done():
		log.Printf("[ERROR] The database persist has done: %v\n", ctx.Err())
	}

	return ctx.Err()
}

func storeCurrencyQuoteDatabase(ctx context.Context, chn chan CurrencyQuote, quote *CurrencyQuote) {
	err := createSqliteDatabaseFile("./database.db")
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return
	}

	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return
	}

	err = createCurrencyQuoteTable(db)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return
	}

	err = insertCurrencyQuoteDatabase(ctx, db, quote)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
		return
	}

	chn <- *quote
}

func createSqliteDatabaseFile(filepath string) error {
	if _, err := os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
		log.Printf("[INFO] Creating '%v' file...\n", filepath)

		file, err := os.Create(filepath)
		if err != nil {
			return err
		}
		file.Close()

		log.Printf("[INFO] The '%v' file has been created\n", filepath)
	}

	return nil
}

func createCurrencyQuoteTable(db *sql.DB) error {
	log.Println("[INFO] Creating currency quote table if not exists...")

	query := `CREATE TABLE IF NOT EXISTS currency_quote (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
		"code" TEXT,
		"codein" TEXT,
		"name" TEXT,
		"high" DECIMAL,
		"low" DECIMAL,
		"varBid" DECIMAL,
		"pctChange" DECIMAL,
		"bid" DECIMAL,
		"ask" DECIMAL,
		"timestamp" TIMESTAMP,
		"create_date" TIMESTAMP
	);`

	statement, err := db.Prepare(query)
	if err != nil {
		return err
	}

	statement.Exec()

	log.Println("[INFO] Currency quote table creation completed")

	return nil
}

func insertCurrencyQuoteDatabase(ctx context.Context, db *sql.DB, quote *CurrencyQuote) error {
	log.Println("[INFO] Inserting currency quote record...")

	insertStudentSQL := `
		INSERT INTO currency_quote(code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, create_date) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	statement, err := db.PrepareContext(ctx, insertStudentSQL)
	if err != nil {
		return err
	}

	_, err = statement.Exec(
		quote.USDBRL.Code,
		quote.USDBRL.CodeIn,
		quote.USDBRL.Name,
		quote.USDBRL.High,
		quote.USDBRL.Low,
		quote.USDBRL.VarBid,
		quote.USDBRL.PctChange,
		quote.USDBRL.Bid,
		quote.USDBRL.Ask,
		quote.USDBRL.Timestamp,
		quote.USDBRL.CreateDate)
	if err != nil {
		return err
	}

	log.Println("[INFO] Currency record insertion completed")

	return nil
}

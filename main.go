package main

import (
	"log"
	"net/http"
	"os"
	"wallet/internal/handler"
	"wallet/internal/repository"
	"wallet/internal/service"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	err := godotenv.Load("config.env")
	if err != nil {
		log.Fatalf("Error loading config.env file: %v", err)
	}

	db, err := repository.NewPostgresDB(repository.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Username: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	repo := repository.NewWalletRepository(db)
	srv := service.NewWalletService(repo)
	hnd := handler.NewWalletHandler(srv)

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/wallet", hnd.HandleWalletOperation).Methods("POST")
	router.HandleFunc("/api/v1/wallets/{id}", hnd.GetBalance).Methods("GET")

	log.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}

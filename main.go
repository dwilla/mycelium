package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dwilla/mycelium/handlers"
	"github.com/dwilla/mycelium/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	cfg := handlers.Config{}
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("warning: assuming default configuration. .env unreadable: %v", err)
	}

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)
	cfg.DB = dbQueries

	mux := http.NewServeMux()
	mux.HandleFunc("/", cfg.HandleMain)
	mux.HandleFunc("GET /app", cfg.HandleApp)
	mux.HandleFunc("GET /auth/email", cfg.CheckEmail)
	mux.HandleFunc("GET /auth/username", cfg.CheckUsername)
	mux.HandleFunc("GET /auth/password", cfg.CheckPassword)

	server := http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: (10 * time.Second),
	}

	log.Printf("Running at: http://localhost%v\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}

package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dwilla/mycelium/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver
)

type config struct {
	DB *database.Queries
}

func main() {
	cfg := config{}
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("warning: assuming default configuration. .env unreadable: %v", err)
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Println("DATABASE_URL environment variable is not set")
		log.Println("Running without CRUD endpoints")
	} else {
		db, err := sql.Open("postgres", dbURL)
		if err != nil {
			log.Fatal(err)
		}
		dbQueries := database.New(db)
		cfg.DB = dbQueries
		log.Println("Connected to database!")
	}

	mux := http.NewServeMux()

	server := http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: (10 * time.Second),
	}

	log.Printf("Running at: http://localhost%v\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}

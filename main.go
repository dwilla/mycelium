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

	if err := godotenv.Load(".env"); err != nil {
		log.Printf("warning: assuming default configuration. .env unreadable: %v", err)
	}

	cfg.Mailgun = os.Getenv("MAILGUN_KEY")
	if cfg.Mailgun == "" {
		log.Fatal("no mailgun key")
	}
	cfg.JwtSecret = os.Getenv("TOKEN_SECRET")
	if cfg.JwtSecret == "" {
		log.Fatal("no token secret")
	}
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)
	cfg.DB = dbQueries

	isProduction := os.Getenv("RENDER") == "true"
	if isProduction {
		cfg.BaseURL = "https://mycelium.chat"
	} else {
		cfg.BaseURL = "https://localhost:" + os.Getenv("PORT")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", cfg.HandleMain)
	mux.HandleFunc("GET /auth/email", cfg.CheckEmail)
	mux.HandleFunc("GET /auth/username", cfg.CheckUsername)
	mux.HandleFunc("GET /auth/password", cfg.CheckPassword)
	mux.HandleFunc("POST /auth/newuser", cfg.HandleNewUser)
	mux.HandleFunc("POST /auth/login", cfg.HandleLogin)
	mux.HandleFunc("POST /email/reset", cfg.SendPassReset)
	mux.HandleFunc("GET /reset/{uuid}", cfg.HandleReset)
	mux.HandleFunc("POST /reset/{uuid}", cfg.HandleResetPost)
	mux.HandleFunc("/auth/logout", cfg.HandleSignOut)
	mux.Handle("/app", cfg.Auth(http.HandlerFunc(cfg.HandleHome)))

	server := &http.Server{
		Addr:              ":" + os.Getenv("PORT"),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	if isProduction {
		log.Println("Server running in production mode.")
		log.Fatal(server.ListenAndServe())
	} else {
		log.Printf("Server running in development mode at: %v\n", cfg.BaseURL)
		log.Fatal(server.ListenAndServeTLS("certs/cert.pem", "certs/key.pem"))
	}
}

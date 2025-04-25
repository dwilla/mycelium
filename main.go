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

	mux := http.NewServeMux()
	mux.HandleFunc("/", cfg.HandleMain)
	mux.HandleFunc("GET /auth/email", cfg.CheckEmail)
	mux.HandleFunc("GET /auth/username", cfg.CheckUsername)
	mux.HandleFunc("GET /auth/password", cfg.CheckPassword)
	mux.HandleFunc("POST /auth/newuser", cfg.HandleNewUser)
	mux.HandleFunc("POST /auth/login", cfg.HandleLogin)
	mux.HandleFunc("/auth/logout", cfg.HandleSignOut)
	mux.Handle("/app", cfg.Auth(http.HandlerFunc(cfg.HandleHome)))

	// Create a single server with proper TLS configuration
	server := &http.Server{
		Addr:              ":" + os.Getenv("PORT"),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Check if we're in production (Render sets this)
	isProduction := os.Getenv("RENDER") == "true"

	if isProduction {
		// In production, Render handles HTTPS termination
		log.Printf("Server running in production mode at: http://localhost:%v\n", os.Getenv("PORT"))
		log.Fatal(server.ListenAndServe())
	} else {
		// In development, use self-signed certificates
		log.Printf("Server running in development mode at: https://localhost:%v\n", os.Getenv("PORT"))
		log.Fatal(server.ListenAndServeTLS("certs/cert.pem", "certs/key.pem"))
	}
}

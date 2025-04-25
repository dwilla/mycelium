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

	server := &http.Server{
		Addr:              ":" + os.Getenv("PORT"),
		Handler:           mux,
		ReadHeaderTimeout: (10 * time.Second),
	}

	errChan := make(chan error)
	go func() {
		err := http.ListenAndServe(":80", http.HandlerFunc(redirect))
		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		if err := <-errChan; err != nil {
			log.Printf("Error in redirect server: %v", err)
		}
	}()

	log.Printf("Server running at: https://localhost%v\n", server.Addr)
	log.Fatal(server.ListenAndServeTLS("certs/cert.pem", "certs/key.pem"))
}

func redirect(w http.ResponseWriter, req *http.Request) { // Redirects to https
	http.Redirect(w, req,
		"https://"+req.Host+req.URL.String(),
		http.StatusMovedPermanently)
}

package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"starvault/core/internal/handlers"
	"starvault/core/internal/repository"
	"starvault/core/internal/services"
)

// Minimal env loader for MVP to avoid extra dependencies
func loadEnv(filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
		return
	}
	lines := string(content)
	for _, line := range splitLines(lines) {
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		var key, val string
		for i := 0; i < len(line); i++ {
			if line[i] == '=' {
				key = line[:i]
				val = line[i+1:]
				break
			}
		}
		if key != "" {
			os.Setenv(key, val)
		}
	}
}

func splitLines(s string) []string {
	var lines []string
	var start int
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func main() {
	// Load .env from parent directory
	loadEnv("../.env")

	dbUser := os.Getenv("POSTGRES_USER")
	dbPass := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	log.Println("Core Service connected to PostgreSQL successfully.")

	userRepo := &repository.UserRepository{DB: db}
	authSvc := &services.AuthService{UserRepo: userRepo}
	authHandler := &handlers.AuthHandler{AuthService: authSvc}

	port := os.Getenv("CORE_PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Core is healthy"))
	})

	http.HandleFunc("/internal/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authHandler.CreateUser(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/internal/users/verify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authHandler.VerifyUser(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Printf("Core Service starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

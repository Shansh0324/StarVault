package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"starvault/core/internal/handlers"
	"starvault/core/internal/repository"
	"starvault/core/internal/services"
	"encoding/hex"
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

	// Vault dependencies
	masterKeyHex := os.Getenv("STARVAULT_MASTER_KEY")
	if masterKeyHex == "" {
		log.Fatal("STARVAULT_MASTER_KEY is required")
	}
	masterKey, err := hex.DecodeString(masterKeyHex)
	if err != nil || len(masterKey) != 32 {
		log.Fatal("STARVAULT_MASTER_KEY must be a valid 64-character hex string")
	}
	encSvc, err := services.NewEncryptionService(masterKey)
	if err != nil {
		log.Fatal("Failed to initialize EncryptionService")
	}
	vaultRepo := &repository.VaultRepository{DB: db}
	vaultSvc := &services.VaultService{VaultRepo: vaultRepo, EncryptSvc: encSvc}
	vaultHandler := &handlers.VaultHandler{VaultService: vaultSvc}

	appRepo := &repository.AppRepository{DB: db}
	appSvc := &services.AppService{AppRepo: appRepo}
	appHandler := &handlers.AppHandler{AppService: appSvc}

	consentRepo := &repository.ConsentRepository{DB: db}
	consentSvc := &services.ConsentService{ConsentRepo: consentRepo, AppRepo: appRepo}
	consentHandler := &handlers.ConsentHandler{ConsentService: consentSvc}

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

	http.HandleFunc("/internal/vault/data", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			vaultHandler.CreateVaultData(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/internal/vault/data/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			vaultHandler.GetVaultData(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/internal/apps", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			appHandler.CreateApp(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/internal/consents", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			consentHandler.CreateConsent(w, r)
		} else if r.Method == http.MethodGet {
			consentHandler.ListConsents(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/internal/consents/check", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			consentHandler.CheckConsent(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/internal/consents/", func(w http.ResponseWriter, r *http.Request) {
		// Basic manual routing: /internal/consents/{id} OR /internal/consents/{id}/revoke
		path := r.URL.Path
		if r.Method == http.MethodGet {
			consentHandler.GetConsent(w, r)
		} else if r.Method == http.MethodPost && strings.HasSuffix(path, "/revoke") {
			consentHandler.RevokeConsent(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Printf("Core Service starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"encoding/json"
	"sync/atomic"

	_ "github.com/lib/pq"
	"starvault/core/internal/handlers"
	"starvault/core/internal/repository"
	"starvault/core/internal/services"
	"starvault/core/internal/blockchain"
	"encoding/hex"

	"github.com/nats-io/nats.go"
)

var isShuttingDown atomic.Bool
var httpRequestsTotal atomic.Uint64

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = "system"
		}
		
		if r.URL.Path != "/health/live" && r.URL.Path != "/health/ready" && r.URL.Path != "/metrics" {
			logData := map[string]string{
				"level":     "info",
				"service":   "core",
				"requestId": reqID,
				"event":     "request_started",
				"path":      r.URL.Path,
			}
			b, _ := json.Marshal(logData)
			log.Println(string(b))
		}
		
		httpRequestsTotal.Add(1)
		
		// Put request ID in context
		ctx := context.WithValue(r.Context(), "request_id", reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

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
			val = strings.TrimSpace(val)
			key = strings.TrimSpace(key)
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

	if dbUser == "" || dbPass == "" || dbName == "" || dbHost == "" {
		log.Fatal("POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB, and POSTGRES_HOST are required")
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbPort, dbName)
	log.Printf("DB Host: %q", dbHost)
	log.Printf("DB User: %q", dbUser)
	log.Printf("DB Pass len: %d", len(dbPass))
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	
	// Hardening: Connection Pool Limits
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	log.Println("Core Service connected to PostgreSQL successfully.")

	userRepo := &repository.UserRepository{DB: db}

	// Vault dependencies
	masterKeyHex := os.Getenv("STARVAULT_MASTER_KEY")
	if masterKeyHex == "" {
		log.Fatal("STARVAULT_MASTER_KEY is required")
	}
	masterKey, err := hex.DecodeString(masterKeyHex)
	if err != nil || len(masterKey) != 32 {
		log.Fatal("STARVAULT_MASTER_KEY must be a valid 64-character hex string")
	}


	vaultRepo := &repository.VaultRepository{DB: db}
	appRepo := &repository.AppRepository{DB: db}
	tokenRepo := &repository.TokenRepository{DB: db}
	consentRepo := &repository.ConsentRepository{DB: db}
	auditRepo := &repository.AuditRepository{DB: db}

	// Blockchain Client Initialization
	rpcURL := os.Getenv("BLOCKCHAIN_RPC_URL")
	privKey := os.Getenv("BLOCKCHAIN_PRIVATE_KEY")
	contractAddr := os.Getenv("SMART_CONTRACT_ADDRESS")
	
	var bcClient *blockchain.Client
	if rpcURL != "" && privKey != "" && contractAddr != "" {
		client, err := blockchain.NewClient(rpcURL, privKey, contractAddr)
		if err != nil {
			log.Printf("WARNING: Failed to initialize Blockchain client: %v", err)
		} else {
			bcClient = client
			log.Println("Blockchain integration enabled.")
		}
	} else {
		log.Println("Blockchain integration disabled (missing env vars).")
	}

	// NATS JetStream Initialization
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Failed to initialize JetStream: %v", err)
	}
	
	// Ensure stream exists
	_, err = js.StreamInfo("AUDIT")
	if err != nil {
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     "AUDIT",
			Subjects: []string{"audit.events"},
			Storage:  nats.FileStorage,
		})
		if err != nil {
			log.Fatalf("Failed to create NATS stream: %v", err)
		}
	}
	log.Println("NATS JetStream integration enabled.")

	auditService := &services.AuditService{
		AuditRepo:        auditRepo,
		BlockchainClient: bcClient,
		JetStream:        js,
		BatchInterval:    60 * time.Second,
		BatchMaxSize:     1000,
	}
	encryptSvc, err := services.NewEncryptionService(masterKey)
	if err != nil {
		log.Fatalf("Failed to initialize encryption service: %v", err)
	}

	appService := &services.AppService{AppRepo: appRepo}
	consentService := &services.ConsentService{
		ConsentRepo:  consentRepo,
		AppRepo:      appRepo,
	}
	tokenService := &services.TokenService{
		TokenRepo:      tokenRepo,
		AppRepo:        appRepo,
		ConsentService: consentService,
		AuditService:   auditService,
	}
	vaultService := &services.VaultService{
		VaultRepo:  vaultRepo,
		EncryptSvc: encryptSvc,
	}
	accessService := &services.AccessService{
		AppRepo:        appRepo,
		ConsentService: consentService,
		VaultService:   vaultService,
		AuditService:   auditService,
		TokenRepo:      tokenRepo,
	}
	riskService := &services.RiskService{DB: db}

	authService := &services.AuthService{
		UserRepo:    userRepo,
		RiskService: riskService,
	}

	webAuthnService, err := services.NewWebAuthnService(db)
	if err != nil {
		log.Printf("WARNING: WebAuthn initialization failed: %v", err)
	}

	authHandler := &handlers.AuthHandler{AuthService: authService}
	vaultHandler := &handlers.VaultHandler{VaultService: vaultService}
	appHandler := &handlers.AppHandler{AppService: appService}
	consentHandler := &handlers.ConsentHandler{ConsentService: consentService}
	auditHandler := &handlers.AuditHandler{AuditRepo: auditRepo}
	accessHandler := &handlers.AccessHandler{AccessService: accessService}
	tokenHandler := &handlers.TokenHandler{TokenService: tokenService}
	webAuthnHandler := &handlers.WebAuthnHandler{WebAuthnService: webAuthnService}

	port := os.Getenv("CORE_PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health/live", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"Core is alive"}`))
	})

	mux.HandleFunc("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		if isShuttingDown.Load() {
			http.Error(w, `{"status":"Shutting down"}`, http.StatusServiceUnavailable)
			return
		}
		if err := db.PingContext(r.Context()); err != nil {
			http.Error(w, `{"status":"Database down"}`, http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"Core is ready"}`))
	})

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "# HELP http_requests_total Total HTTP requests\n")
		fmt.Fprintf(w, "# TYPE http_requests_total counter\n")
		fmt.Fprintf(w, "http_requests_total{service=\"core\"} %d\n", httpRequestsTotal.Load())
	})

	mux.HandleFunc("/internal/auth/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authHandler.CreateUser(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/internal/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authHandler.VerifyUser(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/internal/auth/webauthn/register/begin", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			webAuthnHandler.RegisterBegin(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/internal/auth/webauthn/register/finish", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			webAuthnHandler.RegisterFinish(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/internal/auth/webauthn/login/begin", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			webAuthnHandler.LoginBegin(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/internal/auth/webauthn/login/finish", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			webAuthnHandler.LoginFinish(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/internal/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authHandler.CreateUser(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/internal/users/verify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			authHandler.VerifyUser(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/internal/vault/data", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			vaultHandler.CreateVaultData(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/internal/vault/data/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			vaultHandler.GetVaultData(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/internal/apps", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			appHandler.CreateApp(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/internal/consents", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			consentHandler.CreateConsent(w, r)
		} else if r.Method == http.MethodGet {
			consentHandler.ListConsents(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/internal/consents/check", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			consentHandler.CheckConsent(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/internal/consents/", func(w http.ResponseWriter, r *http.Request) {
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

	mux.HandleFunc("/internal/access/data", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			accessHandler.AccessData(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/internal/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			tokenHandler.IssueToken(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/internal/oauth/revoke", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			tokenHandler.RevokeToken(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/internal/audits/latest", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			auditHandler.GetLatestAudit(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/internal/audits/verify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			auditHandler.VerifyAudit(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Start asynchronous audit + batch workers with context
	workerCtx, workerCancel := context.WithCancel(context.Background())
	go auditService.StartAuditWorker(workerCtx)
	go auditService.StartBatchWorker(workerCtx)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("Core Service starting on port %s...", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Core shutting down gracefully...")
	isShuttingDown.Store(true)

	workerCancel() // stop background worker

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Core Server Shutdown Failed: %+v", err)
	}

	log.Println("Core exited cleanly")
}

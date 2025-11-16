package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bitbucket.org/Amartha/go-megatron/internal/config"
	"bitbucket.org/Amartha/go-megatron/internal/http"
	"bitbucket.org/Amartha/go-megatron/internal/repository"

	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()

	// ASCII banner
	printBanner()

	// Load config
	log.Println("📝 Loading configuration...")
	cfg, err := config.New(ctx)
	if err != nil {
		log.Fatal("❌ Failed to load config:", err)
	}
	log.Printf("✅ Configuration loaded (env: %s)\n", cfg.App.Env)

	// Setup database
	log.Println("🔌 Connecting to database...")
	db, err := setupDatabase(cfg)
	if err != nil {
		log.Fatal("❌ Failed to setup database:", err)
	}
	defer db.Close()
	log.Println("✅ Database connected successfully")

	// Setup repositories
	ruleRepo := repository.NewRuleRepository(db)

	// Setup HTTP API
	apiServer := http.NewAPI(cfg, ruleRepo)
	starter, stopper := apiServer.Start()

	// Start server in goroutine
	go func() {
		if err := starter(); err != nil {
			log.Fatal("❌ Failed to start API server:", err)
		}
	}()

	log.Println("✅ API Server is running")
	log.Println("📍 Endpoints available:")
	log.Println("   - GET  /")
	log.Println("   - GET  /health")
	log.Println("   - POST /api/v1/rules")
	log.Println("   - GET  /api/v1/rules")
	log.Println("   - GET  /api/v1/rules/:name")
	log.Println("   - PUT  /api/v1/rules/:id")
	log.Println("   - DELETE /api/v1/rules/:id")
	log.Println("\n⌨️  Press Ctrl+C to stop")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("\n⏳ Received shutdown signal...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := stopper(shutdownCtx); err != nil {
		log.Fatal("❌ Server forced to shutdown:", err)
	}

	log.Println("✅ Server exited successfully")
}

func setupDatabase(cfg *config.Configuration) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func printBanner() {
	banner := `
╔═══════════════════════════════════════════════╗
║                                               ║
║         GO-MEGATRON API SERVER                ║
║         Rule Management System                ║
║                                               ║
╚═══════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

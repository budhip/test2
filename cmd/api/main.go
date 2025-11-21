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
	"bitbucket.org/Amartha/go-megatron/internal/repositories"

	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()

	printBanner()

	log.Println("📝 Loading configuration...")
	cfg, err := config.New(ctx)
	if err != nil {
		log.Fatal("❌ Failed to load config:", err)
	}
	log.Printf("✅ Configuration loaded (env: %s)\n", cfg.App.Env)

	log.Println("🔌 Connecting to database...")
	db, err := setupDatabase(cfg)
	if err != nil {
		log.Fatal("❌ Failed to setup database:", err)
	}
	defer db.Close()
	log.Println("✅ Database connected successfully")

	ruleRepo := repositories.NewRuleRepository(db)
	acuanRulerepo := repositories.NewAcuanRuleRepository(db)

	apiServer := http.NewAPI(cfg, acuanRulerepo, ruleRepo)
	starter, stopper := apiServer.Start()

	go func() {
		if err := starter(); err != nil {
			log.Fatal("❌ Failed to start API server:", err)
		}
	}()

	log.Println("✅ API Server is running")
	log.Println("📍 Endpoints available:")
	log.Println("\n   🔄 Transformation:")
	log.Println("   - POST /api/v1/transform")
	log.Println("   - POST /api/v1/transform/batch")
	log.Println("   - POST /api/v1/transform/wallet")
	log.Println("\n   📋 Rules Management:")
	log.Println("   - POST   /api/v1/rules")
	log.Println("   - GET    /api/v1/rules")
	log.Println("   - GET    /api/v1/rules/:name")
	log.Println("   - PUT    /api/v1/rules/:id")
	log.Println("   - PATCH  /api/v1/rules/:id/append")
	log.Println("   - DELETE /api/v1/rules/:id")
	log.Println("\n⌨️  Press Ctrl+C to stop")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("\n⏳ Received shutdown signal...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := stopper(shutdownCtx); err != nil {
		log.Fatal("❌ Server forced to shutdown:", err)
	}

	log.Println("✅ Server exited successfully")
}

func setupDatabase(cfg *config.Configuration) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
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

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

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
║    Transformation & Rule Management           ║
║                                               ║
╚═══════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

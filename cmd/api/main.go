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

	ruleDomain "bitbucket.org/Amartha/go-megatron/internal/domain/rule"
	transformationDomain "bitbucket.org/Amartha/go-megatron/internal/domain/transformation"

	ruleCmd "bitbucket.org/Amartha/go-megatron/internal/application/command/rule"
	transformCmd "bitbucket.org/Amartha/go-megatron/internal/application/command/transformation"
	ruleQuery "bitbucket.org/Amartha/go-megatron/internal/application/query/rule"

	"bitbucket.org/Amartha/go-megatron/internal/infrastructure/cache"
	"bitbucket.org/Amartha/go-megatron/internal/infrastructure/engine/grule"

	ruleHandler "bitbucket.org/Amartha/go-megatron/internal/interfaces/http/handler/rule"
	transformHandler "bitbucket.org/Amartha/go-megatron/internal/interfaces/http/handler/transformation"

	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()

	printBanner()
	log.Println("📝 Starting Go-Megatron API Server...")

	cfg := loadConfig()

	log.Println("🔌 Connecting to database...")
	db, err := setupDatabase(cfg.Database)
	if err != nil {
		log.Fatal("❌ Failed to setup database:", err)
	}
	defer db.Close()
	log.Println("✅ Database connected successfully")

	log.Println("🔌 Connecting to Redis...")
	redisCache := cache.NewRedisCache(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	defer redisCache.Close()
	log.Println("✅ Redis connected successfully")

	log.Println("⚙️  Initializing infrastructure layer...")

	ruleRepo := postgres.NewRuleRepository(db)
	transformationRepo := postgres.NewTransformationRepository(db)

	gruleEngine := grule.NewGruleTransformationEngine()
	gruleValidator := grule.NewGruleValidator()

	log.Println("✅ Infrastructure layer initialized")

	log.Println("⚙️  Initializing domain layer...")

	ruleService := ruleDomain.NewRuleService(gruleValidator)

	transformationValidator := transformationDomain.NewTransformationValidator()
	transformationService := transformationDomain.NewTransformationService(
		gruleEngine,
		ruleRepo,
		transformationValidator,
	)

	log.Println("✅ Domain layer initialized")

	log.Println("⚙️  Initializing application layer...")

	createRuleHandler := ruleCmd.NewCreateRuleHandler(ruleRepo, ruleService)
	updateRuleHandler := ruleCmd.NewUpdateRuleHandler(ruleRepo, ruleService)
	appendRuleHandler := ruleCmd.NewAppendRuleHandler(ruleRepo, ruleService)
	deleteRuleHandler := ruleCmd.NewDeleteRuleHandler(ruleRepo)

	getRuleHandler := ruleQuery.NewGetRuleHandler(ruleRepo)
	listRulesHandler := ruleQuery.NewListRulesHandler(ruleRepo)

	transformWalletHandler := transformCmd.NewTransformWalletTransactionHandler(
		transformationService,
		transformationRepo,
	)

	log.Println("✅ Application layer initialized")

	log.Println("⚙️  Initializing interface layer...")

	createRuleHTTPHandler := ruleHandler.NewCreateRuleHTTPHandler(createRuleHandler)
	updateRuleHTTPHandler := ruleHandler.NewUpdateRuleHTTPHandler(updateRuleHandler)
	appendRuleHTTPHandler := ruleHandler.NewAppendRuleHTTPHandler(appendRuleHandler)
	deleteRuleHTTPHandler := ruleHandler.NewDeleteRuleHTTPHandler(deleteRuleHandler)
	getRuleHTTPHandler := ruleHandler.NewGetRuleHTTPHandler(getRuleHandler)
	listRulesHTTPHandler := ruleHandler.NewListRulesHTTPHandler(listRulesHandler)

	transformWalletHTTPHandler := transformHandler.NewTransformWalletHTTPHandler(transformWalletHandler)

	server := http.NewServer(
		cfg.App,
		createRuleHTTPHandler,
		updateRuleHTTPHandler,
		appendRuleHTTPHandler,
		deleteRuleHTTPHandler,
		getRuleHTTPHandler,
		listRulesHTTPHandler,
		transformWalletHTTPHandler,
	)

	log.Println("✅ Interface layer initialized")

	go func() {
		log.Printf("🚀 API Server starting on :%d (env: %s)\n", cfg.App.Port, cfg.App.Env)
		log.Println("   ⚡ Grule engine: ENABLED")
		log.Println("   📦 Clean Architecture: ENABLED")
		log.Println("\n📍 Endpoints available:")
		log.Println("   🔄 Transformation:")
		log.Println("      POST /api/v1/transform/wallet")
		log.Println("\n   📋 Rules Management:")
		log.Println("      POST   /api/v1/rules")
		log.Println("      GET    /api/v1/rules")
		log.Println("      GET    /api/v1/rules/:name")
		log.Println("      PUT    /api/v1/rules/:id")
		log.Println("      PATCH  /api/v1/rules/:id/append")
		log.Println("      DELETE /api/v1/rules/:id")
		log.Println("\n⌨️  Press Ctrl+C to stop")

		if err := server.Start(); err != nil {
			log.Fatal("❌ Failed to start API server:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("\n⏳ Received shutdown signal, gracefully shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal("❌ Server forced to shutdown:", err)
	}

	log.Println("✅ Server exited successfully")
}

func printBanner() {
	banner := `
╔═══════════════════════════════════════════════╗
║                                               ║
║         GO-MEGATRON API SERVER                ║
║    Clean Architecture Implementation          ║
║         Transformation Engine v2.0            ║
║                                               ║
╚═══════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
}

type AppConfig struct {
	Env  string
	Port int
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

func loadConfig() Config {
	return Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", "dev"),
			Port: getEnvInt("APP_PORT", 8080),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "megatron"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
	}
}

func setupDatabase(cfg DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		fmt.Sscanf(value, "%d", &result)
		return result
	}
	return defaultValue
}

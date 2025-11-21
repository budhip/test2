package main

import (
	"bitbucket.org/Amartha/go-megatron/internal/repositories"
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"bitbucket.org/Amartha/go-megatron/internal/config"
	"bitbucket.org/Amartha/go-megatron/internal/kafka/consumer"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/flag"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/graceful"
	"bitbucket.org/Amartha/go-megatron/internal/pkg/metrics"
	"bitbucket.org/Amartha/go-megatron/internal/rules"

	xlog "bitbucket.org/Amartha/go-x/log"
	_ "github.com/lib/pq"
	"github.com/newrelic/go-agent/v3/newrelic"
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

	// Initialize logger
	xlog.Init(cfg.App.Name)
	log.Println("✅ Logger initialized")

	// Setup database for rule loading
	log.Println("🔌 Connecting to database...")
	db, err := setupDatabase(cfg)
	if err != nil {
		log.Fatal("❌ Failed to setup database:", err)
	}
	defer db.Close()
	log.Println("✅ Database connected successfully")

	// Setup rule loader to use database
	log.Println("📚 Initializing rule loader...")
	ruleRepo := repositories.NewRuleRepository(db)

	// Use database loader (or hybrid if you want fallback to files)
	rules.RuleLoaderVariable = rules.NewDatabaseRuleLoader(ruleRepo)
	// Alternative: Use hybrid loader for fallback support
	// rules.RuleLoaderVariable = rules.NewHybridRuleLoader(ruleRepo)

	log.Println("✅ Rule loader initialized (using database)")

	// Setup New Relic (optional)
	var nr *newrelic.Application
	if cfg.NewRelicLicenseKey != "" {
		nr, err = newrelic.NewApplication(
			newrelic.ConfigAppName(cfg.App.Name),
			newrelic.ConfigLicense(cfg.NewRelicLicenseKey),
		)
		if err != nil {
			log.Printf("⚠️  Failed to initialize New Relic: %v", err)
		} else {
			log.Println("✅ New Relic initialized")
		}
	}

	// Setup metrics
	mtc := metrics.New()

	// Setup feature flag
	log.Println("🚩 Initializing feature flags...")
	flagClient, err := flag.New(cfg)
	if err != nil {
		log.Fatal("❌ Failed to initialize feature flag:", err)
	}
	log.Println("✅ Feature flags initialized")

	// Collect all stoppers
	var stoppers []graceful.ProcessStopper

	// Initialize consumers
	// Initialize consumers
	log.Println("🎯 Initializing Kafka consumers...")
	successCount := 0
	for _, consumerName := range consumer.ListConsumerName {
		log.Printf("   - Initializing consumer: %s", consumerName)

		// Use defer recover to catch panics
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("   ❌ Panic in consumer %s: %v", consumerName, r)
				}
			}()

			c, consumerStoppers, err := consumer.New(ctx, consumerName, cfg, nr, mtc, flagClient)
			if err != nil {
				log.Printf("   ❌ Failed to initialize consumer %s: %v", consumerName, err)
				return
			}

			if c != nil {
				stoppers = append(stoppers, consumerStoppers...)
				graceful.StartProcessAtBackground(c.Start())
				log.Printf("   ✅ Consumer %s started", consumerName)
				successCount++
			}
		}()
	}

	log.Printf("✅ Consumers initialized: %d/%d\n", successCount, len(consumer.ListConsumerName))

	if successCount == 0 {
		log.Fatal("❌ No consumers started successfully")
	}

	log.Println("\n🚀 Go-Megatron Consumer is running")
	log.Printf("📊 Active consumers: %d\n", len(consumer.ListConsumerName))
	log.Println("⌨️  Press Ctrl+C to stop\n")

	// Wait for shutdown signal
	graceful.StopProcessAtBackground(cfg.App.GracefulTimeout, stoppers...)
	log.Println("✅ All consumers stopped gracefully")
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
║         GO-MEGATRON CONSUMER                  ║
║         Event Processing System               ║
║                                               ║
╚═══════════════════════════════════════════════╝
`
	fmt.Println(banner)
}

package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"bitbucket.org/Amartha/go-megatron/internal/config"
	_ "github.com/lib/pq"
)

// This script migrates existing file-based rules to database

func main() {
	log.Println("🚀 Rule Migration Tool")
	log.Println("====================")

	ctx := context.Background()

	// Load config
	cfg, err := config.New(ctx)
	if err != nil {
		log.Fatal("❌ Failed to load config:", err)
	}

	// Setup database
	db, err := setupDatabase(cfg)
	if err != nil {
		log.Fatal("❌ Failed to setup database:", err)
	}
	defer db.Close()

	log.Println("✅ Database connected")

	// Setup repository
	repo := repositories.NewRuleRepository(db)

	// Define rules to migrate
	rulesDir := "./rules"
	environments := []string{"dev", "uat", "prod"}
	ruleFiles := []string{"papa.grl", "pas.grl", "lsm.grl", "bre.grl"}

	totalMigrated := 0
	totalFailed := 0

	// Migrate each rule
	for _, env := range environments {
		log.Printf("\n📁 Processing environment: %s", env)

		for _, ruleFile := range ruleFiles {
			path := filepath.Join(rulesDir, env, ruleFile)

			// Check if file exists
			if _, err := os.Stat(path); os.IsNotExist(err) {
				log.Printf("   ⏭️  Skipping %s (file not found)", ruleFile)
				continue
			}

			// Read file content
			content, err := ioutil.ReadFile(path)
			if err != nil {
				log.Printf("   ❌ Failed to read %s: %v", ruleFile, err)
				totalFailed++
				continue
			}

			// Create rule in database
			rule := &repositories.Rule{
				Name:     ruleFile,
				Env:      env,
				Version:  "latest",
				Content:  string(content),
				IsActive: true,
			}

			err = repo.CreateRule(ctx, rule)
			if err != nil {
				log.Printf("   ❌ Failed to migrate %s: %v", ruleFile, err)
				totalFailed++
				continue
			}

			log.Printf("   ✅ Migrated %s (ID: %d)", ruleFile, rule.ID)
			totalMigrated++
		}
	}

	// Summary
	log.Println("\n📊 Migration Summary")
	log.Println("===================")
	log.Printf("✅ Successfully migrated: %d rules", totalMigrated)
	log.Printf("❌ Failed: %d rules", totalFailed)

	if totalFailed == 0 {
		log.Println("\n🎉 Migration completed successfully!")
	} else {
		log.Println("\n⚠️  Migration completed with errors")
	}
}

func setupDatabase(cfg *config.Configuration) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
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

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"lexin-sqlite/internal/config"
	"lexin-sqlite/internal/database"
	"lexin-sqlite/internal/parser"
	"lexin-sqlite/internal/repository"
)

// Version information
const (
	Version   = "0.1.0"
	BuildTime = "dev"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	if cfg.ShowVersion {
		fmt.Printf("lexin-sqlite version %s (built at %s)\n", Version, BuildTime)
		os.Exit(0)
	}

	// Open database
	db, err := database.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	// Create repository
	repo := repository.New(db)

	// Parse XML file
	startTime := time.Now()
	log.Printf("Parsing XML file: %s", cfg.XMLFile)
	dict, err := parser.ParseXMLFile(cfg.XMLFile)
	if err != nil {
		log.Fatalf("Error parsing XML file: %v", err)
	}
	log.Printf("Parsed %d words in %v", len(dict.Words), time.Since(startTime))

	// Verify target language
	if dict.TargetLang != cfg.TargetLang {
		log.Printf("Warning: XML file has target language '%s', but you specified '%s'", dict.TargetLang, cfg.TargetLang)
	}

	// Store data in database
	log.Printf("Storing data in SQLite database: %s", cfg.DBPath)
	startTime = time.Now()
	err = repo.StoreDictionary(context.Background(), dict)
	if err != nil {
		log.Fatalf("Error storing dictionary: %v", err)
	}

	// Get entry count
	entryCount, err := db.CountDictionaryEntries(context.Background(), getDictID(db, dict.BaseLang, dict.TargetLang))
	if err != nil {
		log.Printf("Error counting entries: %v", err)
		entryCount = int64(len(dict.Words))
	}

	log.Printf("Successfully imported %d entries in %v", entryCount, time.Since(startTime))
	log.Printf("Dictionary from %s to %s is now available in %s", dict.BaseLang, dict.TargetLang, cfg.DBPath)
}

func getDictID(db *database.DB, baseLang, targetLang string) int64 {
	dictID, _, _, _, err := db.GetDictionaryByLanguages(context.Background(), baseLang, targetLang)
	if err != nil {
		return 0
	}
	return dictID
}

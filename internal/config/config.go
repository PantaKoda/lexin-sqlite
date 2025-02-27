package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds application configuration
type Config struct {
	XMLFile     string
	DBPath      string
	TargetLang  string
	ShowVersion bool
}

// Load loads configuration from command line arguments
func Load() (*Config, error) {
	config := &Config{}

	flag.StringVar(&config.XMLFile, "file", "", "Path to the XML dictionary file")
	flag.StringVar(&config.DBPath, "db", "lexin.db", "Path to the SQLite database file")
	flag.StringVar(&config.TargetLang, "target", "", "Target language code")
	flag.BoolVar(&config.ShowVersion, "version", false, "Show version information")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -file <xml-file> -target <language-code> [-db <database-path>]\n\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "Examples:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -file swedishenglish.xml -target english\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "  %s -file swedisharabic.xml -target arabic -db custom.db\n\n", os.Args[0])
		fmt.Fprintln(flag.CommandLine.Output(), "Flags:")
		flag.PrintDefaults()
	}

	flag.Parse()

	// Validate required parameters
	if config.ShowVersion {
		return config, nil
	}

	if config.XMLFile == "" {
		return nil, fmt.Errorf("XML file path is required")
	}

	if config.TargetLang == "" {
		return nil, fmt.Errorf("target language code is required")
	}

	// Check if XML file exists
	if _, err := os.Stat(config.XMLFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("XML file does not exist: %s", config.XMLFile)
	}

	// Create database directory if it doesn't exist
	dbDir := filepath.Dir(config.DBPath)
	if dbDir != "." {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	return config, nil
}

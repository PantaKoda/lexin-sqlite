package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "modernc.org/sqlite"
)

// Database schema
const schema = `
-- Dictionary metadata
CREATE TABLE IF NOT EXISTS dictionaries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    base_lang TEXT NOT NULL,
    target_lang TEXT NOT NULL,
    version TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Words
CREATE TABLE IF NOT EXISTS words (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    dictionary_id INTEGER NOT NULL,
    value TEXT NOT NULL,
    variant TEXT,
    type TEXT NOT NULL,
    original_id TEXT NOT NULL,
    variant_id TEXT NOT NULL,
    matching_id TEXT,
    FOREIGN KEY (dictionary_id) REFERENCES dictionaries(id) ON DELETE CASCADE
);

-- Base language information
CREATE TABLE IF NOT EXISTS base_langs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    word_id INTEGER NOT NULL,
    meaning TEXT,
    matching_id TEXT,
    FOREIGN KEY (word_id) REFERENCES words(id) ON DELETE CASCADE
);

-- References (animation, compare, phonetic, see)
CREATE TABLE IF NOT EXISTS references (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    base_lang_id INTEGER NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('animation', 'compare', 'phonetic', 'see')),
    value TEXT NOT NULL,
    matching_id TEXT,
    FOREIGN KEY (base_lang_id) REFERENCES base_langs(id) ON DELETE CASCADE
);

-- Comments for base language
CREATE TABLE IF NOT EXISTS comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    base_lang_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    matching_id TEXT,
    FOREIGN KEY (base_lang_id) REFERENCES base_langs(id) ON DELETE CASCADE
);

-- Explanations for base language
CREATE TABLE IF NOT EXISTS explanations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    base_lang_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    matching_id TEXT,
    FOREIGN KEY (base_lang_id) REFERENCES base_langs(id) ON DELETE CASCADE
);

-- Alternates for base language
CREATE TABLE IF NOT EXISTS alternates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    base_lang_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    FOREIGN KEY (base_lang_id) REFERENCES base_langs(id) ON DELETE CASCADE
);

-- Antonyms
CREATE TABLE IF NOT EXISTS antonyms (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    base_lang_id INTEGER,
    target_lang_id INTEGER,
    value TEXT NOT NULL,
    CHECK (base_lang_id IS NOT NULL OR target_lang_id IS NOT NULL),
    FOREIGN KEY (base_lang_id) REFERENCES base_langs(id) ON DELETE CASCADE,
    FOREIGN KEY (target_lang_id) REFERENCES target_langs(id) ON DELETE CASCADE
);

-- Usage for base language
CREATE TABLE IF NOT EXISTS usages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    base_lang_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    matching_id TEXT,
    FOREIGN KEY (base_lang_id) REFERENCES base_langs(id) ON DELETE CASCADE
);

-- Phonetics
CREATE TABLE IF NOT EXISTS phonetics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    base_lang_id INTEGER NOT NULL,
    content TEXT,
    file TEXT,
    FOREIGN KEY (base_lang_id) REFERENCES base_langs(id) ON DELETE CASCADE
);

-- Illustrations
CREATE TABLE IF NOT EXISTS illustrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    base_lang_id INTEGER NOT NULL,
    type TEXT NOT NULL,
    value TEXT NOT NULL,
    norlexin TEXT,
    FOREIGN KEY (base_lang_id) REFERENCES base_langs(id) ON DELETE CASCADE
);

-- Inflections
CREATE TABLE IF NOT EXISTS inflections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    base_lang_id INTEGER NOT NULL,
    content TEXT,
    FOREIGN KEY (base_lang_id) REFERENCES base_langs(id) ON DELETE CASCADE
);

-- Inflection variants
CREATE TABLE IF NOT EXISTS inflection_variants (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    inflection_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    description TEXT,
    FOREIGN KEY (inflection_id) REFERENCES inflections(id) ON DELETE CASCADE
);

-- Grammar information
CREATE TABLE IF NOT EXISTS graminfos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    base_lang_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    FOREIGN KEY (base_lang_id) REFERENCES base_langs(id) ON DELETE CASCADE
);

-- Examples
CREATE TABLE IF NOT EXISTS examples (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    base_lang_id INTEGER,
    target_lang_id INTEGER,
    content TEXT NOT NULL,
    original_id TEXT NOT NULL,
    matching_id TEXT,
    CHECK (base_lang_id IS NOT NULL OR target_lang_id IS NOT NULL),
    FOREIGN KEY (base_lang_id) REFERENCES base_langs(id) ON DELETE CASCADE,
    FOREIGN KEY (target_lang_id) REFERENCES target_langs(id) ON DELETE CASCADE
);

-- Idioms
CREATE TABLE IF NOT EXISTS idioms (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    base_lang_id INTEGER,
    target_lang_id INTEGER,
    content TEXT NOT NULL,
    original_id TEXT NOT NULL,
    matching_id TEXT,
    CHECK (base_lang_id IS NOT NULL OR target_lang_id IS NOT NULL),
    FOREIGN KEY (base_lang_id) REFERENCES base_langs(id) ON DELETE CASCADE,
    FOREIGN KEY (target_lang_id) REFERENCES target_langs(id) ON DELETE CASCADE
);

-- Compounds
CREATE TABLE IF NOT EXISTS compounds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    base_lang_id INTEGER,
    target_lang_id INTEGER,
    content TEXT,
    original_id TEXT NOT NULL,
    description TEXT,
    matching_id TEXT,
    CHECK (base_lang_id IS NOT NULL OR target_lang_id IS NOT NULL),
    FOREIGN KEY (base_lang_id) REFERENCES base_langs(id) ON DELETE CASCADE,
    FOREIGN KEY (target_lang_id) REFERENCES target_langs(id) ON DELETE CASCADE
);

-- Compound inflections
CREATE TABLE IF NOT EXISTS compound_inflections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    compound_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    FOREIGN KEY (compound_id) REFERENCES compounds(id) ON DELETE CASCADE
);

-- Derivations
CREATE TABLE IF NOT EXISTS derivations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    base_lang_id INTEGER,
    target_lang_id INTEGER,
    content TEXT,
    original_id TEXT NOT NULL,
    description TEXT,
    CHECK (base_lang_id IS NOT NULL OR target_lang_id IS NOT NULL),
    FOREIGN KEY (base_lang_id) REFERENCES base_langs(id) ON DELETE CASCADE,
    FOREIGN KEY (target_lang_id) REFERENCES target_langs(id) ON DELETE CASCADE
);

-- Derivation inflections
CREATE TABLE IF NOT EXISTS derivation_inflections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    derivation_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    FOREIGN KEY (derivation_id) REFERENCES derivations(id) ON DELETE CASCADE
);

-- Indexes
CREATE TABLE IF NOT EXISTS indexes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    base_lang_id INTEGER NOT NULL,
    value TEXT NOT NULL,
    type TEXT CHECK (type IN ('prefix', 'suffix')),
    FOREIGN KEY (base_lang_id) REFERENCES base_langs(id) ON DELETE CASCADE
);

-- Target language information
CREATE TABLE IF NOT EXISTS target_langs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    word_id INTEGER NOT NULL,
    comment TEXT,
    FOREIGN KEY (word_id) REFERENCES words(id) ON DELETE CASCADE
);

-- Translations
CREATE TABLE IF NOT EXISTS translations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    target_lang_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    FOREIGN KEY (target_lang_id) REFERENCES target_langs(id) ON DELETE CASCADE
);

-- Synonyms
CREATE TABLE IF NOT EXISTS synonyms (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    target_lang_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    FOREIGN KEY (target_lang_id) REFERENCES target_langs(id) ON DELETE CASCADE
);

-- Target language comments
CREATE TABLE IF NOT EXISTS target_comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    target_lang_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    FOREIGN KEY (target_lang_id) REFERENCES target_langs(id) ON DELETE CASCADE
);

-- Target language explanations
CREATE TABLE IF NOT EXISTS target_explanations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    target_lang_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    FOREIGN KEY (target_lang_id) REFERENCES target_langs(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_word_value ON words(value);
CREATE INDEX IF NOT EXISTS idx_dictionary_langs ON dictionaries(base_lang, target_lang);
CREATE INDEX IF NOT EXISTS idx_translation_content ON translations(content);
`

// DB is the database wrapper
type DB struct {
	db *sql.DB
}

// New creates a new database connection
func New(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set pragmas for performance
	_, err = db.Exec(`
		PRAGMA journal_mode = WAL;
		PRAGMA synchronous = NORMAL;
		PRAGMA foreign_keys = ON;
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to set pragmas: %w", err)
	}

	// Create schema
	if err := createSchema(db); err != nil {
		return nil, err
	}

	return &DB{db: db}, nil
}

// Close closes the database connection
func (d *DB) Close() error {
	return d.db.Close()
}

// createSchema creates the database schema
func createSchema(db *sql.DB) error {
	// Split statements by semicolon
	statements := strings.Split(schema, ";")
	
	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		
		_, err := db.Exec(statement)
		if err != nil {
			return fmt.Errorf("failed to execute schema statement: %w\nStatement: %s", err, statement)
		}
	}
	
	return nil
}

// GetDB returns the underlying database connection
func (d *DB) GetDB() *sql.DB {
	return d.db
}

// RunInTransaction runs a function within a transaction
func (d *DB) RunInTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	
	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("Transaction failed: %v, rollback failed: %v", err, rbErr)
			return fmt.Errorf("tx failed: %v, rollback failed: %v", err, rbErr)
		}
		return err
	}
	
	return tx.Commit()
}

// GetDictionaryByLanguages gets a dictionary by base and target languages
func (d *DB) GetDictionaryByLanguages(ctx context.Context, baseLang, targetLang string) (int64, string, string, string, error) {
	var id int64
	var base, target, version string
	
	err := d.db.QueryRowContext(ctx, `
		SELECT id, base_lang, target_lang, version 
		FROM dictionaries 
		WHERE base_lang = ? AND target_lang = ?
		LIMIT 1
	`, baseLang, targetLang).Scan(&id, &base, &target, &version)
	
	if err != nil {
		return 0, "", "", "", err
	}
	
	return id, base, target, version, nil
}

// CreateDictionary creates a new dictionary
func (d *DB) CreateDictionary(ctx context.Context, baseLang, targetLang, version string) (int64, error) {
	result, err := d.db.ExecContext(ctx, `
		INSERT INTO dictionaries (base_lang, target_lang, version)
		VALUES (?, ?, ?)
	`, baseLang, targetLang, version)
	
	if err != nil {
		return 0, err
	}
	
	return result.LastInsertId()
}

// CountDictionaryEntries counts the entries in a dictionary
func (d *DB) CountDictionaryEntries(ctx context.Context, dictionaryID int64) (int64, error) {
	var count int64
	
	err := d.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM words WHERE dictionary_id = ?
	`, dictionaryID).Scan(&count)
	
	if err != nil {
		return 0, err
	}
	
	return count, nil
}

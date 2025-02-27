# Lexin SQLite

A Go application that converts Lexin dictionary XML files to a SQLite database for easy querying and integration.

## Overview

Lexin provides multilingual dictionaries with Swedish as the base language and translations to multiple target languages. This tool parses Lexin's XML dictionary files and stores the data in a well-structured SQLite database.

## Features

- Parse Lexin XML dictionary files
- Store all dictionary data in SQLite database
- Preserve relationships between words, translations, examples, etc.
- Optimized database schema for efficient querying
- Simple command-line interface

## Installation

### Prerequisites

- Go 1.16 or higher
- Git (optional)

### Building from source

```bash
# Clone the repository
git clone https://github.com/yourusername/lexin-sqlite.git
cd lexin-sqlite

# Build the application
go build -o bin/lexin-sqlite ./cmd/lexin

# Or use make
make build
```

## Usage

```bash
# Basic usage
./bin/lexin-sqlite -file path/to/dictionary.xml -target targetlanguage

# Examples
./bin/lexin-sqlite -file swedishenglish.xml -target english
./bin/lexin-sqlite -file swedisharabic.xml -target arabic -db dictionaries/arabic.db

# Command-line options
-db string       Path to the SQLite database file (default "lexin.db")
-file string     Path to the XML dictionary file
-target string   Target language code
-version         Show version information
```

## Database Schema

The database schema closely follows the structure of the XML files, with tables for:
* `dictionaries`: Metadata about each dictionary
* `words`: Dictionary entries with original IDs and attributes
* `base_langs`: Information about words in the base language (Swedish)
* `target_langs`: Information about translations
* Additional tables for references, examples, idioms, compounds, inflections, etc.

## Example Queries

```sql
-- Get translations for a specific word
SELECT w.value as word, t.content as translation
FROM words w
JOIN target_langs tl ON w.id = tl.word_id
JOIN translations t ON tl.id = t.target_lang_id
WHERE w.value = 'hus';

-- Find words with examples
SELECT w.value as word, e.content as example
FROM words w
JOIN base_langs bl ON w.id = bl.word_id
JOIN examples e ON bl.id = e.base_lang_id
LIMIT 10;

-- Search for words starting with a prefix
SELECT w.value, t.content
FROM words w
JOIN base_langs bl ON w.id = bl.word_id
JOIN target_langs tl ON w.id = tl.word_id
JOIN translations t ON tl.id = t.target_lang_id
WHERE w.value LIKE 'a%'
ORDER BY w.value
LIMIT 20;
```

## License

MIT License

## Acknowledgements

* Lexin dictionary data is provided by the Swedish Institute for Language and Folklore and the Swedish National Agency for Education
* This project is not affiliated with the Lexin project or its maintainers
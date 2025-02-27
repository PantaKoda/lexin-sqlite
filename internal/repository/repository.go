package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"lexin-sqlite/internal/database"
	"lexin-sqlite/internal/parser"
)

// Repository handles dictionary data storage
type Repository struct {
	db *database.DB
}

// New creates a new repository
func New(db *database.DB) *Repository {
	return &Repository{db: db}
}

// StoreDictionary stores a dictionary in the database
func (r *Repository) StoreDictionary(ctx context.Context, dict *parser.Dictionary) error {
	return r.db.RunInTransaction(ctx, func(tx *sql.Tx) error {
		// Check if dictionary already exists
		dictID, _, _, _, err := r.db.GetDictionaryByLanguages(ctx, dict.BaseLang, dict.TargetLang)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to check if dictionary exists: %w", err)
		}

		if err == sql.ErrNoRows {
			// Create dictionary
			var err error
			dictID, err = r.db.CreateDictionary(ctx, dict.BaseLang, dict.TargetLang, dict.Version)
			if err != nil {
				return fmt.Errorf("failed to create dictionary: %w", err)
			}
		} else {
			log.Printf("Dictionary %s to %s already exists, adding/updating entries", dict.BaseLang, dict.TargetLang)
		}

		// Process each word
		for i, word := range dict.Words {
			if i > 0 && i%1000 == 0 {
				log.Printf("Processed %d words...", i)
			}
			
			if err := storeWord(tx, dictID, word); err != nil {
				return fmt.Errorf("failed to store word %s: %w", word.Value, err)
			}
		}

		return nil
	})
}

// storeWord stores a word and its related data
func storeWord(tx *sql.Tx, dictionaryID int64, word parser.Word) error {
	// Insert word
	wordStmt, err := tx.Prepare(`
		INSERT INTO words (dictionary_id, value, variant, type, original_id, variant_id, matching_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer wordStmt.Close()
	
	wordResult, err := wordStmt.Exec(
		dictionaryID,
		word.Value,
		nullString(word.Variant),
		word.Type,
		word.ID,
		word.VariantID,
		nullString(word.MatchingID),
	)
	if err != nil {
		return err
	}
	
	wordID, err := wordResult.LastInsertId()
	if err != nil {
		return err
	}
	
	// Process base language entries
	for _, baseLang := range word.BaseLangs {
		if err := storeBaseLang(tx, wordID, baseLang); err != nil {
			return fmt.Errorf("failed to store base language: %w", err)
		}
	}
	
	// Process target language entries
	for _, targetLang := range word.TargetLang {
		if err := storeTargetLang(tx, wordID, targetLang); err != nil {
			return fmt.Errorf("failed to store target language: %w", err)
		}
	}
	
	return nil
}

// storeBaseLang stores a BaseLang entry and its related data
func storeBaseLang(tx *sql.Tx, wordID int64, baseLang parser.BaseLang) error {
	// Insert base language
	baseLangStmt, err := tx.Prepare(`
		INSERT INTO base_langs (word_id, meaning, matching_id)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer baseLangStmt.Close()
	
	baseLangResult, err := baseLangStmt.Exec(
		wordID,
		nullString(baseLang.Meaning.Content),
		nullString(baseLang.Meaning.MatchingID),
	)
	if err != nil {
		return err
	}
	
	baseLangID, err := baseLangResult.LastInsertId()
	if err != nil {
		return err
	}
	
	// Store references
	if len(baseLang.References) > 0 {
		refStmt, err := tx.Prepare(`
			INSERT INTO references (base_lang_id, type, value, matching_id)
			VALUES (?, ?, ?, ?)
		`)
		if err != nil {
			return err
		}
		defer refStmt.Close()
		
		for _, ref := range baseLang.References {
			_, err = refStmt.Exec(
				baseLangID,
				ref.Type,
				ref.Value,
				nullString(ref.MatchingID),
			)
			if err != nil {
				return err
			}
		}
	}
	
	// Store comments
	if len(baseLang.Comments) > 0 {
		commentStmt, err := tx.Prepare(`
			INSERT INTO comments (base_lang_id, content, matching_id)
			VALUES (?, ?, ?)
		`)
		if err != nil {
			return err
		}
		defer commentStmt.Close()
		
		for _, comment := range baseLang.Comments {
			_, err = commentStmt.Exec(
				baseLangID,
				comment.Content,
				nullString(comment.MatchingID),
			)
			if err != nil {
				return err
			}
		}
	}
	
	// Store explanations
	if len(baseLang.Explanations) > 0 {
		explStmt, err := tx.Prepare(`
			INSERT INTO explanations (base_lang_id, content, matching_id)
			VALUES (?, ?, ?)
		`)
		if err != nil {
			return err
		}
		defer explStmt.Close()
		
		for _, expl := range baseLang.Explanations {
			_, err = explStmt.Exec(
				baseLangID,
				expl.Content,
				nullString(expl.MatchingID),
			)
			if err != nil {
				return err
			}
		}
	}
	
	// Store alternates
	if len(baseLang.Alternates) > 0 {
		altStmt, err := tx.Prepare(`
			INSERT INTO alternates (base_lang_id, content)
			VALUES (?, ?)
		`)
		if err != nil {
			return err
		}
		defer altStmt.Close()
		
		for _, alt := range baseLang.Alternates {
			_, err = altStmt.Exec(
				baseLangID,
				alt.Content,
			)
			if err != nil {
				return err
			}
		}
	}
	
	// Store antonyms
	if len(baseLang.Antonyms) > 0 {
		antStmt, err := tx.Prepare(`
			INSERT INTO antonyms (base_lang_id, value)
			VALUES (?, ?)
		`)
		if err != nil {
			return err
		}
		defer antStmt.Close()
		
		for _, ant := range baseLang.Antonyms {
			_, err = antStmt.Exec(
				baseLangID,
				ant.Value,
			)
			if err != nil {
				return err
			}
		}
	}
	
	// Store usages
	if len(baseLang.Usages) > 0 {
		usageStmt, err := tx.Prepare(`
			INSERT INTO usages (base_lang_id, content, matching_id)
			VALUES (?, ?, ?)
		`)
		if err != nil {
			return err
		}
		defer usageStmt.Close()
		
		for _, usage := range baseLang.Usages {
			_, err = usageStmt.Exec(
				baseLangID,
				usage.Content,
				nullString(usage.MatchingID),
			)
			if err != nil {
				return err
			}
		}
	}
	
	// Store phonetic
	if baseLang.Phonetic.Content != "" || baseLang.Phonetic.File != "" {
		_, err = tx.Exec(`
			INSERT INTO phonetics (base_lang_id, content, file)
			VALUES (?, ?, ?)
		`,
			baseLangID,
			nullString(baseLang.Phonetic.Content),
			nullString(baseLang.Phonetic.File),
		)
		if err != nil {
			return err
		}
	}
	
	// Store illustrations
	if len(baseLang.Illustrations) > 0 {
		illStmt, err := tx.Prepare(`
			INSERT INTO illustrations (base_lang_id, type, value, norlexin)
			VALUES (?, ?, ?, ?)
		`)
		if err != nil {
			return err
		}
		defer illStmt.Close()
		
		for _, ill := range baseLang.Illustrations {
			_, err = illStmt.Exec(
				baseLangID,
				ill.Type,
				ill.Value,
				nullString(ill.Norlexin),
			)
			if err != nil {
				return err
			}
		}
	}
	
	// Store inflections
	if len(baseLang.Inflections) > 0 {
		inflStmt, err := tx.Prepare(`
			INSERT INTO inflections (base_lang_id, content)
			VALUES (?, ?)
		`)
		if err != nil {
			return err
		}
		defer inflStmt.Close()
		
		variantStmt, err := tx.Prepare(`
			INSERT INTO inflection_variants (inflection_id, content, description)
			VALUES (?, ?, ?)
		`)
		if err != nil {
			return err
		}
		defer variantStmt.Close()
		
		for _, infl := range baseLang.Inflections {
			inflResult, err := inflStmt.Exec(
				baseLangID,
				nullString(infl.Content),
			)
			if err != nil {
				return err
			}
			
			inflID, err := inflResult.LastInsertId()
			if err != nil {
				return err
			}
			
			// Store variants
			for _, variant := range infl.Variants {
				_, err = variantStmt.Exec(
					inflID,
					variant.Content,
					nullString(variant.Description),
				)
				if err != nil {
					return err
				}
			}
		}
	}
	
	// Store grammar info
	if baseLang.Graminfo != "" {
		_, err = tx.Exec(`
			INSERT INTO graminfos (base_lang_id, content)
			VALUES (?, ?)
		`,
			baseLangID,
			baseLang.Graminfo,
		)
		if err != nil {
			return err
		}
	}
	
	// Store examples
	if len(baseLang.Examples) > 0 {
		exampleStmt, err := tx.Prepare(`
			INSERT INTO examples (base_lang_id, content, original_id, matching_id)
			VALUES (?, ?, ?, ?)
		`)
		if err != nil {
			return err
		}
		defer exampleStmt.Close()
		
		for _, example := range baseLang.Examples {
			_, err = exampleStmt.Exec(
				baseLangID,
				example.Content,
				example.ID,
				nullString(example.MatchingID),
			)
			if err != nil {
				return err
			}
		}
	}
	
	// Store idioms
	if len(baseLang.Idioms) > 0 {
		idiomStmt, err := tx.Prepare(`
			INSERT INTO idioms (base_lang_id, content, original_id, matching_id)
			VALUES (?, ?, ?, ?)
		`)
		if err != nil {
			return err
		}
		defer idiomStmt.Close()
		
		for _, idiom := range baseLang.Idioms {
			_, err = idiomStmt.Exec(
				baseLangID,
				idiom.Content,
				idiom.ID,
				nullString(idiom.MatchingID),
			)
			if err != nil {
				return err
			}
		}
	}
	
	// Store compounds
	if len(baseLang.Compounds) > 0 {
		compoundStmt, err := tx.Prepare(`
			INSERT INTO compounds (base_lang_id, content, original_id, description, matching_id)
			VALUES (?, ?, ?, ?, ?)
		`)
		if err != nil {
			return err
		}
		defer compoundStmt.Close()
		
		inflStmt, err := tx.Prepare(`
			INSERT INTO compound_inflections (compound_id, content)
			VALUES (?, ?)
		`)
		if err != nil {
			return err
		}
		defer inflStmt.Close()
		
		for _, compound := range baseLang.Compounds {
			compoundResult, err := compoundStmt.Exec(
				baseLangID,
				nullString(compound.Content),
				compound.ID,
				nullString(compound.Description),
				nullString(compound.MatchingID),
			)
			if err != nil {
				return err
			}
			
			if compound.Inflection != "" {
				compoundID, err := compoundResult.LastInsertId()
				if err != nil {
					return err
				}
				
				_, err = inflStmt.Exec(
					compoundID,
					compound.Inflection,
				)
				if err != nil {
					return err
				}
			}
		}
	}
	
	// Store derivations
	if len(baseLang.Derivations) > 0 {
		derivationStmt, err := tx.Prepare(`
			INSERT INTO derivations (base_lang_id, content, original_id, description)
			VALUES (?, ?, ?, ?)
		`)
		if err != nil {
			return err
		}
		defer derivationStmt.Close()
		
		inflStmt, err := tx.Prepare(`
			INSERT INTO derivation_inflections (derivation_id, content)
			VALUES (?, ?)
		`)
		if err != nil {
			return err
		}
		defer inflStmt.Close()
		
		for _, derivation := range baseLang.Derivations {
			derivationResult, err := derivationStmt.Exec(
				baseLangID,
				nullString(derivation.Content),
				derivation.ID,
				nullString(derivation.Description),
			)
			if err != nil {
				return err
			}
			
			if derivation.Inflection != "" {
				derivationID, err := derivationResult.LastInsertId()
				if err != nil {
					return err
				}
				
				_, err = inflStmt.Exec(
					derivationID,
					derivation.Inflection,
				)
				if err != nil {
					return err
				}
			}
		}
	}
	
	// Store indexes
	if len(baseLang.Indexes) > 0 {
		indexStmt, err := tx.Prepare(`
			INSERT INTO indexes (base_lang_id, value, type)
			VALUES (?, ?, ?)
		`)
		if err != nil {
			return err
		}
		defer indexStmt.Close()
		
		for _, index := range baseLang.Indexes {
			_, err = indexStmt.Exec(
				baseLangID,
				index.Value,
				nullString(index.Type),
			)
			if err != nil {
				return err
			}
		}
	}
	
	return nil
}

// storeTargetLang stores a TargetLang entry and its related data
func storeTargetLang(tx *sql.Tx, wordID int64, targetLang parser.TargetLang) error {
	// Insert target language
	targetLangStmt, err := tx.Prepare(`
		INSERT INTO target_langs (word_id, comment)
		VALUES (?, ?)
	`)
	if err != nil {
		return err
	}
	defer targetLangStmt.Close()
	
	targetLangResult, err := targetLangStmt.Exec(
		wordID,
		nullString(targetLang.Comment),
	)
	if err != nil {
		return err
	}
	
	targetLangID, err := targetLangResult.LastInsertId()
	if err != nil {
		return err
	}
	
	// Store translation
	if targetLang.Translation != "" {
		_, err = tx.Exec(`
			INSERT INTO translations (target_lang_id, content)
			VALUES (?, ?)
		`,
			targetLangID,
			targetLang.Translation,
		)
		if err != nil {
			return err
		}
	}
	
	// Store synonym
	if targetLang.Synonym != "" {
		_, err = tx.Exec(`
			INSERT INTO synonyms (target_lang_id, content)
			VALUES (?, ?)
		`,
			targetLangID,
			targetLang.Synonym,
		)
		if err != nil {
			return err
		}
	}
	
	// Store comment
	if targetLang.CommentElem != "" {
		_, err = tx.Exec(`
			INSERT INTO target_comments (target_lang_id, content)
			VALUES (?, ?)
		`,
			targetLangID,
			targetLang.CommentElem,
		)
		if err != nil {
			return err
		}
	}
	
	// Store explanation
	if targetLang.Explanation != "" {
		_, err = tx.Exec(`
			INSERT INTO target_explanations (target_lang_id, content)
			VALUES (?, ?)
		`,
			targetLangID,
			targetLang.Explanation,
		)
		if err != nil {
			return err
		}
	}
	
	// Store antonyms
	if len(targetLang.Antonyms) > 0 {
		antStmt, err := tx.Prepare(`
			INSERT INTO antonyms (target_lang_id, value)
			VALUES (?, ?)
		`)
		if err != nil {
			return err
		}
		defer antStmt.Close()
		
		for _, ant := range targetLang.Antonyms {
			_, err = antStmt.Exec(
				targetLangID,
				ant.Value,
			)
			if err != nil {
				return err
			}
		}
	}
	
	// Store examples
	if len(targetLang.Examples) > 0 {
		exampleStmt, err := tx.Prepare(`
			INSERT INTO examples (target_lang_id, content, original_id, matching_id)
			VALUES (?, ?, ?, ?)
		`)
		if err != nil {
			return err
		}
		defer exampleStmt.Close()
		
		for _, example := range targetLang.Examples {
			_, err = exampleStmt.Exec(
				targetLangID,
				example.Content,
				example.ID,
				nullString(example.MatchingID),
			)
			if err != nil {
				return err
			}
		}
	}
	
	// Store idioms
	if len(targetLang.Idioms) > 0 {
		idiomStmt, err := tx.Prepare(`
			INSERT INTO idioms (target_lang_id, content, original_id, matching_id)
			VALUES (?, ?, ?, ?)
		`)
		if err != nil {
			return err
		}
		defer idiomStmt.Close()
		
		for _, idiom := range targetLang.Idioms {
			_, err = idiomStmt.Exec(
				targetLangID,
				idiom.Content,
				idiom.ID,
				nullString(idiom.MatchingID),
			)
			if err != nil {
				return err
			}
		}
	}
	
	// Store compounds
	if len(targetLang.Compounds) > 0 {
		compoundStmt, err := tx.Prepare(`
			INSERT INTO compounds (target_lang_id, content, original_id, description, matching_id)
			VALUES (?, ?, ?, ?, ?)
		`)
		if err != nil {
			return err
		}
		defer compoundStmt.Close()
		
		inflStmt, err := tx.Prepare(`
			INSERT INTO compound_inflections (compound_id, content)
			VALUES (?, ?)
		`)
		if err != nil {
			return err
		}
		defer inflStmt.Close()
		
		for _, compound := range targetLang.Compounds {
			compoundResult, err := compoundStmt.Exec(
				targetLangID,
				nullString(compound.Content),
				compound.ID,
				nullString(compound.Description),
				nullString(compound.MatchingID),
			)
			if err != nil {
				return err
			}
			
			if compound.Inflection != "" {
				compoundID, err := compoundResult.LastInsertId()
				if err != nil {
					return err
				}
				
				_, err = inflStmt.Exec(
					compoundID,
					compound.Inflection,
				)
				if err != nil {
					return err
				}
			}
		}
	}
	
	// Store derivations
	if len(targetLang.Derivations) > 0 {
		derivationStmt, err := tx.Prepare(`
			INSERT INTO derivations (target_lang_id, content, original_id, description)
			VALUES (?, ?, ?, ?)
		`)
		if err != nil {
			return err
		}
		defer derivationStmt.Close()
		
		inflStmt, err := tx.Prepare(`
			INSERT INTO derivation_inflections (derivation_id, content)
			VALUES (?, ?)
		`)
		if err != nil {
			return err
		}
		defer inflStmt.Close()
		
		for _, derivation := range targetLang.Derivations {
			derivationResult, err := derivationStmt.Exec(
				targetLangID,
				nullString(derivation.Content),
				derivation.ID,
				nullString(derivation.Description),
			)
			if err != nil {
				return err
			}
			
			if derivation.Inflection != "" {
				derivationID, err := derivationResult.LastInsertId()
				if err != nil {
					return err
				}
				
				_, err = inflStmt.Exec(
					derivationID,
					derivation.Inflection,
				)
				if err != nil {
					return err
				}
			}
		}
	}
	
	return nil
}

// nullString returns a NULL value if the string is empty
func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

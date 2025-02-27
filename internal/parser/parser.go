package parser

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

// Dictionary represents the root XML element
type Dictionary struct {
	XMLName    xml.Name `xml:"Dictionary"`
	BaseLang   string   `xml:"BaseLang,attr"`
	TargetLang string   `xml:"TargetLang,attr"`
	Version    string   `xml:"Version,attr"`
	Words      []Word   `xml:"Word"`
}

// Word represents a dictionary entry
type Word struct {
	Value      string      `xml:"Value,attr"`
	Variant    string      `xml:"Variant,attr"`
	Type       string      `xml:"Type,attr"`
	ID         string      `xml:"ID,attr"`
	VariantID  string      `xml:"VariantID,attr"`
	MatchingID string      `xml:"MatchingID,attr"`
	BaseLangs  []BaseLang  `xml:"BaseLang"`
	TargetLang []TargetLang `xml:"TargetLang"`
}

// BaseLang represents the base language information
type BaseLang struct {
	Meaning      Meaning       `xml:"Meaning"`
	References   []Reference   `xml:"Reference"`
	Comments     []Comment     `xml:"Comment"`
	Explanations []Explanation `xml:"Explanation"`
	Alternates   []Alternate   `xml:"Alternate"`
	Antonyms     []Antonym     `xml:"Antonym"`
	Usages       []Usage       `xml:"Usage"`
	Phonetic     Phonetic      `xml:"Phonetic"`
	Illustrations []Illustration `xml:"Illustration"`
	Inflections  []Inflection   `xml:"Inflection"`
	Graminfo     string        `xml:"Graminfo"`
	Examples     []Example     `xml:"Example"`
	Idioms       []Idiom       `xml:"Idiom"`
	Compounds    []Compound    `xml:"Compound"`
	Derivations  []Derivation  `xml:"Derivation"`
	Indexes      []Index       `xml:"Index"`
}

// Meaning represents the meaning of a word
type Meaning struct {
	Content    string `xml:",chardata"`
	MatchingID string `xml:"MatchingID,attr"`
}

// Reference represents a reference to another word
type Reference struct {
	Type       string `xml:"TYPE,attr"`
	Value      string `xml:"VALUE,attr"`
	MatchingID string `xml:"MatchingID,attr"`
}

// Comment represents a comment on a word
type Comment struct {
	Content    string `xml:",chardata"`
	MatchingID string `xml:"MatchingID,attr"`
}

// Explanation represents an explanation of a word
type Explanation struct {
	Content    string `xml:",chardata"`
	MatchingID string `xml:"MatchingID,attr"`
}

// Alternate represents an alternate form of a word
type Alternate struct {
	Content string `xml:",chardata"`
}

// Antonym represents an antonym of a word
type Antonym struct {
	Value string `xml:"Value,attr"`
}

// Usage represents usage examples of a word
type Usage struct {
	Content    string `xml:",chardata"`
	MatchingID string `xml:"MatchingID,attr"`
}

// Phonetic represents phonetic information
type Phonetic struct {
	Content string `xml:",chardata"`
	File    string `xml:"File,attr"`
}

// Illustration represents an illustration
type Illustration struct {
	Type     string `xml:"TYPE,attr"`
	Value    string `xml:"VALUE,attr"`
	Norlexin string `xml:"Norlexin,attr"`
}

// Inflection represents inflection information
type Inflection struct {
	Content  string    `xml:",chardata"`
	Variants []Variant `xml:"Variant"`
}

// Variant represents a variant of an inflection
type Variant struct {
	Content     string `xml:",chardata"`
	Description string `xml:"Description,attr"`
}

// Example represents an example usage
type Example struct {
	Content    string `xml:",chardata"`
	ID         string `xml:"ID,attr"`
	MatchingID string `xml:"MatchingID,attr"`
}

// Idiom represents an idiom
type Idiom struct {
	Content    string `xml:",chardata"`
	ID         string `xml:"ID,attr"`
	MatchingID string `xml:"MatchingID,attr"`
}

// Compound represents a compound word
type Compound struct {
	Content     string     `xml:",chardata"`
	ID          string     `xml:"ID,attr"`
	Description string     `xml:"Description,attr"`
	MatchingID  string     `xml:"MatchingID,attr"`
	Inflection  string     `xml:"Inflection"`
}

// Derivation represents a derivation of a word
type Derivation struct {
	Content     string `xml:",chardata"`
	ID          string `xml:"ID,attr"`
	Description string `xml:"Description,attr"`
	Inflection  string `xml:"Inflection"`
}

// Index represents an index entry
type Index struct {
	Value string `xml:"Value,attr"`
	Type  string `xml:"type,attr"`
}

// TargetLang represents the target language information
type TargetLang struct {
	Comment     string      `xml:"Comment,attr"`
	Translation string      `xml:"Translation"`
	Synonym     string      `xml:"Synonym"`
	CommentElem string      `xml:"Comment"`
	Explanation string      `xml:"Explanation"`
	Antonyms    []Antonym   `xml:"Antonym"`
	Examples    []Example   `xml:"Example"`
	Idioms      []Idiom     `xml:"Idiom"`
	Compounds   []Compound  `xml:"Compound"`
	Derivations []Derivation `xml:"Derivation"`
}

// ParseXMLFile parses a Lexin XML file into a Dictionary struct
func ParseXMLFile(filePath string) (*Dictionary, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open XML file: %w", err)
	}
	defer file.Close()

	return ParseXML(file)
}

// ParseXML parses XML from a reader into a Dictionary struct
func ParseXML(r io.Reader) (*Dictionary, error) {
	decoder := xml.NewDecoder(r)
	
	var dictionary Dictionary
	if err := decoder.Decode(&dictionary); err != nil {
		return nil, fmt.Errorf("failed to decode XML: %w", err)
	}
	
	return &dictionary, nil
}

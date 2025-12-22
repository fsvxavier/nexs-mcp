package mcp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// SaveConversationContextInput defines input for save_conversation_context tool.
type SaveConversationContextInput struct {
	Context    string   `json:"context"              jsonschema:"conversation context to save as memory"`
	Summary    string   `json:"summary,omitempty"    jsonschema:"brief summary of the context"`
	Tags       []string `json:"tags,omitempty"       jsonschema:"tags for categorization"`
	Importance string   `json:"importance,omitempty" jsonschema:"importance level: low, medium, high, critical"`
	RelatedTo  []string `json:"related_to,omitempty" jsonschema:"IDs of related elements"`
}

// SaveConversationContextOutput defines output for save_conversation_context tool.
type SaveConversationContextOutput struct {
	MemoryID string `json:"memory_id"`
	Saved    bool   `json:"saved"`
	Message  string `json:"message"`
}

// handleSaveConversationContext handles automatic saving of conversation context.
func (s *MCPServer) handleSaveConversationContext(ctx context.Context, req *sdk.CallToolRequest, input SaveConversationContextInput) (*sdk.CallToolResult, SaveConversationContextOutput, error) {
	// Check if auto-save is enabled
	if !s.cfg.AutoSaveMemories {
		return nil, SaveConversationContextOutput{
			Saved:   false,
			Message: "Auto-save memories is disabled",
		}, nil
	}

	// Validate input
	if input.Context == "" || len(input.Context) < 10 {
		return nil, SaveConversationContextOutput{}, errors.New("context must be at least 10 characters")
	}

	// Generate name from summary or first line of context
	name := input.Summary
	if name == "" {
		lines := strings.Split(input.Context, "\n")
		name = lines[0]
		if len(name) > 80 {
			name = name[:80] + "..."
		}
	}

	// Create timestamp-based name
	timestamp := time.Now().Format("2006-01-02 15:04")
	memoryName := "Conversation Context - " + timestamp
	if name != "" {
		memoryName = name
	}

	// Create memory
	memory := domain.NewMemory(memoryName, input.Summary, "1.0.0", "auto-save")
	memory.Content = input.Context
	memory.ComputeHash()

	// Set tags
	tags := input.Tags
	if tags == nil {
		tags = []string{"auto-save", "conversation"}
	} else {
		tags = append(tags, "auto-save", "conversation")
	}

	// Add importance tag if specified
	if input.Importance != "" {
		tags = append(tags, "importance:"+input.Importance)
	}

	// Add search index based on content
	searchTerms := extractKeywords(input.Context, 10)
	memory.SearchIndex = searchTerms

	// Add metadata
	memory.Metadata = map[string]string{
		"auto_saved": "true",
		"saved_at":   time.Now().Format(time.RFC3339),
		"importance": input.Importance,
	}

	// Add related elements if specified
	if len(input.RelatedTo) > 0 {
		memory.Metadata["related_to"] = strings.Join(input.RelatedTo, ",")
	}

	// Set tags
	metadata := memory.GetMetadata()
	metadata.Tags = tags
	memory.SetMetadata(metadata)

	// Validate
	if err := memory.Validate(); err != nil {
		return nil, SaveConversationContextOutput{}, fmt.Errorf("memory validation failed: %w", err)
	}

	// Save to repository
	if err := s.repo.Create(memory); err != nil {
		return nil, SaveConversationContextOutput{}, fmt.Errorf("failed to save conversation context: %w", err)
	}

	output := SaveConversationContextOutput{
		MemoryID: memory.GetID(),
		Saved:    true,
		Message:  "Conversation context saved successfully",
	}

	return nil, output, nil
}

// Stop words by language for keyword extraction
var (
	stopWordsEnglish = map[string]bool{
		"a": true, "an": true, "and": true, "are": true, "as": true, "at": true,
		"be": true, "by": true, "for": true, "from": true, "has": true, "he": true,
		"in": true, "is": true, "it": true, "its": true, "of": true, "on": true,
		"that": true, "the": true, "to": true, "was": true, "will": true, "with": true,
		"she": true, "they": true, "we": true, "you": true, "or": true, "but": true,
		"if": true, "not": true, "this": true, "what": true, "when": true, "where": true,
		"who": true, "which": true, "can": true, "all": true, "any": true, "been": true,
		"being": true, "have": true, "had": true, "do": true, "does": true, "did": true,
		"would": true, "could": true, "should": true, "may": true, "might": true,
	}

	stopWordsPortuguese = map[string]bool{
		"o": true, "os": true, "um": true, "uma": true, "uns": true, "umas": true,
		"de": true, "da": true, "do": true, "dos": true, "das": true, "em": true,
		"no": true, "na": true, "nos": true, "nas": true, "para": true, "pelo": true,
		"pela": true, "pelos": true, "pelas": true, "com": true, "sem": true, "por": true,
		"ao": true, "à": true, "aos": true, "às": true, "foi": true, "ser": true,
		"está": true, "estão": true, "são": true, "essa": true, "esse": true, "esses": true,
		"essas": true, "isto": true, "isso": true, "aquele": true, "aquela": true,
		"mas": true, "mais": true, "como": true, "seu": true, "sua": true, "muito": true,
		"após": true, "antes": true, "durante": true, "sobre": true,
	}

	stopWordsSpanish = map[string]bool{
		"el": true, "la": true, "los": true, "las": true, "un": true, "una": true,
		"unos": true, "unas": true, "del": true, "en": true, "y": true, "que": true,
		"es": true, "por": true, "para": true, "con": true, "se": true, "lo": true,
		"pero": true, "su": true, "al": true, "más": true, "fue": true, "tiene": true,
		"todo": true, "esta": true, "este": true, "eso": true, "ese": true, "son": true,
		"mi": true, "muy": true, "sin": true, "sobre": true, "me": true, "ya": true,
	}

	stopWordsFrench = map[string]bool{
		"le": true, "la": true, "les": true, "un": true, "une": true, "des": true,
		"du": true, "et": true, "à": true, "dans": true, "pour": true, "par": true,
		"sur": true, "avec": true, "est": true, "sont": true, "que": true, "qui": true,
		"ce": true, "au": true, "aux": true, "ces": true, "cette": true, "cet": true,
		"il": true, "elle": true, "on": true, "nous": true, "vous": true, "ils": true,
		"elles": true, "ne": true, "pas": true, "plus": true, "ou": true, "mais": true,
		"comme": true, "tout": true, "son": true,
	}

	stopWordsGerman = map[string]bool{
		"der": true, "die": true, "das": true, "den": true, "dem": true, "des": true,
		"ein": true, "eine": true, "einer": true, "einem": true, "einen": true,
		"und": true, "zu": true, "von": true, "mit": true, "ist": true, "im": true,
		"für": true, "auf": true, "nicht": true, "sich": true, "als": true, "auch": true,
		"bei": true, "nach": true, "aus": true, "oder": true, "aber": true, "sind": true,
		"wird": true, "war": true, "hat": true, "sein": true, "vom": true, "zum": true,
		"zur": true, "am": true, "bis": true, "über": true, "wurde": true, "neue": true,
	}

	stopWordsItalian = map[string]bool{
		"il": true, "lo": true, "i": true, "gli": true, "un": true, "uno": true,
		"di": true, "da": true, "per": true, "con": true, "su": true, "è": true,
		"sono": true, "che": true, "del": true, "al": true, "alla": true, "nel": true,
		"nella": true, "dei": true, "degli": true, "ai": true, "alle": true, "anche": true,
		"si": true, "più": true, "non": true, "questo": true, "quella": true, "dal": true,
		"dalla": true, "sul": true, "sulla": true, "stata": true, "stato": true,
	}

	stopWordsRussian = map[string]bool{
		"и": true, "в": true, "не": true, "на": true, "я": true, "быть": true,
		"с": true, "что": true, "а": true, "по": true, "это": true, "как": true,
		"он": true, "она": true, "они": true, "или": true, "к": true, "у": true,
		"за": true, "из": true, "до": true, "от": true, "но": true, "же": true,
		"бы": true, "так": true, "для": true, "при": true, "то": true, "мы": true,
		"вы": true, "еще": true, "уже": true, "кто": true, "был": true, "была": true,
	}

	stopWordsJapanese = map[string]bool{
		"の": true, "に": true, "は": true, "を": true, "た": true, "が": true,
		"で": true, "て": true, "と": true, "し": true, "れ": true, "さ": true,
		"ある": true, "いる": true, "も": true, "する": true, "から": true, "な": true,
		"こと": true, "として": true, "い": true, "や": true, "れる": true, "など": true,
		"なっ": true, "には": true, "ず": true, "しかし": true, "その": true, "この": true,
	}

	stopWordsChinese = map[string]bool{
		"的": true, "是": true, "在": true, "了": true, "不": true, "和": true,
		"有": true, "人": true, "这": true, "中": true, "大": true, "为": true,
		"上": true, "个": true, "国": true, "我": true, "以": true, "要": true,
		"他": true, "时": true, "来": true, "用": true, "们": true, "到": true,
		"说": true, "子": true, "地": true, "于": true, "出": true, "就": true,
		"分": true, "对": true, "成": true, "会": true, "可": true, "主": true,
	}

	stopWordsArabic = map[string]bool{
		"في": true, "من": true, "على": true, "إلى": true, "أن": true, "هذا": true,
		"هذه": true, "التي": true, "الذي": true, "ما": true, "هو": true, "هي": true,
		"كان": true, "كانت": true, "لم": true, "لا": true, "ان": true, "او": true,
		"مع": true, "عن": true, "قد": true, "كل": true, "به": true, "لها": true,
		"له": true, "بها": true, "عند": true, "غير": true, "بعد": true, "قبل": true,
	}

	stopWordsHindi = map[string]bool{
		"का": true, "की": true, "के": true, "में": true, "है": true, "से": true,
		"को": true, "और": true, "एक": true, "यह": true, "पर": true, "था": true,
		"हैं": true, "कि": true, "जो": true, "साथ": true, "लिए": true, "या": true,
		"इस": true, "थी": true, "ने": true, "तो": true, "अपने": true, "हो": true,
		"कर": true, "ही": true, "गया": true, "गयी": true, "रहा": true, "रही": true,
	}
)

// detectLanguage performs simple language detection based on character patterns and common words
func detectLanguage(text string) string {
	textLower := strings.ToLower(text)
	words := strings.Fields(textLower)

	if len(words) == 0 {
		return "en" // default to English
	}

	// Count matches for each language
	languageScores := make(map[string]int)

	// Check for specific character sets (quick detection)
	for _, char := range text {
		if char >= 0x0600 && char <= 0x06FF { // Arabic
			languageScores["ar"] += 5
		} else if char >= 0x0400 && char <= 0x04FF { // Cyrillic (Russian)
			languageScores["ru"] += 5
		} else if char >= 0x3040 && char <= 0x309F { // Hiragana (Japanese)
			languageScores["ja"] += 5
		} else if char >= 0x4E00 && char <= 0x9FFF { // CJK (Chinese)
			languageScores["zh"] += 5
		} else if char >= 0x0900 && char <= 0x097F { // Devanagari (Hindi)
			languageScores["hi"] += 5
		}
	}

	// Check stop words presence (sample first 50 words for performance)
	sampleSize := 50
	if len(words) < sampleSize {
		sampleSize = len(words)
	}

	for i := 0; i < sampleSize; i++ {
		word := words[i]
		if stopWordsEnglish[word] {
			languageScores["en"]++
		}
		if stopWordsPortuguese[word] {
			languageScores["pt"]++
		}
		if stopWordsSpanish[word] {
			languageScores["es"]++
		}
		if stopWordsFrench[word] {
			languageScores["fr"]++
		}
		if stopWordsGerman[word] {
			languageScores["de"]++
		}
		if stopWordsItalian[word] {
			languageScores["it"]++
		}
		if stopWordsRussian[word] {
			languageScores["ru"]++
		}
		if stopWordsJapanese[word] {
			languageScores["ja"]++
		}
		if stopWordsChinese[word] {
			languageScores["zh"]++
		}
		if stopWordsArabic[word] {
			languageScores["ar"]++
		}
		if stopWordsHindi[word] {
			languageScores["hi"]++
		}
	}

	// Find language with highest score
	maxScore := 0
	detectedLang := "en" // default
	for lang, score := range languageScores {
		if score > maxScore {
			maxScore = score
			detectedLang = lang
		}
	}

	return detectedLang
}

// getStopWords returns the appropriate stop words map based on detected language
func getStopWords(lang string) map[string]bool {
	// Create combined map with detected language + English as fallback
	combined := make(map[string]bool)

	// Always include English (lingua franca for technical terms)
	for word := range stopWordsEnglish {
		combined[word] = true
	}

	// Add language-specific stop words
	var langStopWords map[string]bool
	switch lang {
	case "pt":
		langStopWords = stopWordsPortuguese
	case "es":
		langStopWords = stopWordsSpanish
	case "fr":
		langStopWords = stopWordsFrench
	case "de":
		langStopWords = stopWordsGerman
	case "it":
		langStopWords = stopWordsItalian
	case "ru":
		langStopWords = stopWordsRussian
	case "ja":
		langStopWords = stopWordsJapanese
	case "zh":
		langStopWords = stopWordsChinese
	case "ar":
		langStopWords = stopWordsArabic
	case "hi":
		langStopWords = stopWordsHindi
	default:
		// English only (already added)
		return combined
	}

	// Merge language-specific stop words
	for word := range langStopWords {
		combined[word] = true
	}

	return combined
}

// extractKeywords extracts relevant keywords from text for search indexing.
// Supports multilingual stop words removal (11+ languages) with automatic language detection.
func extractKeywords(text string, maxKeywords int) []string {
	// Simple keyword extraction - can be enhanced with NLP
	words := strings.Fields(strings.ToLower(text))

	// Detect language and get appropriate stop words
	detectedLang := detectLanguage(text)
	stopWords := getStopWords(detectedLang)

	// Count word frequency (excluding stop words)
	wordFreq := make(map[string]int)
	for _, word := range words {
		// Clean word (remove punctuation)
		word = strings.Trim(word, ".,!?;:()[]{}\"'")
		if len(word) < 3 || stopWords[word] {
			continue
		}
		wordFreq[word]++
	}

	// Get top keywords
	type wordCount struct {
		word  string
		count int
	}
	var counts []wordCount
	for word, count := range wordFreq {
		counts = append(counts, wordCount{word, count})
	}

	// Sort by frequency (simple bubble sort for small lists)
	for i := 0; i < len(counts); i++ {
		for j := i + 1; j < len(counts); j++ {
			if counts[j].count > counts[i].count {
				counts[i], counts[j] = counts[j], counts[i]
			}
		}
	}

	// Extract top keywords
	keywords := []string{}
	limit := maxKeywords
	if len(counts) < limit {
		limit = len(counts)
	}
	for i := range limit {
		keywords = append(keywords, counts[i].word)
	}

	return keywords
}

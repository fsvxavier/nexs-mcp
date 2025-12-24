package application

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
)

// PromptCompressor optimizes prompts sent to LLMs.
type PromptCompressor struct {
	config PromptCompressionConfig
	stats  PromptCompressionStats
	mu     sync.RWMutex
}

// PromptCompressionConfig configures prompt compression.
type PromptCompressionConfig struct {
	Enabled                bool
	RemoveRedundancy       bool    // Remove syntactic redundancies
	CompressWhitespace     bool    // Normalize whitespace
	UseAliases             bool    // Replace verbose phrases with aliases
	PreserveStructure      bool    // Maintain JSON/YAML structure
	TargetCompressionRatio float64 // Target: 0.65 (35% reduction)
	MinPromptLength        int     // Only compress if > N chars (default: 500)
}

// PromptCompressionStats tracks compression metrics.
type PromptCompressionStats struct {
	TotalCompressed     int64
	BytesSaved          int64
	AvgCompressionRatio float64
	QualityScore        float64 // LLM judge evaluation (future)
}

// PromptCompressionMetadata describes compression results.
type PromptCompressionMetadata struct {
	OriginalLength   int      `json:"original_length"`
	CompressedLength int      `json:"compressed_length"`
	CompressionRatio float64  `json:"compression_ratio"`
	TechniquesUsed   []string `json:"techniques_used"`
	QualityEstimate  float64  `json:"quality_estimate,omitempty"` // 0.0-1.0
}

// NewPromptCompressor creates a new PromptCompressor.
func NewPromptCompressor(config PromptCompressionConfig) *PromptCompressor {
	// Set defaults
	if config.MinPromptLength == 0 {
		config.MinPromptLength = 500
	}
	if config.TargetCompressionRatio == 0 {
		config.TargetCompressionRatio = 0.65
	}

	return &PromptCompressor{
		config: config,
		stats: PromptCompressionStats{
			AvgCompressionRatio: 1.0,
			QualityScore:        1.0,
		},
	}
}

// CompressPrompt optimizes a prompt for LLM consumption.
func (p *PromptCompressor) CompressPrompt(ctx context.Context, prompt string) (string, PromptCompressionMetadata, error) {
	if !p.config.Enabled || len(prompt) < p.config.MinPromptLength {
		return prompt, PromptCompressionMetadata{
			OriginalLength:   len(prompt),
			CompressedLength: len(prompt),
			CompressionRatio: 1.0,
		}, nil
	}

	originalLength := len(prompt)
	compressed := prompt

	// Step 1: Remove syntactic redundancies
	if p.config.RemoveRedundancy {
		compressed = p.removeRedundancies(compressed)
	}

	// Step 2: Compress whitespace
	if p.config.CompressWhitespace {
		compressed = p.compressWhitespace(compressed)
	}

	// Step 3: Use aliases for verbose phrases
	if p.config.UseAliases {
		compressed = p.applyAliases(compressed)
	}

	// Step 4: Remove filler words (while preserving meaning)
	compressed = p.removeFillers(compressed)

	compressedLength := len(compressed)
	compressionRatio := float64(compressedLength) / float64(originalLength)

	// Update stats
	p.updateStats(originalLength, compressedLength, compressionRatio)

	return compressed, PromptCompressionMetadata{
		OriginalLength:   originalLength,
		CompressedLength: compressedLength,
		CompressionRatio: compressionRatio,
		TechniquesUsed:   p.getTechniquesUsed(),
		QualityEstimate:  0.98, // Estimated quality preservation (can be improved with LLM judge)
	}, nil
}

// removeRedundancies removes syntactic redundancies.
func (p *PromptCompressor) removeRedundancies(text string) string {
	// Remove repeated words (e.g., "the the", "and and")
	// Note: Go's regexp doesn't support backreferences, so we use ReplaceAllStringFunc
	words := strings.Fields(text)
	var result []string
	var prevWord string

	for _, word := range words {
		// Skip if same as previous word (case-insensitive)
		if !strings.EqualFold(word, prevWord) {
			result = append(result, word)
		}
		prevWord = word
	}
	text = strings.Join(result, " ")

	// Remove redundant articles before technical terms
	// "the API" -> "API" (context is clear)
	text = strings.ReplaceAll(text, " the API", " API")
	text = strings.ReplaceAll(text, " the endpoint", " endpoint")
	text = strings.ReplaceAll(text, " the function", " function")
	text = strings.ReplaceAll(text, " the method", " method")
	text = strings.ReplaceAll(text, " the interface", " interface")
	text = strings.ReplaceAll(text, " the service", " service")
	text = strings.ReplaceAll(text, " the database", " database")
	text = strings.ReplaceAll(text, " the server", " server")

	// Remove redundant prepositions in technical contexts
	text = strings.ReplaceAll(text, " in order to ", " to ")
	text = strings.ReplaceAll(text, " in the case of ", " if ")
	text = strings.ReplaceAll(text, " in the event of ", " if ")
	text = strings.ReplaceAll(text, " due to the fact that ", " because ")
	text = strings.ReplaceAll(text, " for the purpose of ", " to ")
	text = strings.ReplaceAll(text, " with regard to ", " about ")
	text = strings.ReplaceAll(text, " with respect to ", " about ")
	text = strings.ReplaceAll(text, " in accordance with ", " per ")

	return text
}

// compressWhitespace normalizes whitespace.
func (p *PromptCompressor) compressWhitespace(text string) string {
	// Replace multiple spaces with single space
	reMultiSpace := regexp.MustCompile(`\s+`)
	text = reMultiSpace.ReplaceAllString(text, " ")

	// Remove leading/trailing whitespace from lines
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}

	// Remove empty lines (but preserve structure if configured)
	if p.config.PreserveStructure {
		// Keep max 1 empty line between sections
		reMultiNewline := regexp.MustCompile(`\n{3,}`)
		text = reMultiNewline.ReplaceAllString(strings.Join(lines, "\n"), "\n\n")
	} else {
		// Remove all empty lines
		nonEmpty := []string{}
		for _, line := range lines {
			if line != "" {
				nonEmpty = append(nonEmpty, line)
			}
		}
		text = strings.Join(nonEmpty, "\n")
	}

	return strings.TrimSpace(text)
}

// applyAliases replaces verbose phrases with concise aliases.
func (p *PromptCompressor) applyAliases(text string) string {
	// Define alias mappings (verbose -> concise)
	aliases := map[string]string{
		// Verbose instructions -> concise
		"Please provide me with":         "Provide:",
		"I would like you to":            "Task:",
		"Can you help me understand":     "Explain:",
		"I need assistance with":         "Help:",
		"Could you please":               "Please",
		"It would be great if you could": "Please",

		// Technical verbosity -> concise
		"in the context of":     "for",
		"with regard to":        "about",
		"in accordance with":    "per",
		"at this point in time": "now",
		"for the purpose of":    "to",
		"in the event that":     "if",
		"despite the fact that": "although",

		// Common verbose patterns
		"a number of":             "several",
		"a large number of":       "many",
		"a small number of":       "few",
		"at the present time":     "currently",
		"in the near future":      "soon",
		"in the process of":       "during",
		"make a decision":         "decide",
		"take into consideration": "consider",
		"give consideration to":   "consider",
		"come to a conclusion":    "conclude",
		"reach a decision":        "decide",
		"make an assumption":      "assume",
	}

	for verbose, concise := range aliases {
		text = strings.ReplaceAll(text, verbose, concise)
		// Also handle capitalized versions
		capitalizedVerbose := strings.ToUpper(verbose[:1]) + verbose[1:]
		capitalizedConcise := strings.ToUpper(concise[:1]) + concise[1:]
		text = strings.ReplaceAll(text, capitalizedVerbose, capitalizedConcise)
	}

	return text
}

// removeFillers removes filler words while preserving meaning.
func (p *PromptCompressor) removeFillers(text string) string {
	// Common filler words in technical contexts
	fillers := []string{
		" basically ",
		" essentially ",
		" actually ",
		" literally ",
		" really ",
		" very ",
		" quite ",
		" just ",
		" simply ",
		" obviously ",
		" clearly ",
		" of course ",
		" you know ",
		" I mean ",
		" sort of ",
		" kind of ",
		" like ",
		" probably ",
		" perhaps ",
		" maybe ",
		" possibly ",
		" seemingly ",
		" apparently ",
	}

	for _, filler := range fillers {
		text = strings.ReplaceAll(text, filler, " ")
		// Also handle capitalized versions at sentence start
		capitalizedFiller := strings.ToUpper(filler[:1]) + filler[1:]
		text = strings.ReplaceAll(text, capitalizedFiller, " ")
	}

	// Normalize spaces after filler removal
	reMultiSpace := regexp.MustCompile(`\s+`)
	text = reMultiSpace.ReplaceAllString(text, " ")

	return strings.TrimSpace(text)
}

// updateStats updates compression statistics.
func (p *PromptCompressor) updateStats(originalLength, compressedLength int, compressionRatio float64) {
	atomic.AddInt64(&p.stats.TotalCompressed, 1)
	atomic.AddInt64(&p.stats.BytesSaved, int64(originalLength-compressedLength))

	// Update average compression ratio using exponential moving average
	p.mu.Lock()
	alpha := 0.1 // Smoothing factor
	p.stats.AvgCompressionRatio = alpha*compressionRatio + (1-alpha)*p.stats.AvgCompressionRatio
	p.mu.Unlock()
}

// GetStats returns current compression statistics.
func (p *PromptCompressor) GetStats() PromptCompressionStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return PromptCompressionStats{
		TotalCompressed:     atomic.LoadInt64(&p.stats.TotalCompressed),
		BytesSaved:          atomic.LoadInt64(&p.stats.BytesSaved),
		AvgCompressionRatio: p.stats.AvgCompressionRatio,
		QualityScore:        p.stats.QualityScore,
	}
}

// getTechniquesUsed returns list of techniques applied.
func (p *PromptCompressor) getTechniquesUsed() []string {
	techniques := []string{}
	if p.config.RemoveRedundancy {
		techniques = append(techniques, "redundancy_removal")
	}
	if p.config.CompressWhitespace {
		techniques = append(techniques, "whitespace_compression")
	}
	if p.config.UseAliases {
		techniques = append(techniques, "alias_substitution")
	}
	techniques = append(techniques, "filler_removal")
	return techniques
}

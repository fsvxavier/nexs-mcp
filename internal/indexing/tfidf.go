package indexing

import (
	"math"
	"sort"
	"strings"
	"unicode"

	"github.com/fsvxavier/nexs-mcp/internal/domain"
)

// Document represents a searchable document with its content and metadata.
type Document struct {
	ID      string
	Type    domain.ElementType
	Name    string
	Content string // Concatenated searchable text
	Terms   map[string]int
}

// SearchResult represents a search result with relevance score.
type SearchResult struct {
	DocumentID string
	Type       domain.ElementType
	Name       string
	Score      float64
	Highlights []string
}

// TFIDFIndex implements TF-IDF (Term Frequency-Inverse Document Frequency) search.
type TFIDFIndex struct {
	documents      map[string]*Document
	idf            map[string]float64 // Inverse Document Frequency
	totalDocuments int
}

// NewTFIDFIndex creates a new TF-IDF search index.
func NewTFIDFIndex() *TFIDFIndex {
	return &TFIDFIndex{
		documents: make(map[string]*Document),
		idf:       make(map[string]float64),
	}
}

// AddDocument adds a document to the index.
func (idx *TFIDFIndex) AddDocument(doc *Document) {
	// Tokenize and count terms
	doc.Terms = tokenizeAndCount(doc.Content)
	idx.documents[doc.ID] = doc
	idx.totalDocuments++

	// Rebuild IDF after adding document
	idx.buildIDF()
}

// RemoveDocument removes a document from the index.
func (idx *TFIDFIndex) RemoveDocument(docID string) {
	if _, exists := idx.documents[docID]; exists {
		delete(idx.documents, docID)
		idx.totalDocuments--
		idx.buildIDF()
	}
}

// Search performs TF-IDF search and returns ranked results.
func (idx *TFIDFIndex) Search(query string, limit int) []SearchResult {
	if len(idx.documents) == 0 {
		return nil
	}

	// Tokenize query
	queryTerms := tokenizeAndCount(query)
	if len(queryTerms) == 0 {
		return nil
	}

	// Calculate scores for each document
	scores := make(map[string]float64)
	for docID, doc := range idx.documents {
		score := idx.calculateSimilarity(queryTerms, doc.Terms)
		if score > 0 {
			scores[docID] = score
		}
	}

	// Sort by score descending
	results := make([]SearchResult, 0, len(scores))
	for docID, score := range scores {
		doc := idx.documents[docID]
		results = append(results, SearchResult{
			DocumentID: docID,
			Type:       doc.Type,
			Name:       doc.Name,
			Score:      score,
			Highlights: extractHighlights(doc.Content, queryTerms, 3),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Limit results
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results
}

// FindSimilar finds documents similar to a given document.
func (idx *TFIDFIndex) FindSimilar(docID string, limit int) []SearchResult {
	doc, exists := idx.documents[docID]
	if !exists {
		return nil
	}

	// Calculate similarity with all other documents
	scores := make(map[string]float64)
	for otherID, otherDoc := range idx.documents {
		if otherID == docID {
			continue // Skip self
		}
		score := idx.calculateSimilarity(doc.Terms, otherDoc.Terms)
		if score > 0 {
			scores[otherID] = score
		}
	}

	// Sort and return
	results := make([]SearchResult, 0, len(scores))
	for otherID, score := range scores {
		otherDoc := idx.documents[otherID]
		results = append(results, SearchResult{
			DocumentID: otherID,
			Type:       otherDoc.Type,
			Name:       otherDoc.Name,
			Score:      score,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results
}

// buildIDF calculates Inverse Document Frequency for all terms.
func (idx *TFIDFIndex) buildIDF() {
	// Count documents containing each term
	termDocCount := make(map[string]int)
	for _, doc := range idx.documents {
		for term := range doc.Terms {
			termDocCount[term]++
		}
	}

	// Calculate IDF: log(total_docs / docs_with_term)
	idx.idf = make(map[string]float64)
	for term, docCount := range termDocCount {
		idx.idf[term] = math.Log(float64(idx.totalDocuments) / float64(docCount))
	}
}

// calculateSimilarity calculates cosine similarity between two term vectors.
func (idx *TFIDFIndex) calculateSimilarity(terms1, terms2 map[string]int) float64 {
	// Calculate TF-IDF vectors
	vec1 := idx.calculateTFIDF(terms1)
	vec2 := idx.calculateTFIDF(terms2)

	// Cosine similarity: dot(vec1, vec2) / (norm(vec1) * norm(vec2))
	dotProduct := 0.0
	norm1 := 0.0
	norm2 := 0.0

	for term, tfidf1 := range vec1 {
		dotProduct += tfidf1 * vec2[term]
		norm1 += tfidf1 * tfidf1
	}

	for _, tfidf2 := range vec2 {
		norm2 += tfidf2 * tfidf2
	}

	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// calculateTFIDF calculates TF-IDF vector for term counts.
func (idx *TFIDFIndex) calculateTFIDF(terms map[string]int) map[string]float64 {
	tfidf := make(map[string]float64)
	maxFreq := 0

	// Find max term frequency for normalization
	for _, count := range terms {
		if count > maxFreq {
			maxFreq = count
		}
	}

	if maxFreq == 0 {
		return tfidf
	}

	// Calculate TF-IDF for each term
	for term, count := range terms {
		tf := float64(count) / float64(maxFreq)
		idf := idx.idf[term]
		tfidf[term] = tf * idf
	}

	return tfidf
}

// tokenizeAndCount tokenizes text and counts term frequencies.
func tokenizeAndCount(text string) map[string]int {
	terms := make(map[string]int)
	text = strings.ToLower(text)

	// Split into words
	var currentWord strings.Builder
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			currentWord.WriteRune(r)
		} else {
			if currentWord.Len() > 0 {
				word := currentWord.String()
				if len(word) >= 2 { // Ignore single characters
					terms[word]++
				}
				currentWord.Reset()
			}
		}
	}

	// Don't forget the last word
	if currentWord.Len() > 0 {
		word := currentWord.String()
		if len(word) >= 2 {
			terms[word]++
		}
	}

	return terms
}

// extractHighlights extracts relevant text snippets containing query terms.
func extractHighlights(content string, queryTerms map[string]int, maxHighlights int) []string {
	content = strings.ToLower(content)
	highlights := make([]string, 0, maxHighlights)
	sentences := strings.Split(content, ".")

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) < 10 {
			continue
		}

		// Check if sentence contains any query terms
		for term := range queryTerms {
			if strings.Contains(sentence, term) {
				highlights = append(highlights, sentence)
				break
			}
		}

		if len(highlights) >= maxHighlights {
			break
		}
	}

	return highlights
}

// GetStats returns statistics about the index.
func (idx *TFIDFIndex) GetStats() map[string]interface{} {
	totalTerms := len(idx.idf)
	avgTermsPerDoc := 0
	if idx.totalDocuments > 0 {
		termCount := 0
		for _, doc := range idx.documents {
			termCount += len(doc.Terms)
		}
		avgTermsPerDoc = termCount / idx.totalDocuments
	}

	typeCount := make(map[domain.ElementType]int)
	for _, doc := range idx.documents {
		typeCount[doc.Type]++
	}

	return map[string]interface{}{
		"total_documents":    idx.totalDocuments,
		"total_unique_terms": totalTerms,
		"avg_terms_per_doc":  avgTermsPerDoc,
		"documents_by_type":  typeCount,
	}
}

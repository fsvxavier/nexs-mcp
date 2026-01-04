//go:build !noonnx && integration
// +build !noonnx,integration

package integration_test

import (
	"context"
	"os"
	"testing"

	"github.com/fsvxavier/nexs-mcp/internal/application"
	"github.com/fsvxavier/nexs-mcp/internal/mcp"
)

// Benchmark entity extraction with real BERT model
// Run with: go test -tags "integration" -bench=BenchmarkONNX -benchmem -benchtime=10x
func BenchmarkONNXBERTProvider_ExtractEntities(b *testing.B) {
	entityModel := "../../models/bert-base-ner/model.onnx"

	if _, err := os.Stat(entityModel); os.IsNotExist(err) {
		b.Skipf("Entity model not found at %s", entityModel)
	}

	config := application.EnhancedNLPConfig{
		EntityModel:         entityModel,
		SentimentModel:      "../../models/distilbert-sentiment/model.onnx",
		EntityConfidenceMin: 0.7,
		BatchSize:           16,
		MaxLength:           512,
		UseGPU:              false,
		EnableFallback:      true,
	}

	provider, err := application.NewONNXBERTProvider(config)
	if err != nil {
		b.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Close()

	if !provider.IsAvailable() {
		b.Skip("ONNX provider not available")
	}

	ctx := context.Background()
	text := "John Smith works at Google in Mountain View, California. He founded the AI Research Lab in 2020 and has published several papers on machine learning."

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := provider.ExtractEntities(ctx, text)
		if err != nil {
			b.Fatalf("ExtractEntities() error = %v", err)
		}
	}
}

func BenchmarkONNXBERTProvider_ExtractEntities_Short(b *testing.B) {
	entityModel := "../../models/bert-base-ner/model.onnx"

	if _, err := os.Stat(entityModel); os.IsNotExist(err) {
		b.Skipf("Entity model not found")
	}

	config := application.EnhancedNLPConfig{
		EntityModel:    entityModel,
		SentimentModel: "../../models/distilbert-sentiment/model.onnx",
		BatchSize:      16,
		MaxLength:      512,
		UseGPU:         false,
	}

	provider, err := application.NewONNXBERTProvider(config)
	if err != nil {
		b.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Close()

	ctx := context.Background()
	text := "Apple Inc. is in Cupertino."

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		provider.ExtractEntities(ctx, text)
	}
}

func BenchmarkONNXBERTProvider_ExtractEntities_Long(b *testing.B) {
	entityModel := "../../models/bert-base-ner/model.onnx"

	if _, err := os.Stat(entityModel); os.IsNotExist(err) {
		b.Skipf("Entity model not found")
	}

	config := application.EnhancedNLPConfig{
		EntityModel:    entityModel,
		SentimentModel: "../../models/distilbert-sentiment/model.onnx",
		BatchSize:      16,
		MaxLength:      512,
		UseGPU:         false,
	}

	provider, err := application.NewONNXBERTProvider(config)
	if err != nil {
		b.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Close()

	ctx := context.Background()
	text := `Barack Obama was born in Honolulu, Hawaii on August 4, 1961. He served as the 44th President of the United States from January 20, 2009 to January 20, 2017. Before his presidency, Obama represented Illinois in the U.S. Senate from 2005 to 2008. He previously served in the Illinois State Senate from 1997 to 2004. Obama graduated from Columbia University in 1983 with a degree in political science and later earned his law degree from Harvard Law School in 1991, where he was the first African-American president of the Harvard Law Review. After graduating, he became a civil rights attorney and taught constitutional law at the University of Chicago Law School from 1992 to 2004.`

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		provider.ExtractEntities(ctx, text)
	}
}

func BenchmarkONNXBERTProvider_ExtractEntitiesBatch(b *testing.B) {
	entityModel := "../../models/bert-base-ner/model.onnx"

	if _, err := os.Stat(entityModel); os.IsNotExist(err) {
		b.Skipf("Entity model not found")
	}

	config := application.EnhancedNLPConfig{
		EntityModel:    entityModel,
		SentimentModel: "../../models/distilbert-sentiment/model.onnx",
		BatchSize:      16,
		MaxLength:      512,
		UseGPU:         false,
	}

	provider, err := application.NewONNXBERTProvider(config)
	if err != nil {
		b.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Close()

	ctx := context.Background()
	texts := []string{
		"Apple Inc. is based in Cupertino, California.",
		"Tim Cook is the CEO of Apple.",
		"Microsoft was founded by Bill Gates.",
		"Google headquarters is in Mountain View.",
		"Amazon CEO is Andy Jassy.",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		provider.ExtractEntitiesBatch(ctx, texts)
	}
}

// Benchmark sentiment analysis with real DistilBERT model
func BenchmarkONNXBERTProvider_AnalyzeSentiment(b *testing.B) {
	sentimentModel := "../../models/distilbert-sentiment/model.onnx"

	if _, err := os.Stat(sentimentModel); os.IsNotExist(err) {
		b.Skipf("Sentiment model not found at %s", sentimentModel)
	}

	config := application.EnhancedNLPConfig{
		EntityModel:        "../../models/bert-base-ner/model.onnx",
		SentimentModel:     sentimentModel,
		SentimentThreshold: 0.6,
		BatchSize:          16,
		MaxLength:          512,
		UseGPU:             false,
		EnableFallback:     true,
	}

	provider, err := application.NewONNXBERTProvider(config)
	if err != nil {
		b.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Close()

	if !provider.IsAvailable() {
		b.Skip("ONNX provider not available")
	}

	ctx := context.Background()
	text := "This is absolutely amazing! I love this product and highly recommend it to everyone."

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := provider.AnalyzeSentiment(ctx, text)
		if err != nil {
			b.Fatalf("AnalyzeSentiment() error = %v", err)
		}
	}
}

func BenchmarkONNXBERTProvider_AnalyzeSentiment_Short(b *testing.B) {
	sentimentModel := "../../models/distilbert-sentiment/model.onnx"

	if _, err := os.Stat(sentimentModel); os.IsNotExist(err) {
		b.Skipf("Sentiment model not found")
	}

	config := application.EnhancedNLPConfig{
		EntityModel:    "../../models/bert-base-ner/model.onnx",
		SentimentModel: sentimentModel,
		BatchSize:      16,
		MaxLength:      512,
		UseGPU:         false,
	}

	provider, err := application.NewONNXBERTProvider(config)
	if err != nil {
		b.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Close()

	ctx := context.Background()
	text := "Great product!"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		provider.AnalyzeSentiment(ctx, text)
	}
}

func BenchmarkONNXBERTProvider_AnalyzeSentiment_Long(b *testing.B) {
	sentimentModel := "../../models/distilbert-sentiment/model.onnx"

	if _, err := os.Stat(sentimentModel); os.IsNotExist(err) {
		b.Skipf("Sentiment model not found")
	}

	config := application.EnhancedNLPConfig{
		EntityModel:    "../../models/bert-base-ner/model.onnx",
		SentimentModel: sentimentModel,
		BatchSize:      16,
		MaxLength:      512,
		UseGPU:         false,
	}

	provider, err := application.NewONNXBERTProvider(config)
	if err != nil {
		b.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Close()

	ctx := context.Background()
	text := `I recently purchased this product and I must say I am extremely impressed with the quality and performance. The design is sleek and modern, fitting perfectly into my workspace. The functionality exceeds my expectations in every way. Setup was incredibly easy and straightforward. Customer service was also outstanding when I had a few questions. I would highly recommend this to anyone looking for a reliable and high-quality solution. It's worth every penny and I'm confident it will last for years to come. Five stars without hesitation!`

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		provider.AnalyzeSentiment(ctx, text)
	}
}

// Benchmark combined operations
func BenchmarkONNXBERTProvider_Combined_EntityAndSentiment(b *testing.B) {
	entityModel := "../../models/bert-base-ner/model.onnx"
	sentimentModel := "../../models/distilbert-sentiment/model.onnx"

	if _, err := os.Stat(entityModel); os.IsNotExist(err) {
		b.Skip("Models not found")
	}

	config := application.EnhancedNLPConfig{
		EntityModel:    entityModel,
		SentimentModel: sentimentModel,
		BatchSize:      16,
		MaxLength:      512,
		UseGPU:         false,
	}

	provider, err := application.NewONNXBERTProvider(config)
	if err != nil {
		b.Fatalf("Failed to create provider: %v", err)
	}
	defer provider.Close()

	ctx := context.Background()
	text := "John Smith at Google created an amazing AI product. The team is thrilled with the results!"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		provider.ExtractEntities(ctx, text)
		provider.AnalyzeSentiment(ctx, text)
	}
}

// Benchmark SentimentAnalyzer service
func BenchmarkSentimentAnalyzer_AnalyzeSentiment(b *testing.B) {
	sentimentModel := "../../models/distilbert-sentiment/model.onnx"

	if _, err := os.Stat(sentimentModel); os.IsNotExist(err) {
		b.Skipf("Sentiment model not found")
	}

	config := application.EnhancedNLPConfig{
		EntityModel:    "../../models/bert-base-ner/model.onnx",
		SentimentModel: sentimentModel,
		BatchSize:      16,
		MaxLength:      512,
		UseGPU:         false,
	}

	provider, _ := application.NewONNXBERTProvider(config)
	defer provider.Close()

	repo := mcp.NewMockElementRepository()
	analyzer := application.NewSentimentAnalyzer(config, repo, provider)

	ctx := context.Background()
	text := "Excellent product! Very satisfied with the purchase."

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		analyzer.AnalyzeText(ctx, text)
	}
}

func BenchmarkSentimentAnalyzer_AnalyzeSentimentBatch(b *testing.B) {
	sentimentModel := "../../models/distilbert-sentiment/model.onnx"

	if _, err := os.Stat(sentimentModel); os.IsNotExist(err) {
		b.Skipf("Sentiment model not found")
	}

	config := application.EnhancedNLPConfig{
		EntityModel:    "../../models/bert-base-ner/model.onnx",
		SentimentModel: sentimentModel,
		BatchSize:      16,
		MaxLength:      512,
		UseGPU:         false,
	}

	provider, _ := application.NewONNXBERTProvider(config)
	defer provider.Close()

	repo := mcp.NewMockElementRepository()
	analyzer := application.NewSentimentAnalyzer(config, repo, provider)

	ctx := context.Background()
	texts := []string{
		"Great experience!",
		"Terrible service.",
		"It's okay.",
		"Absolutely love it!",
		"Very disappointing.",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, text := range texts {
			analyzer.AnalyzeText(ctx, text)
		}
	}
}

// Benchmark EnhancedEntityExtractor service
func BenchmarkEntityExtractor_ExtractEntities(b *testing.B) {
	entityModel := "../../models/bert-base-ner/model.onnx"

	if _, err := os.Stat(entityModel); os.IsNotExist(err) {
		b.Skipf("Entity model not found")
	}

	config := application.EnhancedNLPConfig{
		EntityModel:    entityModel,
		SentimentModel: "../../models/distilbert-sentiment/model.onnx",
		BatchSize:      16,
		MaxLength:      512,
		UseGPU:         false,
	}

	provider, _ := application.NewONNXBERTProvider(config)
	defer provider.Close()

	repo := mcp.NewMockElementRepository()
	extractor := application.NewEnhancedEntityExtractor(config, repo, provider)

	ctx := context.Background()
	text := "Steve Jobs founded Apple Computer in Cupertino, California in 1976."

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		extractor.ExtractFromText(ctx, text)
	}
}

func BenchmarkEntityExtractor_ExtractEntitiesBatch(b *testing.B) {
	entityModel := "../../models/bert-base-ner/model.onnx"

	if _, err := os.Stat(entityModel); os.IsNotExist(err) {
		b.Skipf("Entity model not found")
	}

	config := application.EnhancedNLPConfig{
		EntityModel:    entityModel,
		SentimentModel: "../../models/distilbert-sentiment/model.onnx",
		BatchSize:      16,
		MaxLength:      512,
		UseGPU:         false,
	}

	provider, _ := application.NewONNXBERTProvider(config)
	defer provider.Close()

	repo := mcp.NewMockElementRepository()
	extractor := application.NewEnhancedEntityExtractor(config, repo, provider)

	ctx := context.Background()
	texts := []string{
		"Apple Inc. is in Cupertino.",
		"Google was founded by Larry Page.",
		"Microsoft headquarters is in Redmond.",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, text := range texts {
			extractor.ExtractFromText(ctx, text)
		}
	}
}

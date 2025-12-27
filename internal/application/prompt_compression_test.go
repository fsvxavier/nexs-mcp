package application

import (
	"context"
	"strings"
	"testing"
)

func TestPromptCompressor_CompressPrompt(t *testing.T) {
	tests := []struct {
		name                string
		config              PromptCompressionConfig
		prompt              string
		expectCompression   bool
		minCompressionRatio float64 // Maximum ratio we expect (lower is better)
		maxCompressionRatio float64 // Minimum ratio we accept
	}{
		{
			name: "disabled compression",
			config: PromptCompressionConfig{
				Enabled:         false,
				MinPromptLength: 100,
			},
			prompt:              "This is a test prompt that should not be compressed",
			expectCompression:   false,
			minCompressionRatio: 1.0,
			maxCompressionRatio: 1.0,
		},
		{
			name: "short prompt below threshold",
			config: PromptCompressionConfig{
				Enabled:         true,
				MinPromptLength: 1000,
			},
			prompt:              strings.Repeat("short ", 10),
			expectCompression:   false,
			minCompressionRatio: 1.0,
			maxCompressionRatio: 1.0,
		},
		{
			name: "verbose prompt with redundancies",
			config: PromptCompressionConfig{
				Enabled:          true,
				RemoveRedundancy: true,
				MinPromptLength:  100,
			},
			prompt: "Please provide me with a detailed explanation regarding how the API works " +
				"in the context of authentication. I would like you to include information " +
				"about the authentication process in order to understand it better.",
			expectCompression:   true,
			minCompressionRatio: 0.8,
			maxCompressionRatio: 0.95,
		},
		{
			name: "prompt with fillers",
			config: PromptCompressionConfig{
				Enabled:         true,
				MinPromptLength: 100,
			},
			prompt: "I basically need you to actually explain how this really works. " +
				"It's literally quite important that you simply clarify this for me. " +
				"You know, I mean, it's obviously very critical.",
			expectCompression:   true,
			minCompressionRatio: 0.4,
			maxCompressionRatio: 0.7,
		},
		{
			name: "prompt with excessive whitespace",
			config: PromptCompressionConfig{
				Enabled:            true,
				CompressWhitespace: true,
				MinPromptLength:    100,
			},
			prompt: "This    has    too    many    spaces.  \n\n\n\n  And  many  newlines.\n\n\n  " +
				"It should    be compressed    significantly.",
			expectCompression:   true,
			minCompressionRatio: 0.5,
			maxCompressionRatio: 0.8,
		},
		{
			name: "technical prompt with aliases",
			config: PromptCompressionConfig{
				Enabled:          true,
				RemoveRedundancy: true,
				UseAliases:       true,
				MinPromptLength:  100,
			},
			prompt: "Could you please help me understand with regard to the API in the context of " +
				"authentication at this point in time. For the purpose of learning, " +
				"I need assistance with this.",
			expectCompression:   true,
			minCompressionRatio: 0.4,
			maxCompressionRatio: 0.7,
		},
		{
			name: "full compression with all techniques",
			config: PromptCompressionConfig{
				Enabled:            true,
				RemoveRedundancy:   true,
				CompressWhitespace: true,
				UseAliases:         true,
				PreserveStructure:  false,
				MinPromptLength:    100,
			},
			prompt: "Please provide me with a detailed explanation regarding how the API works " +
				"in the context of authentication. I would like you to basically include " +
				"information about the authentication process in order to understand it better. " +
				"At this point in time, I really need to understand how the API endpoint " +
				"actually functions in the event that there are errors. " +
				"It would be great if you could help me with this.",
			expectCompression:   true,
			minCompressionRatio: 0.4,
			maxCompressionRatio: 0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressor := NewPromptCompressor(tt.config)
			ctx := context.Background()

			compressed, metadata, err := compressor.CompressPrompt(ctx, tt.prompt)

			if err != nil {
				t.Fatalf("CompressPrompt() error = %v", err)
			}

			if tt.expectCompression {
				if metadata.CompressionRatio < tt.minCompressionRatio || metadata.CompressionRatio > tt.maxCompressionRatio {
					t.Errorf("compression ratio %f not in expected range [%f, %f]",
						metadata.CompressionRatio, tt.minCompressionRatio, tt.maxCompressionRatio)
				}

				if len(compressed) >= len(tt.prompt) {
					t.Errorf("compressed length (%d) should be smaller than original (%d)",
						len(compressed), len(tt.prompt))
				}

				t.Logf("Original: %d chars\nCompressed: %d chars\nRatio: %.2f%%\nSaved: %d chars",
					metadata.OriginalLength, metadata.CompressedLength,
					metadata.CompressionRatio*100,
					metadata.OriginalLength-metadata.CompressedLength)
				t.Logf("Techniques: %v", metadata.TechniquesUsed)
			} else if metadata.CompressionRatio != 1.0 {
				t.Errorf("expected no compression (ratio 1.0), got %f", metadata.CompressionRatio)
			}
		})
	}
}

func TestPromptCompressor_RemoveRedundancies(t *testing.T) {
	config := PromptCompressionConfig{
		Enabled:          true,
		RemoveRedundancy: true,
		MinPromptLength:  100,
	}

	compressor := NewPromptCompressor(config)

	tests := []struct {
		name     string
		input    string
		contains []string // Strings that should NOT be in output
	}{
		{
			name:     "remove repeated words",
			input:    "the the function works works well", //nolint:dupword // intentional duplicate words for testing
			contains: []string{"the the", "works works"},  //nolint:dupword // testing duplicate word removal
		},
		{
			name:     "remove redundant articles",
			input:    "call the API and use the endpoint with the function",
			contains: []string{" the API", " the endpoint", " the function"},
		},
		{
			name:     "remove redundant prepositions",
			input:    "we need to authenticate in order to access the API",
			contains: []string{" in order to "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compressor.removeRedundancies(tt.input)

			for _, substr := range tt.contains {
				if strings.Contains(result, substr) {
					t.Errorf("output should not contain '%s', but got: %s", substr, result)
				}
			}

			t.Logf("Input:  %s", tt.input)
			t.Logf("Output: %s", result)
		})
	}
}

func TestPromptCompressor_RemoveFillers(t *testing.T) {
	config := PromptCompressionConfig{
		Enabled:         true,
		MinPromptLength: 100,
	}

	compressor := NewPromptCompressor(config)

	tests := []struct {
		name     string
		input    string
		contains []string // Filler words that should be removed
	}{
		{
			name:     "remove common fillers",
			input:    "I basically need you to actually explain how this really works",
			contains: []string{" basically ", " actually ", " really "},
		},
		{
			name:     "remove intensifiers",
			input:    "It's literally quite important that you simply clarify this",
			contains: []string{" literally ", " quite ", " simply "},
		},
		{
			name:     "remove hedging words",
			input:    "You probably should maybe consider possibly using this approach",
			contains: []string{" probably ", " maybe ", " possibly "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compressor.removeFillers(tt.input)

			for _, filler := range tt.contains {
				if strings.Contains(result, filler) {
					t.Errorf("output should not contain filler '%s', but got: %s", filler, result)
				}
			}

			// Result should be shorter
			if len(result) >= len(tt.input) {
				t.Errorf("expected shorter output, got %d chars vs %d original", len(result), len(tt.input))
			}

			t.Logf("Input:  %s", tt.input)
			t.Logf("Output: %s", result)
			t.Logf("Saved:  %d chars (%.1f%%)", len(tt.input)-len(result),
				float64(len(tt.input)-len(result))/float64(len(tt.input))*100)
		})
	}
}

func TestPromptCompressor_ApplyAliases(t *testing.T) {
	config := PromptCompressionConfig{
		Enabled:         true,
		UseAliases:      true,
		MinPromptLength: 100,
	}

	compressor := NewPromptCompressor(config)

	tests := []struct {
		name             string
		input            string
		shouldContain    string
		shouldNotContain string
	}{
		{
			name:             "verbose instruction to concise",
			input:            "Please provide me with the details",
			shouldContain:    "Provide:",
			shouldNotContain: "Please provide me with",
		},
		{
			name:             "task instruction",
			input:            "I would like you to explain this",
			shouldContain:    "Task:",
			shouldNotContain: "I would like you to",
		},
		{
			name:             "verbose phrase to concise",
			input:            "in the context of authentication",
			shouldContain:    "for authentication",
			shouldNotContain: "in the context of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compressor.applyAliases(tt.input)

			if !strings.Contains(result, tt.shouldContain) {
				t.Errorf("output should contain '%s', but got: %s", tt.shouldContain, result)
			}

			if strings.Contains(result, tt.shouldNotContain) {
				t.Errorf("output should not contain '%s', but got: %s", tt.shouldNotContain, result)
			}

			t.Logf("Input:  %s", tt.input)
			t.Logf("Output: %s", result)
		})
	}
}

func TestPromptCompressor_CompressWhitespace(t *testing.T) {
	config := PromptCompressionConfig{
		Enabled:            true,
		CompressWhitespace: true,
		PreserveStructure:  false,
		MinPromptLength:    100,
	}

	compressor := NewPromptCompressor(config)

	input := "This    has    too    many    spaces.\n\n\n\nAnd  many  newlines.\n\n\n  " +
		"It should    be compressed."

	result := compressor.compressWhitespace(input)

	// Should not have multiple consecutive spaces
	if strings.Contains(result, "  ") {
		t.Errorf("result should not contain multiple spaces: %s", result)
	}

	// Should not have multiple consecutive newlines (when PreserveStructure=false)
	if strings.Contains(result, "\n\n") {
		t.Errorf("result should not contain multiple newlines: %s", result)
	}

	// Should be shorter
	if len(result) >= len(input) {
		t.Errorf("expected compression, got %d chars vs %d original",
			len(result), len(input))
	}

	t.Logf("Input:  %d chars", len(input))
	t.Logf("Output: %d chars", len(result))
	t.Logf("Saved:  %.1f%%", float64(len(input)-len(result))/float64(len(input))*100)
}

func TestPromptCompressor_Stats(t *testing.T) {
	config := PromptCompressionConfig{
		Enabled:          true,
		RemoveRedundancy: true,
		UseAliases:       true,
		MinPromptLength:  100,
	}

	compressor := NewPromptCompressor(config)
	ctx := context.Background()

	// Process multiple prompts
	prompts := []string{
		"Please provide me with a detailed explanation regarding how the API works in the context of authentication.",
		"I would like you to basically include information about the authentication process in order to understand it better.",
		"Could you please help me understand with regard to the API in the event that there are errors.",
	}

	for _, prompt := range prompts {
		_, _, err := compressor.CompressPrompt(ctx, prompt)
		if err != nil {
			t.Fatalf("CompressPrompt() error = %v", err)
		}
	}

	stats := compressor.GetStats()

	// Some prompts may be below threshold
	if stats.TotalCompressed == 0 {
		t.Error("expected at least some compressed prompts")
	}

	if stats.BytesSaved == 0 {
		t.Error("expected some bytes saved")
	}

	if stats.AvgCompressionRatio >= 1.0 {
		t.Errorf("average compression ratio should be < 1.0, got %f", stats.AvgCompressionRatio)
	}

	t.Logf("Total compressed: %d", stats.TotalCompressed)
	t.Logf("Bytes saved: %d", stats.BytesSaved)
	t.Logf("Avg compression ratio: %.2f%%", stats.AvgCompressionRatio*100)
	t.Logf("Quality score: %.2f", stats.QualityScore)
}

func BenchmarkPromptCompressor_Full(b *testing.B) {
	config := PromptCompressionConfig{
		Enabled:            true,
		RemoveRedundancy:   true,
		CompressWhitespace: true,
		UseAliases:         true,
		PreserveStructure:  false,
		MinPromptLength:    100,
	}

	compressor := NewPromptCompressor(config)
	ctx := context.Background()

	prompt := "Please provide me with a detailed explanation regarding how the API works " +
		"in the context of authentication. I would like you to basically include " +
		"information about the authentication process in order to understand it better. " +
		"At this point in time, I really need to understand how the API endpoint " +
		"actually functions in the event that there are errors. " +
		"It would be great if you could help me with this. " +
		strings.Repeat("Additional context with fillers like basically and really. ", 10)

	b.ResetTimer()
	for range b.N {
		_, _, err := compressor.CompressPrompt(ctx, prompt)
		if err != nil {
			b.Fatalf("CompressPrompt() error = %v", err)
		}
	}
}

func BenchmarkPromptCompressor_NoCompression(b *testing.B) {
	config := PromptCompressionConfig{
		Enabled:         false,
		MinPromptLength: 100,
	}

	compressor := NewPromptCompressor(config)
	ctx := context.Background()

	prompt := strings.Repeat("This is a test prompt without compression. ", 20)

	b.ResetTimer()
	for range b.N {
		_, _, err := compressor.CompressPrompt(ctx, prompt)
		if err != nil {
			b.Fatalf("CompressPrompt() error = %v", err)
		}
	}
}

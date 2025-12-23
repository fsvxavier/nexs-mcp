package quality

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// downloadTestModel downloads production ONNX models for testing
// Downloads both MS MARCO (default) and Paraphrase-Multilingual (configurable)
// Returns path to MS MARCO (default model)
func downloadTestModel(t *testing.T) string {
	t.Helper()

	// Create test models directory
	testModelDir := filepath.Join("testdata", "models")
	if err := os.MkdirAll(testModelDir, 0755); err != nil {
		t.Fatalf("Failed to create test model directory: %v", err)
	}

	// Production models to download
	models := []struct {
		name     string
		url      string
		filename string
		size     string
	}{
		{
			name:     "MS MARCO MiniLM-L-6-v2 (default)",
			url:      "https://huggingface.co/sentence-transformers/ms-marco-MiniLM-L-6-v2/resolve/main/onnx/model.onnx",
			filename: "ms-marco-MiniLM-L-6-v2.onnx",
			size:     "~23MB",
		},
		{
			name:     "Paraphrase-Multilingual-MiniLM-L12-v2 (configurable)",
			url:      "https://huggingface.co/sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2/resolve/main/onnx/model.onnx",
			filename: "paraphrase-multilingual-MiniLM-L12-v2.onnx",
			size:     "~470MB",
		},
	}

	defaultModelPath := ""

	// Download each model
	for i, model := range models {
		modelPath := filepath.Join(testModelDir, model.filename)

		// Check if model already exists
		if _, err := os.Stat(modelPath); err == nil {
			t.Logf("✓ Using cached %s at %s", model.name, modelPath)
			if i == 0 {
				defaultModelPath = modelPath
			}
			continue
		}

		t.Logf("Downloading %s (%s) from HuggingFace...", model.name, model.size)

		// Download model
		resp, err := http.Get(model.url)
		if err != nil {
			t.Logf("⚠ Failed to download %s (network error): %v", model.name, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Logf("⚠ Failed to download %s (HTTP %d)", model.name, resp.StatusCode)
			continue
		}

		// Create file
		out, err := os.Create(modelPath)
		if err != nil {
			t.Logf("⚠ Failed to create model file for %s: %v", model.name, err)
			continue
		}

		// Copy data
		_, err = io.Copy(out, resp.Body)
		out.Close()
		if err != nil {
			os.Remove(modelPath)
			t.Logf("⚠ Failed to save %s: %v", model.name, err)
			continue
		}

		t.Logf("✓ %s downloaded successfully to %s", model.name, modelPath)

		// Save default model path (MS MARCO)
		if i == 0 {
			defaultModelPath = modelPath
		}
	}

	// Return default model path (MS MARCO) or empty if download failed
	if defaultModelPath == "" {
		t.Skipf("Failed to download test models")
	}

	return defaultModelPath
}

// createMockONNXModel creates a minimal mock ONNX model for testing
// This is a fallback if we can't download a real model
func createMockONNXModel(t *testing.T) string {
	t.Helper()

	// Create testdata directory
	testModelDir := filepath.Join("testdata", "models")
	if err := os.MkdirAll(testModelDir, 0755); err != nil {
		t.Fatalf("Failed to create test model directory: %v", err)
	}

	mockPath := filepath.Join(testModelDir, "mock.onnx")

	// Minimal ONNX model structure (protobuf format)
	// This is a valid but minimal ONNX file that won't work for inference
	// but will test model loading logic
	mockData := []byte{
		0x08, 0x07, // IR version 7
		0x12, 0x0a, 0x62, 0x61, 0x63, 0x6b, 0x65, 0x6e, 0x64, 0x2d, 0x74, 0x65, 0x73, 0x74,
	}

	if err := os.WriteFile(mockPath, mockData, 0644); err != nil {
		t.Skipf("Failed to create mock model: %v", err)
		return ""
	}

	return mockPath
}

// ensureTestModel ensures a test model is available, downloading if necessary
// Prioritizes production models: MS MARCO (default) and Paraphrase-Multilingual (configurable)
func ensureTestModel(t *testing.T) string {
	t.Helper()

	// Try to use production MS MARCO model first (default model)
	prodModelMSMarco := "../../models/ms-marco-MiniLM-L-6-v2/model.onnx"
	if _, err := os.Stat(prodModelMSMarco); err == nil {
		t.Logf("Using production MS MARCO model (default): %s", prodModelMSMarco)
		return prodModelMSMarco
	}

	// Try Paraphrase-Multilingual as fallback (configurable model)
	prodModelParaphrase := "../../models/paraphrase-multilingual-MiniLM-L12-v2/model.onnx"
	if _, err := os.Stat(prodModelParaphrase); err == nil {
		t.Logf("Using production Paraphrase-Multilingual model (configurable): %s", prodModelParaphrase)
		return prodModelParaphrase
	}

	// Check if test model already exists
	testModelDir := filepath.Join("testdata", "models")
	testModel := filepath.Join(testModelDir, "ms-marco-MiniLM-L-6-v2.onnx")
	if _, err := os.Stat(testModel); err == nil {
		t.Logf("Using cached test model: %s", testModel)
		return testModel
	}

	// Skip if we can't download (network not available or in CI)
	if os.Getenv("CI") != "" || os.Getenv("SKIP_DOWNLOAD") != "" {
		t.Skip("Test model not available and download is disabled (CI or SKIP_DOWNLOAD set)")
		return ""
	}

	// Try to download test model
	t.Logf("Attempting to download test model (this may take a moment)...")
	if modelPath := downloadTestModel(t); modelPath != "" {
		return modelPath
	}

	// If download fails, skip test
	t.Skip("No ONNX model available for testing (download failed and production models not found)")
	return ""
}

// getTestModelConfig returns a config with a test model path
// Defaults to MS MARCO (production default), can be configured for Paraphrase-Multilingual
func getTestModelConfig(t *testing.T) *Config {
	modelPath := ensureTestModel(t)

	// Determine model type based on path
	config := &Config{
		ONNXModelPath: modelPath,
	}

	// Configure based on which production model is being used
	if strings.Contains(modelPath, "ms-marco") {
		// MS MARCO MiniLM-L-6-v2 (default production model)
		config.RequiresTokenTypeIds = true
		config.ONNXModelType = "reranker"
		config.ONNXOutputName = "logits"
		config.ONNXOutputShape = []int64{1, 1}
		t.Logf("Configured for MS MARCO (default production model)")
	} else if strings.Contains(modelPath, "paraphrase-multilingual") {
		// Paraphrase-Multilingual MiniLM-L12-v2 (configurable production model)
		config.RequiresTokenTypeIds = true
		config.ONNXModelType = "embedder"
		config.ONNXOutputName = "last_hidden_state"
		config.ONNXOutputShape = []int64{1, 512, 384}
		t.Logf("Configured for Paraphrase-Multilingual (configurable production model)")
	}

	return config
}

// cleanupTestModels removes downloaded test models (optional, for cleanup)
func cleanupTestModels(t *testing.T) {
	testModelDir := filepath.Join("testdata", "models")
	if err := os.RemoveAll(testModelDir); err != nil {
		t.Logf("Warning: Failed to cleanup test models: %v", err)
	}
}

// isONNXRuntimeAvailable checks if ONNX Runtime is installed
func isONNXRuntimeAvailable() bool {
	// Try to load ONNX runtime library
	// This is a simple check - actual initialization happens in NewONNXScorer
	return true // Will fail gracefully in NewONNXScorer if not available
}

// skipIfONNXNotAvailable skips test if ONNX runtime is not available
func skipIfONNXNotAvailable(t *testing.T) {
	if !isONNXRuntimeAvailable() {
		t.Skip("ONNX Runtime not available (CGO or library not installed)")
	}
}

// assertValidScore checks if a score is valid
func assertValidScore(t *testing.T, score *Score, methodName string) {
	t.Helper()

	if score == nil {
		t.Fatal("Score is nil")
	}

	if score.Value < 0 || score.Value > 1 {
		t.Errorf("Score value %f out of valid range [0, 1]", score.Value)
	}

	if score.Confidence < 0 || score.Confidence > 1 {
		t.Errorf("Confidence %f out of valid range [0, 1]", score.Confidence)
	}

	if score.Method != methodName {
		t.Errorf("Expected method '%s', got '%s'", methodName, score.Method)
	}

	if score.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

// testWithTimeout runs a test function with timeout
func testWithTimeout(t *testing.T, name string, timeoutSec int, fn func(*testing.T)) {
	t.Helper()

	done := make(chan bool)
	go func() {
		fn(t)
		done <- true
	}()

	select {
	case <-done:
		// Test completed
	case <-time.After(time.Duration(timeoutSec) * time.Second):
		t.Fatalf("Test %s timed out after %d seconds", name, timeoutSec)
	}
}

// skipIfNetworkUnavailable skips test if network is not available
func skipIfNetworkUnavailable(t *testing.T) {
	t.Helper()

	resp, err := http.Head("https://www.google.com")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Skip("Network not available for downloading test models")
	}
}

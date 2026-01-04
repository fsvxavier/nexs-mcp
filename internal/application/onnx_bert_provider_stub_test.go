//go:build noonnx
// +build noonnx

package application

import (
	"testing"
)

func TestONNXBERTProvider_StubBehavior(t *testing.T) {
	config := EnhancedNLPConfig{
		EntityModel:    "test.onnx",
		SentimentModel: "test.onnx",
	}

	provider, err := NewONNXBERTProvider(config)
	if err != nil {
		t.Fatalf("NewONNXBERTProvider() error = %v", err)
	}

	if provider == nil {
		t.Fatal("Provider is nil")
	}

	// Stub should always report unavailable
	if provider.IsAvailable() {
		t.Error("Stub provider should not be available")
	}
}

func TestONNXBERTProvider_StubClose(t *testing.T) {
	provider, _ := NewONNXBERTProvider(EnhancedNLPConfig{})

	err := provider.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

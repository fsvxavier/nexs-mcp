//go:build !noonnx
// +build !noonnx

package main

import (
	ort "github.com/yalue/onnxruntime_go"
)

// isONNXAvailable checks if ONNX Runtime is available.
func isONNXAvailable() bool {
	// Try to initialize ONNX Runtime
	if err := ort.InitializeEnvironment(); err != nil {
		return false
	}
	_ = ort.DestroyEnvironment() // Ignore error on cleanup check
	return true
}

// getONNXStatus returns a status message about ONNX availability.
func getONNXStatus() string {
	if isONNXAvailable() {
		return "enabled (ONNX Runtime loaded successfully)"
	}
	return "disabled (ONNX Runtime not available)"
}

//go:build noonnx
// +build noonnx

package main

// isONNXAvailable always returns false when built without ONNX
func isONNXAvailable() bool {
	return false
}

// getONNXStatus returns a status message about ONNX availability
func getONNXStatus() string {
	return "disabled (built without ONNX support)"
}

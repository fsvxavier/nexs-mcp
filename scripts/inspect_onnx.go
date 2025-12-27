//go:build !noonnx
// +build !noonnx

package main

import (
	"fmt"

	ort "github.com/yalue/onnxruntime_go"
)

func main() {
	if err := ort.InitializeEnvironment(); err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer func() { _ = ort.DestroyEnvironment() }()

	models := []string{
		"../models/distiluse-base-multilingual-cased-v1/model.onnx",
		"../models/distiluse-base-multilingual-cased-v2/model.onnx",
		"../models/paraphrase-multilingual-MiniLM-L12-v2/model.onnx",
		"../models/ms-marco-MiniLM-L-6-v2/model.onnx",
	}

	for _, modelPath := range models {
		fmt.Printf("\n=== %s ===\n", modelPath)

		// Get input/output info
		inputs, outputs, err := ort.GetInputOutputInfo(modelPath)
		if err != nil {
			fmt.Printf("Error getting input/output info: %v\n", err)
			continue
		}

		fmt.Println("\nInputs:")
		for i, info := range inputs {
			fmt.Printf("  %d: Name='%s', Type=%v, Dims=%v\n", i, info.Name, info.DataType, info.Dimensions)
		}

		fmt.Println("\nOutputs:")
		for i, info := range outputs {
			fmt.Printf("  %d: Name='%s', Type=%v, Dims=%v\n", i, info.Name, info.DataType, info.Dimensions)
		}
	}
}

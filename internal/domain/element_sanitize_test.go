package domain

import (
	"testing"
)

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name",
			input:    "Simple Name",
			expected: "simple_name",
		},
		{
			name:     "name with forward slash",
			input:    "CI/CD",
			expected: "ci_cd",
		},
		{
			name:     "name with backslash",
			input:    "Path\\To\\File",
			expected: "path_to_file",
		},
		{
			name:     "name with colon",
			input:    "Time: 12:30",
			expected: "time_12_30",
		},
		{
			name:     "name with asterisk",
			input:    "Test * Wildcard",
			expected: "test_wildcard",
		},
		{
			name:     "name with question mark",
			input:    "What? How?",
			expected: "what_how",
		},
		{
			name:     "name with quotes",
			input:    "\"Quoted\"",
			expected: "quoted",
		},
		{
			name:     "name with angle brackets",
			input:    "Name <Tag>",
			expected: "name_tag",
		},
		{
			name:     "name with pipe",
			input:    "Option | Choice",
			expected: "option_choice",
		},
		{
			name:     "name with parentheses",
			input:    "Cloud Architecture (AWS/GCP/Azure)",
			expected: "cloud_architecture_aws_gcp_azure",
		},
		{
			name:     "name with accents",
			input:    "Configuração",
			expected: "configuracao",
		},
		{
			name:     "name with multiple spaces",
			input:    "Multiple    Spaces",
			expected: "multiple_spaces",
		},
		{
			name:     "name with leading/trailing underscores",
			input:    "_Test_",
			expected: "test",
		},
		{
			name:     "complex name",
			input:    "API Design (REST/gRPC)",
			expected: "api_design_rest_grpc",
		},
		{
			name:     "name with special chars",
			input:    "Test@#$%^&*()Name",
			expected: "test_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeName(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeName(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateElementIDSanitization(t *testing.T) {
	tests := []struct {
		name        string
		elementType ElementType
		inputName   string
		wantPrefix  string
	}{
		{
			name:        "skill with forward slash",
			elementType: SkillElement,
			inputName:   "CI/CD",
			wantPrefix:  "skill_ci_cd_",
		},
		{
			name:        "skill with parentheses and slash",
			elementType: SkillElement,
			inputName:   "Cloud Architecture (AWS/GCP/Azure)",
			wantPrefix:  "skill_cloud_architecture_aws_gcp_azure_",
		},
		{
			name:        "persona with special chars",
			elementType: PersonaElement,
			inputName:   "Senior Engineer - Go/Rust",
			wantPrefix:  "persona_senior_engineer_go_rust_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateElementID(tt.elementType, tt.inputName)

			// Check if result starts with expected prefix
			if len(result) < len(tt.wantPrefix) {
				t.Errorf("GenerateElementID() result too short: %q", result)
				return
			}

			prefix := result[:len(tt.wantPrefix)]
			if prefix != tt.wantPrefix {
				t.Errorf("GenerateElementID() prefix = %q, want %q", prefix, tt.wantPrefix)
			}

			// Verify no filesystem-problematic characters in result
			problematicChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
			for _, char := range problematicChars {
				if contains := contains(result, char); contains {
					t.Errorf("GenerateElementID() contains problematic char %q in result: %q", char, result)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	for i := 0; i < len(s); i++ {
		if i+len(substr) <= len(s) && s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

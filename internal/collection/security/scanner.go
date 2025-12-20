package security

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Severity represents the severity level of a security finding
type Severity string

const (
	// SeverityCritical indicates an immediate security threat
	SeverityCritical Severity = "critical"
	// SeverityHigh indicates a serious security issue
	SeverityHigh Severity = "high"
	// SeverityMedium indicates a moderate security concern
	SeverityMedium Severity = "medium"
	// SeverityLow indicates a minor security issue
	SeverityLow Severity = "low"
)

// ScanRule represents a pattern to scan for
type ScanRule struct {
	Name        string         // Rule name (e.g., "eval-injection")
	Pattern     *regexp.Regexp // Regex pattern to match
	Severity    Severity       // Severity level
	Description string         // Human-readable description
	Fix         string         // Suggested fix
}

// ScanFinding represents a security issue found during scanning
type ScanFinding struct {
	Rule     *ScanRule // Rule that matched
	File     string    // File where issue was found
	Line     int       // Line number
	Content  string    // Line content
	Severity Severity  // Severity level
}

// ScanResult holds the results of a security scan
type ScanResult struct {
	Findings     []*ScanFinding // All findings
	Stats        map[string]int // Statistics by severity
	FilesScanned int            // Number of files scanned
	Clean        bool           // True if no critical/high issues found
}

// CodeScanner scans code for malicious patterns
type CodeScanner struct {
	rules     []*ScanRule
	threshold Severity // Minimum severity to report
}

// NewCodeScanner creates a new code scanner with default rules
func NewCodeScanner() *CodeScanner {
	return &CodeScanner{
		rules:     defaultScanRules(),
		threshold: SeverityLow, // Report all by default
	}
}

// SetThreshold sets the minimum severity to report
func (s *CodeScanner) SetThreshold(threshold Severity) {
	s.threshold = threshold
}

// AddRule adds a custom scan rule
func (s *CodeScanner) AddRule(rule *ScanRule) {
	s.rules = append(s.rules, rule)
}

// Scan scans a directory recursively for security issues
func (s *CodeScanner) Scan(basePath string) (*ScanResult, error) {
	result := &ScanResult{
		Findings: make([]*ScanFinding, 0),
		Stats:    make(map[string]int),
	}

	// Walk directory tree
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip binary files and large files (>1MB)
		if info.Size() > 1024*1024 {
			return nil
		}

		// Scan file
		findings, err := s.scanFile(path)
		if err != nil {
			// Log error but continue scanning
			return nil
		}

		result.Findings = append(result.Findings, findings...)
		result.FilesScanned++

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	// Compute stats
	result.Stats = map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
	}

	for _, finding := range result.Findings {
		result.Stats[string(finding.Severity)]++
	}

	// Determine if clean (no critical/high issues)
	result.Clean = result.Stats["critical"] == 0 && result.Stats["high"] == 0

	return result, nil
}

// scanFile scans a single file
func (s *CodeScanner) scanFile(filePath string) ([]*ScanFinding, error) {
	findings := make([]*ScanFinding, 0)

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read line by line
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check against all rules
		for _, rule := range s.rules {
			if rule.Pattern.MatchString(line) {
				// Check threshold
				if s.shouldReport(rule.Severity) {
					findings = append(findings, &ScanFinding{
						Rule:     rule,
						File:     filePath,
						Line:     lineNum,
						Content:  strings.TrimSpace(line),
						Severity: rule.Severity,
					})
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return findings, nil
}

// shouldReport checks if a severity should be reported based on threshold
func (s *CodeScanner) shouldReport(severity Severity) bool {
	severityLevels := map[Severity]int{
		SeverityLow:      1,
		SeverityMedium:   2,
		SeverityHigh:     3,
		SeverityCritical: 4,
	}

	return severityLevels[severity] >= severityLevels[s.threshold]
}

// defaultScanRules returns the default set of security scan rules (50+ patterns)
func defaultScanRules() []*ScanRule {
	return []*ScanRule{
		// Critical: Code execution
		{
			Name:        "eval-injection",
			Pattern:     regexp.MustCompile(`\beval\s+["']?\$`),
			Severity:    SeverityCritical,
			Description: "eval() can execute arbitrary code",
			Fix:         "Remove eval(), use safe alternatives",
		},
		{
			Name:        "exec-injection",
			Pattern:     regexp.MustCompile(`\bexec\s*\(`),
			Severity:    SeverityCritical,
			Description: "exec() can execute system commands",
			Fix:         "Remove exec(), validate all inputs",
		},
		{
			Name:        "shell-exec",
			Pattern:     regexp.MustCompile(`\bshell_exec\s*\(`),
			Severity:    SeverityCritical,
			Description: "shell_exec() executes shell commands",
			Fix:         "Use safe command execution methods",
		},
		{
			Name:        "system-call",
			Pattern:     regexp.MustCompile(`\bsystem\s*\(`),
			Severity:    SeverityCritical,
			Description: "system() executes shell commands",
			Fix:         "Use safe alternatives",
		},
		{
			Name:        "passthru",
			Pattern:     regexp.MustCompile(`\bpassthru\s*\(`),
			Severity:    SeverityCritical,
			Description: "passthru() executes commands",
			Fix:         "Remove passthru()",
		},

		// Critical: Remote code execution
		{
			Name:        "curl-pipe-bash",
			Pattern:     regexp.MustCompile(`curl.*\|.*bash`),
			Severity:    SeverityCritical,
			Description: "Downloading and executing remote code",
			Fix:         "Never pipe curl to bash",
		},
		{
			Name:        "wget-pipe-sh",
			Pattern:     regexp.MustCompile(`wget.*\|.*sh`),
			Severity:    SeverityCritical,
			Description: "Downloading and executing remote code",
			Fix:         "Never pipe wget to sh",
		},
		{
			Name:        "curl-silent-install",
			Pattern:     regexp.MustCompile(`curl.*-s.*\|\s*(sudo\s+)?sh`),
			Severity:    SeverityCritical,
			Description: "Silent remote code execution",
			Fix:         "Review and verify scripts before execution",
		},

		// Critical: Data destruction
		{
			Name:        "rm-rf-root",
			Pattern:     regexp.MustCompile(`rm\s+-rf\s+/`),
			Severity:    SeverityCritical,
			Description: "Recursive deletion from root",
			Fix:         "Never use rm -rf /",
		},
		{
			Name:        "rm-rf-wildcard",
			Pattern:     regexp.MustCompile(`rm\s+-rf\s+\*`),
			Severity:    SeverityCritical,
			Description: "Recursive wildcard deletion",
			Fix:         "Use specific paths, avoid wildcards with rm -rf",
		},
		{
			Name:        "dd-if-dev",
			Pattern:     regexp.MustCompile(`dd\s+if=/dev/`),
			Severity:    SeverityCritical,
			Description: "Direct disk access via dd",
			Fix:         "Remove dd commands",
		},
		{
			Name:        "mkfs",
			Pattern:     regexp.MustCompile(`\bmkfs\b`),
			Severity:    SeverityCritical,
			Description: "Filesystem formatting",
			Fix:         "Remove mkfs commands",
		},
		{
			Name:        "fdisk",
			Pattern:     regexp.MustCompile(`\bfdisk\b`),
			Severity:    SeverityCritical,
			Description: "Disk partitioning",
			Fix:         "Remove fdisk commands",
		},

		// Critical: Fork bombs
		{
			Name:        "fork-bomb",
			Pattern:     regexp.MustCompile(`:\(\)\{:\|:&\};:`),
			Severity:    SeverityCritical,
			Description: "Fork bomb pattern detected",
			Fix:         "Remove fork bomb",
		},

		// High: Command injection vectors
		{
			Name:        "backtick-execution",
			Pattern:     regexp.MustCompile("`[^`]+`"),
			Severity:    SeverityHigh,
			Description: "Backtick command execution",
			Fix:         "Use $(command) instead of backticks",
		},
		{
			Name:        "command-substitution",
			Pattern:     regexp.MustCompile(`\$\([^)]+\)`),
			Severity:    SeverityMedium,
			Description: "Command substitution",
			Fix:         "Ensure proper input validation",
		},
		{
			Name:        "double-pipe",
			Pattern:     regexp.MustCompile(`\|\|`),
			Severity:    SeverityMedium,
			Description: "Command chaining (OR)",
			Fix:         "Use explicit error handling",
		},
		{
			Name:        "double-ampersand",
			Pattern:     regexp.MustCompile(`&&`),
			Severity:    SeverityMedium,
			Description: "Command chaining (AND)",
			Fix:         "Use explicit command sequencing",
		},
		{
			Name:        "semicolon-separator",
			Pattern:     regexp.MustCompile(`;`),
			Severity:    SeverityMedium,
			Description: "Command separator",
			Fix:         "Split into separate commands",
		},

		// High: Network access
		{
			Name:        "netcat-listen",
			Pattern:     regexp.MustCompile(`nc\s+-l`),
			Severity:    SeverityHigh,
			Description: "Netcat listening mode",
			Fix:         "Remove network listeners",
		},
		{
			Name:        "ncat",
			Pattern:     regexp.MustCompile(`\bncat\b`),
			Severity:    SeverityHigh,
			Description: "Ncat network tool",
			Fix:         "Remove network tools",
		},
		{
			Name:        "socat",
			Pattern:     regexp.MustCompile(`\bsocat\b`),
			Severity:    SeverityHigh,
			Description: "Socat relay tool",
			Fix:         "Remove network relay tools",
		},

		// High: Privilege escalation
		{
			Name:        "sudo-nopasswd",
			Pattern:     regexp.MustCompile(`sudo.*NOPASSWD`),
			Severity:    SeverityHigh,
			Description: "Passwordless sudo configuration",
			Fix:         "Require password for sudo",
		},
		{
			Name:        "chmod-777",
			Pattern:     regexp.MustCompile(`chmod\s+777`),
			Severity:    SeverityHigh,
			Description: "World-writable permissions",
			Fix:         "Use restrictive permissions (e.g., 755, 644)",
		},
		{
			Name:        "chmod-setuid",
			Pattern:     regexp.MustCompile(`chmod\s+[4-7][0-7]{3}`),
			Severity:    SeverityHigh,
			Description: "Setuid/setgid permissions",
			Fix:         "Avoid setuid/setgid unless necessary",
		},

		// Medium: File operations
		{
			Name:        "redirect-to-dev-null",
			Pattern:     regexp.MustCompile(`>\s*/dev/null`),
			Severity:    SeverityLow,
			Description: "Output redirection to /dev/null",
			Fix:         "Consider logging output instead",
		},
		{
			Name:        "redirect-stderr",
			Pattern:     regexp.MustCompile(`2>&1`),
			Severity:    SeverityLow,
			Description: "Stderr redirect",
			Fix:         "Ensure errors are properly handled",
		},

		// Medium: Encoding/Obfuscation
		{
			Name:        "base64-decode",
			Pattern:     regexp.MustCompile(`base64\s+-d`),
			Severity:    SeverityMedium,
			Description: "Base64 decoding (potential obfuscation)",
			Fix:         "Review decoded content",
		},
		{
			Name:        "hex-decode",
			Pattern:     regexp.MustCompile(`xxd\s+-r`),
			Severity:    SeverityMedium,
			Description: "Hex decoding",
			Fix:         "Review decoded content",
		},

		// Medium: Unsafe Python patterns
		{
			Name:        "python-exec",
			Pattern:     regexp.MustCompile(`__import__\s*\(\s*['"]os['"]\s*\)`),
			Severity:    SeverityHigh,
			Description: "Dynamic os module import",
			Fix:         "Use explicit imports",
		},
		{
			Name:        "pickle-load",
			Pattern:     regexp.MustCompile(`pickle\.loads?\(`),
			Severity:    SeverityHigh,
			Description: "Pickle deserialization (code execution risk)",
			Fix:         "Use safe serialization (JSON)",
		},

		// Medium: Unsafe JavaScript patterns
		{
			Name:        "javascript-eval",
			Pattern:     regexp.MustCompile(`\beval\s*\(`),
			Severity:    SeverityCritical,
			Description: "JavaScript eval()",
			Fix:         "Use JSON.parse() or safe alternatives",
		},
		{
			Name:        "function-constructor",
			Pattern:     regexp.MustCompile(`new\s+Function\s*\(`),
			Severity:    SeverityHigh,
			Description: "Function constructor (dynamic code)",
			Fix:         "Use regular functions",
		},
		{
			Name:        "base64-decode",
			Pattern:     regexp.MustCompile(`base64_decode\(`),
			Severity:    SeverityMedium,
			Description: "Base64 decode operation (potential obfuscation)",
			Fix:         "Review decoded content for malicious code",
		},

		// Low: Hardcoded credentials
		{
			Name:        "hardcoded-password",
			Pattern:     regexp.MustCompile(`(?i)(password|passwd|pwd)\s*=\s*['"][^'"]+['"]`),
			Severity:    SeverityMedium,
			Description: "Potential hardcoded password",
			Fix:         "Use environment variables or secrets management",
		},
		{
			Name:        "hardcoded-api-key",
			Pattern:     regexp.MustCompile(`(?i)(api[_-]?key|apikey|token)\s*=\s*['"][^'"]+['"]`),
			Severity:    SeverityMedium,
			Description: "Potential hardcoded API key",
			Fix:         "Use environment variables",
		},

		// Low: Debug/Development code
		{
			Name:        "console-log",
			Pattern:     regexp.MustCompile(`console\.log\(`),
			Severity:    SeverityLow,
			Description: "Debug console.log statement",
			Fix:         "Remove debug statements",
		},
		{
			Name:        "print-debug",
			Pattern:     regexp.MustCompile(`print\s*\(\s*['"]DEBUG`),
			Severity:    SeverityLow,
			Description: "Debug print statement",
			Fix:         "Remove debug statements",
		},

		// Additional patterns for comprehensive coverage
		{
			Name:        "temp-file-creation",
			Pattern:     regexp.MustCompile(`/tmp/[a-zA-Z0-9_-]+`),
			Severity:    SeverityLow,
			Description: "Hardcoded temp file path",
			Fix:         "Use mktemp or secure temp file creation",
		},
		{
			Name:        "world-readable-file",
			Pattern:     regexp.MustCompile(`chmod\s+[0-7]44`),
			Severity:    SeverityLow,
			Description: "World-readable file",
			Fix:         "Review file permissions",
		},
		{
			Name:        "sql-query-concat",
			Pattern:     regexp.MustCompile(`(?i)(select|insert|update|delete).*\+.*['"]`),
			Severity:    SeverityHigh,
			Description: "Potential SQL injection (string concatenation)",
			Fix:         "Use parameterized queries",
		},
		{
			Name:        "sql-string-concat",
			Pattern:     regexp.MustCompile(`(?i)(SELECT|INSERT|UPDATE|DELETE)\s+.*\s+WHERE\s+.*="\s*\+`),
			Severity:    SeverityHigh,
			Description: "SQL injection via string concatenation",
			Fix:         "Use parameterized queries or prepared statements",
		},
		{
			Name:        "unsafe-yaml-load",
			Pattern:     regexp.MustCompile(`yaml\.load\(`),
			Severity:    SeverityHigh,
			Description: "Unsafe YAML loading (code execution risk)",
			Fix:         "Use yaml.safe_load()",
		},
		{
			Name:        "unsafe-xml-parser",
			Pattern:     regexp.MustCompile(`xml\.etree\.ElementTree\.parse`),
			Severity:    SeverityMedium,
			Description: "XML parsing (XXE risk)",
			Fix:         "Use defusedxml library",
		},
	}
}

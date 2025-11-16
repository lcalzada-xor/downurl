package jsanalyzer

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"
)

// Beautifier beautifies minified JavaScript
type Beautifier struct {
	indentSize int
	indentChar string
}

// NewBeautifier creates a new JavaScript beautifier
func NewBeautifier() *Beautifier {
	return &Beautifier{
		indentSize: 2,
		indentChar: " ",
	}
}

// Beautify beautifies minified JavaScript code
func (b *Beautifier) Beautify(code string) string {
	var result strings.Builder
	indentLevel := 0
	inString := false
	inComment := false
	inRegex := false
	stringChar := rune(0)
	prevChar := rune(0)

	indent := func() string {
		return strings.Repeat(b.indentChar, indentLevel*b.indentSize)
	}

	for i, char := range code {
		// Handle string literals
		if !inComment && !inRegex {
			if (char == '"' || char == '\'' || char == '`') && prevChar != '\\' {
				if !inString {
					inString = true
					stringChar = char
				} else if char == stringChar {
					inString = false
					stringChar = 0
				}
			}
		}

		// Handle comments
		if !inString && !inRegex {
			if i < len(code)-1 {
				nextChar := rune(code[i+1])
				if char == '/' && nextChar == '/' {
					inComment = true
				}
			}
		}

		// End of line comment
		if inComment && char == '\n' {
			inComment = false
		}

		// Skip processing if in string, comment, or regex
		if inString || inComment {
			result.WriteRune(char)
			prevChar = char
			continue
		}

		// Handle different characters
		switch char {
		case '{':
			result.WriteRune(char)
			indentLevel++
			result.WriteString("\n")
			result.WriteString(indent())

		case '}':
			indentLevel--
			if indentLevel < 0 {
				indentLevel = 0
			}
			// Remove trailing whitespace before }
			resultStr := result.String()
			resultStr = strings.TrimRight(resultStr, " \t")
			result.Reset()
			result.WriteString(resultStr)
			result.WriteString("\n")
			result.WriteString(indent())
			result.WriteRune(char)

		case ';':
			result.WriteRune(char)
			// Add newline after semicolon (if not in for loop)
			if prevChar != ')' || !b.isInForLoop(code, i) {
				result.WriteString("\n")
				result.WriteString(indent())
			}

		case ',':
			result.WriteRune(char)
			result.WriteRune(' ')

		case ':':
			result.WriteRune(char)
			result.WriteRune(' ')

		case '\n', '\r':
			// Skip original newlines
			continue

		case ' ', '\t':
			// Normalize whitespace
			if prevChar != ' ' && prevChar != '\t' {
				result.WriteRune(' ')
			}

		default:
			result.WriteRune(char)
		}

		prevChar = char
	}

	return result.String()
}

// isInForLoop checks if position is inside a for loop declaration
func (b *Beautifier) isInForLoop(code string, pos int) bool {
	// Simple heuristic: look back for 'for' keyword
	lookBack := 50
	start := pos - lookBack
	if start < 0 {
		start = 0
	}

	snippet := code[start:pos]
	return strings.Contains(snippet, "for")
}

// IsMinified checks if JavaScript code appears to be minified
func IsMinified(code string) bool {
	if len(code) == 0 {
		return false
	}

	// Read first 1000 characters to analyze
	sample := code
	if len(code) > 1000 {
		sample = code[:1000]
	}

	// Count newlines
	newlineCount := strings.Count(sample, "\n")

	// Count average line length
	lines := strings.Split(sample, "\n")
	if len(lines) == 0 {
		return false
	}

	totalLength := 0
	for _, line := range lines {
		totalLength += len(line)
	}
	avgLineLength := totalLength / len(lines)

	// Heuristics:
	// - Very few newlines (< 5 in 1000 chars)
	// - Very long average line length (> 200)
	if newlineCount < 5 || avgLineLength > 200 {
		return true
	}

	return false
}

// StringExtractor extracts strings from JavaScript code
type StringExtractor struct {
	minLength int
	pattern   *regexp.Regexp
}

// NewStringExtractor creates a new string extractor
func NewStringExtractor(minLength int, pattern string) *StringExtractor {
	var regex *regexp.Regexp
	if pattern != "" {
		regex = regexp.MustCompile("(?i)" + pattern)
	}

	return &StringExtractor{
		minLength: minLength,
		pattern:   regex,
	}
}

// Extract extracts strings from JavaScript code
func (s *StringExtractor) Extract(code string) []string {
	var strings []string
	seen := make(map[string]bool)

	// Extract strings from different quote types
	patterns := []string{
		`"([^"\\]*(\\.[^"\\]*)*)"`,  // Double quotes
		`'([^'\\]*(\\.[^'\\]*)*)'`,  // Single quotes
		"`([^`\\\\]*(\\\\.[^`\\\\]*)*)`", // Template literals
	}

	for _, pattern := range patterns {
		regex := regexp.MustCompile(pattern)
		matches := regex.FindAllStringSubmatch(code, -1)

		for _, match := range matches {
			if len(match) < 2 {
				continue
			}

			str := match[1]

			// Filter by length
			if len(str) < s.minLength {
				continue
			}

			// Filter by pattern if specified
			if s.pattern != nil && !s.pattern.MatchString(str) {
				continue
			}

			// Avoid duplicates
			if seen[str] {
				continue
			}
			seen[str] = true

			strings = append(strings, str)
		}
	}

	return strings
}

// ExtractFromFile extracts strings from a JavaScript file
func (s *StringExtractor) ExtractFromFile(filepath string) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var content strings.Builder
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content.WriteString(scanner.Text())
		content.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return s.Extract(content.String()), nil
}

// DetectObfuscation detects if JavaScript code is obfuscated
func DetectObfuscation(code string) bool {
	// Analyze first 2000 characters
	sample := code
	if len(code) > 2000 {
		sample = code[:2000]
	}

	// Check for common obfuscation patterns
	obfuscationPatterns := []string{
		`eval\s*\(`,                    // eval usage
		`Function\s*\(`,                // Function constructor
		`fromCharCode`,                 // String encoding
		`\\x[0-9a-fA-F]{2}`,           // Hex encoding
		`\\u[0-9a-fA-F]{4}`,           // Unicode encoding
		`atob\s*\(`,                   // Base64 decode
		`_0x[a-fA-F0-9]+`,             // Common obfuscator variable pattern
	}

	count := 0
	for _, pattern := range obfuscationPatterns {
		regex := regexp.MustCompile(pattern)
		if regex.MatchString(sample) {
			count++
		}
	}

	// If multiple obfuscation patterns found, likely obfuscated
	return count >= 2
}

// CalculateComplexity calculates code complexity score
func CalculateComplexity(code string) int {
	complexity := 0

	// Count control flow structures
	patterns := map[string]int{
		`\bif\b`:       1,
		`\belse\b`:     1,
		`\bfor\b`:      2,
		`\bwhile\b`:    2,
		`\bswitch\b`:   2,
		`\bcase\b`:     1,
		`\bcatch\b`:    1,
		`\bfunction\b`: 1,
		`=>`:           1, // Arrow functions
	}

	for pattern, weight := range patterns {
		regex := regexp.MustCompile(pattern)
		matches := regex.FindAllString(code, -1)
		complexity += len(matches) * weight
	}

	return complexity
}

// CountLines counts non-empty lines of code
func CountLines(code string) int {
	lines := strings.Split(code, "\n")
	count := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip empty lines and comment-only lines
		if trimmed != "" && !strings.HasPrefix(trimmed, "//") {
			count++
		}
	}

	return count
}

// ExtractFunctions extracts function names from JavaScript code
func ExtractFunctions(code string) []string {
	var functions []string
	seen := make(map[string]bool)

	// Pattern for function declarations
	patterns := []string{
		`function\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\(`,                   // function name()
		`([a-zA-Z_$][a-zA-Z0-9_$]*)\s*:\s*function\s*\(`,              // name: function()
		`([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*function\s*\(`,              // name = function()
		`([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*\([^)]*\)\s*=>`,            // arrow functions
		`([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\([^)]*\)\s*\{`,                // ES6 method shorthand
	}

	for _, pattern := range patterns {
		regex := regexp.MustCompile(pattern)
		matches := regex.FindAllStringSubmatch(code, -1)

		for _, match := range matches {
			if len(match) < 2 {
				continue
			}

			funcName := match[1]

			// Skip common keywords
			keywords := map[string]bool{
				"if": true, "else": true, "for": true, "while": true,
				"switch": true, "case": true, "return": true,
			}

			if keywords[funcName] {
				continue
			}

			// Avoid duplicates
			if seen[funcName] {
				continue
			}
			seen[funcName] = true

			functions = append(functions, funcName)
		}
	}

	return functions
}

// ExtractVariables extracts variable names from JavaScript code
func ExtractVariables(code string) []string {
	var variables []string
	seen := make(map[string]bool)

	// Patterns for variable declarations
	patterns := []string{
		`\b(?:var|let|const)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)`,
	}

	for _, pattern := range patterns {
		regex := regexp.MustCompile(pattern)
		matches := regex.FindAllStringSubmatch(code, -1)

		for _, match := range matches {
			if len(match) < 2 {
				continue
			}

			varName := match[1]

			// Avoid duplicates
			if seen[varName] {
				continue
			}
			seen[varName] = true

			variables = append(variables, varName)
		}
	}

	return variables
}

// IsWhitespace checks if a rune is whitespace
func IsWhitespace(r rune) bool {
	return unicode.IsSpace(r)
}

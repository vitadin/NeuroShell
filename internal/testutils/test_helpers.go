package testutils

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDataGenerator provides common test data
type TestDataGenerator struct{}

// NewTestDataGenerator creates a new test data generator
func NewTestDataGenerator() *TestDataGenerator {
	return &TestDataGenerator{}
}

// BasicVariables returns a set of basic variables for testing
func (g *TestDataGenerator) BasicVariables() map[string]string {
	return map[string]string{
		"name":    "test",
		"status":  "working",
		"message": "hello world",
		"count":   "42",
	}
}

// InterpolationTestCases returns test cases for variable interpolation
func (g *TestDataGenerator) InterpolationTestCases() []InterpolationTestCase {
	return []InterpolationTestCase{
		{
			Name:     "simple variable",
			Input:    "Hello ${name}",
			Expected: "Hello test",
			Variables: map[string]string{
				"name": "test",
			},
		},
		{
			Name:     "multiple variables",
			Input:    "${greeting}, ${name}!",
			Expected: "Hello, World!",
			Variables: map[string]string{
				"greeting": "Hello",
				"name":     "World",
			},
		},
		{
			Name:      "system variable",
			Input:     "User: ${@user}",
			Expected:  "User: testuser",
			Variables: map[string]string{},
		},
		{
			Name:     "nested variables",
			Input:    "${prefix}${middle}${suffix}",
			Expected: "HelloWorldTest",
			Variables: map[string]string{
				"prefix": "Hello",
				"middle": "World",
				"suffix": "Test",
			},
		},
		{
			Name:      "no variables",
			Input:     "plain text",
			Expected:  "plain text",
			Variables: map[string]string{},
		},
		{
			Name:     "empty variable",
			Input:    "Value: ${empty}",
			Expected: "Value: ",
			Variables: map[string]string{
				"empty": "",
			},
		},
	}
}

// InterpolationTestCase represents a test case for interpolation
type InterpolationTestCase struct {
	Name      string
	Input     string
	Expected  string
	Variables map[string]string
	ShouldErr bool
}

// ScriptTestData returns test script content
func (g *TestDataGenerator) ScriptTestData() map[string]string {
	return map[string]string{
		"basic.neuro": `# Basic script
\set[name="test"]
\get[name]`,

		"variables.neuro": `# Variable interpolation
\set[greeting="Hello"]
\set[name="World"]
\set[message="${greeting}, ${name}!"]
\get[message]`,

		"system.neuro": `# System variables
\get[@user]
\get[#test_mode]`,

		"invalid.neuro": `# Invalid command
\invalid[param="value"]`,

		"empty.neuro": `# Empty script
`,
	}
}

// AssertionHelpers provides common assertion patterns
type AssertionHelpers struct {
	t *testing.T
}

// NewAssertionHelpers creates assertion helpers for a test
func NewAssertionHelpers(t *testing.T) *AssertionHelpers {
	return &AssertionHelpers{t: t}
}

// AssertVariableEquals checks if a variable has the expected value
func (h *AssertionHelpers) AssertVariableEquals(ctx *MockContext, name, expected string) {
	actual, err := ctx.GetVariable(name)
	require.NoError(h.t, err, "Getting variable %s should not error", name)
	assert.Equal(h.t, expected, actual, "Variable %s should equal %s", name, expected)
}

// AssertVariableNotFound checks if a variable is not found
func (h *AssertionHelpers) AssertVariableNotFound(ctx *MockContext, name string) {
	_, err := ctx.GetVariable(name)
	assert.Error(h.t, err, "Variable %s should not be found", name)
	assert.Contains(h.t, err.Error(), "not found", "Error should indicate variable not found")
}

// AssertMapEquals compares two string maps
func (h *AssertionHelpers) AssertMapEquals(expected, actual map[string]string) {
	assert.Equal(h.t, len(expected), len(actual), "Maps should have same length")
	for k, v := range expected {
		assert.Equal(h.t, v, actual[k], "Value for key %s should match", k)
	}
}

// FileHelpers provides utilities for working with test files
type FileHelpers struct{}

// NewFileHelpers creates a new file helpers instance
func NewFileHelpers() *FileHelpers {
	return &FileHelpers{}
}

// CreateTempFile creates a temporary file with given content
func (f *FileHelpers) CreateTempFile(t *testing.T, filename, content string) string {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, filename)

	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err, "Should create temp file successfully")

	return filePath
}

// CreateTempDir creates a temporary directory structure
func (f *FileHelpers) CreateTempDir(t *testing.T, files map[string]string) string {
	tmpDir := t.TempDir()

	for filename, content := range files {
		filePath := filepath.Join(tmpDir, filename)

		// Create directory if needed
		dir := filepath.Dir(filePath)
		if dir != tmpDir {
			err := os.MkdirAll(dir, 0755)
			require.NoError(t, err, "Should create directory %s", dir)
		}

		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err, "Should create file %s", filename)
	}

	return tmpDir
}

// BenchmarkHelpers provides utilities for benchmark tests
type BenchmarkHelpers struct{}

// NewBenchmarkHelpers creates a new benchmark helpers instance
func NewBenchmarkHelpers() *BenchmarkHelpers {
	return &BenchmarkHelpers{}
}

// RunBenchmarkN runs a function N times for benchmarking
func (b *BenchmarkHelpers) RunBenchmarkN(b2 *testing.B, fn func()) {
	b2.ResetTimer()
	for i := 0; i < b2.N; i++ {
		fn()
	}
}

// GenerateLargeVariableSet creates a large set of variables for performance testing
func (b *BenchmarkHelpers) GenerateLargeVariableSet(count int) map[string]string {
	vars := make(map[string]string, count)
	for i := 0; i < count; i++ {
		vars[fmt.Sprintf("var_%d", i)] = fmt.Sprintf("value_%d", i)
	}
	return vars
}

// GenerateComplexInterpolationString creates a string with many variables for testing
func (b *BenchmarkHelpers) GenerateComplexInterpolationString(varCount int) string {
	result := ""
	for i := 0; i < varCount; i++ {
		result += fmt.Sprintf("${var_%d} ", i)
	}
	return result
}

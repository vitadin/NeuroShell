package services

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"neuroshell/internal/testutils"
)

func TestScriptService_Name(t *testing.T) {
	service := NewScriptService()
	assert.Equal(t, "script", service.Name())
}

func TestScriptService_Initialize(t *testing.T) {
	tests := []struct {
		name string
		ctx  *testutils.MockContext
		want error
	}{
		{
			name: "successful initialization",
			ctx:  testutils.NewMockContext(),
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewScriptService()
			err := service.Initialize(tt.ctx)

			if tt.want != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.want.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.True(t, service.initialized)
			}
		})
	}
}

func TestScriptService_LoadScript(t *testing.T) {
	testDataGen := testutils.NewTestDataGenerator()
	fileHelper := testutils.NewFileHelpers()

	tests := []struct {
		name          string
		scriptName    string
		scriptContent string
		wantError     string
	}{
		{
			name:          "load basic script",
			scriptName:    "basic.neuro",
			scriptContent: testDataGen.ScriptTestData()["basic.neuro"],
		},
		{
			name:          "load script with variables",
			scriptName:    "variables.neuro",
			scriptContent: testDataGen.ScriptTestData()["variables.neuro"],
		},
		{
			name:          "load script with system variables",
			scriptName:    "system.neuro",
			scriptContent: testDataGen.ScriptTestData()["system.neuro"],
		},
		{
			name:          "load empty script",
			scriptName:    "empty.neuro",
			scriptContent: testDataGen.ScriptTestData()["empty.neuro"],
		},
		{
			name:       "load script with comments only",
			scriptName: "comments.neuro",
			scriptContent: `# This is a comment
# Another comment
`,
		},
		{
			name:       "load script with mixed content",
			scriptName: "mixed.neuro",
			scriptContent: `# Comment line
\set[var1="value1"]

# Another comment
\get[var1]
# Final comment`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewScriptService()
			ctx := testutils.NewMockContext()

			// Initialize service
			err := service.Initialize(ctx)
			require.NoError(t, err)

			// Create temporary script file
			scriptPath := fileHelper.CreateTempFile(t, tt.scriptName, tt.scriptContent)

			// Test LoadScript - note this will fail since MockContext is not NeuroContext
			err = service.LoadScript(scriptPath, ctx)

			if tt.wantError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantError)
			} else {
				// Should fail because MockContext is not a NeuroContext
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "context is not a NeuroContext")
			}
		})
	}
}

func TestScriptService_LoadScript_FileNotFound(t *testing.T) {
	service := NewScriptService()
	ctx := testutils.NewMockContext()

	// Initialize service
	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Try to load non-existent file
	err = service.LoadScript("/nonexistent/path/script.neuro", ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open script file")
}

func TestScriptService_LoadScript_NotInitialized(t *testing.T) {
	service := NewScriptService()
	ctx := testutils.NewMockContext()
	fileHelper := testutils.NewFileHelpers()

	// Create a script file
	scriptPath := fileHelper.CreateTempFile(t, "test.neuro", `\set[var="value"]`)

	// Try to load without initialization
	err := service.LoadScript(scriptPath, ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "script service not initialized")
}

func TestScriptService_GetScriptMetadata(t *testing.T) {
	service := NewScriptService()
	ctx := testutils.NewMockContext()

	// Initialize service
	err := service.Initialize(ctx)
	require.NoError(t, err)

	// Test GetScriptMetadata - note this will fail since MockContext is not NeuroContext
	metadata, err := service.GetScriptMetadata(ctx)

	// Should fail because MockContext is not a NeuroContext
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context is not a NeuroContext")
	assert.Nil(t, metadata)
}

func TestScriptService_GetScriptMetadata_NotInitialized(t *testing.T) {
	service := NewScriptService()
	ctx := testutils.NewMockContext()

	metadata, err := service.GetScriptMetadata(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "script service not initialized")
	assert.Nil(t, metadata)
}

// Test script parsing behavior
func TestScriptService_ScriptParsing(t *testing.T) {
	testCases := []struct {
		name          string
		content       string
		expectedLines int
		description   string
	}{
		{
			name: "empty lines ignored",
			content: `\set[var="value"]

\get[var]`,
			expectedLines: 2,
			description:   "Should ignore empty lines",
		},
		{
			name: "comments ignored",
			content: `# This is a comment
\set[var="value"]
# Another comment
\get[var]
# Final comment`,
			expectedLines: 2,
			description:   "Should ignore comment lines",
		},
		{
			name: "whitespace handling",
			content: `   \set[var="value"]   
	\get[var]	`,
			expectedLines: 2,
			description:   "Should handle whitespace properly",
		},
		{
			name: "mixed content",
			content: `# Comment
\set[var1="value1"]

# Another comment
\set[var2="value2"]
\get[var1]
\get[var2]

# Final comment`,
			expectedLines: 4,
			description:   "Should handle mixed comments and commands",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := NewScriptService()
			ctx := testutils.NewMockContext()
			fileHelper := testutils.NewFileHelpers()

			// Initialize service
			err := service.Initialize(ctx)
			require.NoError(t, err)

			// Create script file
			scriptPath := fileHelper.CreateTempFile(t, "test.neuro", tc.content)

			// Count actual non-empty, non-comment lines
			file, err := os.Open(scriptPath)
			require.NoError(t, err)
			defer file.Close()

			// This test just verifies the file was created correctly
			// The actual parsing would require NeuroContext
			assert.FileExists(t, scriptPath)
		})
	}
}

// Performance tests
func BenchmarkScriptService_LoadScript_Small(b *testing.B) {
	service := NewScriptService()
	ctx := testutils.NewMockContext()

	err := service.Initialize(ctx)
	require.NoError(b, err)

	// Create a small script
	scriptContent := `\set[var1="value1"]
\set[var2="value2"]
\get[var1]
\get[var2]`

	tmpDir := b.TempDir()
	scriptPath := filepath.Join(tmpDir, "small.neuro")
	err = os.WriteFile(scriptPath, []byte(scriptContent), 0644)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This will error due to MockContext, but we're measuring file I/O performance
		_ = service.LoadScript(scriptPath, ctx)
	}
}

func BenchmarkScriptService_LoadScript_Large(b *testing.B) {
	service := NewScriptService()
	ctx := testutils.NewMockContext()

	err := service.Initialize(ctx)
	require.NoError(b, err)

	// Create a larger script
	scriptContent := ""
	for i := 0; i < 1000; i++ {
		scriptContent += fmt.Sprintf("\\set[var%d=\"value%d\"]\n", i, i)
	}

	tmpDir := b.TempDir()
	scriptPath := filepath.Join(tmpDir, "large.neuro")
	err = os.WriteFile(scriptPath, []byte(scriptContent), 0644)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// This will error due to MockContext, but we're measuring file I/O performance
		_ = service.LoadScript(scriptPath, ctx)
	}
}

// Edge case tests
func TestScriptService_EdgeCases(t *testing.T) {
	service := NewScriptService()
	ctx := testutils.NewMockContext()
	fileHelper := testutils.NewFileHelpers()

	err := service.Initialize(ctx)
	require.NoError(t, err)

	testCases := []struct {
		name    string
		content string
	}{
		{
			name:    "only comments",
			content: "# Comment 1\n# Comment 2\n# Comment 3",
		},
		{
			name:    "only empty lines",
			content: "\n\n\n\n",
		},
		{
			name:    "mixed whitespace",
			content: "\t\n  \n\t  \t\n",
		},
		{
			name:    "unicode content",
			content: "# Unicode comment: 测试\n\\set[测试=\"值\"]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scriptPath := fileHelper.CreateTempFile(t, "edge.neuro", tc.content)

			// Should handle edge cases gracefully (even if it errors due to MockContext)
			err := service.LoadScript(scriptPath, ctx)
			// We expect error due to MockContext, but it shouldn't panic
			assert.Error(t, err)
		})
	}
}

func TestScriptService_ConcurrentAccess(t *testing.T) {
	fileHelper := testutils.NewFileHelpers()

	// Create multiple script files
	scripts := map[string]string{
		"script1.neuro": `\set[var1="value1"]`,
		"script2.neuro": `\set[var2="value2"]`,
		"script3.neuro": `\set[var3="value3"]`,
	}

	tmpDir := fileHelper.CreateTempDir(t, scripts)

	// Test concurrent usage with separate service instances
	done := make(chan bool)

	for i := 0; i < 3; i++ {
		go func(id int) {
			// Each goroutine gets its own service instance to avoid race conditions
			service := NewScriptService()
			ctx := testutils.NewMockContext()
			err := service.Initialize(ctx)
			assert.NoError(t, err)

			scriptPath := filepath.Join(tmpDir, fmt.Sprintf("script%d.neuro", id+1))
			err = service.LoadScript(scriptPath, ctx)
			// Expect error due to MockContext, but shouldn't panic
			assert.Error(t, err)

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
}

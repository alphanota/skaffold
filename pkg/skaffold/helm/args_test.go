/*
Copyright 2022 The Skaffold Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package helm

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GoogleContainerTools/skaffold/v2/testutil"
)

func TestSanitizeFilePath(t *testing.T) {
	tests := []struct {
		description string
		isWindowsOS bool
		input       string
		output      string
	}{
		{
			description: "unescaped relative path on Windows",
			isWindowsOS: true,
			input:       `.\foo\not.escaped.relative.yaml`,
			output:      `.\\foo\\not.escaped.relative.yaml`,
		},
		{
			description: "unescaped absolute path on Windows",
			isWindowsOS: true,
			input:       `C:\Users\foo\not.escaped.abs.yaml`,
			output:      `C:\\Users\\foo\\not.escaped.abs.yaml`,
		},
		{
			description: "escaped relative path on Windows",
			isWindowsOS: true,
			input:       `.\\foo\\escaped.relative.yaml`,
			output:      `.\\foo\\escaped.relative.yaml`,
		},
		{
			description: "escaped absolute path on Windows",
			isWindowsOS: true,
			input:       `C:\\Users\\foo\\escaped.abs.yaml`,
			output:      `C:\\Users\\foo\\escaped.abs.yaml`,
		},
		{
			description: "escaped relative path with spaces on Windows",
			isWindowsOS: true,
			input:       `.\\foo bar\\escaped.spaces.relative.yaml`,
			output:      `".\\foo bar\\escaped.spaces.relative.yaml"`,
		},
		{
			description: "escaped absolute path with spaces on Windows",
			isWindowsOS: true,
			input:       `C:\\Users\\foo bar\\escaped.spaces.abs.yaml`,
			output:      `"C:\\Users\\foo bar\\escaped.spaces.abs.yaml"`,
		},
		{
			description: "unescaped relative path with spaces on Windows",
			isWindowsOS: true,
			input:       `.\foo bar\not.escaped.spaces.relative.yaml`,
			output:      `".\\foo bar\\not.escaped.spaces.relative.yaml"`,
		},
		{
			description: "unescaped absolute path with spaces on Windows",
			isWindowsOS: true,
			input:       `C:\Users\foo bar\not.escaped.spaces.abs.yaml`,
			output:      `"C:\\Users\\foo bar\\not.escaped.spaces.abs.yaml"`,
		},
		{
			description: "relative path on non-Windows",
			input:       `./foo/spaces.relative.yaml`,
			output:      `./foo/spaces.relative.yaml`,
		},
		{
			description: "absolute path on non-Windows",
			input:       `z/foo/spaces.abs.yaml`,
			output:      `z/foo/spaces.abs.yaml`,
		},
		{
			description: "relative path with spaces on non-Windows",
			input:       `./foo bar/spaces.relative.yaml`,
			output:      `"./foo bar/spaces.relative.yaml"`,
		},
		{
			description: "absolute path with spaces on non-Windows",
			input:       `z/foo bar/spaces.abs.yaml`,
			output:      `"z/foo bar/spaces.abs.yaml"`,
		},
		{
			description: "unescaped relative dir path on Windows",
			isWindowsOS: true,
			input:       `.\foo\not_escaped_relative\`,
			output:      `.\\foo\\not_escaped_relative\\`,
		},
		{
			description: "escaped relative dir path on Windows",
			isWindowsOS: true,
			input:       `.\\foo\\escaped_relative\\`,
			output:      `.\\foo\\escaped_relative\\`,
		},
		{
			description: "unescaped relative dir path on non-Windows",
			input:       `./foo/not_escaped_relative/`,
			output:      `./foo/not_escaped_relative/`,
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			output := SanitizeFilePath(test.input, test.isWindowsOS)
			t.CheckDeepEqual(test.output, output)
		})
	}
}

// MockGsutil implements gcs.Gsutil with a controlled Copy method
type mockGsutil struct{}

// Copy extracts the filename and returns success or failure based on the filename.
func (m *mockGsutil) Copy(ctx context.Context, src string, dst string, recursive bool) error {
	// Extract the filename from the source path
	filename := filepath.Base(src)

	// If the filename contains "not_exist", return an error (simulate failure)
	if strings.Contains(filename, "not_exist") {
		return fmt.Errorf("file not found: %s", src)
	}

	// Otherwise, return success
	return nil
}

func TestExtractValueFileFromGCS(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		description string
		src         string
		dst         string
		shouldErr   bool
	}{
		{
			description: "copy file from GCS",
			src:         "gs://bucket/path/to/values.yaml",
			dst:         filepath.Join(tempDir, "values.yaml"),
			shouldErr:   false,
		},
		{
			description: "fail to copy file from GCS",
			src:         "gs://bucket/path/to/not_exist_values.yaml",
			dst:         filepath.Join(tempDir, "not_exist_values.yaml"),
			shouldErr:   true,
		},
	}

	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			result, err := extractValueFileFromGCS(test.src, tempDir, &mockGsutil{})
			// Check if the error status matches the expected result
			t.CheckError(test.shouldErr, err)

			// If no error, validate the copied file path
			if err == nil {
				t.CheckDeepEqual(result, test.dst)
			}
		})
	}
}

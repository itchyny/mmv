package mmv

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMove(t *testing.T) {
	testCases := []struct {
		name     string
		files    map[string]string
		contents map[string]string
		expected map[string]string
		err      error
	}{
		{
			name:  "nothing",
			files: nil,
		},
		{
			name: "one file",
			files: map[string]string{
				"foo": "bar",
			},
			contents: map[string]string{
				"foo": "0",
			},
			expected: map[string]string{
				"bar": "0",
			},
		},
		{
			name: "two files",
			files: map[string]string{
				"foo": "qux",
				"bar": "quxx",
			},
			contents: map[string]string{
				"foo": "0",
				"bar": "1",
				"baz": "2",
			},
			expected: map[string]string{
				"qux":  "0",
				"quxx": "1",
				"baz":  "2",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "mvv-"+tc.name+"-")
			defer os.RemoveAll(dir)
			require.NoError(t, os.Chdir(dir))
			require.NoError(t, err)
			require.NoError(t, setupFiles(tc.contents))
			require.NoError(t, Move(tc.files))
			assert.Equal(t, tc.expected, fileContents("."))
		})
	}
}

func setupFiles(contents map[string]string) error {
	for f, cnt := range contents {
		if err := ioutil.WriteFile(f, []byte(cnt), 0600); err != nil {
			return err
		}
	}
	return nil
}

func fileContents(dir string) map[string]string {
	m := make(map[string]string)
	fis, _ := ioutil.ReadDir(dir)
	for _, fi := range fis {
		if fi.IsDir() {
			for k, v := range fileContents(filepath.Join(dir, fi.Name())) {
				m[k] = v
			}
		} else {
			path := filepath.Join(dir, fi.Name())
			cnt, _ := ioutil.ReadFile(path)
			m[path] = string(cnt)
		}
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

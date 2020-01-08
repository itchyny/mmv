package mmv

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRename(t *testing.T) {
	testCases := []struct {
		name     string
		files    map[string]string
		contents map[string]string
		expected map[string]string
		cnt      int
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
			cnt: 1,
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
			cnt: 2,
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
		{
			name: "swap two files",
			files: map[string]string{
				"foo": "bar",
				"bar": "foo",
			},
			cnt: 3,
			contents: map[string]string{
				"foo": "0",
				"bar": "1",
				"baz": "2",
			},
			expected: map[string]string{
				"bar": "0",
				"foo": "1",
				"baz": "2",
			},
		},
		{
			name: "two swaps",
			files: map[string]string{
				"foo": "bar",
				"bar": "foo",
				"baz": "qux",
				"qux": "baz",
			},
			cnt: 6,
			contents: map[string]string{
				"foo": "0",
				"bar": "1",
				"baz": "2",
				"qux": "3",
			},
			expected: map[string]string{
				"bar": "0",
				"foo": "1",
				"qux": "2",
				"baz": "3",
			},
		},
		{
			name: "three files",
			files: map[string]string{
				"foo": "bar",
				"bar": "baz",
				"baz": "qux",
			},
			cnt: 3,
			contents: map[string]string{
				"foo": "0",
				"bar": "1",
				"baz": "2",
			},
			expected: map[string]string{
				"bar": "0",
				"baz": "1",
				"qux": "2",
			},
		},
		{
			name: "cycle three files",
			files: map[string]string{
				"foo": "bar",
				"bar": "baz",
				"baz": "foo",
			},
			cnt: 4,
			contents: map[string]string{
				"foo": "0",
				"bar": "1",
				"baz": "2",
			},
			expected: map[string]string{
				"bar": "0",
				"baz": "1",
				"foo": "2",
			},
		},
		{
			name: "empty source path error",
			files: map[string]string{
				"foo": "baz",
				"":    "baz",
			},
			contents: map[string]string{
				"foo": "0",
				"bar": "1",
			},
			err: &emptyPathError{},
		},
		{
			name: "empty destination path error",
			files: map[string]string{
				"foo": "baz",
				"bar": "",
			},
			contents: map[string]string{
				"foo": "0",
				"bar": "1",
			},
			err: &emptyPathError{},
		},
		{
			name: "same destination error",
			files: map[string]string{
				"foo": "baz",
				"bar": "baz",
				"baz": "qux",
			},
			contents: map[string]string{
				"foo": "0",
				"bar": "1",
				"baz": "2",
			},
			err: &sameDestinationError{"baz"},
		},
		{
			name: "clean source path",
			files: map[string]string{
				"foo":  "bar",
				"bar/": "foo/",
			},
			cnt: 3,
			contents: map[string]string{
				"foo": "0",
				"bar": "1",
			},
			expected: map[string]string{
				"bar": "0",
				"foo": "1",
			},
		},
		{
			name: "cleaned path same destination error",
			files: map[string]string{
				"foo": "baz",
				"bar": "foo/../baz",
			},
			contents: map[string]string{
				"foo": "0",
				"bar": "1",
			},
			err: &sameDestinationError{"baz"},
		},
		{
			name: "same source and destination",
			files: map[string]string{
				"foo/": "foo",
				"bar/": "bar",
			},
			cnt: 0,
			contents: map[string]string{
				"foo": "0",
				"bar": "1",
			},
			expected: map[string]string{
				"foo": "0",
				"bar": "1",
			},
		},
		{
			name: "same source and destination with error",
			files: map[string]string{
				"foo/": "foo/",
				"bar/": "foo",
			},
			err: &sameDestinationError{"foo"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "mvv-"+tc.name+"-")
			defer os.RemoveAll(dir)
			require.NoError(t, os.Chdir(dir))
			require.NoError(t, err)
			require.NoError(t, setupFiles(tc.contents))
			rs, _ := buildRenames(tc.files)
			assert.Equal(t, tc.cnt, len(rs))
			got := Rename(tc.files)
			if tc.err == nil {
				require.NoError(t, got)
				assert.Equal(t, tc.expected, fileContents("."))
			} else {
				assert.Equal(t, tc.err, got)
			}
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

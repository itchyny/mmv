package mmv

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestRename(t *testing.T) {
	testCases := []struct {
		name     string
		files    map[string]string
		contents map[string]string
		expected map[string]string
		count    int
		err      string
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
			count: 1,
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
				"bar": "quux",
			},
			count: 2,
			contents: map[string]string{
				"foo": "0",
				"bar": "1",
				"baz": "2",
			},
			expected: map[string]string{
				"qux":  "0",
				"quux": "1",
				"baz":  "2",
			},
		},
		{
			name: "swap two files",
			files: map[string]string{
				"foo": "bar",
				"bar": "foo",
			},
			count: 3,
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
			count: 6,
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
			count: 3,
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
			count: 4,
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
			expected: map[string]string{
				"foo": "0",
				"bar": "1",
			},
			err: "empty path error",
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
			expected: map[string]string{
				"foo": "0",
				"bar": "1",
			},
			err: "empty path error",
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
			expected: map[string]string{
				"foo": "0",
				"bar": "1",
				"baz": "2",
			},
			err: "duplicate destination: baz",
		},
		{
			name: "clean source path",
			files: map[string]string{
				"foo":  "bar",
				"bar/": "foo/",
			},
			count: 3,
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
			name: "cleaned path same source error",
			files: map[string]string{
				"foo":        "baz",
				"bar/../foo": "bar",
			},
			contents: map[string]string{
				"foo": "0",
				"bar": "1",
			},
			expected: map[string]string{
				"foo": "0",
				"bar": "1",
			},
			err: "duplicate source: foo",
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
			expected: map[string]string{
				"foo": "0",
				"bar": "1",
			},
			err: "duplicate destination: baz",
		},
		{
			name: "same source and destination",
			files: map[string]string{
				"foo/": "foo",
				"bar/": "bar",
			},
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
			err: "duplicate destination: foo",
		},
		{
			name: "undo on error",
			files: map[string]string{
				"foo":  "bar",
				"bar":  "foo",
				"baz":  "qux",
				"qux":  "quux",
				"quux": "baz",
			},
			count: 7,
			contents: map[string]string{
				"foo": "0",
				"bar": "1",
				"baz": "2",
				"qux": "3",
			},
			expected: map[string]string{
				"foo": "0",
				"bar": "1",
				"baz": "2",
				"qux": "3",
			},
			err: "quux: ", // no such file or directory
		},
		{
			name: "create destination directory",
			files: map[string]string{
				"foo": "x/foo",
				"bar": "x/bar",
				"baz": "a/b/c/baz",
			},
			count: 3,
			contents: map[string]string{
				"foo": "0",
				"bar": "1",
				"baz": "2",
			},
			expected: map[string]string{
				"x/foo":     "0",
				"x/bar":     "1",
				"a/b/c/baz": "2",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "mmv-")
			if err != nil {
				t.Fatalf("ioutil.TempDir returned an error: %s", err)
			}
			defer os.RemoveAll(dir)
			if err := os.Chdir(dir); err != nil {
				t.Fatalf("os.Chdir returned an error: %s", err)
			}
			if err := setupFiles(tc.contents); err != nil {
				t.Fatalf("setupFiles returned an error: %s", err)
			}
			rs, _ := buildRenames(clone(tc.files))
			if got := len(rs); got != tc.count {
				t.Errorf("expected: %d, got: %d", tc.count, got)
			}
			err = Rename(tc.files)
			if tc.err == "" {
				if err != nil {
					t.Errorf("Rename returned an error: %s", err)
				}
			} else if !strings.Contains(err.Error(), tc.err) {
				t.Errorf("error should contain: %s, got: %s", tc.err, err)
			}
			if got := fileContents("."); !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected: %v, got: %v", tc.expected, got)
			}
		})
	}
}

func setupFiles(contents map[string]string) error {
	for f, cnt := range contents {
		if err := ioutil.WriteFile(f, []byte(cnt), 0o600); err != nil {
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
			m[filepath.ToSlash(path)] = string(cnt)
		}
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

func clone(xs map[string]string) map[string]string {
	ys := make(map[string]string, len(xs))
	for k, v := range xs {
		ys[k] = v
	}
	return ys
}

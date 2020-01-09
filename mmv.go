package mmv

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
)

// Rename multiple files.
func Rename(files map[string]string, dryRun bool) error {
	rs, err := buildRenames(files)
	if err != nil {
		return err
	}
	if dryRun {
		for _, r := range rs {
			fmt.Printf("%s => %s\n", r.src, r.dst)
		}
		return nil
	}
	for i, r := range rs {
		if err := doRename(r.src, r.dst); err != nil {
			// undo on error not to leave the temporary files
			// this does not undo directory creation
			for i--; i >= 0; i-- {
				if r = rs[i]; os.Rename(r.dst, r.src) != nil {
					// something wrong happens so give up not to overwrite files
					break
				}
			}
			return err
		}
	}
	return nil
}

// rename with creating the destination directory
func doRename(src, dst string) (err error) {
	// first of all, try renaming the file, which will succeed in most cases
	if err = os.Rename(src, dst); err != nil && os.IsNotExist(err) {
		// check the source file existence to exit without creating the destination
		// directory when the both source file and destination directory do not exist
		if _, err := os.Stat(src); err != nil {
			return err
		}
		// create the destination directory
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return err
		}
		// try renaming again
		return os.Rename(src, dst)
	}
	return
}

type rename struct {
	src, dst string
}

type emptyPathError struct{}

func (err *emptyPathError) Error() string {
	return "empty path error"
}

type sameSourceError struct {
	path string
}

func (err *sameSourceError) Error() string {
	return fmt.Sprintf("duplicate source: %s", err.path)
}

type sameDestinationError struct {
	path string
}

func (err *sameDestinationError) Error() string {
	return fmt.Sprintf("duplicate destination: %s", err.path)
}

func buildRenames(files map[string]string) ([]rename, error) {
	revs := make(map[string]string, len(files)) // reverse of files

	// list the current rename sources
	srcs := make([]string, 0, len(files))
	for src := range files {
		srcs = append(srcs, src)
	}

	// clean the paths and check duplication
	for _, src := range srcs {
		dst := files[src]
		if src == "" || dst == "" {
			return nil, &emptyPathError{}
		}
		if d := filepath.Clean(src); d != src {
			delete(files, src)
			src = d
			if _, ok := files[src]; ok {
				return nil, &sameSourceError{src}
			}
			files[src] = dst
		}
		if d := filepath.Clean(dst); d != dst {
			dst = d
			files[src] = dst
		}
		if _, ok := revs[dst]; ok {
			return nil, &sameDestinationError{dst}
		}
		revs[dst] = src
	}

	// remove source == destination
	for src, dst := range files {
		if src == dst {
			delete(files, src)
			delete(revs, dst)
		}
	}

	// list the renames
	var i int
	rs := make([]rename, 0, 2*len(files))
	vs := make(map[string]int, len(files))
	for _, dst := range files {
		if vs[dst] > 0 {
			continue
		}
		i++ // connected component identifier

		// mark the nodes in the connected component and check cycle
		var cycle bool
		for {
			vs[dst] = i
			if x, ok := files[dst]; ok {
				dst = x
				if vs[x] > 0 {
					cycle = vs[x] == i
					break
				}
			} else {
				break
			}
		}

		// if there is a cycle, rename to a temporary file
		var tmp string
		if cycle {
			tmp = randomPath(filepath.Dir(dst))
			rs = append(rs, rename{dst, tmp})
			vs[dst]--
		}

		// rename from the leaf node
		for {
			if src, ok := revs[dst]; ok && (!cycle || vs[src] == i) {
				rs = append(rs, rename{src, dst})
				if !cycle {
					vs[dst] = i
				}
				dst = src
			} else {
				break
			}
		}

		// if there is a cycle, rename the temporary file
		if cycle {
			rs = append(rs, rename{tmp, dst})
		}
	}
	return rs, nil
}

func randomPath(dir string) string {
	for {
		path := filepath.Join(dir, fmt.Sprint(rand.Uint64()))
		if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
			return path
		}
	}
}

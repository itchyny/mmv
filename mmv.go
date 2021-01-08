package mmv

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

// Rename multiple files.
func Rename(files map[string]string) error {
	rs, err := buildRenames(files)
	if err != nil {
		return err
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
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
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
	return "duplicate source: " + err.path
}

type sameDestinationError struct {
	path string
}

func (err *sameDestinationError) Error() string {
	return "duplicate destination: " + err.path
}

type invalidRenameError struct {
	src, dst string
}

func (err *invalidRenameError) Error() string {
	return "invalid rename: " + err.src + ", " + err.dst
}

type temporaryPathError struct {
	dir string
}

func (err *temporaryPathError) Error() string {
	return "failed to create a temporary path: " + err.dir
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
		if k, l := len(src), len(dst); k > l && src[l] == filepath.Separator && src[:l] == dst ||
			k < l && dst[k] == filepath.Separator && dst[:k] == src {
			return nil, &invalidRenameError{src, dst}
		}
		revs[dst] = src
	}

	// group paths by directory depth
	srcdepths := make([][]string, 1)
	dstdepths := make([][]string, 1)
	for src, dst := range files {
		// group source paths by directory depth
		i := strings.Count(src, string(filepath.Separator))
		if len(srcdepths) <= i {
			xs := make([][]string, i*2)
			copy(xs, srcdepths)
			srcdepths = xs
		}
		srcdepths[i] = append(srcdepths[i], src)
		// group destination paths by directory depth
		i = strings.Count(dst, string(filepath.Separator))
		if len(dstdepths) <= i {
			xs := make([][]string, i*2)
			copy(xs, dstdepths)
			dstdepths = xs
		}
		dstdepths[i] = append(dstdepths[i], dst)
	}

	// result renames
	count := len(files)
	rs := make([]rename, 0, 2*count)

	// check if any parent directory will be moved
	for i := len(srcdepths) - 1; i >= 0; i-- {
	L:
		for _, src := range srcdepths[i] {
			for j := 0; j < i; j++ {
				for _, s := range srcdepths[j] {
					if k := len(s); len(src) > k && src[k] == filepath.Separator && src[:k] == s {
						if d := files[s]; s != d {
							if dst, l := files[src], len(d); i == j+1 && len(dst) > l && dst[:l] == d && dst[l:] == src[k:] {
								// skip moving a file when it moves along with the closest parent directory
								delete(files, src)
								delete(revs, dst)
							} else {
								// move to a temporary path before any parent directory is moved
								tmp, err := temporaryPath(filepath.Dir(s))
								if err != nil {
									return nil, err
								}
								rs = append(rs, rename{src, tmp})
								files[tmp] = files[dst]
								delete(files, src)
								revs[dst] = tmp
							}
							continue L
						}
					}
				}
			}
			// remove if source path is equal to destination path
			if dst := files[src]; src == dst {
				delete(files, src)
				delete(revs, dst)
			}
		}
	}

	// list renames in increasing destination directory depth order
	i, vs := 0, make(map[string]int, count)
	for _, dsts := range dstdepths {
		for _, dst := range dsts {
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
				var err error
				tmp, err = temporaryPath(filepath.Dir(dst))
				if err != nil {
					return nil, err
				}
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
	}
	return rs, nil
}

// create a temporary path where there is no file currently
func temporaryPath(dir string) (string, error) {
	for i := 0; i < 256; i++ {
		path := filepath.Join(dir, fmt.Sprint(rand.Uint64()))
		if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
			return path, nil
		}
	}
	return "", &temporaryPathError{dir}
}

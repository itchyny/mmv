package mmv

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
)

// Rename multiple files.
func Rename(files map[string]string) error {
	rs, err := buildRenames(files)
	if err != nil {
		return err
	}
	for _, r := range rs {
		if err := os.Rename(r.src, r.dst); err != nil {
			return err
		}
	}
	return nil
}

type rename struct {
	src, dst string
}

type emptyPathError struct{}

func (err *emptyPathError) Error() string {
	return "empty path error"
}

type sameDestinationError struct {
	path string
}

func (err *sameDestinationError) Error() string {
	return fmt.Sprintf("duplicate destination: %s", err.path)
}

func buildRenames(files map[string]string) ([]rename, error) {
	rs := make([]rename, 0, 2*len(files))
	revs := make(map[string]string, len(files))
	srcs := make([]string, 0, len(files))
	for src := range files {
		srcs = append(srcs, src)
	}
	for _, src := range srcs {
		dst := files[src]
		if src == "" || dst == "" {
			return nil, &emptyPathError{}
		}
		if d := filepath.Clean(src); d != src {
			delete(files, src)
			src = d
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
	for src, dst := range files {
		if src == dst {
			delete(files, src)
			delete(revs, dst)
		}
	}
	var i int
	vs := make(map[string]int, len(files))
	for _, dst := range files {
		if vs[dst] > 0 {
			continue
		}
		i++
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
		var tmp string
		if cycle {
			tmp = randomPath(filepath.Dir(dst))
			rs = append(rs, rename{dst, tmp})
			vs[dst]--
		}
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

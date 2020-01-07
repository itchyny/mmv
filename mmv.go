package mmv

import "os"

// Move multiple files.
func Move(files map[string]string) error {
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

func buildRenames(files map[string]string) ([]rename, error) {
	rs := make([]rename, 0, 2*len(files))
	for src, dst := range files {
		rs = append(rs, rename{src, dst})
	}
	return rs, nil
}

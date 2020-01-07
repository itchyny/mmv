package mmv

import "os"

// Move multiple files.
func Move(files map[string]string) error {
	for k, v := range files {
		if err := os.Rename(k, v); err != nil {
			return err
		}
	}
	return nil
}

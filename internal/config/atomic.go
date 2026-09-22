package config

import (
	"errors"
	"os"
	"path/filepath"
)

// WriteFileAtomic writes data to path so that a crash or power loss leaves
// either the old file or the new one, never a truncated mix: the bytes go to a
// temporary file in the same directory, are synced, and then renamed over the
// original (a same-volume rename replaces it in one step on every platform).
func WriteFileAtomic(path string, data []byte, perm os.FileMode) (err error) {
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() {
		// Leave nothing behind on failure; after a successful rename the
		// temporary path no longer exists.
		if err != nil {
			err = errors.Join(err, os.Remove(tmpPath))
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		return errors.Join(err, tmp.Close())
	}
	if err := tmp.Sync(); err != nil {
		return errors.Join(err, tmp.Close())
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpPath, perm); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

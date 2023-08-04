package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"time"
)

// UserHomeDir returns the home directory path of the current user by using
// os.UserHomeDir() and user.Current() as fallback.
func UserHomeDir() string {
	home, _ := os.UserHomeDir()
	if len(home) > 0 {
		return home
	}

	currUser, _ := user.Current()
	if currUser != nil {
		home = currUser.HomeDir
	}

	return home
}

// StoreCacheFile writes the provided object to a temporary cache file before
// renaming it to the specified filename, returning any errors encountered
// during this process.
func StoreCacheFile(filename string, obj interface{}, fileMode os.FileMode) error {
	if len(filename) == 0 {
		return fmt.Errorf("filename is blank")
	}

	tmpFilename := filename + ".tmp-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	if err := writeCacheFile(tmpFilename, fileMode, obj); err != nil {
		return err
	}

	if err := os.Rename(tmpFilename, filename); err != nil {
		return fmt.Errorf("failed to replace old cache file, %w", err)
	}

	return nil
}

func writeCacheFile(filename string, fileMode os.FileMode, obj interface{}) error {
	var f *os.File
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, fileMode)
	if err != nil {
		return fmt.Errorf("failed to create cached file %w", err)
	}

	defer func() {
		closeErr := f.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("failed to close cached file, %w", closeErr)
		}
	}()

	encoder := json.NewEncoder(f)

	if err = encoder.Encode(obj); err != nil {
		return fmt.Errorf("failed to serialize cached file, %w", err)
	}

	return nil
}

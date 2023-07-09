package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"time"
)

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

func StoreCacheFile(filename string, obj interface{}, fileMode os.FileMode) (err error) {
	tmpFilename := filename + ".tmp-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	if err := writeCacheFile(tmpFilename, fileMode, obj); err != nil {
		return err
	}

	if err := os.Rename(tmpFilename, filename); err != nil {
		return fmt.Errorf("failed to replace old cache file, %w", err)
	}

	return nil
}

func writeCacheFile(filename string, fileMode os.FileMode, obj interface{}) (err error) {
	var f *os.File
	f, err = os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, fileMode)
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
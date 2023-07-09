package utils

import (
	"fmt"
	"os"
	"os/exec"
)

// Last return the last element of a slice.
func Last[E any](s []E) (E, bool) {
	if len(s) == 0 {
		var zero E
		return zero, false
	}
	return s[len(s)-1], true
}

func CheckErr(err interface{}) {
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func OpenUrlInBrowser(url string) error {
	fmt.Printf("Opening URL in default browser: %s\n", url)
	err := exec.Command("open", url).Start()
	if err != nil {
		return err
	}
	return nil
}

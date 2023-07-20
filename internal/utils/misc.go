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

// CheckErr if err is not nil then print the stderr and exist.
func CheckErr(err interface{}) {
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// OpenUrlInBrowser open a URL in the default browser.
// Only works on macOS right now.
func OpenUrlInBrowser(url string) error {
	fmt.Printf("Opening URL in default browser: %s\n", url)
	err := exec.Command("open", url).Start()
	if err != nil {
		return err
	}
	return nil
}

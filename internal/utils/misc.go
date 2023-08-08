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

// CheckErr if msg is not nil then print the stderr and exit.
func CheckErr(msg interface{}) {
	if msg != nil {
		switch t := msg.(type) {
		case error:
			if t.Error() == "^C" {
				fmt.Println("Cancelled")
			} else {
				_, _ = fmt.Fprintln(os.Stderr, "Error:", msg)
			}
		default:
			_, _ = fmt.Fprintln(os.Stderr, "Error:", msg)
		}
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

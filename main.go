package main

import (
	"fmt"
	"os"

	"going/cmd"
)

func main() {
	root := cmd.NewCmdRoot()
	if err := root.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

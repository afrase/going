package main

import (
	"going/cmd"
	"going/internal/utils"
)

var version = "dev"

func main() {
	root := cmd.NewCmdRoot(version)
	err := root.Execute()
	utils.CheckErr(err)
}

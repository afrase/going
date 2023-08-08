package main

import (
	"going/cmd"
	"going/internal/utils"
)

func main() {
	root := cmd.NewCmdRoot()
	err := root.Execute()
	utils.CheckErr(err)
}

package main

import (
	"os"

	"github.com/catrossim/monbeat/cmd"

	_ "github.com/catrossim/monbeat/include"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

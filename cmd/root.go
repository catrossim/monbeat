package cmd

import (
	"github.com/catrossim/monbeat/beater"

	_ "github.com/catrossim/monbeat/include"
	cmd "github.com/elastic/beats/libbeat/cmd"
)

// Name of this beat
var Name = "monbeat"

// RootCmd to handle beats cli
var RootCmd = cmd.GenRootCmd(Name, "", beater.New)

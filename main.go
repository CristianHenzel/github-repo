package main

import (
	"github.com/CristianHenzel/github-repo/cmd"
)

// Version holds the application version
// It gets filled automatically at build time
var Version string

// BuildDate holds the date and time at which the application was build
// It gets filled automatically at build time
var BuildDate string

func main() {
	cmd.Version = Version
	cmd.BuildDate = BuildDate
	cmd.Execute()
}

package main

import (
	"github.com/martins6/acolyte/cmd"
	"github.com/martins6/acolyte/internal/updater"
)

var Version = "dev"

func main() {
	updater.Version = Version
	cmd.SetVersion(Version)
	cmd.Execute()
}

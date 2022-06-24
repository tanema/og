package main

import (
	_ "embed"

	"github.com/tanema/og/cmd"
)

//go:embed VERSION
var version string

func main() {
	cmd.Execute(version)
}

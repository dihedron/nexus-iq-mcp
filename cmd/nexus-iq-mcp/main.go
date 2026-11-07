package main

import (
	"os"

	"github.com/dihedron/nexus-iq-mcp/cmd/nexus-iq-mcp/command"
	"github.com/jessevdk/go-flags"
)

func main() {
	os.Exit(run())
}

// run parses the command line and runs the command; it returns the exit
// code so that the deferred cleanup runs before the process exits.
func run() int {
	defer cleanup()

	options := command.Commands{}
	parser := flags.NewParser(&options, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			return 0
		}
		return 1
	}
	return 0
}

// Command rit is Ritari, a long-term memory for AI agents.
package main

import (
	"os"

	"golang.org/x/term"

	"github.com/cndreisbach/ritari/internal/cli"
)

func main() {
	io := cli.DefaultIO(term.IsTerminal(int(os.Stdout.Fd())))
	os.Exit(cli.Main(os.Args[1:], io))
}

package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"

	"github.com/cndreisbach/ritari/internal/cli"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"rit": func() { os.Exit(cli.Main(os.Args[1:], cli.DefaultIO(false))) },
	})
}

// TestAcceptance runs the scenarios of SPECIFICATION.md §12 against the real
// command line, with a pinned clock and actor.
func TestAcceptance(t *testing.T) {
	repo, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	testscript.Run(t, testscript.Params{
		Dir: "testdata/scripts",
		Setup: func(env *testscript.Env) error {
			env.Setenv("RITARI_CONFIG", env.WorkDir+"/config.yaml")
			env.Setenv("RITARI_TEST_NOW", "2026-07-29T12:00:00Z")
			env.Setenv("RITARI_AGENT", "claude-code")
			env.Setenv("TZ", "UTC")
			env.Setenv("HOME", env.WorkDir)
			return nil
		},
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			// okflint checks a bundle with the third-party `okf lint`
			// (github.com/abcubed3/okf), skipping if it is not installed.
			"okflint": func(ts *testscript.TestScript, neg bool, args []string) {
				if len(args) != 1 {
					ts.Fatalf("usage: okflint <bundle>")
				}
				bin, err := exec.LookPath("okf")
				if err != nil {
					ts.Logf("okf not installed; skipping OKF conformance check")
					return
				}
				cmd := exec.Command(bin, "lint", ts.MkAbs(args[0]))
				cmd.Dir = repo // mise shims resolve versions from the repo's mise.toml
				out, err := cmd.CombinedOutput()
				ts.Logf("okf lint:\n%s", out)
				failed := err != nil || strings.Contains(string(out), "[ERROR]")
				if failed != neg {
					ts.Fatalf("okf lint: unexpected result (err=%v)", err)
				}
			},
		},
	})
}

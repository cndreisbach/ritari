package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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
	baseManifest, err := os.ReadFile(filepath.Join("testdata", "okf-base.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	testscript.Run(t, testscript.Params{
		Dir: "testdata/scripts",
		Setup: func(env *testscript.Env) error {
			env.Setenv("RITARI_CONFIG", env.WorkDir+"/config.yaml")
			env.Setenv("RITARI_TEST_NOW", "2026-07-29T12:00:00Z")
			// Clear agent and model markers so the developer's environment cannot
			// change the stamped actor; each scenario sets what it needs.
			for _, k := range []string{
				"ROVODEV_CLI", "ROVO_DEV_CLI_VERSION",
				"CLAUDE_CODE_CHILD_SESSION", "CLAUDECODE",
				"CURSOR_AGENT",
				"CODEX_THREAD_ID", "CODEX_CI", "CODEX_SANDBOX",
				"AI_AGENT", "PI_CODING_AGENT", "PI_MODEL",
				"RITARI_AGENT_MODEL",
			} {
				env.Setenv(k, "")
			}
			env.Setenv("RITARI_AGENT", "claude-code")
			env.Setenv("TZ", "UTC")
			env.Setenv("HOME", env.WorkDir)
			return nil
		},
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			// okflint checks a bundle with okflint (github.com/mattdav/okflint)
			// against testdata/okf-base.yaml, failing on OKF errors. It is
			// skipped if okflint is not installed (an error instead when
			// RITARI_REQUIRE_OKFLINT is set, as in CI) or RITARI_SKIP_OKFLINT is set.
			"okflint": func(ts *testscript.TestScript, neg bool, args []string) {
				if len(args) != 1 {
					ts.Fatalf("usage: okflint <bundle>")
				}
				if os.Getenv("RITARI_SKIP_OKFLINT") != "" {
					ts.Logf("RITARI_SKIP_OKFLINT is set; skipping OKF conformance check")
					return
				}
				bin, err := exec.LookPath("okflint")
				if err != nil {
					if os.Getenv("RITARI_REQUIRE_OKFLINT") != "" {
						ts.Fatalf("okflint not found, and RITARI_REQUIRE_OKFLINT is set")
					}
					ts.Logf("okflint not installed; skipping OKF conformance check")
					return
				}
				// okflint only validates files under a manifest root, so point a
				// copy of the shipped manifest at this bundle.
				manifest := strings.Replace(string(baseManifest), `- path: "."`, "- path: "+strconv.Quote(ts.MkAbs(args[0])), 1)
				if manifest == string(baseManifest) {
					ts.Fatalf("testdata/okf-base.yaml has no `- path: \".\"` root to replace")
				}
				mpath := ts.MkAbs("okflint-" + filepath.Base(args[0]) + ".yaml")
				if err := os.WriteFile(mpath, []byte(manifest), 0o644); err != nil {
					ts.Fatalf("writing okflint manifest: %v", err)
				}
				cmd := exec.Command(bin, "validate", "--manifest", mpath)
				cmd.Dir = repo // mise shims resolve versions from the repo's mise.toml
				out, err := cmd.CombinedOutput()
				ts.Logf("okflint validate:\n%s", out)
				if failed := err != nil; failed != neg {
					ts.Fatalf("okflint validate: unexpected result (err=%v)", err)
				}
			},
		},
	})
}

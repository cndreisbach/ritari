// Package cli implements the `rit` command line (§9).
package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/alecthomas/kong"

	"github.com/cndreisbach/ritari/internal/actor"
	"github.com/cndreisbach/ritari/internal/bundle"
	"github.com/cndreisbach/ritari/internal/note"
	"github.com/cndreisbach/ritari/internal/output"
	"github.com/cndreisbach/ritari/internal/registry"
	"github.com/cndreisbach/ritari/internal/screen"
)

// CLI is the command tree.
type CLI struct {
	Output      string `short:"o" placeholder:"text|yaml|json" help:"Output format (default: yaml unless stdout is a terminal)."`
	ProjectName string `name:"project" short:"p" placeholder:"NAME" help:"Project to operate on (default: RITARI_PROJECT, then the registry default)."`

	Note    NoteCmd    `cmd:"" help:"Write, read, edit, and manage notes."`
	Read    ViewCmd    `cmd:"" help:"Print a note's raw markdown."`
	Search  SearchCmd  `cmd:"" help:"Search notes by text and structured filters."`
	Context ContextCmd `cmd:"" help:"Follow links outward from a note."`
	Daily   DailyCmd   `cmd:"" help:"Open or write a day's note."`
	Brief   BriefCmd   `cmd:"" help:"Print the latest daily notes and open tasks, for the start of a session."`
	Recent  RecentCmd  `cmd:"" help:"List recently modified notes."`
	Ls      LsCmd      `cmd:"" help:"List concept IDs."`
	Index   IndexCmd   `cmd:"" help:"Regenerate or check index.md files."`
	Lint    LintCmd    `cmd:"" help:"Report conformance errors and hygiene warnings."`
	Types   TypesCmd   `cmd:"" help:"List the type vocabulary."`
	Project ProjectCmd `cmd:"" help:"Manage the project registry."`
	Config  ConfigCmd  `cmd:"" help:"Get or set configuration."`
	Install InstallCmd `cmd:"" help:"Install agent integrations."`
	Skills  SkillsCmd  `cmd:"" hidden:"" help:"Old spelling of install skills."`
	Version VersionCmd `cmd:"" help:"Print the version."`
}

// IO is the process environment a run uses.
type IO struct {
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
	Getenv     func(string) string
	IsTerminal bool
}

// Ctx is passed to every command's Run.
type Ctx struct {
	IO
	Machine     bool
	JSON        bool
	projectFlag string
	reg         *registry.Registry
}

// reported is an error whose failure document has already been emitted.
type reported struct{ msg string }

func (r reported) Error() string { return r.msg }

type exitCode int

// Main runs the CLI and returns the process exit status.
func Main(args []string, io IO) (code int) {
	var cli CLI
	parser, err := kong.New(&cli,
		kong.Name("rit"),
		kong.Description("Ritari: a long-term memory for AI agents, kept as OKF markdown files."),
		kong.Writers(io.Stdout, io.Stderr),
		kong.Exit(func(c int) { panic(exitCode(c)) }),
		kong.UsageOnError(),
	)
	if err != nil {
		warn(io, err)
		return 2
	}
	defer func() {
		if r := recover(); r != nil {
			if c, ok := r.(exitCode); ok {
				code = int(c)
				return
			}
			panic(r)
		}
	}()
	kctx, err := parser.Parse(args)
	if err != nil {
		warn(io, err)
		return 2
	}
	format, err := mode(cli.Output, io)
	if err != nil {
		warn(io, err)
		return 2
	}
	machine, jsonOut := format != "text", format == "json"
	ctx := &Ctx{IO: io, Machine: machine, JSON: jsonOut, projectFlag: cli.ProjectName}
	if err := kctx.Run(ctx); err != nil {
		var r reported
		if !errors.As(err, &r) && machine {
			doc := output.New().Set("success", false).Set("error", err.Error())
			if jsonOut {
				_ = output.WriteJSON(io.Stdout, doc)
			} else {
				_ = output.Write(io.Stdout, doc)
			}
		}
		warn(io, err)
		return 1
	}
	return 0
}

func warn(io IO, err error) {
	_, _ = fmt.Fprintf(io.Stderr, "rit: %s\n", screen.Escape(err.Error()))
}

func mode(flag string, io IO) (string, error) {
	v := strings.TrimSpace(flag)
	if v == "" {
		v = strings.TrimSpace(io.Getenv("RITARI_OUTPUT"))
	}
	switch strings.ToLower(v) {
	case "":
		if io.IsTerminal {
			return "text", nil
		}
		return "yaml", nil
	case "text", "yaml", "json":
		return strings.ToLower(v), nil
	}
	return "", fmt.Errorf("unknown output format %q (want text, yaml, or json)", v)
}

// Emit writes machine-mode documents.
func (c *Ctx) Emit(docs ...*output.Doc) error {
	if !c.Machine {
		return nil
	}
	if c.JSON {
		return output.WriteJSON(c.Stdout, docs...)
	}
	return output.Write(c.Stdout, docs...)
}

// Printf writes human text (text mode only). Invisible characters are
// escaped so a reader sees them (§8.1).
func (c *Ctx) Printf(format string, a ...any) {
	if !c.Machine {
		_, _ = io.WriteString(c.Stdout, screen.Escape(fmt.Sprintf(format, a...)))
	}
}

// hidden screens a note and names its worst finding (§8.1).
func hidden(n *note.Note) ([]screen.Finding, string) {
	fs := screen.Note(n)
	return fs, screen.Worst(fs)
}

// hiddenMarker is the text-mode mark for a note's worst finding, with a
// leading space, or "".
func hiddenMarker(worst string) string {
	switch worst {
	case screen.HiddenInstruction:
		return "  [HIDDEN INSTRUCTION]"
	case screen.HiddenContent:
		return "  [hidden content]"
	}
	return ""
}

// Registry loads the project registry.
func (c *Ctx) Registry() (*registry.Registry, error) {
	if c.reg != nil {
		return c.reg, nil
	}
	p, err := registry.Path(c.Getenv)
	if err != nil {
		return nil, err
	}
	r, err := registry.Load(p)
	if err != nil {
		return nil, err
	}
	c.reg = r
	return r, nil
}

// Root is the active project's bundle directory.
func (c *Ctx) Root() (string, error) {
	r, err := c.Registry()
	if err != nil {
		return "", err
	}
	p, err := r.Active(c.projectFlag, c.Getenv)
	if err != nil {
		return "", err
	}
	return p.Path, nil
}

// Open loads the active bundle.
func (c *Ctx) Open() (*bundle.Bundle, error) {
	root, err := c.Root()
	if err != nil {
		return nil, err
	}
	return bundle.Open(root)
}

// Env is the identity and clock for writes.
func (c *Ctx) Env() (bundle.Env, error) {
	r, err := c.Registry()
	if err != nil {
		return bundle.Env{}, err
	}
	env := bundle.Env{Actor: actor.Detect(actor.Lookup{
		Getenv: c.Getenv, Author: r.Author, GitEmail: actor.GitEmail, Username: actor.Username,
	})}
	// RITARI_TEST_NOW pins the clock for the acceptance tests.
	if v := c.Getenv("RITARI_TEST_NOW"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return bundle.Env{}, fmt.Errorf("RITARI_TEST_NOW: %w", err)
		}
		env.Now = func() time.Time { return t }
	}
	return env, nil
}

// readStdinIf returns stdin's contents when v is "-", else v.
func (c *Ctx) readStdinIf(v string) (string, error) {
	if v != "-" {
		return v, nil
	}
	data, err := io.ReadAll(c.Stdin)
	return string(data), err
}

func parseMeta(specs []string) ([][2]string, error) {
	var out [][2]string
	for _, s := range specs {
		k, v, ok := strings.Cut(s, "=")
		if !ok || strings.TrimSpace(k) == "" {
			return nil, fmt.Errorf("--meta %q is not key=value", s)
		}
		out = append(out, [2]string{strings.TrimSpace(k), v})
	}
	return out, nil
}

func splitTags(s string) []string {
	var out []string
	for _, t := range strings.Split(s, ",") {
		if t = strings.TrimSpace(t); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// VersionCmd prints the version.
type VersionCmd struct{}

// version is set at link time for release builds:
// -ldflags "-X github.com/cndreisbach/ritari/internal/cli.version=1.2.3".
var version string

// Version is the build's version string.
func Version() string {
	if version != "" {
		return version
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	if v := info.Main.Version; v != "" && v != "(devel)" {
		return v
	}
	rev, dirty := "", false
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}
	if rev == "" {
		return "dev"
	}
	if len(rev) > 12 {
		rev = rev[:12]
	}
	if dirty {
		rev += "-dirty"
	}
	return rev
}

// Run prints the version.
func (VersionCmd) Run(c *Ctx) error {
	v := Version()
	c.Printf("rit %s\n", v)
	return c.Emit(output.New().Set("success", true).Set("version", v))
}

// DefaultIO is the real process environment.
func DefaultIO(isTerminal bool) IO {
	return IO{Stdin: os.Stdin, Stdout: os.Stdout, Stderr: os.Stderr, Getenv: os.Getenv, IsTerminal: isTerminal}
}

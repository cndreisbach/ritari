package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/cndreisbach/ritari/internal/actor"
	"github.com/cndreisbach/ritari/internal/bundle"
	"github.com/cndreisbach/ritari/internal/output"
	"github.com/cndreisbach/ritari/internal/registry"
	"github.com/cndreisbach/ritari/internal/skills"
)

// ProjectCmd groups the project subcommands.
type ProjectCmd struct {
	List    ProjectListCmd    `cmd:"" help:"List registered projects."`
	Init    ProjectInitCmd    `cmd:"" help:"Create a new bundle and register it."`
	Add     ProjectAddCmd     `cmd:"" help:"Register an existing bundle."`
	Remove  ProjectRemoveCmd  `cmd:"" help:"Unregister a project (files are untouched)."`
	Default ProjectDefaultCmd `cmd:"" help:"Set the default project."`
}

// ProjectListCmd is `project list`.
type ProjectListCmd struct{}

// Run executes `project list`.
func (ProjectListCmd) Run(c *Ctx) error {
	r, err := c.Registry()
	if err != nil {
		return err
	}
	if len(r.Projects) == 0 {
		c.Printf("No projects. Create one: rit project init <name> <path>\n")
		return c.Emit(output.New().Set("success", true).Set("count", 0).Set("projects", output.ExplicitList{}))
	}
	docs := []*output.Doc{output.New().Set("success", true).Set("count", len(r.Projects)).Set("default", r.Default)}
	for _, p := range r.Projects {
		mark := " "
		if p.Name == r.Default {
			mark = "*"
		}
		c.Printf("%s %s  %s\n", mark, p.Name, p.Path)
		docs = append(docs, output.New().Set("name", p.Name).Set("path", p.Path).Set("is_default", p.Name == r.Default))
	}
	return c.Emit(docs...)
}

const sourcesReadme = `This directory holds raw material (transcripts, exports, PDFs) that notes
cite. Two rules keep the knowledge base a valid OKF bundle:

1. Never put .md files here. Anything ending in .md is a concept to OKF tools.
2. Never put material in hidden (dot-) directories. Other OKF tools descend
   into them and would treat a stray .md file there as a conformance error.

Cite a file from a note with its bundle path, for example:

    rit note write --title "..." --source "Call transcript /sources/call.txt"
`

// ProjectInitCmd is `project init`.
type ProjectInitCmd struct {
	Name    string `arg:"" help:"Project name."`
	Path    string `arg:"" help:"Directory for the new bundle."`
	Default bool   `help:"Make it the default project."`
}

// Run executes `project init`.
func (cmd *ProjectInitCmd) Run(c *Ctx) error {
	r, err := c.Registry()
	if err != nil {
		return err
	}
	if _, ok := r.Find(cmd.Name); ok {
		return fmt.Errorf("project %q already exists", cmd.Name)
	}
	abs, err := registry.ExpandPath(cmd.Path)
	if err != nil {
		return err
	}
	if info, err := os.Stat(abs); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%s is not a directory", abs)
		}
		b, err := bundle.Open(abs)
		if err != nil {
			return err
		}
		if len(b.Files) > 0 {
			return fmt.Errorf("%s already holds notes; use `rit project add %s %s` to register it", abs, cmd.Name, abs)
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(abs, "sources")); errors.Is(err, fs.ErrNotExist) {
		if err := bundle.WriteFileAtomic(filepath.Join(abs, "sources", "README.txt"), []byte(sourcesReadme)); err != nil {
			return err
		}
	}
	b, err := bundle.Open(abs)
	if err != nil {
		return err
	}
	if _, err := b.SyncIndexes(false); err != nil {
		return err
	}
	p, err := r.Add(cmd.Name, abs, cmd.Default)
	if err != nil {
		return err
	}
	if err := r.Save(); err != nil {
		return err
	}
	c.Printf("Created project %s at %s\n", p.Name, p.Path)
	return c.Emit(output.New().Set("success", true).Set("status", "created").Set("name", p.Name).Set("path", p.Path))
}

// ProjectAddCmd is `project add`.
type ProjectAddCmd struct {
	Name    string `arg:"" help:"Project name."`
	Path    string `arg:"" help:"Existing bundle directory."`
	Default bool   `help:"Make it the default project."`
}

// Run executes `project add`.
func (cmd *ProjectAddCmd) Run(c *Ctx) error {
	r, err := c.Registry()
	if err != nil {
		return err
	}
	abs, err := registry.ExpandPath(cmd.Path)
	if err != nil {
		return err
	}
	info, err := os.Stat(abs)
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%s does not exist; use `rit project init %s %s` to create it", abs, cmd.Name, abs)
	} else if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", abs)
	}
	p, err := r.Add(cmd.Name, abs, cmd.Default)
	if err != nil {
		return err
	}
	if err := r.Save(); err != nil {
		return err
	}
	c.Printf("Added project %s at %s\n", p.Name, p.Path)
	return c.Emit(output.New().Set("success", true).Set("status", "added").Set("name", p.Name).Set("path", p.Path))
}

// ProjectRemoveCmd is `project remove`.
type ProjectRemoveCmd struct {
	Name string `arg:"" help:"Project name."`
}

// Run executes `project remove`.
func (cmd *ProjectRemoveCmd) Run(c *Ctx) error {
	r, err := c.Registry()
	if err != nil {
		return err
	}
	if err := r.Remove(cmd.Name); err != nil {
		return err
	}
	if err := r.Save(); err != nil {
		return err
	}
	c.Printf("Removed project %s\n", cmd.Name)
	return c.Emit(output.New().Set("success", true).Set("status", "removed").Set("name", cmd.Name))
}

// ProjectDefaultCmd is `project default`.
type ProjectDefaultCmd struct {
	Name string `arg:"" help:"Project name."`
}

// Run executes `project default`.
func (cmd *ProjectDefaultCmd) Run(c *Ctx) error {
	r, err := c.Registry()
	if err != nil {
		return err
	}
	if err := r.SetDefault(cmd.Name); err != nil {
		return err
	}
	if err := r.Save(); err != nil {
		return err
	}
	c.Printf("Default project is now %s\n", cmd.Name)
	return c.Emit(output.New().Set("success", true).Set("status", "default").Set("name", cmd.Name))
}

// ConfigCmd groups configuration subcommands.
type ConfigCmd struct {
	Author ConfigAuthorCmd `cmd:"" help:"Get or set the configured author identity (human:<id>)."`
}

// ConfigAuthorCmd is `config author`.
type ConfigAuthorCmd struct {
	Value string `arg:"" optional:"" help:"New author, conventionally human:<id>."`
}

// Run executes `config author`.
func (cmd *ConfigAuthorCmd) Run(c *Ctx) error {
	r, err := c.Registry()
	if err != nil {
		return err
	}
	if cmd.Value == "" {
		c.Printf("%s\n", r.Author)
		return c.Emit(output.New().Set("success", true).Set("author", r.Author))
	}
	r.Author = actor.Normalize(cmd.Value)
	if err := r.Save(); err != nil {
		return err
	}
	c.Printf("Author set to %s\n", r.Author)
	return c.Emit(output.New().Set("success", true).Set("status", "set").Set("author", r.Author))
}

// SkillsCmd groups skill subcommands.
type SkillsCmd struct {
	Install SkillsInstallCmd `cmd:"" help:"Install the agent skills."`
}

// SkillsInstallCmd is `skills install`.
type SkillsInstallCmd struct {
	Agent string `help:"Target agent: agents (~/.agents/skills) or claude (~/.claude/skills)."`
	Dir   string `help:"Install into this directory instead."`
	Force bool   `help:"Overwrite existing SKILL.md files."`
}

// Run executes `skills install`.
func (cmd *SkillsInstallCmd) Run(c *Ctx) error {
	if cmd.Agent != "" && cmd.Dir != "" {
		return errors.New("--agent and --dir are alternatives; pass one")
	}
	dir := cmd.Dir
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		switch cmd.Agent {
		case "", "agents":
			dir = filepath.Join(home, ".agents", "skills")
		case "claude":
			dir = filepath.Join(home, ".claude", "skills")
		default:
			return fmt.Errorf("unknown agent %q (known: agents, claude); use --dir for anything else", cmd.Agent)
		}
	}
	dir, err := registry.ExpandPath(dir)
	if err != nil {
		return err
	}
	installed, err := skills.Install(dir, cmd.Force)
	if err != nil {
		return err
	}
	if len(installed) == 0 {
		c.Printf("All skills already present in %s (use --force to overwrite).\n", dir)
	} else {
		c.Printf("Installed skills to %s: %s\n", dir, strings.Join(installed, ", "))
	}
	return c.Emit(output.New().Set("success", true).Set("dir", dir).Set("count", len(installed)).
		Set("installed", output.ExplicitList(installed)))
}

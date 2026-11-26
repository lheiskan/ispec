// Package harness installs ispec into an LLM agent harness.
//
// It owns placement — where skills go, what config is edited, what the agent
// is told at session start — for each supported harness. It does not own
// judgment: that lives in skills/*/SKILL.md, which this package only copies.
// It does not compute facts about a canon either; that is package status.
//
// Targets are named and looked up, rather than assumed, so that adding a
// second harness is a new file here instead of a rewrite. Claude Code is the
// only target today, which makes the registry look like overhead — but the
// alternative is hardcoding ~/.claude into the install path and discovering
// later that every caller assumed it.
package harness

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SkillsSubdir is where embedded skills live inside the asset tree.
const SkillsSubdir = "skills"

// SessionStartSubcommand is what an installed hook invokes.
const SessionStartSubcommand = "harness session-start"

// Default is the target assumed when --target is not given.
const Default = "claude-code"

// backupSuffix marks the one-time copy taken before the first edit of a
// config file ispec did not create.
const backupSuffix = ".ispec-bak"

// Assets is the embedded asset tree: skills/<name>/SKILL.md. An interface
// rather than embed.FS so tests can supply a fixture.
type Assets interface {
	fs.ReadDirFS
	fs.ReadFileFS
}

// Action is what an install step did, for reporting.
type Action string

const (
	Installed Action = "installed"
	Updated   Action = "updated"
	Unchanged Action = "unchanged"
	Removed   Action = "removed"
	Missing   Action = "missing"
	Stale     Action = "stale"
	WouldSet  Action = "would set"
)

// NeedsWork reports whether an action means the install is not current.
// `harness check` exits non-zero on any of these so a script can gate on it.
func (a Action) NeedsWork() bool {
	return a == Missing || a == Stale
}

// Step is one reported unit of work.
type Step struct {
	Action Action
	Detail string
}

// Options configures an install.
type Options struct {
	// Dir is the harness's config directory. Empty means the target's own
	// default. Point it at a scratch directory to rehearse an install —
	// which is the only safe way to test one, since the real file is the
	// user's live agent configuration.
	Dir string

	// HookCommand is the command a session-start hook should run. Empty
	// means the running executable plus " harness session-start". An
	// absolute path is the point: a hook runs in a non-interactive shell
	// whose $PATH may not contain the install prefix.
	HookCommand string

	// DryRun reports what would change and writes nothing.
	DryRun bool
}

// Target is one supported agent harness.
type Target interface {
	// Name is the --target value.
	Name() string

	// Describe is a one-line summary for help output.
	Describe() string

	// DefaultDir is the harness's own config directory.
	DefaultDir() (string, error)

	// Install writes skills and config. It must be idempotent.
	Install(Assets, Options) ([]Step, error)

	// Check reports what is installed and whether it is current, writing
	// nothing.
	Check(Assets, Options) ([]Step, error)

	// Uninstall removes only what Install wrote. It takes the assets
	// because the set of skills to remove is the set this binary ships.
	Uninstall(Assets, Options) ([]Step, error)
}

// registry holds the supported targets by name.
var registry = map[string]Target{
	claudeCode{}.Name(): claudeCode{},
}

// Lookup resolves a target by name. An unknown name is an error listing what
// is accepted — never a silent fallback to the default, which would install
// Claude Code's layout for someone who asked for something else.
func Lookup(name string) (Target, error) {
	if name == "" {
		name = Default
	}
	t, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown harness %q (supported: %s)", name, strings.Join(Names(), ", "))
	}
	return t, nil
}

// Names lists the supported target names in deterministic order.
func Names() []string {
	out := make([]string, 0, len(registry))
	for n := range registry {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// Targets lists the supported targets in name order.
func Targets() []Target {
	out := make([]Target, 0, len(registry))
	for _, n := range Names() {
		out = append(out, registry[n])
	}
	return out
}

// SkillNames lists the skill directories in the asset tree.
func SkillNames(a Assets) ([]string, error) {
	entries, err := a.ReadDir(SkillsSubdir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out, nil
}

// skillBody reads one embedded SKILL.md. The embed path separator is always
// "/", regardless of host platform.
func skillBody(a Assets, name string) ([]byte, error) {
	return a.ReadFile(SkillsSubdir + "/" + name + "/SKILL.md")
}

// selfCommand builds the absolute invocation of this binary plus a
// subcommand. Quoted if the path contains spaces, since the result is handed
// to a shell.
//
// It refuses to record an ephemeral path. `go run ./cmd/ispec harness install`
// resolves to a binary in Go's build cache, which is deleted when the run
// ends: the hook would then fail at the start of every session on the
// machine, with an error pointing at a directory that no longer exists. That
// is a bad enough outcome, and a confusing enough one to diagnose, to be
// worth blocking outright.
func selfCommand(sub string) (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	if isEphemeralPath(exe) {
		return "", fmt.Errorf(
			"this binary is running from a temporary build directory (%s), so the hook "+
				"it installed would break as soon as that directory is cleaned up.\n"+
				"Install a real binary first:\n"+
				"  go install ./cmd/ispec && \"$(go env GOBIN GOPATH | head -1)/bin/ispec\" harness install\n"+
				"or just: make install", exe)
	}
	if strings.ContainsAny(exe, " \t") {
		exe = `"` + exe + `"`
	}
	return exe + " " + sub, nil
}

// isEphemeralPath reports whether a path lives in Go's build cache, which is
// where `go run` puts the binary it builds.
func isEphemeralPath(exe string) bool {
	slashed := filepath.ToSlash(exe)
	return strings.Contains(slashed, "/go-build") && strings.Contains(slashed, "/exe/")
}

// hookCommand resolves the command a hook should run.
func hookCommand(opts Options) (string, error) {
	if opts.HookCommand != "" {
		return opts.HookCommand, nil
	}
	return selfCommand(SessionStartSubcommand)
}

// actionOrWould downgrades a write action under --dry-run.
func actionOrWould(opts Options, a Action) Action {
	if opts.DryRun {
		return WouldSet
	}
	return a
}

// writeFileAtomic writes via a temporary file in the destination directory,
// then renames. A crash mid-write leaves the original intact rather than a
// truncated config file.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

// replaceFile creates the parent directory if needed, then writes data to
// path, replacing whatever was there. The rename inside writeFileAtomic is
// what does the replacing, so no separate unlink is needed.
func replaceFile(path string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return writeFileAtomic(path, data, perm)
}

// homeDir is the user's home directory.
func homeDir() (string, error) {
	return os.UserHomeDir()
}

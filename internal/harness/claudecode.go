package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// claudeCode installs into Claude Code: skills as directories under
// <dir>/skills, and a SessionStart hook in <dir>/settings.json.
type claudeCode struct{}

func (claudeCode) Name() string { return "claude-code" }

func (claudeCode) Describe() string {
	return "Claude Code — skills under ~/.claude/skills, SessionStart hook in ~/.claude/settings.json"
}

func (claudeCode) DefaultDir() (string, error) {
	if d := os.Getenv("CLAUDE_DIR"); d != "" {
		return d, nil
	}
	h, err := homeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, ".claude"), nil
}

// resolveDir returns the config directory an operation should act on.
func (t claudeCode) resolveDir(opts Options) (string, error) {
	if opts.Dir != "" {
		return opts.Dir, nil
	}
	return t.DefaultDir()
}

// settingsPath is the file the hook install edits.
func (t claudeCode) settingsPath(opts Options) (string, error) {
	dir, err := t.resolveDir(opts)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "settings.json"), nil
}

func (t claudeCode) skillPath(dir, name string) string {
	return filepath.Join(dir, "skills", name, "SKILL.md")
}

func (t claudeCode) Install(a Assets, opts Options) ([]Step, error) {
	dir, err := t.resolveDir(opts)
	if err != nil {
		return nil, err
	}
	names, err := SkillNames(a)
	if err != nil {
		return nil, err
	}

	var steps []Step
	for _, name := range names {
		want, err := skillBody(a, name)
		if err != nil {
			return nil, err
		}
		dest := t.skillPath(dir, name)

		switch got, readErr := os.ReadFile(dest); {
		case readErr == nil && string(got) == string(want):
			steps = append(steps, Step{Unchanged, "skill " + name})
			continue
		case readErr == nil:
			steps = append(steps, Step{actionOrWould(opts, Updated), "skill " + name})
		default:
			steps = append(steps, Step{actionOrWould(opts, Installed), "skill " + name})
		}
		if opts.DryRun {
			continue
		}
		if err := replaceFile(dest, want, 0o644); err != nil {
			return nil, err
		}
	}

	hookStep, err := t.installHook(opts)
	if err != nil {
		return nil, err
	}
	return append(steps, hookStep), nil
}

func (t claudeCode) Check(a Assets, opts Options) ([]Step, error) {
	dir, err := t.resolveDir(opts)
	if err != nil {
		return nil, err
	}
	names, err := SkillNames(a)
	if err != nil {
		return nil, err
	}

	var steps []Step
	for _, name := range names {
		want, err := skillBody(a, name)
		if err != nil {
			return nil, err
		}
		dest := t.skillPath(dir, name)
		got, readErr := os.ReadFile(dest)
		switch {
		case os.IsNotExist(readErr):
			steps = append(steps, Step{Missing, "skill " + name})
		case readErr != nil:
			return nil, readErr
		case string(got) != string(want):
			steps = append(steps, Step{Stale, "skill " + name + " (differs from this binary's copy)"})
		default:
			steps = append(steps, Step{Unchanged, "skill " + name})
		}
	}

	want, err := hookCommand(opts)
	if err != nil {
		return nil, err
	}
	path, err := t.settingsPath(opts)
	if err != nil {
		return nil, err
	}
	found, err := t.findHook(path)
	switch {
	case err != nil:
		return nil, err
	case found == "":
		steps = append(steps, Step{Missing, "session-start hook"})
	case found != want:
		steps = append(steps, Step{Stale, fmt.Sprintf("session-start hook points at %s", found)})
	default:
		steps = append(steps, Step{Unchanged, "session-start hook"})
	}
	return steps, nil
}

func (t claudeCode) Uninstall(a Assets, opts Options) ([]Step, error) {
	dir, err := t.resolveDir(opts)
	if err != nil {
		return nil, err
	}
	names, err := SkillNames(a)
	if err != nil {
		return nil, err
	}

	var steps []Step
	for _, name := range names {
		dest := t.skillPath(dir, name)

		if _, statErr := os.Lstat(dest); statErr != nil {
			steps = append(steps, Step{Missing, "skill " + name})
			continue
		}
		steps = append(steps, Step{actionOrWould(opts, Removed), "skill " + name})
		if opts.DryRun {
			continue
		}
		if err := os.Remove(dest); err != nil {
			return nil, err
		}
		// Remove the skill directory only if nothing else lives in it.
		// A user who added files beside SKILL.md keeps them.
		if entries, err := os.ReadDir(filepath.Dir(dest)); err == nil && len(entries) == 0 {
			_ = os.Remove(filepath.Dir(dest))
		}
	}

	hookStep, err := t.removeHook(opts)
	if err != nil {
		return nil, err
	}
	return append(steps, hookStep), nil
}

// installHook merges ispec's SessionStart group into settings.json,
// re-emitting everything it does not own from the original bytes.
func (t claudeCode) installHook(opts Options) (Step, error) {
	path, err := t.settingsPath(opts)
	if err != nil {
		return Step{}, err
	}
	command, err := hookCommand(opts)
	if err != nil {
		return Step{}, err
	}

	original, had, err := readSettings(path)
	if err != nil {
		return Step{}, err
	}

	updated, err := mergeHook(original, command)
	if err != nil {
		return Step{}, err
	}
	if string(updated) == string(original) {
		return Step{Unchanged, "session-start hook"}, nil
	}
	if opts.DryRun {
		return Step{WouldSet, "session-start hook -> " + command}, nil
	}
	if err := backupOnce(path, original, had); err != nil {
		return Step{}, err
	}
	if err := replaceFile(path, updated, 0o644); err != nil {
		return Step{}, err
	}
	return Step{Installed, "session-start hook -> " + command}, nil
}

// removeHook drops ispec's SessionStart group, leaving every other group and
// every other member of the file alone.
func (t claudeCode) removeHook(opts Options) (Step, error) {
	path, err := t.settingsPath(opts)
	if err != nil {
		return Step{}, err
	}
	original, had, err := readSettings(path)
	if err != nil {
		return Step{}, err
	}
	if !had {
		return Step{Missing, "session-start hook"}, nil
	}
	updated, removed, err := dropHook(original)
	if err != nil {
		return Step{}, err
	}
	if !removed {
		return Step{Missing, "session-start hook"}, nil
	}
	if opts.DryRun {
		return Step{WouldSet, "remove session-start hook"}, nil
	}
	if err := backupOnce(path, original, had); err != nil {
		return Step{}, err
	}
	if err := replaceFile(path, updated, 0o644); err != nil {
		return Step{}, err
	}
	return Step{Removed, "session-start hook"}, nil
}

// findHook returns the command of ispec's installed hook, or "".
func (t claudeCode) findHook(path string) (string, error) {
	original, had, err := readSettings(path)
	if err != nil || !had {
		return "", err
	}
	groups, err := sessionStartGroups(original)
	if err != nil {
		return "", err
	}
	for _, g := range groups {
		if cmd, ok := ourCommand(g); ok {
			return cmd, nil
		}
	}
	return "", nil
}

// readSettings reads a config file, reporting whether it existed. A file that
// is not valid JSON is never overwritten: it is more likely a config ispec
// does not understand than one it should replace.
func readSettings(path string) (content []byte, had bool, err error) {
	b, readErr := os.ReadFile(path)
	if os.IsNotExist(readErr) {
		return []byte("{}"), false, nil
	}
	if readErr != nil {
		return nil, false, readErr
	}
	if !json.Valid(b) {
		return nil, false, fmt.Errorf("%s is not valid JSON; refusing to modify it", path)
	}
	return b, true, nil
}

// backupOnce copies a pre-existing config file aside before the first edit.
// Only once: a second install must not overwrite the pristine backup with an
// already-modified file.
func backupOnce(path string, original []byte, had bool) error {
	if !had {
		return nil
	}
	bak := path + backupSuffix
	if _, err := os.Stat(bak); !os.IsNotExist(err) {
		return err
	}
	return replaceFile(bak, original, 0o644)
}

// mergeHook rewrites only .hooks.SessionStart, repointing ispec's own group
// if present rather than appending a duplicate.
func mergeHook(original []byte, command string) ([]byte, error) {
	top, err := decodeObject(original)
	if err != nil {
		return nil, err
	}
	hooks, groups, err := hooksAndGroups(top)
	if err != nil {
		return nil, err
	}

	kept := make([]json.RawMessage, 0, len(groups)+1)
	for _, g := range groups {
		if _, ok := ourCommand(g); ok {
			continue // repoint rather than duplicate
		}
		kept = append(kept, g)
	}
	ours, err := hookGroup(command)
	if err != nil {
		return nil, err
	}
	kept = append(kept, ours)

	return writeGroups(top, hooks, kept)
}

// dropHook removes ispec's group from .hooks.SessionStart.
func dropHook(original []byte) ([]byte, bool, error) {
	top, err := decodeObject(original)
	if err != nil {
		return nil, false, err
	}
	hooks, groups, err := hooksAndGroups(top)
	if err != nil {
		return nil, false, err
	}
	kept := make([]json.RawMessage, 0, len(groups))
	removed := false
	for _, g := range groups {
		if _, ok := ourCommand(g); ok {
			removed = true
			continue
		}
		kept = append(kept, g)
	}
	if !removed {
		return original, false, nil
	}
	out, err := writeGroups(top, hooks, kept)
	return out, true, err
}

// hooksAndGroups pulls .hooks and .hooks.SessionStart out of the top object.
func hooksAndGroups(top object) (object, []json.RawMessage, error) {
	var hooks object
	if raw, ok := top.get("hooks"); ok {
		var err error
		if hooks, err = decodeObject(raw); err != nil {
			return nil, nil, fmt.Errorf("hooks: %w", err)
		}
	} else {
		hooks = object{}
	}

	var groups []json.RawMessage
	if raw, ok := hooks.get("SessionStart"); ok {
		if err := json.Unmarshal(raw, &groups); err != nil {
			return nil, nil, fmt.Errorf("hooks.SessionStart: %w", err)
		}
	}
	return hooks, groups, nil
}

// writeGroups splices the group list back into the file at the right depth.
// settings.json nesting is { hooks: { SessionStart: [ ... ] } }, so the array
// sits two levels in.
func writeGroups(top, hooks object, groups []json.RawMessage) ([]byte, error) {
	groupsRaw, err := encodeArray(groups, 2)
	if err != nil {
		return nil, err
	}
	hooks = hooks.set("SessionStart", groupsRaw)
	hooksRaw, err := hooks.encode(1)
	if err != nil {
		return nil, err
	}
	top = top.set("hooks", hooksRaw)
	out, err := top.encode(0)
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}

// hookGroup builds the one group ispec owns, with members in the order every
// existing settings file already uses.
func hookGroup(command string) (json.RawMessage, error) {
	quoted, err := marshalString(command)
	if err != nil {
		return nil, err
	}
	entry, err := object{}.
		set("type", json.RawMessage(`"command"`)).
		set("command", quoted).
		encode(4)
	if err != nil {
		return nil, err
	}
	entries, err := encodeArray([]json.RawMessage{entry}, 3)
	if err != nil {
		return nil, err
	}
	return object{}.
		set("matcher", json.RawMessage(`""`)).
		set("hooks", entries).
		encode(2)
}

// ourCommand reports whether a SessionStart group is ispec's, returning its
// command.
//
// Ownership keys on the *binary name*, not just the subcommand. Any tool that
// installs an agent hook the same way plausibly spells it
// `<its-binary> harness session-start`, so matching the subcommand alone would
// make ispec hijack a sibling tool's hook on install and delete it on
// uninstall.
func ourCommand(group json.RawMessage) (string, bool) {
	var g struct {
		Hooks []struct {
			Type    string `json:"type"`
			Command string `json:"command"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(group, &g); err != nil {
		return "", false
	}
	for _, h := range g.Hooks {
		if isOurCommand(h.Command) {
			return h.Command, true
		}
	}
	return "", false
}

// isOurCommand matches "<path-to-ispec> harness session-start".
//
// Both separators are normalized by hand rather than with filepath: that
// package follows the *host* convention, so on Unix it would leave a
// backslashed Windows path intact and treat the whole thing as one filename.
// A settings file can be read on a different platform than it was written on.
func isOurCommand(command string) bool {
	exe, rest, ok := splitCommand(command)
	if !ok || rest != SessionStartSubcommand {
		return false
	}
	base := strings.ToLower(exe)
	if i := strings.LastIndexAny(base, `/\`); i >= 0 {
		base = base[i+1:]
	}
	return base == "ispec" || base == "ispec.exe"
}

// splitCommand splits a command into its executable and the remaining
// arguments, honouring a leading double-quoted path.
func splitCommand(command string) (exe, rest string, ok bool) {
	command = strings.TrimSpace(command)
	if command == "" {
		return "", "", false
	}
	if command[0] == '"' {
		end := strings.Index(command[1:], `"`)
		if end < 0 {
			return "", "", false
		}
		return command[1 : 1+end], strings.TrimSpace(command[end+2:]), true
	}
	i := strings.IndexAny(command, " \t")
	if i < 0 {
		return command, "", true
	}
	return command[:i], strings.TrimSpace(command[i+1:]), true
}

// sessionStartGroups reads .hooks.SessionStart from raw settings bytes.
func sessionStartGroups(original []byte) ([]json.RawMessage, error) {
	top, err := decodeObject(original)
	if err != nil {
		return nil, err
	}
	_, groups, err := hooksAndGroups(top)
	return groups, err
}

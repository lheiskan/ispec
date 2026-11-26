package harness

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// testAssets is a stand-in for the embedded skill tree.
func testAssets() Assets {
	return fstest.MapFS{
		"skills/alpha/SKILL.md": &fstest.MapFile{Data: []byte("alpha body\n")},
		"skills/beta/SKILL.md":  &fstest.MapFile{Data: []byte("beta body\n")},
	}
}

// settingsFixture mirrors the shape of a real settings.json: members ispec
// does not own, a sibling tool's SessionStart hook, and a field this binary
// has never heard of (timeout).
const settingsFixture = `{
  "env": {
    "FOO": "bar"
  },
  "hooks": {
    "SessionStart": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "/usr/local/bin/othertool harness session-start"
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "echo pre",
            "timeout": 5
          }
        ]
      }
    ]
  },
  "effortLevel": "high"
}
`

func writeFixture(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte(settingsFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func install(t *testing.T, dir string, opts Options) []Step {
	t.Helper()
	opts.Dir = dir
	if opts.HookCommand == "" {
		opts.HookCommand = "/opt/ispec/ispec " + SessionStartSubcommand
	}
	steps, err := claudeCode{}.Install(testAssets(), opts)
	if err != nil {
		t.Fatalf("install: %v", err)
	}
	return steps
}

func TestIsOurCommand(t *testing.T) {
	cases := []struct {
		command string
		want    bool
	}{
		{"/opt/bin/ispec harness session-start", true},
		{"ispec harness session-start", true},
		{`"/Applications/My Tools/ispec" harness session-start`, true},
		{`C:\tools\ispec.exe harness session-start`, true},

		// Any tool installing an agent hook this way plausibly spells it
		// `<its-binary> harness session-start`. Matching the subcommand alone
		// would make ispec adopt such a hook and then delete it.
		{"/usr/local/bin/othertool harness session-start", false},
		{"/usr/local/bin/othertoolispec harness session-start", false},

		{"/opt/bin/ispec lint", false},
		{"/opt/bin/ispec harness install", false},
		{"", false},
	}
	for _, c := range cases {
		if got := isOurCommand(c.command); got != c.want {
			t.Errorf("isOurCommand(%q) = %v, want %v", c.command, got, c.want)
		}
	}
}

func TestInstallPreservesForeignSettings(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir)
	install(t, dir, Options{})

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	// Member order must be untouched: reordering a user's config on install
	// turns a one-line change into an unreviewable diff.
	wantOrder := []string{`"env"`, `"hooks"`, `"effortLevel"`}
	last := -1
	for _, key := range wantOrder {
		i := strings.Index(string(got), key)
		if i < 0 {
			t.Fatalf("key %s missing after install:\n%s", key, got)
		}
		if i < last {
			t.Errorf("key %s moved; member order not preserved:\n%s", key, got)
		}
		last = i
	}

	// A field this binary does not model must survive verbatim.
	if !strings.Contains(string(got), `"timeout": 5`) {
		t.Errorf("unknown field timeout was dropped:\n%s", got)
	}

	var parsed struct {
		Hooks struct {
			SessionStart []json.RawMessage `json:"SessionStart"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(got, &parsed); err != nil {
		t.Fatalf("result is not valid JSON: %v\n%s", err, got)
	}
	if n := len(parsed.Hooks.SessionStart); n != 2 {
		t.Fatalf("SessionStart has %d group(s), want 2 (the sibling tool's plus ours)", n)
	}
	if _, ours := ourCommand(parsed.Hooks.SessionStart[0]); ours {
		t.Error("ispec's group replaced the sibling tool's group rather than appending")
	}
	if _, ours := ourCommand(parsed.Hooks.SessionStart[1]); !ours {
		t.Error("ispec's group was not appended")
	}
}

func TestInstallIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir)

	install(t, dir, Options{})
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	steps := install(t, dir, Options{})
	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Errorf("second install changed the file:\n%s", second)
	}
	for _, s := range steps {
		if s.Action != Unchanged {
			t.Errorf("second install reported %q for %q, want unchanged", s.Action, s.Detail)
		}
	}
}

func TestInstallBacksUpOnceAndRefusesInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir)

	install(t, dir, Options{})
	bak, err := os.ReadFile(path + backupSuffix)
	if err != nil {
		t.Fatalf("no backup taken: %v", err)
	}
	if string(bak) != settingsFixture {
		t.Error("backup is not the pristine original")
	}

	// A second install must not overwrite the pristine backup with an
	// already-modified file.
	install(t, dir, Options{})
	again, err := os.ReadFile(path + backupSuffix)
	if err != nil {
		t.Fatal(err)
	}
	if string(again) != settingsFixture {
		t.Error("second install overwrote the backup")
	}

	bad := t.TempDir()
	if err := os.WriteFile(filepath.Join(bad, "settings.json"), []byte("{ not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := (claudeCode{}).Install(testAssets(), Options{Dir: bad, HookCommand: "/opt/ispec/ispec " + SessionStartSubcommand}); err == nil {
		t.Error("install overwrote a settings file it could not parse")
	}
}

func TestDryRunWritesNothing(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir)

	install(t, dir, Options{DryRun: true})

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != settingsFixture {
		t.Errorf("dry run modified settings.json:\n%s", got)
	}
	if _, err := os.Stat(filepath.Join(dir, "skills")); !os.IsNotExist(err) {
		t.Error("dry run wrote skills")
	}
}

func TestUninstallLeavesForeignHook(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir)
	install(t, dir, Options{})

	if _, err := (claudeCode{}).Uninstall(testAssets(), Options{Dir: dir}); err != nil {
		t.Fatalf("uninstall: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "othertool harness session-start") {
		t.Errorf("uninstall removed a sibling tool's hook:\n%s", got)
	}
	if strings.Contains(string(got), "/opt/ispec/ispec") {
		t.Errorf("uninstall left ispec's own hook:\n%s", got)
	}
	if !strings.Contains(string(got), `"timeout": 5`) {
		t.Errorf("uninstall dropped an unmodelled field:\n%s", got)
	}
}

func TestCheckReportsMissingThenCurrent(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir)
	opts := Options{Dir: dir, HookCommand: "/opt/ispec/ispec " + SessionStartSubcommand}

	steps, err := claudeCode{}.Check(testAssets(), opts)
	if err != nil {
		t.Fatal(err)
	}
	needs := 0
	for _, s := range steps {
		if s.Action.NeedsWork() {
			needs++
		}
	}
	if needs != 3 { // two skills plus the hook
		t.Errorf("check on an empty dir found %d item(s) needing work, want 3", needs)
	}

	install(t, dir, Options{})
	steps, err = claudeCode{}.Check(testAssets(), opts)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range steps {
		if s.Action.NeedsWork() {
			t.Errorf("check after install reports %q for %q", s.Action, s.Detail)
		}
	}
}

func TestLookupRejectsUnknownTarget(t *testing.T) {
	if _, err := Lookup("cursor"); err == nil {
		t.Error("unknown target accepted; it would get Claude Code's layout")
	}
	if got, err := Lookup(""); err != nil || got.Name() != Default {
		t.Errorf("Lookup(\"\") = %v, %v; want the default target", got, err)
	}
}

func TestDecodeObjectRoundTrip(t *testing.T) {
	obj, err := decodeObject([]byte(settingsFixture))
	if err != nil {
		t.Fatal(err)
	}
	out, err := obj.encode(0)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(out) {
		t.Fatalf("round trip produced invalid JSON:\n%s", out)
	}

	var before, after any
	if err := json.Unmarshal([]byte(settingsFixture), &before); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(out, &after); err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(before)
	a, _ := json.Marshal(after)
	if string(a) != string(b) {
		t.Errorf("round trip changed the document:\n got %s\nwant %s", a, b)
	}

	if _, err := decodeObject([]byte(`[1,2]`)); err == nil {
		t.Error("a JSON array was accepted as an object")
	}
	if obj, err := decodeObject(nil); err != nil || len(obj) != 0 {
		t.Errorf("decodeObject(nil) = %v, %v; want an empty object", obj, err)
	}
}

func TestIsEphemeralPath(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		// What `go run` produces; recording this in a hook would break
		// every session once the build cache is cleaned.
		{"/private/var/folders/m6/T/go-build3830492811/b001/exe/ispec", true},
		{"/tmp/go-build123/b001/exe/ispec", true},

		{"/Users/someone/go/bin/ispec", false},
		{"/usr/local/bin/ispec", false},
		{"/Users/someone/src/go-buildtools/bin/ispec", false},
	}
	for _, c := range cases {
		if got := isEphemeralPath(c.path); got != c.want {
			t.Errorf("isEphemeralPath(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

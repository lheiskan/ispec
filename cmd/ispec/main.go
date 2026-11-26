// Command ispec maintains and checks a project's ispec canon, and installs
// ispec's workflow skills into an LLM agent harness.
//
// Usage:
//
//	ispec lint    [-root DIR] [-canon-dir NAME]
//	ispec status  [-root DIR] [-canon-dir NAME]
//	ispec harness install|check|uninstall|session-start [flags]
//
// Day-to-day canon work happens through the agent, which reads and writes the
// cards as ordinary markdown files. This binary covers the two things file
// access cannot do: validating the reference graph, and installing itself.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	ispec "github.com/lheiskan/ispec"
	"github.com/lheiskan/ispec/internal/harness"
	"github.com/lheiskan/ispec/internal/lint"
	"github.com/lheiskan/ispec/internal/status"
)

// exit codes. 2 is reserved for "no canon here", distinct from 1 for "the
// canon is broken", so a script can tell an unbootstrapped project from a
// failing one.
const (
	exitOK       = 0
	exitProblems = 1
	exitNoCanon  = 2
)

func main() {
	os.Exit(run())
}

// run dispatches one invocation and returns the process exit code.
//
// Separated from main so that os.Exit is called exactly once, at a point
// where every deferred function has already run. Scattered os.Exit calls skip
// defers silently, which is a trap for any cleanup added later. It also means
// every exit path is a single `return` naming the code it produces, which
// matters because those codes are documented surface that scripts gate on.
func run() int {
	if len(os.Args) < 2 {
		usage(os.Stderr)
		return exitProblems
	}

	sub, args := os.Args[1], os.Args[2:]

	var err error
	switch sub {
	case "lint":
		err = cmdLint(args)
	case "status":
		err = cmdStatus(args)
	case "harness":
		err = cmdHarness(args)
	case "help", "-h", "--help":
		usage(os.Stdout)
		return exitOK
	default:
		fmt.Fprintf(os.Stderr, "ispec: unknown command %q\n\n", sub)
		usage(os.Stderr)
		return exitProblems
	}

	if err == nil {
		return exitOK
	}

	var ec exitError
	if errors.As(err, &ec) {
		if ec.msg != "" {
			fmt.Fprintf(os.Stderr, "ispec: %s\n", ec.msg)
		}
		return ec.code
	}
	fmt.Fprintf(os.Stderr, "ispec: %v\n", err)
	return exitProblems
}

// exitError carries a specific exit code out of a subcommand.
type exitError struct {
	code int
	msg  string
}

func (e exitError) Error() string {
	if e.msg == "" {
		return fmt.Sprintf("exit %d", e.code)
	}
	return e.msg
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "ispec — requirements, specs and decisions that stay aligned with the code")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  ispec lint     [-root DIR] [-canon-dir NAME]   Validate the canon's schema and references")
	fmt.Fprintln(w, "  ispec status   [-root DIR] [-canon-dir NAME]   Report open sessions, deferred work, lint state")
	fmt.Fprintln(w, "  ispec harness  <command>                       Install ispec into an agent harness")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Run `ispec harness help` for the harness commands.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Exit codes: 0 clean, 1 problems found, 2 no canon directory.")
}

// canonFlags registers the two flags every canon-reading command takes.
func canonFlags(fs *flag.FlagSet) (root, canonDir *string) {
	root = fs.String("root", ".", "project root containing the canon directory")
	canonDir = fs.String("canon-dir", lint.DefaultCanonDir, "canon directory name under root")
	return
}

// noCanon converts a missing-canon error into the dedicated exit code.
func noCanon(err error) error {
	if errors.Is(err, lint.ErrNoCanonDir) {
		return exitError{code: exitNoCanon, msg: err.Error()}
	}
	return err
}

func cmdLint(args []string) error {
	fs := flag.NewFlagSet("lint", flag.ExitOnError)
	root, canonDir := canonFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}

	res, err := lint.Check(*root, *canonDir)
	if err != nil {
		return noCanon(err)
	}

	for _, i := range res.Issues {
		rel, relErr := filepath.Rel(*root, i.Path)
		if relErr != nil {
			rel = i.Path
		}
		fmt.Printf("%s: [%s] %s\n", rel, i.Severity, i.Msg)
	}
	fmt.Printf("\n%d card(s) checked. %d error(s), %d warning(s).\n",
		len(res.Cards), res.Errors, res.Warnings)

	if res.Errors > 0 {
		return exitError{code: exitProblems}
	}
	return nil
}

func cmdStatus(args []string) error {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	root, canonDir := canonFlags(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}

	st, err := status.Compute(*root, *canonDir)
	if err != nil {
		return noCanon(err)
	}
	st.Write(os.Stdout)

	if !st.Clean() {
		return exitError{code: exitProblems}
	}
	return nil
}

func harnessUsage(w io.Writer) {
	fmt.Fprintln(w, "ispec harness — install ispec into an LLM agent harness")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  install         Write the workflow skills and the session-start hook. Idempotent.")
	fmt.Fprintln(w, "  check           Report what is installed and whether it matches this binary.")
	fmt.Fprintln(w, "  uninstall       Remove only what install wrote.")
	fmt.Fprintln(w, "  session-start   Print the session-start directive. What the installed hook runs.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  --target NAME   Agent harness to act on (default: "+harness.Default+")")
	fmt.Fprintln(w, "  --dir PATH      Config directory of the target harness, overriding its default")
	fmt.Fprintln(w, "  --dry-run       Report what would change; write nothing")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Supported harnesses:")
	for _, t := range harness.Targets() {
		fmt.Fprintf(w, "  %-14s %s\n", t.Name(), t.Describe())
	}
}

func cmdHarness(args []string) error {
	if len(args) == 0 {
		harnessUsage(os.Stderr)
		return exitError{code: exitProblems}
	}
	sub := args[0]
	if sub == "help" || sub == "-h" || sub == "--help" {
		harnessUsage(os.Stdout)
		return nil
	}

	fs := flag.NewFlagSet("harness "+sub, flag.ExitOnError)
	target := fs.String("target", harness.Default, "agent harness to act on")
	dir := fs.String("dir", "", "config directory of the target harness")
	dryRun := fs.Bool("dry-run", false, "report what would change; write nothing")
	// session-start reads a canon rather than writing a harness, so it takes
	// the canon flags instead. Registering them for every harness subcommand
	// keeps one flag set rather than branching before parsing.
	root, canonDir := canonFlags(fs)
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	if sub == "session-start" {
		harness.SessionStart(os.Stdout, *root, *canonDir)
		return nil
	}

	t, err := harness.Lookup(*target)
	if err != nil {
		return err
	}
	opts := harness.Options{Dir: *dir, DryRun: *dryRun}

	var steps []harness.Step
	switch sub {
	case "install":
		steps, err = t.Install(&ispec.Skills, opts)
	case "check":
		steps, err = t.Check(&ispec.Skills, opts)
	case "uninstall":
		steps, err = t.Uninstall(&ispec.Skills, opts)
	default:
		fmt.Fprintf(os.Stderr, "ispec harness: unknown command %q\n\n", sub)
		harnessUsage(os.Stderr)
		return exitError{code: exitProblems}
	}
	if err != nil {
		return err
	}

	where := opts.Dir
	if where == "" {
		if where, err = t.DefaultDir(); err != nil {
			return err
		}
	}
	fmt.Printf("%s (%s)\n", t.Name(), where)

	needsWork := 0
	for _, s := range steps {
		fmt.Printf("  %-12s %s\n", s.Action, s.Detail)
		if s.Action.NeedsWork() {
			needsWork++
		}
	}

	// `check` is meant to be usable as a gate, so it reports through its exit
	// code as well as its output.
	if sub == "check" && needsWork > 0 {
		fmt.Printf("\n%d item(s) need installing. Run `ispec harness install`.\n", needsWork)
		return exitError{code: exitProblems}
	}
	return nil
}

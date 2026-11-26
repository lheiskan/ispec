package harness

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/lheiskan/ispec/internal/lint"
	"github.com/lheiskan/ispec/internal/status"
)

// SessionStart writes the directive an agent is given at the start of a
// session in the project rooted at root.
//
// It prints nothing when the project has no canon. The hook is installed
// user-globally and therefore fires in every session in every repository,
// the overwhelming majority of which do not use ispec. A tool that announced
// itself in all of them would be noise that trains the user to ignore it —
// and worse, would push an unasked-for workflow onto unrelated work. Silence
// is the correct output for "not an ispec project".
//
// It returns no error for a broken canon either. A hook that fails loudly at
// session start degrades the session it was supposed to help; a canon with
// lint errors is reported as a fact for the agent to act on.
func SessionStart(w io.Writer, root, canonDir string) {
	st, err := status.Compute(root, canonDir)
	if err != nil {
		if errors.Is(err, lint.ErrNoCanonDir) {
			return
		}
		fmt.Fprintf(w, "## ispec\n\nCanon at %s could not be read: %v\n", lint.CanonPath(root, canonDir), err)
		return
	}

	var b strings.Builder
	b.WriteString("## ispec canon active\n\n")
	st.Write(&b)
	b.WriteString("\n")
	b.WriteString(guidance(st))
	fmt.Fprint(w, b.String())
}

// guidance turns the status into the one or two things that actually change
// what the agent should do next. Ordered most-blocking first.
func guidance(st status.Status) string {
	var b strings.Builder

	if n := len(st.OpenSessions); n > 0 {
		fmt.Fprintf(&b, "**%d session(s) are still open** (%s). Before starting new work, read the open session card(s) and establish whether this request continues that work or is a new unit. Do not silently open a second session on top of an unfinished one; if the open session is in fact finished, run `ispec-close-session` on it first.\n\n",
			n, strings.Join(st.OpenSessions, ", "))
	}

	if !st.Clean() {
		fmt.Fprintf(&b, "**The canon does not lint clean** (%d error(s)). Run `ispec lint` and fix the canon before adding to it, or the new cards inherit a broken reference graph.\n\n",
			st.Errors)
	}

	b.WriteString("Invoke the `ispec-start-session` skill at the start of any substantive change here — it classifies the work, agrees a contract with the user before coding, and reconciles code against canon afterwards. Skip it only for chores that touch nothing in canon.\n")
	return b.String()
}

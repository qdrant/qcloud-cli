package output

import (
	"io"
	"os"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/term"
)

// ColorEnabled reports whether w is a terminal that supports colors, honoring
// NO_COLOR and TERM=dumb.
func ColorEnabled(w io.Writer) bool {
	return colorprofile.Detect(w, os.Environ()) >= colorprofile.ANSI
}

// TerminalWidth returns the width of the terminal w writes to, or fallback
// when w is not a terminal.
func TerminalWidth(w io.Writer, fallback int) int {
	f, ok := w.(*os.File)
	if !ok || !term.IsTerminal(f.Fd()) {
		return fallback
	}

	width, _, err := term.GetSize(f.Fd())
	if err != nil || width <= 0 {
		return fallback
	}

	return width
}

package textcell

import "github.com/gdamore/tcell/v2"

// Option is a function that modifies e at initialization.
type Option func(e *Editor)

// WithString returns an Option that initializes e with s.
// Each LF in s will create a new line.
func WithString(s string) Option {
	return func(e *Editor) {
		e.lines = e.lines[:0]
		e.lines = append(e.lines, newLine())
		ln := e.lines[0]
		for _, char := range []rune(s) {
			if char == '\n' {
				e.lines = append(e.lines, newLine())
				ln = e.lines[len(e.lines)-1]
				continue
			}
			ln.buf = append(ln.buf, char)
		}
	}
}

// WithStyle returns an Option that initializes e with default style s.
// Default is tcell.StyleDefault.
func WithStyle(s tcell.Style) Option {
	return func(e *Editor) {
		e.styleDefault = s
	}
}

// WithFocus returns an Option that initializes e with focus.
func WithFocus() Option {
	return func(e *Editor) {
		e.Focus()
	}
}

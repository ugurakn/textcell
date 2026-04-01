package textcell

import "github.com/gdamore/tcell/v2"

// Option is a function that modifies e at initialization.
type Option func(e *Editor)

// WithString returns an Option that initializes [Editor] with s.
// Each LF in s will create a new line.
func WithString(s string) Option {
	return func(e *Editor) {
		e.lines = e.lines[:0]
		e.lines = append(e.lines, newLine(e.maxWidth))
		ln := e.lines[0]
		for _, char := range []rune(s) {
			if char == '\n' {
				e.lines = append(e.lines, newLine(e.maxWidth))
				ln = e.lines[len(e.lines)-1]
				continue
			}
			ln.buf = append(ln.buf, char)
		}
	}
}

// WithStyle returns an Option that initializes [Editor] with default style s.
// Default is tcell.StyleDefault.
func WithStyle(s tcell.Style) Option {
	return func(e *Editor) {
		e.styleDefault = s
	}
}

// WithFocus returns an Option that initializes [Editor] with focus.
func WithFocus() Option {
	return func(e *Editor) {
		e.Focus()
	}
}

// WithNoOpts returns an option that removes all options (including itself) from [Editor].
func WithNoOpts() Option {
	return func(e *Editor) {
		e.opts = e.opts[:0]
	}
}

// WithMaxW returns an option that sets mw as [Editor] max width.
func WithMaxW(mw int) Option {
	return func(e *Editor) {
		e.maxWidth = mw
	}
}

// WithMaxH returns an option that sets mh as [Editor] max height.
func WithMaxH(mh int) Option {
	return func(e *Editor) {
		e.maxHeight = mh
	}
}

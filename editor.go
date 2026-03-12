package main

import "github.com/gdamore/tcell/v2"

const (
	maxCharsOnLine = 64
)

// Line represents a single line of text.
type Line struct {
	buf    []rune
	length int
	y      int
}

func NewLine() *Line {
	l := new(Line)
	l.buf = make([]rune, maxCharsOnLine) // can hold 64 chars by default
	l.length = 0
	l.y = 0
	return l
}

// WriteChar adds char to line buffer at cursor position cx.
// cx is assumed to be a legal position.
func (ln *Line) WriteChar(char rune, cx int) {
	// // TODO bounds checking (do not grow buf)

	// if cx points at the end of buf, just append.
	// otherwise, insert.
	if cx == ln.length {
		ln.buf[cx] = char
	} else {
		ln.buf = append(ln.buf[:cx], append([]rune{char}, ln.buf[cx:]...)...)
	}
	ln.length++
}

// Backspace deletes the char that is to the left of cursor position cx.
func (ln *Line) Backspace(cx int) bool {
	if ln.length == 0 || cx == 0 {
		return false
	}

	ln.buf = append(ln.buf[:cx-1], ln.buf[cx:]...)
	ln.length--
	return true
}

func (ln *Line) Show(screen tcell.Screen) {
	for i := range ln.length {
		screen.SetContent(
			i,
			ln.y,
			ln.buf[i],
			nil,
			tcell.StyleDefault,
		)
	}
}

// // Cursor
type Cursor struct {
	x, y int
}

func NewCursor(x, y int) *Cursor {
	return &Cursor{x, y}
}

func (c *Cursor) Right(lnLen int) {
	// cursor can't move into 'empty' area to the right of its line
	if c.x == lnLen {
		return
	}
	c.x++
}

func (c *Cursor) Left() {
	if c.x <= 0 {
		return
	}
	c.x--
}

func (c *Cursor) Show(screen tcell.Screen) {
	screen.ShowCursor(c.x, c.y)
}

package main

import "github.com/gdamore/tcell/v2"

const (
	maxCharsOnLine = 64
)

// Line represents a single line of text.
type Line struct {
	buf []rune
	y   int
}

func NewLine() *Line {
	l := new(Line)
	l.buf = make([]rune, 0, maxCharsOnLine) // can hold 64 chars by default
	l.y = 0
	return l
}

func (ln *Line) WriteChar(char rune) {
	// TODO bounds checking (do not grow buf)
	ln.buf = append(ln.buf, char)
}

// TODO must accept a cursor position when cursor can move.
func (ln *Line) Backspace() bool {
	if len(ln.buf) == 0 {
		return false
	}
	ln.buf = ln.buf[:len(ln.buf)-1]
	return true
}

func (ln *Line) Show(screen tcell.Screen) {
	for i := range len(ln.buf) {
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

func (c *Cursor) Right() {
	// if c.x < maxCharsOnLine-1 {
	// 	c.x++
	// }
	c.x++
}

func (c *Cursor) Left() {
	if c.x > 0 {
		c.x--
	}
}

func (c *Cursor) Show(screen tcell.Screen) {
	screen.ShowCursor(c.x, c.y)
}

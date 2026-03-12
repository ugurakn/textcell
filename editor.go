package main

import "github.com/gdamore/tcell/v2"

// Line represents a single line of text.
type Line struct {
	buf []rune
	// x is the current cursor position.
	x, y int
}

func NewLine() *Line {
	l := new(Line)
	l.buf = make([]rune, 0, 64) // can hold 64 chars by default
	l.x = 0
	l.y = 0
	return l
}

func (ln *Line) WriteChar(char rune) {
	// TODO bounds checking (do not grow buf)
	ln.buf = append(ln.buf, char)
	ln.x++
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

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

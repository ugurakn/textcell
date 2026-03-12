package textcell

import (
	"strings"

	"github.com/gdamore/tcell/v2"
)

const (
	MaxCharsOnLine = 32
)

type Editor struct {
	lines    []*line
	screen   tcell.Screen
	cursor   *cursor
	x, y     int
	scrollX  int
	hasFocus bool
}

// NewEditor creates and returns a new Editor without focus.
func NewEditor(baseX, baseY int, screen tcell.Screen) *Editor {
	e := new(Editor)
	e.cursor = newCursor()
	e.lines = make([]*line, 0, 32)
	e.lines = append(e.lines, newLine())
	e.x, e.y = baseX, baseY
	e.scrollX = 0
	e.hasFocus = false
	e.screen = screen
	return e
}

// // Editor: Public API

// Reset resets editor to initial state.
func (e *Editor) Reset() {
	e = NewEditor(e.x, e.y, e.screen)
}

// String returns the text in all the lines separated by sep.
func (e *Editor) String(sep rune) string {
	var b strings.Builder
	for i, ln := range e.lines {
		b.WriteString(string(ln.buf))
		if i < len(e.lines)-1 {
			b.WriteRune(sep)
		}
	}
	return b.String()
}

// Show puts text and cursor on screen.
// Cursor will be showed only if editor has focus.
func (e *Editor) Show() {
	e.ShowText()
	if e.hasFocus {
		e.ShowCursor()
	}
}

// ProcessEvent handles keypress events.
// Do not call this method if e is unfocused.
func (e *Editor) ProcessEvent(ev tcell.Event) {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyRune:
			e.WriteChar(ev.Rune())
		case tcell.KeyRight:
			e.CurRight()
		case tcell.KeyLeft:
			e.CurLeft()
		case tcell.KeyDown:
			e.CurDown()
		case tcell.KeyUp:
			e.CurUp()
		case tcell.KeyBackspace:
			e.Backspace()
		case tcell.KeyEnter:
			e.NewLine()
		case tcell.KeyESC:
			e.Unfocus()
		}
	}
}

// Focus focuses e and shows cursor.
func (e *Editor) Focus() {
	e.hasFocus = true
}

// Unfocus unfocuses e and hides cursor.
// Pressing ESC will call Unfocus.
func (e *Editor) Unfocus() {
	e.hasFocus = false
	e.screen.HideCursor()
}

func (e *Editor) HasFocus() bool {
	return e.hasFocus
}

// // Editor: cursor methods

func (e *Editor) CurRight() {
	c := e.cursor
	if c.x == e.currentLine().len() {
		// move to next line if it exists
		if c.y == len(e.lines)-1 {
			return
		}
		e.setCurCol(0)
		c.y++
		return
	}
	// incr scrollX if cursor moved beyond MaxCharOnLine limit
	e.setCurCol(c.x + 1)
}

func (e *Editor) CurLeft() {
	c := e.cursor
	if c.x == 0 {
		// move to prev line if it exists
		if c.y == 0 {
			return
		}
		e.setCurCol(e.prevLine().len())
		c.y--
		return
	}
	e.setCurCol(c.x - 1)
}

func (e *Editor) CurDown() {
	if e.cursor.y == len(e.lines)-1 {
		return
	}
	e.cursor.down(e.lines[e.cursor.y+1].len())
	// TEMP because down modifies cursor.x:
	e.calcScrollX()
}

func (e *Editor) CurUp() {
	if e.cursor.y == 0 {
		return
	}
	e.cursor.up(e.lines[e.cursor.y-1].len())
	// TEMP because up modifies cursor.x:
	e.calcScrollX()
}

func (e *Editor) setCurCol(x int) {
	e.cursor.setCol(x)
	e.calcScrollX()
}

func (e *Editor) ShowCursor() {
	e.cursor.show(e.x-e.scrollX, e.y, e.screen)
}

// // Editor: line methods

// NewLine creates a new line under cursor y
// and moves cursor to new line.
func (e *Editor) NewLine() {
	newLn := newLine()
	if len(e.lines)-1 == e.cursor.y {
		e.lines = append(e.lines, newLn)
	} else {
		e.lines = append(
			e.lines[:e.cursor.y+1],
			append([]*line{newLn}, e.lines[e.cursor.y+1:]...)...,
		)
	}
	// carry text to right of cursor to the new line
	curLine := e.currentLine()
	newLn.buf = append(newLn.buf, curLine.buf[e.cursor.x:]...)
	curLine.buf = curLine.buf[:e.cursor.x]
	// reposition cursor to start of new line
	e.setCurCol(0)
	e.cursor.y++
}

// WriteChar appends char to current line and moves cursor right.
func (e *Editor) WriteChar(char rune) {
	e.currentLine().writeChar(char, e.cursor.x)
	e.CurRight()
}

func (e *Editor) Backspace() {
	if e.cursor.x == 0 {
		if e.cursor.y == 0 {
			return
		}
		buf := e.removeLine()
		e.cursor.y--
		e.setCurCol(e.currentLine().len())
		if len(buf) > 0 {
			e.currentLine().append(buf)
		}
		return
	}
	e.currentLine().backspace(e.cursor.x)
	e.CurLeft()
}

func (e *Editor) ShowText() {
	for i := range e.lines {
		e.lines[i].show(e.x, e.y+i, e.scrollX, e.screen)
	}
}

// // Editor: helper methods

// currentLine returns the line the cursor is on.
func (e *Editor) currentLine() *line {
	return e.lines[e.cursor.y]
}

// prevLine returns the line above the current line
// or nil if cursor is already on the top line.
func (e *Editor) prevLine() *line {
	if e.cursor.y == 0 {
		return nil
	}
	return e.lines[e.cursor.y-1]
}

// removeLine removes the current line and returns its buf.
func (e *Editor) removeLine() []rune {
	buf := e.currentLine().buf
	// if current line is the last line, reslice.
	if e.cursor.y == len(e.lines)-1 {
		e.lines = e.lines[:len(e.lines)-1]
	} else {
		e.lines = append(e.lines[:e.cursor.y], e.lines[e.cursor.y+1:]...)
	}
	return buf
}

// calcScrollX calculates horizontal scroll position.
// It should be called only after cursor movement to right on a line.
func (e *Editor) calcScrollX() {
	e.scrollX = max(0, e.cursor.x-MaxCharsOnLine+1)
}

// line represents a single line of text.
type line struct {
	buf []rune
}

func newLine() *line {
	ln := new(line)
	ln.buf = make([]rune, 0, MaxCharsOnLine)
	return ln
}

// writeChar adds char to line buffer at cursor position cx.
// cx is assumed to be a legal position.
func (ln *line) writeChar(char rune, cx int) {
	// // TODO bounds checking (do not grow buf)

	// if cx points at the end of buf, just append.
	// otherwise, insert.
	if cx == ln.len() {
		ln.buf = append(ln.buf, char)
	} else {
		ln.buf = append(ln.buf[:cx], append([]rune{char}, ln.buf[cx:]...)...)
	}
}

// backspace deletes the char that is to the left of cursor position cx.
// caller must guarantee cx > 0.
func (ln *line) backspace(cx int) {
	ln.buf = append(ln.buf[:cx-1], ln.buf[cx:]...)
}

// append appends the contents of buf to its own buffer.
func (ln *line) append(buf []rune) {
	ln.buf = append(ln.buf, buf...)
}

func (ln *line) show(baseX, baseY int, scrollX int, screen tcell.Screen) {
	for i, char := range ln.buf[min(ln.len(), scrollX):min(ln.len(), scrollX+MaxCharsOnLine)] {
		screen.SetContent(
			baseX+i,
			baseY,
			char,
			nil,
			tcell.StyleDefault,
		)
	}
}

// len returns the length of line buffer.
func (ln *line) len() int {
	return len(ln.buf)
}

// // cursor
type cursor struct {
	x, y    int
	goalCol int
}

func newCursor() *cursor {
	return &cursor{
		x:       0,
		y:       0,
		goalCol: 0,
	}
}

func (c *cursor) down(lnLen int) {
	c.y++
	c.x = min(c.goalCol, lnLen)
}

func (c *cursor) up(lnLen int) {
	c.y--
	c.x = min(c.goalCol, lnLen)
}

// setCol sets cursor col and goalCol to x.
func (c *cursor) setCol(x int) {
	c.x = x
	c.goalCol = x
}

func (c *cursor) show(xOffset, yOffset int, screen tcell.Screen) {
	screen.ShowCursor(c.x+xOffset, c.y+yOffset)
}

package textcell

import "github.com/gdamore/tcell/v2"

const (
	MaxCharsOnLine = 32
)

type Editor struct {
	lines  []*line
	cursor *cursor
	x, y   int
}

func NewEditor(baseX, baseY int) *Editor {
	e := new(Editor)
	e.cursor = newCursor(baseX, baseY)
	e.lines = make([]*line, 0, 32)
	e.lines = append(e.lines, newLine(baseX, baseY))
	e.x, e.y = baseX, baseY
	return e
}

// // Editor: cursor methods

func (e *Editor) CurRight() {
	e.cursor.right(e.currentLine().length)
}

func (e *Editor) CurLeft() {
	e.cursor.left()
}

// TODO
func (e *Editor) CurDown() {
	e.cursor.down(len(e.lines))
}

func (e *Editor) CurUp() {
	e.cursor.up()
}

func (e *Editor) ShowCursor(screen tcell.Screen) {
	e.cursor.show(screen)
}

// // Editor: line methods

// NewLine creates a new line under cursor y
// and moves cursor to new line.
func (e *Editor) NewLine() {
	newLn := newLine(e.x, e.y+e.cursor.y+1)
	if len(e.lines)-1 == e.cursor.y {
		e.lines = append(e.lines, newLn)
	} else {
		e.lines = append(
			e.lines[:e.cursor.y+1],
			append([]*line{newLn}, e.lines[e.cursor.y+1:]...)...,
		)
		// update y for lines after new line
		for _, ln := range e.lines[e.cursor.y+2:] {
			ln.baseY++
		}
	}

	// reposition cursor to start of new line
	e.cursor.x = 0
	e.cursor.y++
}

func (e *Editor) WriteChar(char rune) {
	e.currentLine().writeChar(char, e.cursor.x)
	e.CurRight()
}
func (e *Editor) Backspace() {
	if ok := e.currentLine().backspace(e.cursor.x); ok {
		e.CurLeft()
	}
}
func (e *Editor) ShowText(screen tcell.Screen) {
	for i := range e.lines {
		e.lines[i].show(screen)
	}
}

// // Editor: helper methods

// currentLine returns the line the cursor is on.
func (e *Editor) currentLine() *line {
	return e.lines[e.cursor.y]
}

// line represents a single line of text.
type line struct {
	buf          []rune
	length       int
	baseX, baseY int
}

func newLine(baseX, baseY int) *line {
	ln := new(line)
	ln.buf = make([]rune, 0, MaxCharsOnLine)
	ln.length = 0
	ln.baseX = baseX
	ln.baseY = baseY
	return ln
}

// writeChar adds char to line buffer at cursor position cx.
// cx is assumed to be a legal position.
func (ln *line) writeChar(char rune, cx int) {
	// // TODO bounds checking (do not grow buf)

	// if cx points at the end of buf, just append.
	// otherwise, insert.
	if cx == ln.length {
		ln.buf = append(ln.buf, char)
	} else {
		ln.buf = append(ln.buf[:cx], append([]rune{char}, ln.buf[cx:]...)...)
	}
	ln.length++
}

// backspace deletes the char that is to the left of cursor position cx.
func (ln *line) backspace(cx int) bool {
	if ln.length == 0 || cx == 0 {
		return false
	}

	ln.buf = append(ln.buf[:cx-1], ln.buf[cx:]...)
	ln.length--
	return true
}

func (ln *line) show(screen tcell.Screen) {
	for i := range ln.length {
		screen.SetContent(
			ln.baseX+i,
			ln.baseY,
			ln.buf[i],
			nil,
			tcell.StyleDefault,
		)
	}
}

// // cursor
type cursor struct {
	baseX, baseY int
	x, y         int
}

func newCursor(x, y int) *cursor {
	return &cursor{
		baseX: x,
		baseY: y,
		x:     0,
		y:     0,
	}
}

func (c *cursor) right(lnLen int) {
	// cursor can't move into 'empty' area to the right of its line
	if c.x == lnLen {
		return
	}
	c.x++
}

func (c *cursor) left() {
	if c.x <= 0 {
		return
	}
	c.x--
}

func (c *cursor) down(numLines int) {
	if c.y == numLines-1 {
		return
	}
	c.y++
	// TEMP move cursor to line start
	c.x = 0
}

func (c *cursor) up() {
	if c.y == 0 {
		return
	}
	c.y--
	// TEMP move cursor to line start
	c.x = 0
}

func (c *cursor) show(screen tcell.Screen) {
	screen.ShowCursor(c.x+c.baseX, c.y+c.baseY)
}

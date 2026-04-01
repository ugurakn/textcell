package textcell

import (
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

// charPos represents a character at a given x(=col) and y(=line) position.
type charPos struct {
	ln, col int
	char    rune
}

type Editor struct {
	styleDefault tcell.Style
	// styleHighlight is used for highlighting selected text.
	// its fg and bg colors are reversed from styleDefault.
	styleHighlight tcell.Style
	opts           []Option
	lines          []*line
	screen         tcell.Screen
	cursor         *cursor
	selected       *selectedText
	// maxWidth and maxHeight are the maximum number of visible characters on a line
	// and maximum number of visible lines on screen, respectively.
	maxWidth, maxHeight int
	x, y                int
	scrollX, scrollY    int
	hasFocus            bool
}

// NewEditor creates and returns a new Editor without focus.
func NewEditor(baseX, baseY, maxWidth, maxHeight int, screen tcell.Screen, opts ...Option) *Editor {
	if maxWidth <= 0 || maxHeight <= 0 {
		panic("maxWidth and maxHeight must be > 0.")
	}
	e := new(Editor)
	e.screen = screen
	e.x, e.y = baseX, baseY
	e.maxWidth, e.maxHeight = maxWidth, maxHeight
	e.setInitState()

	e.hasFocus = false
	e.styleDefault = tcell.StyleDefault
	e.styleHighlight = e.styleDefault.Reverse(true)
	e.opts = opts
	e.applyOpts()
	return e
}

// // Editor: Public API

// Reset resets e to its initial state.
// If any opts are provided, they override current ones.
// If no opts are provided, old opts are applied.
// Use Option WithNoOpts to remove all options.
func (e *Editor) Reset(opts ...Option) {
	e.setInitState()
	if len(opts) > 0 {
		e.opts = opts
	}
	e.applyOpts()
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

func (e *Editor) ShowText() {
	start := e.scrollY
	end := min(len(e.lines), e.scrollY+e.maxHeight)
	for i, ln := range e.lines[start:end] {
		ln.show(e.x, e.y+i, e.scrollX, e.maxWidth, e.styleDefault, e.screen)
	}

	if e.selected != nil {
		e.highlightSelected(start, end)
	}
}

// highlightSelected highlights the selected portion of on-screen text.
func (e *Editor) highlightSelected(fVisibleLn, lVisibleLn int) {
	var ln *line
	var offsetX, offsetY int
	for _, sla := range e.selected.lines {
		if sla.y < fVisibleLn {
			continue
		}
		if sla.y >= lVisibleLn {
			break
		}
		ln = e.lines[sla.y]
		offsetX = max(0, sla.start-ln.fVisible)
		offsetY = sla.y - fVisibleLn
		for i := sla.start; i < sla.end; i++ {
			if i < ln.fVisible {
				continue
			}
			if i >= ln.lVisible {
				break
			}
			e.screen.SetContent(
				e.x+offsetX,
				e.y+offsetY,
				ln.buf[i],
				nil,
				e.styleHighlight,
			)
			offsetX++
		}
	}
}

// Show puts text and cursor on screen.
// Cursor will be showed only if editor has focus.
func (e *Editor) Show() {
	e.ShowText()
	if e.hasFocus {
		e.ShowCursor()
	}
}

func (e *Editor) ShowCursor() {
	e.cursor.show(e.x-e.scrollX, e.y-e.scrollY, e.screen)
}

// ProcessEvent handles keypress events.
// Do not call this method if e is unfocused.
func (e *Editor) ProcessEvent(ev tcell.Event) {
	modCtrl := false
	modShift := false

	switch ev := ev.(type) {
	case *tcell.EventKey:
		// modShift
		if ev.Modifiers()&tcell.ModShift != 0 {
			modShift = true
			if e.selected == nil {
				e.startSelected()
			}
			// defer until after cursor movement
			defer e.setSelected()
		}
		// modCtrl
		if ev.Modifiers()&tcell.ModCtrl != 0 {
			modCtrl = true
		}

		switch ev.Key() {
		case tcell.KeyRune:
			e.WriteChar(ev.Rune())
		case tcell.KeyRight:
			if e.selected != nil && !modShift {
				e.clearSelected()
			}
			if modCtrl {
				e.CurRightWord()
				break
			}
			e.CurRight()
		case tcell.KeyLeft:
			if e.selected != nil && !modShift {
				e.clearSelected()
			}
			if modCtrl {
				e.CurLeftWord()
				break
			}
			e.CurLeft()
		case tcell.KeyDown:
			if e.selected != nil && !modShift {
				e.clearSelected()
			}
			e.CurDown()
		case tcell.KeyUp:
			if e.selected != nil && !modShift {
				e.clearSelected()
			}
			e.CurUp()
		case tcell.KeyBackspace:
			if e.selected != nil {
				e.backspaceSelected()
				break
			}
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
		e.setCurLine(c.y + 1)
		return
	}
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
		e.setCurLine(c.y - 1)
		return
	}
	e.setCurCol(c.x - 1)
}

// CurRightWord moves cursor to right by a whole word.
// A word is defined as any contiguous sequence of letters and digits.
func (e *Editor) CurRightWord() {
	cp := e.MoveCurByWord(e.nextCharPos)
	e.setCurCol(cp.col)
	e.setCurLine(cp.ln)
}

// CurLeftWord moves cursor to left by a whole word.
// A word is defined as any contiguous sequence of letters and digits.
func (e *Editor) CurLeftWord() {
	cp := e.MoveCurByWord(e.prevCharPos)
	// move onto first char of word if possible
	if cp.char != 0 {
		e.nextCharPos(cp)
	}
	e.setCurCol(cp.col)
	e.setCurLine(cp.ln)
}

// MoveCurByWord returns charPos for cursor movement by a whole word.
//
// (right: word_ left: _word if possible to move beyond word)
func (e *Editor) MoveCurByWord(getCharPos func(cp *charPos)) *charPos {
	cp := &charPos{ln: e.cursor.y, col: e.cursor.x}
	processChar := func() bool {
		getCharPos(cp)
		if cp.char == 0 {
			return true
		}
		return false
	}

	// find next word if not on one
	for {
		if done := processChar(); done {
			return cp
		}
		if isAlpha(cp.char) {
			break
		}
	}

	// move past word
	for {
		if done := processChar(); done {
			return cp
		}
		if !isAlpha(cp.char) {
			break
		}
	}
	return cp
}

func (e *Editor) CurDown() {
	if e.cursor.y == len(e.lines)-1 {
		e.setCurCol(e.currentLine().len())
		return
	}
	e.setCurLine(e.cursor.y + 1)
	e.setCurCol_noModifGC(min(e.cursor.goalCol, e.currentLine().len()))
}

func (e *Editor) CurUp() {
	if e.cursor.y == 0 {
		e.setCurCol(0)
		return
	}
	e.setCurLine(e.cursor.y - 1)
	e.setCurCol_noModifGC(min(e.cursor.goalCol, e.currentLine().len()))
}

// setCurCol sets cursor col to x and modifies
// goal column and horizontal scroll accordingly.
func (e *Editor) setCurCol(x int) {
	e.cursor.x = x
	e.cursor.goalCol = x
	e.calcScrollX()
}

// setCurCol_noModifGC is the same as setCurCol except for
// that it does not modify cursor goal column.
func (e *Editor) setCurCol_noModifGC(x int) {
	e.cursor.x = x
	e.calcScrollX()
}

// setCurLine sets cursor line to y and modifies
// vertical scroll accordingly.
func (e *Editor) setCurLine(y int) {
	e.cursor.y = y
	e.calcScrollY()
}

// // Editor: line methods

// NewLine creates a new line under cursor y
// and moves cursor to new line.
func (e *Editor) NewLine() {
	newLn := newLine(e.maxWidth)
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
	e.setCurLine(e.cursor.y + 1)
}

// WriteChar appends char to current line and moves cursor right.
func (e *Editor) WriteChar(char rune) {
	e.currentLine().writeChar(char, e.cursor.x)
	e.CurRight()
}

func (e *Editor) Backspace() {
	if e.cursor.x == 0 {
		e.backspaceToPrevLn()
		return
	}
	e.currentLine().backspace(e.cursor.x-1, e.cursor.x)
	e.CurLeft()
}

// backspaceToPrevLn assumes cursor col is 0.
func (e *Editor) backspaceToPrevLn() {
	if e.cursor.y == 0 {
		return
	}
	buf := e.removeLine(e.cursor.y)
	e.setCurLine(e.cursor.y - 1)
	e.setCurCol(e.currentLine().len())
	if len(buf) > 0 {
		e.currentLine().append(buf)
	}
}

// // Editor: selectedText methods

func (e *Editor) startSelected() {
	e.selected = newSelectedText(e.cursor.x, e.cursor.y)
}

func (e *Editor) clearSelected() {
	e.selected = nil
}

// setSelected sets currently selected areas on every line.
// e.selected is assumed to have been already initialized with pivot point.
func (e *Editor) setSelected() {
	if e.selected == nil {
		return
	}
	cx, cy := e.cursor.x, e.cursor.y
	px, py := e.selected.pivotX, e.selected.pivotY
	sel := e.selected

	// reset selected
	sel.lines = sel.lines[:0]

	var lnCount int
	sel.dir, lnCount = sel.calcDirLnCount(cx, cy)

	var start, end, firstLnY int
	switch sel.dir {
	case 0: // cursor on pivot point, nothing selected
		return
	case 'R':
		start = px
		end = cx
		firstLnY = py
	case 'L':
		start = cx
		end = px
		firstLnY = cy
	}

	// first line
	sla := &selectedLnArea{start: start, end: e.lines[firstLnY].len(), y: firstLnY}
	sel.lines = append(sel.lines, sla)
	if lnCount == 1 {
		sla.end = end
		return
	}
	// last line
	slaLast := &selectedLnArea{start: 0, end: end, y: firstLnY + lnCount - 1}
	if lnCount == 2 {
		sel.lines = append(sel.lines, slaLast)
		return
	}
	//  middle line(s)
	for i := range lnCount - 2 {
		lnY := firstLnY + i + 1
		sla = &selectedLnArea{start: 0, end: e.lines[lnY].len(), y: lnY}
		sel.lines = append(sel.lines, sla)
	}

	sel.lines = append(sel.lines, slaLast)
}

// backspaceSelected deletes all selected text.
func (e *Editor) backspaceSelected() {
	sel := e.selected
	if sel.dir == 0 {
		return
	}

	// remove selected areas from line buffers
	for _, sla := range sel.lines {
		e.lines[sla.y].backspace(sla.start, sla.end)
	}

	// remove empty lines from editor
	var lnIdx int
	var buf []rune
	offset := 0
	for _, sla := range sel.lines[1:] {
		lnIdx = sla.y + offset
		buf = e.removeLine(lnIdx)
		offset--
	}

	// carry remaining text from last line to left boundary
	if len(buf) > 0 {
		e.lines[sel.lines[0].y].append(buf)
	}

	// reposition cursor to left boundary
	e.setCurCol(sel.lines[0].start)
	e.setCurLine(sel.lines[0].y)

	e.clearSelected()
}

// // Editor: helper methods

func (e *Editor) setInitState() {
	e.cursor = newCursor()
	e.lines = make([]*line, 0, e.maxHeight)
	e.lines = append(e.lines, newLine(e.maxWidth))
	e.scrollX, e.scrollY = 0, 0
	e.selected = nil
}

func (e *Editor) applyOpts() {
	for _, opt := range e.opts {
		opt(e)
	}
}

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

// removeLine removes the line at given idx and returns its buf.
func (e *Editor) removeLine(idx int) []rune {
	buf := e.lines[idx].buf
	// if current line is the last line, reslice.
	if idx == len(e.lines)-1 {
		e.lines = e.lines[:len(e.lines)-1]
	} else {
		e.lines = append(e.lines[:idx], e.lines[idx+1:]...)
	}
	return buf
}

// calcScrollX calculates horizontal scroll position.
// It should be called after every cursor col change.
func (e *Editor) calcScrollX() {
	e.scrollX = max(0, e.cursor.x-e.maxWidth+1)
}

// calcScrollY calculates vertical scroll position.
// It should be called after every cursor line change.
func (e *Editor) calcScrollY() {
	e.scrollY = max(0, e.cursor.y-e.maxHeight+1)
}

// nextCharPos modifies cp to represent the char after cp.
// if no next char exists, it will always return
// char=0 with ln and col unchanged.
func (e *Editor) nextCharPos(cp *charPos) {
	cp.col += 1
	if cp.col > e.lines[cp.ln].len() {
		if cp.ln == len(e.lines)-1 { // end of text
			cp.col--
			cp.char = 0
			return
		}
		// go to next line
		cp.ln++
		cp.col = 0
	}

	if cp.col == e.lines[cp.ln].len() {
		// cursor at end-of-line represented as whitespace
		cp.char = ' '
		return
	}

	cp.char = e.lines[cp.ln].buf[cp.col]
}

// prevCharPos modifies cp to represent the char before cp.
// if no prev char exists, it will always return
// char=0 with ln and col unchanged.
func (e *Editor) prevCharPos(cp *charPos) {
	if cp.col-1 < 0 {
		if cp.ln == 0 { // left boundary
			cp.char = 0
			return
		}
		// go to above line
		cp.ln--
		cp.col = e.lines[cp.ln].len()
		cp.char = ' '
		return
	}

	cp.col--
	cp.char = e.lines[cp.ln].buf[cp.col]
}

// line represents a single line of text.
type line struct {
	buf []rune
	// fVisible is the idx for first visible char in buf.
	// lVisible is the idx for last visible char in buf+1.
	// set by [line].show.
	fVisible, lVisible int
}

func newLine(initCap int) *line {
	ln := new(line)
	ln.buf = make([]rune, 0, initCap)
	return ln
}

// writeChar adds char to line buffer at cursor position cx.
// cx is assumed to be a legal position.
func (ln *line) writeChar(char rune, cx int) {
	// if cx points at the end of buf, just append.
	// otherwise, insert.
	if cx == ln.len() {
		ln.buf = append(ln.buf, char)
	} else {
		ln.buf = append(ln.buf[:cx], append([]rune{char}, ln.buf[cx:]...)...)
	}
}

// backspace deletes all chars in buf[start:end].
func (ln *line) backspace(start, end int) {
	ln.buf = append(ln.buf[:start], ln.buf[end:]...)
}

// append appends the contents of buf to its own buffer.
func (ln *line) append(buf []rune) {
	ln.buf = append(ln.buf, buf...)
}

func (ln *line) show(baseX, baseY int, scrollX int, maxWidth int, style tcell.Style, screen tcell.Screen) {
	ln.fVisible = min(ln.len(), scrollX)
	ln.lVisible = min(ln.len(), scrollX+maxWidth)
	for i, char := range ln.buf[ln.fVisible:ln.lVisible] {
		screen.SetContent(
			baseX+i,
			baseY,
			char,
			nil,
			style,
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

func (c *cursor) show(xOffset, yOffset int, screen tcell.Screen) {
	screen.ShowCursor(c.x+xOffset, c.y+yOffset)
}

// // Utils

func isAlpha(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

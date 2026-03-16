package textcell

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type Debug struct {
	e      *Editor
	screen tcell.Screen
	x, y   int
}

func NewDebug(baseX, baseY int, e *Editor, screen tcell.Screen) *Debug {
	d := new(Debug)
	d.x, d.y = baseX, baseY
	d.e = e
	d.screen = screen
	return d
}

func (d *Debug) ShowInfo() {
	var b strings.Builder
	fmt.Fprintf(&b, "scrollX:%d | ", d.e.scrollX)
	fmt.Fprintf(&b, "logCurPos:(%d,%d) | ", d.e.cursor.x, d.e.cursor.y)
	fmt.Fprintf(&b, "curLnLen:%d | ", d.e.currentLine().len())
	// fmt.Fprintf(&b, "curLnFL:(%d,%d) | ", d.e.currentLine().fVisible, d.e.currentLine().lVisible)
	// fmt.Fprintf(&b, "curLineText:%s | ", string(d.e.currentLine().buf))
	d.screen.PutStr(
		d.x,
		d.y-2,
		b.String(),
	)
}

func (d *Debug) DrawLineEnd() {
	for i := range d.e.lines {
		d.screen.SetContent(
			d.x+MaxCharsOnLine,
			d.y+i,
			'|',
			nil,
			tcell.StyleDefault,
		)
	}
}

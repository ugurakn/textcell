package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

func main() {
	screen, err := tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	defer screen.Fini()

	err = screen.Init()
	if err != nil {
		panic(err)
	}

	firstLine := NewLine()
	cursor := NewCursor(0, 0)

	// first render
	cursor.Show(screen)
	screen.Show()

	running := true
	for running {
		// screen.Clear()
		debug_clearScreen(screen)

		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyRune:
				firstLine.WriteChar(ev.Rune(), cursor.x)
				cursor.Right(firstLine.length)
			case tcell.KeyRight:
				cursor.Right(firstLine.length)
			case tcell.KeyLeft:
				cursor.Left()
			case tcell.KeyBackspace:
				if ok := firstLine.Backspace(cursor.x); ok {
					cursor.Left()
				}
			case tcell.KeyEsc:
				running = false
			}
		}

		// render text
		firstLine.Show(screen)

		// DEBUG
		// debug_showCursorCoords(cursor, screen)

		cursor.Show(screen)
		screen.Show()
	}
}

func debug_clearScreen(screen tcell.Screen) {
	screen.Fill('.', tcell.StyleDefault)
}

func debug_showCursorCoords(c *Cursor, screen tcell.Screen) {
	screen.PutStr(0, 10, fmt.Sprintf(
		"cursorCoords: (%d, %d)",
		c.x,
		c.y,
	))
}

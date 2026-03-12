package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// TEMP config as global const
const (
	BaseX = 2
	BaseY = 2

	MaxCharsOnLine = 32
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

	editor := NewEditor(BaseX, BaseY)

	// first render
	editor.ShowCursor(screen)
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
				editor.WriteChar(ev.Rune())
			case tcell.KeyRight:
				editor.CurRight()
			case tcell.KeyLeft:
				editor.CurLeft()
			case tcell.KeyBackspace:
				editor.Backspace()
			case tcell.KeyEsc:
				running = false
			}
		}

		// render text and cursor
		editor.ShowText(screen)
		editor.ShowCursor(screen)

		// DEBUG
		// debug_showCursorCoords(cursor, screen)

		screen.Show()
	}
}

func debug_clearScreen(screen tcell.Screen) {
	screen.Fill('.', tcell.StyleDefault)
}

func debug_showCursorCoords(c *cursor, screen tcell.Screen) {
	screen.PutStr(0, 10, fmt.Sprintf(
		"cursorCoords: (%d, %d)",
		c.x,
		c.y,
	))
}

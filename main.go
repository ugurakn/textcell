package main

import (
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
		screen.Clear()

		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyRune:
				firstLine.WriteChar(ev.Rune())
				cursor.Right()
			case tcell.KeyRight:
				cursor.Right()
			case tcell.KeyLeft:
				cursor.Left()
			case tcell.KeyBackspace:
				if ok := firstLine.Backspace(); ok {
					cursor.Left()
				}
			case tcell.KeyEsc:
				running = false
			}
		}

		// render text
		firstLine.Show(screen)

		cursor.Show(screen)
		screen.Show()
	}
}

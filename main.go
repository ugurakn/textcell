package main

import "github.com/gdamore/tcell/v2"

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
	cursorX := 0

	// first render
	screen.ShowCursor(cursorX, 0)
	screen.Show()

	running := true
	for running {
		screen.Clear()

		var char rune
		ev := screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyRune:
				char = ev.Rune()
			case tcell.KeyEsc:
				running = false
			}
		}

		if char > 0 {
			firstLine.WriteChar(char)
			cursorX++
		}

		// render text
		firstLine.Show(screen)

		screen.ShowCursor(cursorX, 0)

		screen.Show()
	}
}

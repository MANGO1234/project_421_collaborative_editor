package main

import (
	"./buffer"
	"github.com/nsf/termbox-go"
)

func drawLines(lines *buffer.Line, height int) {
	y := 0
	for lines != nil && y < height {
		x := 0
		for _, ch := range lines.Bytes {
			if ch == '\t' {
				for i := 0; i < 4; i++ {
					termbox.SetCell(x+i, y, ' ', termbox.ColorWhite, termbox.ColorDefault)
				}
				x += 4
			} else if ch != '\n' {
				termbox.SetCell(x, y, rune(ch), termbox.ColorWhite, termbox.ColorDefault)
				x++
			}
		}
		y++
		lines = lines.Next
	}
}

func redrawPrompt(prompt *buffer.Prompt, width, height int) {
	cursorX, cursorY, lines := prompt.GetDisplayInformation(width-1, height)
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	drawLines(lines, height)
	termbox.SetCursor(cursorX, cursorY)
	termbox.Flush()
}

func InitPrompt() error {
	err := termbox.Init()
	if err != nil {
		return err
	}

	width, height := termbox.Size()
	termbox.SetInputMode(termbox.InputEsc)
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	prompt := buffer.NewPrompt("TESTING\nTTT")
	redrawPrompt(prompt, width, height)

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyCtrlC:
				return nil
			case termbox.KeyBackspace:
				prompt.Delete()
				redrawPrompt(prompt, width, height)
			case termbox.KeySpace:
				prompt.Insert(' ')
				redrawPrompt(prompt, width, height)
			case termbox.KeyTab:
				prompt.Insert('\t')
				redrawPrompt(prompt, width, height)
			case termbox.KeyEnter:
				// ACTION
				redrawPrompt(prompt, width, height)
			default:
				if ev.Key == 0 && ev.Ch <= 256 {
					prompt.Insert(byte(ev.Ch))
					redrawPrompt(prompt, width, height)
				}
			}
		case termbox.EventResize:
			width = ev.Width
			height = ev.Height
			redrawPrompt(prompt, width, height)
		case termbox.EventInterrupt:
			redrawPrompt(prompt, width, height)
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}

func ClosePrompt() {
	termbox.Close()
}

func main() {
	InitPrompt()
	ClosePrompt()
}

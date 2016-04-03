package gui

import (
	"bytes"
	"github.com/nsf/termbox-go"
	"strconv"
)

func DrawLines(lines *Line) {
	y := 0
	for lines != nil {
		x := 0
		for _, ch := range lines.bytes {
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
		lines = lines.next
	}
}

func InitEditor() error {
	err := termbox.Init()
	if err != nil {
		return err
	}

	width, _ := termbox.Size()
	termbox.SetInputMode(termbox.InputEsc)
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	build := bytes.Buffer{}
	for i := 0; i < 60; i++ {
		build.WriteString(strconv.Itoa(i))
		build.WriteString("\t")
		build.WriteString(strconv.Itoa(i))
		build.WriteString(" ")
		build.WriteString(strconv.Itoa(i))
		build.WriteString("\n")
	}
	buf := StringToBuffer(build.String(), width-1)
	DrawLines(buf.Lines())
	termbox.SetCursor(0, 0)
	termbox.Flush()

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyCtrlC:
				return nil
			case termbox.KeyArrowLeft:
				buf.MoveLeft()
				x, y := buf.GetCursorPosition()
				termbox.SetCursor(x, y)
			case termbox.KeyArrowRight:
				buf.MoveRight()
				x, y := buf.GetCursorPosition()
				termbox.SetCursor(x, y)
			case termbox.KeyArrowUp:
				buf.MoveUp()
				x, y := buf.GetCursorPosition()
				termbox.SetCursor(x, y)
			case termbox.KeyArrowDown:
				buf.MoveDown()
				x, y := buf.GetCursorPosition()
				termbox.SetCursor(x, y)
			default:
				termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
				DrawLines(buf.Lines())
				termbox.Flush()
			}
		case termbox.EventResize:
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			DrawLines(buf.Lines())
			termbox.Flush()
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}

func CloseEditor() {
	termbox.Close()
}

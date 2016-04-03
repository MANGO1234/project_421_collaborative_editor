package gui

import (
	"bytes"
	"github.com/nsf/termbox-go"
	"strconv"
)

func DrawLines(lines *Line, height int) {
	y := 0
	for lines != nil && y < height {
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

	width, height := termbox.Size()
	termbox.SetInputMode(termbox.InputEsc)
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	build := bytes.Buffer{}
	for i := 0; i < 60; i++ {
		build.WriteString(strconv.Itoa(i))
		build.WriteString("\t")
		build.WriteString(strconv.Itoa(i))
		build.WriteString(" ")
		build.WriteString(strconv.Itoa(i))
		build.WriteString(strconv.Itoa(i))
		build.WriteString(strconv.Itoa(i))
		build.WriteString(strconv.Itoa(i))
		build.WriteString(strconv.Itoa(i))
		build.WriteString(strconv.Itoa(i))
		build.WriteString(strconv.Itoa(i))
		build.WriteString(strconv.Itoa(i))
		build.WriteString(strconv.Itoa(i))
		build.WriteString(strconv.Itoa(i))
		build.WriteString(strconv.Itoa(i))
		build.WriteString(strconv.Itoa(i))
		build.WriteString(strconv.Itoa(i))
		build.WriteString(strconv.Itoa(i))
		build.WriteString(strconv.Itoa(i))
		build.WriteString(strconv.Itoa(i))
		build.WriteString(strconv.Itoa(i))
		build.WriteString(strconv.Itoa(i))
		build.WriteString("\n")
	}
	buf := StringToBuffer(build.String(), width)
	DrawLines(buf.Lines(), height)
	termbox.SetCursor(0, 0)
	termbox.Flush()

	var cursorX, cursorY int
	var lines *Line
	screenY := 0
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyCtrlC:
				return nil
			case termbox.KeyArrowLeft:
				buf.MoveLeft()
				screenY, cursorX, cursorY, lines = buf.GetDisplayInformation(screenY, height)
				termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
				DrawLines(lines, height)
				termbox.SetCursor(cursorX, cursorY)
				termbox.Flush()
			case termbox.KeyArrowRight:
				buf.MoveRight()
				screenY, cursorX, cursorY, lines = buf.GetDisplayInformation(screenY, height)
				termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
				DrawLines(lines, height)
				termbox.SetCursor(cursorX, cursorY)
				termbox.Flush()
			case termbox.KeyArrowUp:
				buf.MoveUp()
				screenY, cursorX, cursorY, lines = buf.GetDisplayInformation(screenY, height)
				termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
				DrawLines(lines, height)
				termbox.SetCursor(cursorX, cursorY)
				termbox.Flush()
			case termbox.KeyArrowDown:
				buf.MoveDown()
				screenY, cursorX, cursorY, lines = buf.GetDisplayInformation(screenY, height)
				termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
				DrawLines(lines, height)
				termbox.SetCursor(cursorX, cursorY)
				termbox.Flush()
			case termbox.KeyBackspace:
				buf.BackspaceAtCurrent()
				screenY, cursorX, cursorY, lines = buf.GetDisplayInformation(screenY, height)
				termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
				DrawLines(lines, height)
				termbox.SetCursor(cursorX, cursorY)
				termbox.Flush()
			case termbox.KeyDelete:
				buf.DeleteAtCurrent()
				screenY, cursorX, cursorY, lines = buf.GetDisplayInformation(screenY, height)
				termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
				DrawLines(lines, height)
				termbox.SetCursor(cursorX, cursorY)
				termbox.Flush()
			case termbox.KeySpace:
				buf.InsertAtCurrent(' ')
				screenY, cursorX, cursorY, lines = buf.GetDisplayInformation(screenY, height)
				termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
				DrawLines(lines, height)
				termbox.SetCursor(cursorX, cursorY)
				termbox.Flush()
			case termbox.KeyTab:
				buf.InsertAtCurrent('\t')
				screenY, cursorX, cursorY, lines = buf.GetDisplayInformation(screenY, height)
				termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
				DrawLines(lines, height)
				termbox.SetCursor(cursorX, cursorY)
				termbox.Flush()
			case termbox.KeyEnter:
				buf.InsertAtCurrent('\n')
				screenY, cursorX, cursorY, lines = buf.GetDisplayInformation(screenY, height)
				termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
				DrawLines(lines, height)
				termbox.SetCursor(cursorX, cursorY)
				termbox.Flush()
			default:
				if ev.Key == 0 && ev.Ch <= 256 {
					buf.InsertAtCurrent(byte(ev.Ch))
					screenY, cursorX, cursorY, lines = buf.GetDisplayInformation(screenY, height)
					termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
					DrawLines(lines, height)
					termbox.SetCursor(cursorX, cursorY)
					termbox.Flush()
				}
			}
		case termbox.EventResize:
			width = ev.Width
			height = ev.Height
			buf.Resize(width)
			// this is a bug within the library, without this call clear would panic when
			// cursor is outside of resized window
			termbox.HideCursor()
			screenY, cursorX, cursorY, lines = buf.GetDisplayInformation(screenY, height)
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			DrawLines(lines, height)
			termbox.SetCursor(cursorX, cursorY)
			termbox.Flush()
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}

func CloseEditor() {
	termbox.Close()
}

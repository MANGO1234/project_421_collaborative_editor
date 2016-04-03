package gui

import (
	"github.com/nsf/termbox-go"
)

func DrawBuffer(buffer [][]byte) {
	for y, line := range buffer {
		x := 0
		for _, ch := range line {
			if ch == '\t' {
				x += 4
			} else {
				termbox.SetCell(x, y, rune(ch), termbox.ColorWhite, termbox.ColorDefault)
				x++
			}
		}
	}
}

func DrawLines(lines *Line) {
	y := 0
	for lines != nil {
		x := 0
		for _, ch := range lines.bytes {
			if ch == '\t' {
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

	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	k := make([][]byte, 0, 4)
	v := make([]byte, 0, 4)
	v = append(v, 'a')
	v = append(v, 'b')
	v = append(v, 'c')
	k = append(k, v)
	v = make([]byte, 0, 4)
	v = append(v, '\t')
	v = append(v, 'd')
	v = append(v, 'e')
	v = append(v, 'f')
	k = append(k, v)
	buf := StringToBuffer("abc abc\ndef def def aasdfa asdfa asdfa asdfa asdfas \t asdfasdf", 20)
	DrawLines(buf.Lines())
	//	DrawBuffer(k)
	termbox.SetCursor(0, 0)
	termbox.Flush()

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyCtrlC {
				return nil
			}
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			DrawLines(buf.Lines())
			termbox.Flush()
		case termbox.EventResize:
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			DrawBuffer(k)
			termbox.Flush()
		case termbox.EventMouse:
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			DrawBuffer(k)
			termbox.Flush()
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}

func CloseEditor() {
	termbox.Close()
}

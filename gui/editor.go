package gui

import (
	"../buffer"
	. "../common"
	"../documentmanager"
	"github.com/nsf/termbox-go"
)

var docModel *documentmanager.DocumentModel

func DrawLines(lines *buffer.Line, height int) {
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

func InitEditor(siteId SiteId) error {
	err := termbox.Init()
	if err != nil {
		return err
	}

	width, height := termbox.Size()
	termbox.SetInputMode(termbox.InputEsc)
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	docModel = documentmanager.NewDocumentModel(siteId, width-1, func() {
		termbox.Interrupt()
	})
	DrawLines(docModel.Buffer.Lines(), height)
	termbox.SetCursor(0, 0)
	termbox.Flush()

	var cursorX, cursorY int
	var lines *buffer.Line
	screenY := 0
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyCtrlC:
				return nil
			case termbox.KeyArrowLeft:
				docModel.Buffer.MoveLeft()
				screenY, cursorX, cursorY, lines = docModel.Buffer.GetDisplayInformation(screenY, height)
				termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
				DrawLines(lines, height)
				termbox.SetCursor(cursorX, cursorY)
				termbox.Flush()
			case termbox.KeyArrowRight:
				docModel.Buffer.MoveRight()
				screenY, cursorX, cursorY, lines = docModel.Buffer.GetDisplayInformation(screenY, height)
				termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
				DrawLines(lines, height)
				termbox.SetCursor(cursorX, cursorY)
				termbox.Flush()
			case termbox.KeyArrowUp:
				docModel.Buffer.MoveUp()
				screenY, cursorX, cursorY, lines = docModel.Buffer.GetDisplayInformation(screenY, height)
				termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
				DrawLines(lines, height)
				termbox.SetCursor(cursorX, cursorY)
				termbox.Flush()
			case termbox.KeyArrowDown:
				docModel.Buffer.MoveDown()
				screenY, cursorX, cursorY, lines = docModel.Buffer.GetDisplayInformation(screenY, height)
				termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
				DrawLines(lines, height)
				termbox.SetCursor(cursorX, cursorY)
				termbox.Flush()
			case termbox.KeyBackspace:
				docModel.LocalBackspace()
				screenY, cursorX, cursorY, lines = docModel.Buffer.GetDisplayInformation(screenY, height)
				termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
				DrawLines(lines, height)
				termbox.SetCursor(cursorX, cursorY)
				termbox.Flush()
			case termbox.KeyDelete:
				docModel.LocalDelete()
				screenY, cursorX, cursorY, lines = docModel.Buffer.GetDisplayInformation(screenY, height)
				termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
				DrawLines(lines, height)
				termbox.SetCursor(cursorX, cursorY)
				termbox.Flush()
			case termbox.KeySpace:
				docModel.LocalInsert(' ')
				screenY, cursorX, cursorY, lines = docModel.Buffer.GetDisplayInformation(screenY, height)
				termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
				DrawLines(lines, height)
				termbox.SetCursor(cursorX, cursorY)
				termbox.Flush()
			case termbox.KeyTab:
				docModel.LocalInsert('\t')
				screenY, cursorX, cursorY, lines = docModel.Buffer.GetDisplayInformation(screenY, height)
				termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
				DrawLines(lines, height)
				termbox.SetCursor(cursorX, cursorY)
				termbox.Flush()
			case termbox.KeyEnter:
				docModel.LocalInsert('\n')
				screenY, cursorX, cursorY, lines = docModel.Buffer.GetDisplayInformation(screenY, height)
				termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
				DrawLines(lines, height)
				termbox.SetCursor(cursorX, cursorY)
				termbox.Flush()
			default:
				if ev.Key == 0 && ev.Ch <= 256 {
					docModel.LocalInsert(byte(ev.Ch))
					screenY, cursorX, cursorY, lines = docModel.Buffer.GetDisplayInformation(screenY, height)
					termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
					DrawLines(lines, height)
					termbox.SetCursor(cursorX, cursorY)
					termbox.Flush()
				}
			}
		case termbox.EventResize:
			width = ev.Width
			height = ev.Height
			docModel.Buffer.Resize(width - 1)
			// this is a bug within the library, without this call clear would panic when
			// cursor is outside of resized window
			termbox.HideCursor()
			screenY, cursorX, cursorY, lines = docModel.Buffer.GetDisplayInformation(screenY, height)
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			DrawLines(lines, height)
			termbox.SetCursor(cursorX, cursorY)
			termbox.Flush()
		case termbox.EventInterrupt:
			screenY, cursorX, cursorY, lines = docModel.Buffer.GetDisplayInformation(screenY, height)
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			DrawLines(lines, height)
			termbox.SetCursor(cursorX, cursorY)
			termbox.Flush()
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}

func Model() *documentmanager.DocumentModel {
	return docModel
}

func ForceUpdate() {
	termbox.Interrupt()
}

func CloseEditor() {
	termbox.Close()
}

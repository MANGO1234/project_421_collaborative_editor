package gui

import (
	"../buffer"
	. "../common"
	"../documentmanager"
	"github.com/nsf/termbox-go"
)

var docModel *documentmanager.DocumentModel

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

func redrawEditor(screenY, height int) int {
	screenY, cursorX, cursorY, lines := docModel.Buffer.GetDisplayInformation(screenY, height)
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	drawLines(lines, height)
	termbox.SetCursor(cursorX, cursorY)
	termbox.Flush()
	return screenY
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
	screenY := 0
	screenY = redrawEditor(screenY, height)

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyCtrlC:
				return nil
			case termbox.KeyArrowLeft:
				docModel.Buffer.MoveLeft()
				screenY = redrawEditor(screenY, height)
			case termbox.KeyArrowRight:
				docModel.Buffer.MoveRight()
				screenY = redrawEditor(screenY, height)
			case termbox.KeyArrowUp:
				docModel.Buffer.MoveUp()
				screenY = redrawEditor(screenY, height)
			case termbox.KeyArrowDown:
				docModel.Buffer.MoveDown()
				screenY = redrawEditor(screenY, height)
			case termbox.KeyBackspace:
				docModel.LocalBackspace()
				screenY = redrawEditor(screenY, height)
			case termbox.KeyDelete:
				docModel.LocalDelete()
				screenY = redrawEditor(screenY, height)
			case termbox.KeySpace:
				docModel.LocalInsert(' ')
				screenY = redrawEditor(screenY, height)
			case termbox.KeyTab:
				docModel.LocalInsert('\t')
				screenY = redrawEditor(screenY, height)
			case termbox.KeyEnter:
				docModel.LocalInsert('\n')
				screenY = redrawEditor(screenY, height)
			default:
				if ev.Key == 0 && ev.Ch <= 256 {
					docModel.LocalInsert(byte(ev.Ch))
					screenY = redrawEditor(screenY, height)
				}
			}
		case termbox.EventResize:
			width = ev.Width
			height = ev.Height
			docModel.Buffer.Resize(width - 1)
			// this is a bug within the library, without this call clear would panic when
			// cursor is outside of resized window
			termbox.HideCursor()
			screenY = redrawEditor(screenY, height)
		case termbox.EventInterrupt:
			// remote operations
			screenY = redrawEditor(screenY, height)
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

package gui

import (
	"../buffer"
	. "../common"
	"../documentmanager"
	"../network"
	"../treedoc2"
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

func newDocument(siteId SiteId) *documentmanager.DocumentModel {
	width, _ := termbox.Size()
	return documentmanager.NewDocumentModel(siteId, width-1, func() {
		termbox.Interrupt()
	})
}

func broadcastTreeDocOperation(id string) func(version uint32, operation treedoc2.Operation) {
	return func(uint32, treedoc2.Operation) {}
}

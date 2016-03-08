package main

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"log"
)

func delMsg(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView("msg"); err != nil {
		return err
	}
	if err := g.SetCurrentView("side"); err != nil {
		return err
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func nextView(g *gocui.Gui, v *gocui.View) error {
	if v == nil || v.Name() == "side" {
		return g.SetCurrentView("main")
	}
	return g.SetCurrentView("side")
}

type MenuOption struct {
	Label string
}

var options = []MenuOption{
	MenuOption{"New File"},
	MenuOption{"Open Plaintext"},
	MenuOption{"Save Plaintext"},
	MenuOption{"Open Treedoc"},
	MenuOption{"Save Treedoc"},
}

var optionIndex = 0

func optionDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil && optionIndex < len(options)-1 {
		if err := v.SetCursor(0, optionIndex+1); err != nil {
			return err
		}
		optionIndex++
	}
	return nil
}

func optionUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil && optionIndex > 0 {
		if err := v.SetCursor(0, optionIndex-1); err != nil {
			return err
		}
		optionIndex--
	}
	return nil
}

func executeOption(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("msg", maxX/2-30, maxY/2, maxX/2+30, maxY/2+2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, options[optionIndex].Label)
		if err := g.SetCurrentView("msg"); err != nil {
			return err
		}
	}
	return nil
}

func simpleEditor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
	case key == gocui.KeySpace:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	case key == gocui.KeyDelete:
		v.EditDelete(false)
	case key == gocui.KeyEnter:
		v.EditNewLine()
	case key == gocui.KeyArrowDown:
		_, y := v.Cursor()
		if _, err := v.Line(y + 1); err == nil {
			v.MoveCursor(0, 1, false)
		}
	case key == gocui.KeyArrowUp:
		v.MoveCursor(0, -1, false)
	case key == gocui.KeyArrowLeft:
		v.MoveCursor(-1, 0, false)
	case key == gocui.KeyArrowRight:
		v.MoveCursor(1, 0, false)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("side", -1, -1, 20, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.Highlight = true
		for i, option := range options {
			fmt.Fprint(v, i+1)
			fmt.Fprint(v, ". ")
			fmt.Fprintln(v, option.Label)
		}
	}

	g.Editor = gocui.EditorFunc(simpleEditor)

	if v, err := g.SetView("main", 20, -1, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Editable = true
		v.Wrap = true
		if err := g.SetCurrentView("main"); err != nil {
			return err
		}
	}
	return nil
}

func keybindings(g *gocui.Gui) error {
	// quitting
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyEsc, gocui.ModNone, quit); err != nil {
		return err
	}

	// switching views
	if err := g.SetKeybinding("side", gocui.KeyCtrlSpace, gocui.ModNone, nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding("main", gocui.KeyCtrlSpace, gocui.ModNone, nextView); err != nil {
		return err
	}

	// options
	if err := g.SetKeybinding("side", gocui.KeyArrowDown, gocui.ModNone, optionDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("side", gocui.KeyArrowUp, gocui.ModNone, optionUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("side", gocui.KeyEnter, gocui.ModNone, executeOption); err != nil {
		return err
	}
	if err := g.SetKeybinding("msg", gocui.KeyEnter, gocui.ModNone, delMsg); err != nil {
		return err
	}

	return nil
}

func main() {
	g := gocui.NewGui()
	if err := g.Init(); err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetLayout(layout)
	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}
	g.SelBgColor = gocui.ColorGreen
	g.SelFgColor = gocui.ColorBlack
	g.Cursor = true

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

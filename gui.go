package main

import (
	"./buffer"
	"github.com/nsf/termbox-go"
	"strconv"
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

const MENU = 0
const MENU_RETRY = 1
const CONNECT = 10
const DISCONNECT = 20
const EXIT = 30

var AppState struct {
	State     int
	Connected bool
}

func doAction(input string) {
	if AppState.State == MENU || AppState.State == MENU_RETRY {
		n, err := strconv.Atoi(input)
		if err != nil {
			AppState.State = MENU_RETRY
		} else if n == 1 {
			// todo connect/disconnect
			AppState.Connected = !AppState.Connected
			AppState.State = MENU
		} else if n == 2 {
			AppState.State = EXIT
		} else {
			AppState.State = MENU_RETRY
		}
	}
}

func getPrompt() *buffer.Prompt {
	if AppState.State == MENU || AppState.State == MENU_RETRY {
		str := ""
		if AppState.Connected {
			str += "1. Disconnect\n"
		} else {
			str += "1. Connect\n"
		}
		str += "2. Exit\n"
		str += "\n"
		if AppState.State == MENU_RETRY {
			str += "Input is not a valid input, please try again: "
		} else {
			str += "Enter number to execture action: "
		}
		return buffer.NewPrompt(str)
	}
	panic("UNKNOWN STATE " + strconv.Itoa(AppState.State))
}

func InitPrompt() error {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}

	width, height := termbox.Size()
	termbox.SetInputMode(termbox.InputEsc)
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	prompt := getPrompt()
	redrawPrompt(prompt, width, height)

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyCtrlC:
				AppState.State = EXIT
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
				doAction(prompt.ToString())
				return nil
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
	for {
		if AppState.State == EXIT {
			return
		}
		InitPrompt()
		ClosePrompt()
	}
}

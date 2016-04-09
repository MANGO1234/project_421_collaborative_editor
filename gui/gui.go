package gui

import (
	"../buffer"
	. "../common"
	"../documentmanager"
	"github.com/nsf/termbox-go"
	"strconv"
)

// States
const STATE_MENU = 0
const STATE_MENU_RETRY = 1
const STATE_EXIT = 10
const STATE_DOCUMENT = 20

// Menu Options
const OPTION_EXIT = "Exit"
const OPTION_CONNECT = "Connect"
const OPTION_DISCONNECT = "Disconnect"
const OPTION_NEW_DOCUMENT = "New Document"
const OPTION_CLOSE_DOCUMENT = "Close Document"

var appState struct {
	State       int
	Connected   bool
	DocModel    *documentmanager.DocumentModel
	ScreenY     int
	MenuOptions []string
}

func doAction(input string) {
	if appState.State == STATE_MENU || appState.State == STATE_MENU_RETRY {
		n, err := strconv.Atoi(input)
		if err != nil || n < 1 || n > len(appState.MenuOptions) {
			appState.State = STATE_MENU_RETRY
			return
		}
		if appState.MenuOptions[n-1] == OPTION_EXIT {
			appState.State = STATE_EXIT
		} else if appState.MenuOptions[n-1] == OPTION_CONNECT {
			appState.Connected = true
		} else if appState.MenuOptions[n-1] == OPTION_DISCONNECT {
			appState.Connected = false
		} else if appState.MenuOptions[n-1] == OPTION_NEW_DOCUMENT {
			appState.DocModel = newDocument(StringToSiteId("aaaaaaaaaaaaaaaa"))
			appState.ScreenY = 0
			appState.State = STATE_DOCUMENT
		} else if appState.MenuOptions[n-1] == OPTION_CLOSE_DOCUMENT {
			appState.DocModel = nil
		} else {
			appState.State = STATE_MENU_RETRY
		}
	}
}

func getPrompt() *buffer.Prompt {
	if appState.State == STATE_MENU || appState.State == STATE_MENU_RETRY {
		options := make([]string, 0, 10)
		options = append(options, OPTION_EXIT)
		if appState.Connected {
			options = append(options, OPTION_DISCONNECT)
		} else {
			options = append(options, OPTION_CONNECT)
		}
		if appState.DocModel == nil {
			options = append(options, OPTION_NEW_DOCUMENT)
		} else {
			options = append(options, OPTION_CLOSE_DOCUMENT)
		}
		appState.MenuOptions = options

		str := ""
		for i, option := range options {
			str += strconv.Itoa(i+1) + ". " + option + "\n"
		}
		str += "\n"
		if appState.DocModel != nil {
			str += "Esc to switch between menu and editing the document\n"
			str += "\n"
		}
		if appState.State == STATE_MENU_RETRY {
			str += "Input is not a valid input, please try again: "
		} else {
			str += "Enter number to exectute option: "
		}
		return buffer.NewPrompt(str)
	}
	panic("UNKNOWN STATE " + strconv.Itoa(appState.State))
}

func StartMainLoop() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	appState.State = STATE_MENU
	for {
		if appState.State == STATE_EXIT {
			break
		}
		if appState.State == STATE_DOCUMENT {
			DrawEditor()
		} else {
			DrawPrompt(getPrompt())
		}
	}
}

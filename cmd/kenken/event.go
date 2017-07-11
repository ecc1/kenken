package main

import (
	"unsafe"

	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/glib"
)

func buttonpress(x, y int) func(*glib.CallbackContext) {
	return func(ctx *glib.CallbackContext) {
		arg := ctx.Args(0)
		ev := *(**gdk.EventButton)(unsafe.Pointer(&arg))
		switch ev.Button {
		case 1: // left
			if isConstant(x, y) {
				return
			}
			addNotes(x, y)
		case 3: // right
			tryAgain()
		}
	}
}

func keypress(x, y int) func(*glib.CallbackContext) {
	return func(ctx *glib.CallbackContext) {
		if isConstant(x, y) {
			return
		}
		arg := ctx.Args(0)
		ev := *(**gdk.EventKey)(unsafe.Pointer(&arg))
		n := int(ev.Keyval) - '0'
		if 1 <= n && n <= game.size {
			updateCell(x, y, n)
			return
		}
		n = decodeKey(uint(ev.Keyval))
		switch n {
		case -1:
			return
		case 0:
			clearAll(x, y)
		default:
			if 1 <= n && n <= game.size {
				updateNote(x, y, n)
			}
		}
	}
}

func decodeKey(c uint) int {
	switch c {
	case '!':
		return 1
	case '@':
		return 2
	case '#':
		return 3
	case '$':
		return 4
	case '%':
		return 5
	case '^':
		return 6
	case '&':
		return 7
	case '*':
		return 8
	case '(':
		return 9
	case ' ', gdk.KEY_Delete:
		return 0
	default:
		return -1
	}
}

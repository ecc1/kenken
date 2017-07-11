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

var keycode = map[uint]int{
	'!':            1,
	'@':            2,
	'#':            3,
	'$':            4,
	'%':            5,
	'^':            6,
	'&':            7,
	'*':            8,
	'(':            9,
	' ':            0,
	gdk.KEY_Delete: 0,
}

func decodeKey(c uint) int {
	n, known := keycode[c]
	if !known {
		return -1
	}
	return n
}

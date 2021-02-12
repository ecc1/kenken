package main

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

func focus(x, y int, eb *gtk.EventBox) {
	if isConstant(x, y) {
		return
	}
	eb.GrabFocus()
}

func keyPress(x, y int, w gtk.IWidget, e *gdk.Event) {
	if isConstant(x, y) {
		return
	}
	ek := gdk.EventKeyNewFromEvent(e)
	k := ek.KeyVal()
	n, known := keycode[k]
	if !known {
		return
	}
	// Space/backspace/delete: clear cell.
	if n == 0 {
		clearAll(x, y)
		return
	}
	if n > size {
		return
	}
	if ek.State() != 0 {
		// Modified number key (control or shift): update notes.
		updateNote(x, y, n)
		return
	}
	// Unmodified number key: update cell.
	updateCell(x, y, n)
}

var keycode = map[uint]int{
	' ': 0, gdk.KEY_BackSpace: 0, gdk.KEY_Delete: 0,
	// Sorry, this is specific to US keyboard layouts.
	'1': 1, '!': 1,
	'2': 2, '@': 2,
	'3': 3, '#': 3,
	'4': 4, '$': 4,
	'5': 5, '%': 5,
	'6': 6, '^': 6,
	'7': 7, '&': 7,
	'8': 8, '*': 8,
	'9': 9, '(': 9,
}

func buttonPress(x, y int, w gtk.IWidget, e *gdk.Event) {
	switch gdk.EventButtonNewFromEvent(e).Button() {
	case 1: // left
		if isConstant(x, y) {
			return
		}
		addNotes(x, y)
	case 3: // right
		tryAgain()
	}
}

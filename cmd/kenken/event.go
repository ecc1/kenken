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
	k := gdk.EventKeyNewFromEvent(e).KeyVal()
	n := int(k - '0')
	if 1 <= n && n <= size {
		updateCell(x, y, n)
		return
	}
	n, known := keycode[k]
	if !known {
		return
	}
	if n == 0 {
		clearAll(x, y)
		return
	}
	if 1 <= n && n <= size {
		updateNote(x, y, n)
	}
}

var keycode = map[uint]int{
	gdk.KEY_BackSpace: 0,
	gdk.KEY_Delete:    0,
	' ':               0,
	'!':               1,
	'@':               2,
	'#':               3,
	'$':               4,
	'%':               5,
	'^':               6,
	'&':               7,
	'*':               8,
	'(':               9,
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

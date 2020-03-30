package main

import (
	"fmt"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

const (
	textFont = "Sans"

	// These are relative to a unit-square cell size.
	lineWidth     = 0.025
	innerSep      = 0.050
	largeFontSize = 0.400
	// Must be small enough for "123456789" to fit in a cell.
	smallFontSize = 0.160
)

var (
	window *gtk.Window
	grid   *gtk.Grid
)

func initUI() {
	gtk.Init(nil)
	window, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	window.SetTitle(fmt.Sprintf("KenKen %s", title))
	setGeometry(window)
	window.Connect("destroy", gtk.MainQuit)
	window.Add(makeGrid())
	window.ShowAll()
}

func setGeometry(win *gtk.Window) {
	d, _ := win.GetScreen().GetDisplay()
	m, _ := d.GetPrimaryMonitor()
	r := m.GetGeometry()
	sz := size * min(r.GetWidth(), r.GetHeight()) / 10
	win.SetDefaultSize(sz, sz)
	var g gdk.Geometry
	g.SetMinAspect(1)
	g.SetMaxAspect(1)
	win.SetGeometryHints(nil, g, gdk.HINT_ASPECT)
	win.SetPosition(gtk.WIN_POS_MOUSE)
}

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}

func makeGrid() gtk.IWidget {
	grid, _ = gtk.GridNew()
	grid.SetRowHomogeneous(true)
	grid.SetColumnHomogeneous(true)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			attachCell(x, y)
		}
	}
	a, _ := gtk.AspectFrameNew("", 0.5, 0.5, 1, false)
	a.Add(grid)
	return a
}

func attachCell(x, y int) {
	d, _ := gtk.DrawingAreaNew()
	d.SetHExpand(true)
	d.SetVExpand(true)
	d.Connect("draw", func(d *gtk.DrawingArea, c *cairo.Context) { drawCell(x, y, d, c) })
	eb, _ := gtk.EventBoxNew()
	eb.Add(d)
	eb.SetCanFocus(true)
	eb.Connect("enter-notify-event", func(eb *gtk.EventBox) { focus(x, y, eb) })
	eb.Connect("key-press-event", func(eb *gtk.EventBox, e *gdk.Event) { keyPress(x, y, eb, e) })
	eb.Connect("button-press-event", func(eb *gtk.EventBox, e *gdk.Event) { buttonPress(x, y, eb, e) })
	grid.Attach(eb, x+1, y+1, 1, 1)
}

func drawCell(x, y int, d *gtk.DrawingArea, c *cairo.Context) {
	sc, _ := d.GetStyleContext()
	cv := sc.GetColor(gtk.STATE_FLAG_NORMAL).Floats()
	c.SetSourceRGBA(cv[0], cv[1], cv[2], cv[3])
	// Transform cell to unit square.
	cellSize := float64(d.GetAllocatedWidth())
	c.Scale(cellSize, cellSize)
	// Cage lines.
	c.SetLineWidth(lineWidth)
	if x == 0 || puzzle.Vertical[y][x-1] {
		c.MoveTo(0, 0)
		c.LineTo(0, 1)
	}
	if x == size-1 {
		c.MoveTo(1, 0)
		c.LineTo(1, 1)
	}
	if y == 0 || puzzle.Horizontal[x][y-1] {
		c.MoveTo(0, 0)
		c.LineTo(1, 0)
	}
	if y == size-1 {
		c.MoveTo(0, 1)
		c.LineTo(1, 1)
	}
	c.Stroke()
	// Answer.
	c.SelectFontFace(textFont, cairo.FONT_SLANT_NORMAL, cairo.FONT_WEIGHT_NORMAL)
	c.SetFontSize(largeFontSize)
	n := cells[y][x]
	if n != 0 {
		num := fmt.Sprintf("%d", n)
		t := c.TextExtents(num)
		c.MoveTo(0.5-t.Width/2, 0.5+t.Height/2)
		c.ShowText(num)
	}
	if isConstant(x, y) {
		return
	}
	// Clue.
	c.SelectFontFace(textFont, cairo.FONT_SLANT_NORMAL, cairo.FONT_WEIGHT_NORMAL)
	c.SetFontSize(smallFontSize)
	clue := clueString(x, y)
	if clue != "" {
		t := c.TextExtents(clue)
		c.MoveTo(innerSep, innerSep+t.Height)
		c.ShowText(clue)
	}
	// Notes.
	noteStr := noteStrings[y][x]
	if noteStr != "" {
		t := c.TextExtents(noteStr)
		c.MoveTo(1-innerSep-t.Width, 1-innerSep)
		c.ShowText(noteStr)
	}
}

func redraw(x, y int) {
	w, _ := grid.GetChildAt(x+1, y+1)
	w.QueueDraw()
}

func runUI() {
	gtk.Main()
}

var winning = false

func winnerWinner() {
	if winning {
		return
	}
	winning = true
	dialog := gtk.MessageDialogNewWithMarkup(window, gtk.DIALOG_MODAL, gtk.MESSAGE_INFO, gtk.BUTTONS_OK, "")
	dialog.SetMarkup("<b>Correct!</b>")
	dialog.Run()
	dialog.Destroy()
	gtk.MainQuit()
}

var losing = false

func tryAgain() {
	if losing {
		return
	}
	losing = true
	dialog := gtk.MessageDialogNewWithMarkup(window, gtk.DIALOG_MODAL, gtk.MESSAGE_WARNING, gtk.BUTTONS_YES_NO, "")
	dialog.SetMarkup("<b>Try again?</b>")
	if dialog.Run() == gtk.RESPONSE_YES {
		restartGame()
	}
	dialog.Destroy()
	losing = false
}

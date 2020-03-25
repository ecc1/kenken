package main

import (
	"fmt"

	"github.com/ecc1/kenken"
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
	smallFontSize = 0.160
)

var (
	puzzle   *kenken.Puzzle
	size     int
	cellSize float64
	grid     *gtk.Grid
)

func main() {
	var title string
	puzzle, title = kenken.ReadPuzzle()
	size = puzzle.Size()
	gtk.Init(nil)
	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetTitle(fmt.Sprintf("KenKen %s", title))
	setGeometry(win)
	win.Connect("destroy", gtk.MainQuit)
	win.Connect("size-allocate", resize)
	grid, _ = gtk.GridNew()
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			attachCell(x, y)
		}
	}
	win.Add(grid)
	win.ShowAll()
	gtk.Main()
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
	win.SetPosition(gtk.WIN_POS_CENTER_ALWAYS)
}

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}

func attachCell(x int, y int) {
	d, _ := gtk.DrawingAreaNew()
	d.SetHExpand(true)
	d.SetVExpand(true)
	d.Connect("draw", func(d *gtk.DrawingArea, c *cairo.Context) { drawCell(x, y, d, c) })
	grid.Attach(d, x+1, y+1, 1, 1)
}

func drawCell(x int, y int, d *gtk.DrawingArea, c *cairo.Context) {
	sc, _ := d.GetStyleContext()
	cv := sc.GetColor(gtk.STATE_FLAG_NORMAL).Floats()
	c.SetSourceRGBA(cv[0], cv[1], cv[2], cv[3])
	// Transform cell to unit square.
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
	c.SelectFontFace(textFont, cairo.FONT_SLANT_NORMAL, cairo.FONT_WEIGHT_NORMAL)
	c.SetFontSize(largeFontSize)
	num := fmt.Sprintf("%d", puzzle.Answer[y][x])
	t := c.TextExtents(num)
	c.MoveTo(0.5-t.Width/2, 0.5+t.Height/2)
	c.ShowText(num)
	op := puzzle.Operation[y][x]
	switch op {
	case kenken.None, kenken.Given:
		return
	}
	c.SetFontSize(smallFontSize)
	clue := fmt.Sprintf("%d %s", puzzle.Clue[y][x], op.Symbol())
	t = c.TextExtents(clue)
	c.MoveTo(innerSep, innerSep+t.Height)
	c.ShowText(clue)
}

func resize(arg interface{}) {
	a := grid.GetAllocation()
	cellSize = float64(a.GetWidth()) / float64(puzzle.Size())
}

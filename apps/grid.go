package main

import (
	"fmt"
	"path/filepath"

	"github.com/ecc1/kenken"
	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/gtk"
)

const (
	cellSize      = 100
	cageThickness = 4
)

func main() {
	k, path := kenken.ReadPuzzle()
	size := k.Size()

	gtk.Init(nil)

	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	window.SetTitle(fmt.Sprintf("KenKen %s", filepath.Base(path)))
	window.Connect("destroy", gtk.MainQuit)
	window.SetResizable(false)

	table := gtk.NewTable(uint(size), uint(size), true)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			vbox := gtk.NewVBox(false, 0)
			vbox.SetSizeRequest(cellSize, cellSize)

			// Clue, if any
			str := ""
			switch k.Operation[y][x] {
			case kenken.None, kenken.Given:
			default:
				str = fmt.Sprintf("%d %s", k.Clue[y][x], k.Operation[y][x].Symbol())
			}
			label := gtk.NewLabel(str)
			label.ModifyFontEasy("DejaVu Sans 12")
			label.ModifyFG(gtk.STATE_INSENSITIVE, white)
			align := gtk.NewAlignment(0, 0, 0, 0)
			align.Add(label)
			vbox.PackStart(align, false, false, 0)

			// Notes
			label = gtk.NewLabel("")
			label.ModifyFontEasy("DejaVu Sans 12")
			align = gtk.NewAlignment(1, 0, 0, 0)
			align.Add(label)
			vbox.PackEnd(align, false, false, 0)

			// Cell contents
			label = gtk.NewLabel(fmt.Sprintf("%d", k.Answer[y][x]))
			label.ModifyFontEasy("DejaVu Sans 28")
			label.ModifyFG(gtk.STATE_INSENSITIVE, white)
			vbox.Add(label)

			button := gtk.NewButton()
			button.SetRelief(gtk.RELIEF_NONE)
			button.SetSensitive(false)
			button.Add(vbox)

			table.AttachDefaults(button, uint(x), uint(x+1), uint(y), uint(y+1))
		}
	}
	window.Add(table)
	window.Connect("configure-event", func() {
		drawCages(k, table.GetWindow().GetDrawable(), table.GetAllocation())
	})
	window.ShowAll()

	gtk.Main()
}

var white = gdk.NewColor("white")

func drawCages(k *kenken.Puzzle, d *gdk.Drawable, a *gtk.Allocation) {
	size := k.Size()
	w := a.Width / size
	h := a.Height / size
	gc := gdk.NewGC(d)
	gc.SetRgbFgColor(white)

	t := cageThickness

	d.DrawRectangle(gc, true, 0, 0, t, a.Height)
	d.DrawRectangle(gc, true, 0, 0, a.Width, t)
	d.DrawRectangle(gc, true, 0, a.Height-t, a.Width, t)
	d.DrawRectangle(gc, true, a.Width-t, 0, t, a.Height)

	for j := 0; j < size; j++ {
		y := j*h - t/2
		for i := 0; i < size; i++ {
			x := i*w - t/2
			if i > 0 && k.Vertical[j][i-1] {
				d.DrawRectangle(gc, true, x, y, t, h+t/2)
			}
			if j > 0 && k.Horizontal[i][j-1] {
				d.DrawRectangle(gc, true, x, y, w+t/2, t)
			}
		}
	}
}

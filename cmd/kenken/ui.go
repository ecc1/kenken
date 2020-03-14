package main

import (
	"fmt"
	"path/filepath"

	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/gtk"
)

const (
	cellSize  = 100
	lineWidth = 4
	largeFont = "DejaVu Sans 28"
	smallFont = "DejaVu Sans 12"
)

var (
	table     *gtk.Table
	cellLabel [][]*gtk.Label
	noteLabel [][]*gtk.Label

	white = gdk.NewColor("white")
)

func initUI() {
	gtk.Init(nil)
	size := puzzle.Size()
	table = gtk.NewTable(uint(size), uint(size), true)
	cellLabel = make([][]*gtk.Label, size)
	noteLabel = make([][]*gtk.Label, size)
	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	window.SetTitle(fmt.Sprintf("KenKen %s", filepath.Base(path)))
	window.SetPosition(gtk.WIN_POS_CENTER_ALWAYS)
	window.SetResizable(false)
	window.Connect("destroy", gtk.MainQuit)
	for y := 0; y < size; y++ {
		cellLabel[y] = make([]*gtk.Label, size)
		noteLabel[y] = make([]*gtk.Label, size)
		for x := 0; x < size; x++ {
			attachCell(x, y)
		}
	}
	window.Add(table)
	window.Connect("configure-event", drawCages)
	window.Connect("expose-event", drawCages)
	window.ShowAll()
}

func runUI() {
	gtk.Main()
}

func attachCell(x, y int) {
	vbox := gtk.NewVBox(false, 0)
	vbox.SetSizeRequest(cellSize, cellSize)

	// Clue, if any
	label := gtk.NewLabel(clueString(x, y))
	label.ModifyFontEasy(smallFont)
	align := gtk.NewAlignment(0, 0, 0, 0)
	align.Add(label)
	vbox.PackStart(align, false, false, 0)

	// Cell contents
	str := ""
	if isConstant(x, y) {
		str = fmt.Sprintf("%d", puzzle.Answer[y][x])
	}
	label = gtk.NewLabel(str)
	cellLabel[y][x] = label
	label.ModifyFontEasy(largeFont)
	label.ModifyFG(gtk.STATE_INSENSITIVE, white)
	vbox.Add(label)

	// Notes
	label = gtk.NewLabel("")
	noteLabel[y][x] = label
	label.ModifyFontEasy(smallFont)
	align = gtk.NewAlignment(1, 0, 0, 0)
	align.Add(label)
	vbox.PackEnd(align, false, false, 0)

	button := gtk.NewButton()
	button.SetRelief(gtk.RELIEF_NONE)
	if isConstant(x, y) {
		button.SetSensitive(false)
	}
	button.Add(vbox)
	button.Connect("enter-notify-event", button.GrabFocus)
	button.Connect("expose-event", drawCages)
	button.Connect("key-press-event", keypress(x, y))
	button.Connect("button-press-event", buttonpress(x, y))

	table.AttachDefaults(button, uint(x), uint(x+1), uint(y), uint(y+1))
}

func setCellLabel(x, y, n int) {
	cellLabel[y][x].SetText(fmt.Sprintf("%d", n))
}

func clearCellLabel(x, y int) {
	cellLabel[y][x].SetText("")
}

func setNoteLabel(x, y int, digits string) {
	noteLabel[y][x].SetText(digits)
}

func drawCages() {
	a := table.GetAllocation()
	size := puzzle.Size()
	w := a.Width / size
	h := a.Height / size
	d := table.GetWindow().GetDrawable()
	gc := gdk.NewGC(d)
	gc.SetRgbFgColor(white)

	t := lineWidth

	d.DrawRectangle(gc, true, 0, 0, t, a.Height)
	d.DrawRectangle(gc, true, 0, 0, a.Width, t)
	d.DrawRectangle(gc, true, 0, a.Height-t, a.Width, t)
	d.DrawRectangle(gc, true, a.Width-t, 0, t, a.Height)

	for j := 0; j < size; j++ {
		y := j*h - t/2
		for i := 0; i < size; i++ {
			x := i*w - t/2
			if i > 0 && puzzle.Vertical[j][i-1] {
				d.DrawRectangle(gc, true, x, y, t, h+t/2)
			}
			if j > 0 && puzzle.Horizontal[i][j-1] {
				d.DrawRectangle(gc, true, x, y, w+t/2, t)
			}
		}
	}
}

var winning = false

func winnerWinner() {
	if winning {
		return
	}
	winning = true
	dialog := gtk.NewMessageDialogWithMarkup(
		table.GetTopLevelAsWindow(), gtk.DIALOG_MODAL,
		gtk.MESSAGE_INFO, gtk.BUTTONS_OK, "<b>Correct!</b>",
	)
	dialog.Response(gtk.MainQuit)
	dialog.Run()
}

var losing = false

func tryAgain() {
	if losing {
		return
	}
	losing = true
	dialog := gtk.NewMessageDialogWithMarkup(
		table.GetTopLevelAsWindow(), gtk.DIALOG_MODAL,
		gtk.MESSAGE_WARNING, gtk.BUTTONS_YES_NO, "<b>Try again?</b>",
	)
	if dialog.Run() == gtk.RESPONSE_YES {
		restartGame()
	}
	dialog.Destroy()
	losing = false
}

package main

import (
	"fmt"
	"path/filepath"
	"unsafe"

	"github.com/ecc1/kenken"
	"github.com/mattn/go-gtk/gdk"
	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
)

const (
	cellSize  = 100
	lineWidth = 4
	largeFont = "DejaVu Sans 28"
	smallFont = "DejaVu Sans 12"
)

var game struct {
	k         *kenken.KenKen
	size      int
	cell      [][]int
	cellLabel [][]*gtk.Label
	note      [][][]bool
	noteLabel [][]*gtk.Label
	table     *gtk.Table
}

var (
	autoPromote = true
	white       = gdk.NewColor("white")
)

func main() {
	k, path := kenken.ReadPuzzle()

	initGame(k)

	gtk.Init(nil)

	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	window.SetTitle(fmt.Sprintf("KenKen %s", filepath.Base(path)))
	window.Connect("destroy", gtk.MainQuit)
	window.SetResizable(false)

	table := gtk.NewTable(uint(game.size), uint(game.size), true)
	game.table = table
	for y := 0; y < game.size; y++ {
		for x := 0; x < game.size; x++ {
			attachCell(x, y)
		}
	}
	window.Add(table)
	window.Connect("configure-event", drawCages)
	window.Connect("expose-event", drawCages)

	window.ShowAll()

	gtk.Main()
}

func initGame(k *kenken.KenKen) {
	game.k = k
	game.size = k.Size()
	game.cell = make([][]int, game.size)
	game.cellLabel = make([][]*gtk.Label, game.size)
	game.note = make([][][]bool, game.size)
	game.noteLabel = make([][]*gtk.Label, game.size)
	for y := 0; y < game.size; y++ {
		game.cell[y] = make([]int, game.size)
		game.cellLabel[y] = make([]*gtk.Label, game.size)
		game.note[y] = make([][]bool, game.size)
		for x := 0; x < game.size; x++ {
			if isConstant(x, y) {
				game.cell[y][x] = k.Answer[y][x]
			}
			game.note[y][x] = make([]bool, game.size)
		}
		game.noteLabel[y] = make([]*gtk.Label, game.size)
	}
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
		str = fmt.Sprintf("%d", game.k.Answer[y][x])
	}
	label = gtk.NewLabel(str)
	game.cellLabel[y][x] = label
	label.ModifyFontEasy(largeFont)
	label.ModifyFG(gtk.STATE_INSENSITIVE, white)
	vbox.Add(label)

	// Notes
	label = gtk.NewLabel("")
	game.noteLabel[y][x] = label
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
	button.Connect("enter-notify-event", func() { button.GrabFocus() })
	button.Connect("expose-event", drawCages)
	button.Connect("clicked", click(x, y))
	button.Connect("key-press-event", keypress(x, y))

	game.table.AttachDefaults(button, uint(x), uint(x+1), uint(y), uint(y+1))
}

func clueString(x, y int) string {
	switch game.k.Operation[y][x] {
	case kenken.None, kenken.Given:
		return ""
	default:
		return fmt.Sprintf("%d %s", game.k.Clue[y][x], game.k.Operation[y][x].Symbol())
	}
}

func ignore(_ *glib.CallbackContext) {
}

func click(x, y int) func(*glib.CallbackContext) {
	if isConstant(x, y) {
		return ignore
	}
	return func(_ *glib.CallbackContext) {
		addNotes(x, y)
	}
}

func addNotes(x, y int) {
	assertNonConstant(x, y)
	for i := 1; i <= game.size; i++ {
		if game.cell[y][x] == 0 && !inRowOrColumn(x, y, i) {
			game.note[y][x][i-1] = true
		}
	}
	setNoteLabel(x, y, true)
}

func keypress(x, y int) func(*glib.CallbackContext) {
	if isConstant(x, y) {
		return ignore
	}
	return func(ctx *glib.CallbackContext) {
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

func isConstant(x, y int) bool {
	return game.k.Operation[y][x] == kenken.Given
}

func assertNonConstant(x, y int) {
	if isConstant(x, y) {
		panic(fmt.Sprintf("modifying constant cell (%d, %d)", x, y))
	}
}

func updateCell(x, y, n int) {
	assertNonConstant(x, y)
	game.cell[y][x] = n
	game.cellLabel[y][x].SetText(fmt.Sprintf("%d", n))
	clearNotes(x, y)
	removeOtherNotes(x, y, n)
	if isSolved() {
		winnerWinner()
	}
}

func clearNotes(x, y int) {
	assertNonConstant(x, y)
	for i := 1; i <= game.size; i++ {
		game.note[y][x][i-1] = false
	}
	setNoteLabel(x, y, false)
}

func clearAll(x, y int) {
	assertNonConstant(x, y)
	game.cell[y][x] = 0
	game.cellLabel[y][x].SetText("")
	clearNotes(x, y)
}

func removeOtherNotes(x, y, n int) {
	assertNonConstant(x, y)
	for i := 1; i <= game.size; i++ {
		removeNote(i-1, y, n)
		removeNote(x, i-1, n)
	}
}

// For simplicity tHis can be called on constant cells too.
func removeNote(x, y, n int) {
	if !isConstant(x, y) {
		game.note[y][x][n-1] = false
		setNoteLabel(x, y, true)
	}
}

func updateNote(x, y, n int) {
	assertNonConstant(x, y)
	// Don't allow note if the cell is already set,
	// or if that value is already present in the row or column.
	if game.cell[y][x] != 0 || inRowOrColumn(x, y, n) {
		return
	}
	game.note[y][x][n-1] = !game.note[y][x][n-1]
	setNoteLabel(x, y, false)
}

func inRowOrColumn(x, y, n int) bool {
	for i := 1; i <= game.size; i++ {
		if game.cell[y][i-1] == n || game.cell[i-1][x] == n {
			return true
		}
	}
	return false
}

func setNoteLabel(x, y int, promote bool) {
	assertNonConstant(x, y)
	notes := ""
	p := 0
	count := 0
	for i := 1; i <= game.size; i++ {
		if game.note[y][x][i-1] {
			notes += fmt.Sprintf("%d", i)
			p = i
			count++
		}
	}
	if autoPromote && count == 1 && promote {
		updateCell(x, y, p)
	} else {
		game.noteLabel[y][x].SetText(notes)
	}
}

func isSolved() bool {
	for j := 0; j < game.size; j++ {
		for i := 0; i < game.size; i++ {
			if game.cell[j][i] != game.k.Answer[j][i] {
				return false
			}
		}
	}
	return true
}

func drawCages() {
	a := game.table.GetAllocation()
	w := a.Width / game.size
	h := a.Height / game.size
	d := game.table.GetWindow().GetDrawable()
	gc := gdk.NewGC(d)
	gc.SetRgbFgColor(white)

	t := lineWidth

	d.DrawRectangle(gc, true, 0, 0, t, a.Height)
	d.DrawRectangle(gc, true, 0, 0, a.Width, t)
	d.DrawRectangle(gc, true, 0, a.Height-t, a.Width, t)
	d.DrawRectangle(gc, true, a.Width-t, 0, t, a.Height)

	for j := 0; j < game.size; j++ {
		y := j*h - t/2
		for i := 0; i < game.size; i++ {
			x := i*w - t/2
			if i > 0 && game.k.Vertical[j][i-1] {
				d.DrawRectangle(gc, true, x, y, t, h+t/2)
			}
			if j > 0 && game.k.Horizontal[i][j-1] {
				d.DrawRectangle(gc, true, x, y, w+t/2, t)
			}
		}
	}
}

func winnerWinner() {
	dialog := gtk.NewDialog()
	dialog.SetSizeRequest(200, 50)
	dialog.SetTitle("Congratulations!")
	button := gtk.NewButtonWithLabel("Done")
	button.Connect("clicked", gtk.MainQuit)
	dialog.GetVBox().Add(button)
	dialog.ShowAll()
}

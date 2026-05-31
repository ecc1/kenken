# KenKen in the browser

A static, dependency-free browser version of the GTK app in
[`../cmd/kenken`](../cmd/kenken). It plays puzzles in the same `A/T/S/V/H` text
format read by [`../read.go`](../read.go) — the files downloaded by
[`../cmd/kenken/play-kenken.sh`](../cmd/kenken/play-kenken.sh).

## Running

Just open `index.html` in a browser — no server or build step required. Then:

- **Open file** to load a downloaded puzzle `.txt`, or drag-and-drop it onto the
  page (or click **Demo** for a bundled 6×6).
- Tap/click a cell, then a number (on-screen keypad or the keyboard).
- **Notes** toggles pencil-mark mode; on a keyboard, hold Shift/Ctrl with a
  number instead. A note left alone in a cell is auto-promoted to the answer.
- **Candidates** fills in every number still possible for the selected cell.
- **Clear** empties the selected cell (or Space/Backspace/Delete).
- **Restart** clears your entries (or right-click the board).

Works with a desktop keyboard/mouse and with touch on mobile; the layout reflows
between portrait and landscape.

## Editing

The app is written in TypeScript (`app.ts`) and compiled to a plain classic
script (`app.js`), which is committed so the page runs with no build. After
changing `app.ts`, recompile:

```sh
tsc -p tsconfig.json
```

## Out of scope

Puzzle *generation* (the `sgt-keen`/`keensolver` binaries used by the Go default
path) needs native executables and isn't available in the browser. `loadText` is
the single ingestion point, so a small server endpoint serving generated puzzles
(as this text format or JSON) could be added later without reworking the app.

## Samples

`samples/sample3.txt`, `samples/sample6.txt` and `samples/sample8.txt` are valid
puzzles in the file format (transcribed from the test data in
[`../sgt_test.go`](../sgt_test.go)); they load in the app and also render with
`go run ./cmd/showkenken samples/sample6.txt`.

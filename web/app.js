"use strict";
// KenKen in the browser.
//
// A static, dependency-free port of the GTK app in ../cmd/kenken.  It loads a
// puzzle file in the A/T/S/V/H text format (see ../read.go), renders it on a
// <canvas>, and supports interactive play with either a desktop keyboard/mouse
// or an on-screen keypad for touch devices.
//
// The code is organized into sections, each porting one Go source file:
//   parser    <- read.go
//   model     <- kenken.go
//   Game      <- cmd/kenken/game.go
//   Renderer  <- cmd/kenken/ui.go (drawCell)
//   input/UI  <- cmd/kenken/event.go + ui.go dialogs + cmd/kenken/main.go
// ops maps an Operation to its display glyph (<- kenken.go `ops`).
const OPS = ["?", "?", "+", "−", "×", "∕"]; // ?, ?, +, minus, times, division slash
function symbol(op) {
    return OPS[op];
}
function puzzleSize(p) {
    return p.answer.length;
}
function isConstant(p, x, y) {
    return p.operation[y][x] === 1 /* Op.Given */;
}
function clueString(p, x, y) {
    const op = p.operation[y][x];
    if (op === 0 /* Op.None */ || op === 1 /* Op.Given */) {
        return "";
    }
    return `${p.clue[y][x]} ${symbol(op)}`;
}
// ---------------------------------------------------------------------------
// Parser (<- read.go)
// ---------------------------------------------------------------------------
class MalformedPuzzleError extends Error {
    constructor() {
        super("malformed KenKen file");
        this.name = "MalformedPuzzleError";
    }
}
// Scanner yields lines like Go's bufio.Scanner.
class Scanner {
    constructor(text) {
        this.i = 0;
        this.lines = text.split(/\r?\n/);
    }
    scan() {
        if (this.i >= this.lines.length) {
            return null;
        }
        return this.lines[this.i++];
    }
}
function fields(line) {
    const t = line.trim();
    return t === "" ? [] : t.split(/\s+/);
}
// readMatrix ports read.go readMatrix.  For square matrices nRows == nCols;
// otherwise (V/H) nRows == nCols+1, so an N x (N-1) matrix is read.
function readMatrix(s, label, square) {
    const head = s.scan();
    if (head === null || head.trim() !== label) {
        return null;
    }
    const first = s.scan();
    if (first === null) {
        return null;
    }
    const line = fields(first);
    const nCols = line.length;
    if (nCols === 0) {
        return null;
    }
    const nRows = square ? nCols : nCols + 1;
    const m = new Array(nRows);
    m[0] = line;
    for (let i = 1; i < nRows; i++) {
        const t = s.scan();
        if (t === null) {
            return null;
        }
        m[i] = fields(t);
        if (m[i].length !== nCols) {
            return null;
        }
    }
    return m;
}
function intMatrix(s, label) {
    const m = readMatrix(s, label, true);
    if (m === null) {
        return null;
    }
    return m.map((row) => row.map((t) => {
        const v = Number(t);
        if (!Number.isInteger(v)) {
            throw new MalformedPuzzleError();
        }
        return v;
    }));
}
function boolMatrix(s, label) {
    const m = readMatrix(s, label, false);
    if (m === null) {
        return null;
    }
    return m.map((row) => row.map((t) => {
        switch (t) {
            case "0":
                return false;
            case "1":
                return true;
            default:
                throw new MalformedPuzzleError();
        }
    }));
}
function parseOperation(s) {
    switch (s) {
        case "0":
            return 0 /* Op.None */;
        case "1":
            return 1 /* Op.Given */;
        case "+":
            return 2 /* Op.Sum */;
        case "-":
            return 3 /* Op.Difference */;
        case "*":
            return 4 /* Op.Product */;
        case "/":
            return 5 /* Op.Quotient */;
        default:
            throw new MalformedPuzzleError();
    }
}
function opMatrix(s, label) {
    const m = readMatrix(s, label, true);
    if (m === null) {
        return null;
    }
    return m.map((row) => row.map(parseOperation));
}
// parsePuzzleText ports read.go Read.  Throws MalformedPuzzleError on bad input.
function parsePuzzleText(text) {
    const s = new Scanner(text);
    const answer = intMatrix(s, "A");
    const clue = intMatrix(s, "T");
    const operation = opMatrix(s, "S");
    const vertical = boolMatrix(s, "V");
    const horizontal = boolMatrix(s, "H");
    if (!answer || !clue || !operation || !vertical || !horizontal) {
        throw new MalformedPuzzleError();
    }
    return { answer, clue, operation, vertical, horizontal };
}
// ---------------------------------------------------------------------------
// Game state (<- cmd/kenken/game.go)
// ---------------------------------------------------------------------------
const autoPromote = true;
function clamp(v, lo, hi) {
    return v < lo ? lo : v > hi ? hi : v;
}
class Game {
    constructor(puzzle, hooks) {
        this.puzzle = puzzle;
        this.hooks = hooks;
        this.cells = [];
        this.notes = [];
        this.noteStrings = [];
        this.selected = null;
        this.notesMode = false;
        this.size = puzzleSize(puzzle);
        this.initGame();
    }
    initGame() {
        for (let y = 0; y < this.size; y++) {
            this.cells[y] = new Array(this.size).fill(0);
            this.notes[y] = [];
            this.noteStrings[y] = new Array(this.size).fill("");
            for (let x = 0; x < this.size; x++) {
                if (isConstant(this.puzzle, x, y)) {
                    this.cells[y][x] = this.puzzle.answer[y][x];
                }
                this.notes[y][x] = new Array(this.size).fill(false);
            }
        }
    }
    constant(x, y) {
        return isConstant(this.puzzle, x, y);
    }
    assertNonConstant(x, y) {
        if (this.constant(x, y)) {
            throw new Error(`modifying constant cell (${x}, ${y})`);
        }
    }
    changed() {
        this.hooks.onChange();
    }
    addNotes(x, y) {
        this.assertNonConstant(x, y);
        for (let i = 1; i <= this.size; i++) {
            if (this.cells[y][x] === 0 && !this.inRowOrColumn(x, y, i)) {
                this.notes[y][x][i - 1] = true;
            }
        }
        this.updateNoteLabel(x, y, true);
    }
    updateCell(x, y, n) {
        this.assertNonConstant(x, y);
        this.cells[y][x] = n;
        this.clearNotes(x, y);
        this.removeOtherNotes(x, y, n);
        this.changed();
        const [done, correct] = this.gameStatus();
        if (done) {
            if (correct) {
                this.hooks.onWin();
            }
            else {
                this.hooks.onLose();
            }
        }
    }
    clearNotes(x, y) {
        this.assertNonConstant(x, y);
        for (let i = 1; i <= this.size; i++) {
            this.notes[y][x][i - 1] = false;
        }
        this.updateNoteLabel(x, y, false);
    }
    clearAll(x, y) {
        this.assertNonConstant(x, y);
        this.cells[y][x] = 0;
        this.clearNotes(x, y);
    }
    removeOtherNotes(x, y, n) {
        this.assertNonConstant(x, y);
        for (let i = 1; i <= this.size; i++) {
            this.removeNote(i - 1, y, n);
            this.removeNote(x, i - 1, n);
        }
    }
    // Safe to call on constant cells (it no-ops), like read.go removeNote.
    removeNote(x, y, n) {
        if (!this.constant(x, y)) {
            this.notes[y][x][n - 1] = false;
            this.updateNoteLabel(x, y, true);
        }
    }
    updateNote(x, y, n) {
        this.assertNonConstant(x, y);
        if (this.cells[y][x] !== 0 || this.inRowOrColumn(x, y, n)) {
            return;
        }
        this.notes[y][x][n - 1] = !this.notes[y][x][n - 1];
        this.updateNoteLabel(x, y, false);
        this.changed();
    }
    inRowOrColumn(x, y, n) {
        for (let i = 1; i <= this.size; i++) {
            if (this.cells[y][i - 1] === n || this.cells[i - 1][x] === n) {
                return true;
            }
        }
        return false;
    }
    updateNoteLabel(x, y, promote) {
        this.assertNonConstant(x, y);
        let noteStr = "";
        let p = 0;
        let count = 0;
        for (let i = 1; i <= this.size; i++) {
            if (this.notes[y][x][i - 1]) {
                noteStr += String(i);
                p = i;
                count++;
            }
        }
        if (autoPromote && count === 1 && promote) {
            this.updateCell(x, y, p);
        }
        else {
            this.noteStrings[y][x] = noteStr;
        }
        this.changed();
    }
    gameStatus() {
        let correct = true;
        for (let j = 0; j < this.size; j++) {
            for (let i = 0; i < this.size; i++) {
                if (this.cells[j][i] === 0) {
                    return [false, correct];
                }
                if (this.cells[j][i] !== this.puzzle.answer[j][i]) {
                    correct = false;
                }
            }
        }
        return [true, correct];
    }
    restartGame() {
        for (let y = 0; y < this.size; y++) {
            for (let x = 0; x < this.size; x++) {
                if (!this.constant(x, y)) {
                    this.clearAll(x, y);
                }
            }
        }
        this.changed();
    }
    // --- High-level intents shared by keyboard, pointer and on-screen keypad ---
    select(x, y) {
        this.selected = { x, y };
        this.changed();
    }
    moveSelection(dx, dy) {
        if (!this.selected) {
            this.selected = { x: 0, y: 0 };
        }
        else {
            this.selected = {
                x: clamp(this.selected.x + dx, 0, this.size - 1),
                y: clamp(this.selected.y + dy, 0, this.size - 1),
            };
        }
        this.changed();
    }
    // enterDigit sets the selected cell, or toggles a note when forceNote (a
    // keyboard modifier) or notesMode (the on-screen toggle) is active.
    enterDigit(n, forceNote = false) {
        const sel = this.selected;
        if (!sel || this.constant(sel.x, sel.y) || n < 1 || n > this.size) {
            return;
        }
        if (forceNote || this.notesMode) {
            this.updateNote(sel.x, sel.y, n);
        }
        else {
            this.updateCell(sel.x, sel.y, n);
        }
    }
    clearSelected() {
        const sel = this.selected;
        if (!sel || this.constant(sel.x, sel.y)) {
            return;
        }
        this.clearAll(sel.x, sel.y);
        this.changed();
    }
    addCandidates() {
        const sel = this.selected;
        if (!sel || this.constant(sel.x, sel.y)) {
            return;
        }
        this.addNotes(sel.x, sel.y);
    }
    toggleNotesMode() {
        this.notesMode = !this.notesMode;
        return this.notesMode;
    }
}
// ---------------------------------------------------------------------------
// Renderer (<- cmd/kenken/ui.go drawCell)
// ---------------------------------------------------------------------------
// These are relative to a unit-square cell size (<- ui.go).
const lineWidth = 0.025;
const innerSep = 0.05;
const largeFontSize = 0.4;
const smallFontSize = 0.16;
const textFont = '"DejaVu Sans", sans-serif';
class Renderer {
    constructor(canvas, wrap, game) {
        this.canvas = canvas;
        this.wrap = wrap;
        this.game = game;
        this.boardPx = 0;
        this.cellSize = 0; // CSS pixels per cell
        const ctx = canvas.getContext("2d");
        if (!ctx) {
            throw new Error("2d canvas context unavailable");
        }
        this.ctx = ctx;
    }
    layout() {
        const dpr = window.devicePixelRatio || 1;
        const avail = Math.min(this.wrap.clientWidth, this.wrap.clientHeight);
        this.boardPx = Math.max(1, Math.floor(avail));
        this.cellSize = this.boardPx / this.game.size;
        this.canvas.style.width = `${this.boardPx}px`;
        this.canvas.style.height = `${this.boardPx}px`;
        this.canvas.width = Math.floor(this.boardPx * dpr);
        this.canvas.height = Math.floor(this.boardPx * dpr);
        this.ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    }
    // cellAt maps CSS pixel coordinates (relative to the canvas) to a cell.
    cellAt(px, py) {
        const x = Math.floor(px / this.cellSize);
        const y = Math.floor(py / this.cellSize);
        if (x < 0 || x >= this.game.size || y < 0 || y >= this.game.size) {
            return null;
        }
        return { x, y };
    }
    draw() {
        const ctx = this.ctx;
        const g = this.game;
        ctx.clearRect(0, 0, this.boardPx, this.boardPx);
        ctx.fillStyle = "#ffffff";
        ctx.fillRect(0, 0, this.boardPx, this.boardPx);
        // Selection highlight (no GTK equivalent; needed without per-cell focus).
        const sel = g.selected;
        if (sel) {
            ctx.fillStyle = g.notesMode ? "#ffe7b3" : "#cfe8ff";
            ctx.fillRect(sel.x * this.cellSize, sel.y * this.cellSize, this.cellSize, this.cellSize);
        }
        for (let y = 0; y < g.size; y++) {
            for (let x = 0; x < g.size; x++) {
                this.drawCell(x, y);
            }
        }
    }
    drawCell(x, y) {
        const ctx = this.ctx;
        const g = this.game;
        const p = g.puzzle;
        const cs = this.cellSize;
        const ox = x * cs;
        const oy = y * cs;
        ctx.strokeStyle = "#000000";
        ctx.fillStyle = "#000000";
        ctx.lineWidth = lineWidth * cs;
        ctx.lineCap = "square";
        // Cage lines (same conditionals as ui.go, including the H transpose).
        ctx.beginPath();
        if (x === 0 || p.vertical[y][x - 1]) {
            ctx.moveTo(ox, oy);
            ctx.lineTo(ox, oy + cs);
        }
        if (x === g.size - 1) {
            ctx.moveTo(ox + cs, oy);
            ctx.lineTo(ox + cs, oy + cs);
        }
        if (y === 0 || p.horizontal[x][y - 1]) {
            ctx.moveTo(ox, oy);
            ctx.lineTo(ox + cs, oy);
        }
        if (y === g.size - 1) {
            ctx.moveTo(ox, oy + cs);
            ctx.lineTo(ox + cs, oy + cs);
        }
        ctx.stroke();
        // Answer (centered).
        const n = g.cells[y][x];
        if (n !== 0) {
            ctx.font = `${largeFontSize * cs}px ${textFont}`;
            ctx.textAlign = "center";
            ctx.textBaseline = "middle";
            ctx.fillText(String(n), ox + cs / 2, oy + cs / 2);
        }
        if (isConstant(p, x, y)) {
            return;
        }
        // Clue (top-left).
        const clue = clueString(p, x, y);
        if (clue !== "") {
            ctx.font = `${smallFontSize * cs}px ${textFont}`;
            ctx.textAlign = "left";
            ctx.textBaseline = "top";
            ctx.fillText(clue, ox + innerSep * cs, oy + innerSep * cs);
        }
        // Notes (bottom-right).
        const noteStr = g.noteStrings[y][x];
        if (noteStr !== "") {
            ctx.font = `${smallFontSize * cs}px ${textFont}`;
            ctx.textAlign = "right";
            ctx.textBaseline = "alphabetic";
            ctx.fillText(noteStr, ox + cs - innerSep * cs, oy + cs - innerSep * cs);
        }
    }
}
// ---------------------------------------------------------------------------
// Input + UI wiring (<- event.go, ui.go dialogs, main.go)
// ---------------------------------------------------------------------------
// keycode maps a KeyboardEvent.key to a digit (<- event.go keycode).  Space,
// Backspace and Delete map to 0 (clear).  Shifted US symbols map to their digit
// so note-toggling works on a US layout.
const KEYCODE = {
    " ": 0,
    Backspace: 0,
    Delete: 0,
    "1": 1, "!": 1,
    "2": 2, "@": 2,
    "3": 3, "#": 3,
    "4": 4, "$": 4,
    "5": 5, "%": 5,
    "6": 6, "^": 6,
    "7": 7, "&": 7,
    "8": 8, "*": 8,
    "9": 9, "(": 9,
};
let game = null;
let renderer = null;
// Coalesce redraws across a cascade of state changes (<- per-cell QueueDraw).
let drawScheduled = false;
function scheduleDraw() {
    if (drawScheduled) {
        return;
    }
    drawScheduled = true;
    requestAnimationFrame(() => {
        drawScheduled = false;
        renderer?.draw();
    });
}
// Win/lose dialogs (<- ui.go winnerWinner / tryAgain), with the same
// re-entrancy guards.
const ui = {
    winning: false,
    losing: false,
    winnerWinner() {
        if (this.winning) {
            return;
        }
        this.winning = true;
        showModal("Correct!", [{ label: "OK", action: () => { } }]);
    },
    tryAgain() {
        if (this.losing) {
            return;
        }
        this.losing = true;
        showModal("Try again?", [
            { label: "Yes", action: () => game?.restartGame() },
            { label: "No", action: () => { } },
        ], () => {
            this.losing = false;
        });
    },
};
const hooks = {
    onChange: scheduleDraw,
    onWin: () => ui.winnerWinner(),
    onLose: () => ui.tryAgain(),
};
function byId(id) {
    const e = document.getElementById(id);
    if (!e) {
        throw new Error(`missing element #${id}`);
    }
    return e;
}
function buildKeypad(size) {
    const digits = byId("digits");
    digits.innerHTML = "";
    for (let n = 1; n <= size; n++) {
        const b = document.createElement("button");
        b.className = "key digit";
        b.type = "button";
        b.textContent = String(n);
        b.addEventListener("click", () => game?.enterDigit(n));
        digits.appendChild(b);
    }
}
function startGame(puzzle, name) {
    const canvas = byId("board");
    const wrap = byId("board-wrap");
    game = new Game(puzzle, hooks);
    renderer = new Renderer(canvas, wrap, game);
    buildKeypad(game.size);
    byId("notes-toggle").classList.remove("active");
    renderer.layout();
    scheduleDraw();
    document.title = name ? `KenKen — ${name}` : "KenKen";
    setStatus(name ? `Playing ${name}` : "Puzzle loaded", false);
}
function loadText(text, name) {
    try {
        const puzzle = parsePuzzleText(text);
        startGame(puzzle, name);
    }
    catch (err) {
        const msg = err instanceof Error ? err.message : String(err);
        setStatus(`Could not load ${name || "puzzle"}: ${msg}`, true);
    }
}
function readFile(file) {
    const reader = new FileReader();
    reader.onload = () => loadText(String(reader.result), file.name);
    reader.onerror = () => setStatus(`Could not read ${file.name}`, true);
    reader.readAsText(file);
}
function setStatus(text, isError) {
    const status = byId("status");
    status.textContent = text;
    status.classList.toggle("error", isError);
}
function showModal(message, buttons, onClose) {
    const overlay = document.createElement("div");
    overlay.className = "modal-overlay";
    const box = document.createElement("div");
    box.className = "modal";
    const msg = document.createElement("p");
    msg.className = "modal-message";
    msg.textContent = message;
    box.appendChild(msg);
    const row = document.createElement("div");
    row.className = "modal-buttons";
    for (const b of buttons) {
        const btn = document.createElement("button");
        btn.className = "key";
        btn.type = "button";
        btn.textContent = b.label;
        btn.addEventListener("click", () => {
            document.body.removeChild(overlay);
            b.action();
            onClose?.();
        });
        row.appendChild(btn);
    }
    box.appendChild(row);
    overlay.appendChild(box);
    document.body.appendChild(overlay);
}
function onKeyDown(e) {
    if (!game) {
        return;
    }
    switch (e.key) {
        case "ArrowLeft":
            game.moveSelection(-1, 0);
            e.preventDefault();
            return;
        case "ArrowRight":
            game.moveSelection(1, 0);
            e.preventDefault();
            return;
        case "ArrowUp":
            game.moveSelection(0, -1);
            e.preventDefault();
            return;
        case "ArrowDown":
            game.moveSelection(0, 1);
            e.preventDefault();
            return;
    }
    const n = KEYCODE[e.key];
    if (n === undefined) {
        return;
    }
    e.preventDefault();
    if (n === 0) {
        game.clearSelected();
        return;
    }
    // A modifier means "note", mirroring event.go's ek.State() != 0 check.
    game.enterDigit(n, e.shiftKey || e.ctrlKey);
}
function init() {
    const canvas = byId("board");
    const fileInput = byId("file-input");
    const dropZone = byId("app");
    // Pointer selection (mouse + touch unified).
    canvas.addEventListener("pointerdown", (e) => {
        if (!game || !renderer) {
            return;
        }
        const rect = canvas.getBoundingClientRect();
        const cell = renderer.cellAt(e.clientX - rect.left, e.clientY - rect.top);
        if (cell) {
            game.select(cell.x, cell.y);
        }
    });
    // Right-click restarts, for parity with event.go button 3.
    canvas.addEventListener("contextmenu", (e) => {
        e.preventDefault();
        ui.tryAgain();
    });
    // On-screen controls.
    byId("notes-toggle").addEventListener("click", (e) => {
        if (!game) {
            return;
        }
        e.currentTarget.classList.toggle("active", game.toggleNotesMode());
        scheduleDraw();
    });
    byId("clear-btn").addEventListener("click", () => game?.clearSelected());
    byId("candidates-btn").addEventListener("click", () => game?.addCandidates());
    byId("restart-btn").addEventListener("click", () => ui.tryAgain());
    byId("demo-btn").addEventListener("click", () => loadText(DEMO_PUZZLE, "demo 6×6"));
    // File loading.
    fileInput.addEventListener("change", () => {
        const f = fileInput.files?.[0];
        if (f) {
            readFile(f);
        }
        fileInput.value = "";
    });
    dropZone.addEventListener("dragover", (e) => {
        e.preventDefault();
        dropZone.classList.add("dragging");
    });
    dropZone.addEventListener("dragleave", () => dropZone.classList.remove("dragging"));
    dropZone.addEventListener("drop", (e) => {
        e.preventDefault();
        dropZone.classList.remove("dragging");
        const f = e.dataTransfer?.files?.[0];
        if (f) {
            readFile(f);
        }
    });
    window.addEventListener("keydown", onKeyDown);
    window.addEventListener("resize", () => {
        if (renderer) {
            renderer.layout();
            scheduleDraw();
        }
    });
    setStatus("Open a downloaded KenKen .txt file, or try the demo.", false);
}
// A bundled 6x6 puzzle so the app is playable without finding a file first
// (matches web/samples/sample6.txt).
const DEMO_PUZZLE = `A
6 2 4 1 3 5
3 4 5 6 1 2
2 1 3 5 6 4
4 6 1 2 5 3
1 5 2 3 4 6
5 3 6 4 2 1
T
3 0 24 0 3 3
12 0 8 0 0 0
6 0 0 17 0 0
0 6 0 0 2 0
6 3 0 1 2 6
0 2 0 0 0 0
S
/ 0 * 0 / -
* 0 + 0 0 0
+ 0 0 + 0 0
0 * 0 0 - 0
+ - 0 - / *
0 / 0 0 0 0
V
0 1 0 1 1
0 1 1 1 1
1 1 1 0 0
1 0 1 1 0
1 0 1 1 1
1 0 1 1 1
H
1 1 0 1 0
1 0 1 1 1
1 0 1 1 1
0 1 0 1 0
0 1 1 1 0
0 1 1 1 0
`;
if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
}
else {
    init();
}

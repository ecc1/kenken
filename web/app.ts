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

// ---------------------------------------------------------------------------
// Model (<- kenken.go)
// ---------------------------------------------------------------------------

const enum Op {
	None = 0,
	Given = 1,
	Sum = 2,
	Difference = 3,
	Product = 4,
	Quotient = 5,
}

interface Puzzle {
	answer: number[][]; // N x N
	clue: number[][]; // N x N
	operation: Op[][]; // N x N
	// N rows of N-1 columns.  vertical[y][x] is a heavy line between (x,y) and (x+1,y).
	vertical: boolean[][];
	// N rows of N-1 columns, transposed: horizontal[x][y] is a heavy line between (x,y) and (x,y+1).
	horizontal: boolean[][];
}

// ops maps an Operation to its display glyph (<- kenken.go `ops`).
const OPS = ["?", "?", "+", "−", "×", "∕"]; // ?, ?, +, minus, times, division slash

function symbol(op: Op): string {
	return OPS[op];
}

function puzzleSize(p: Puzzle): number {
	return p.answer.length;
}

function isConstant(p: Puzzle, x: number, y: number): boolean {
	return p.operation[y][x] === Op.Given;
}

function clueString(p: Puzzle, x: number, y: number): string {
	const op = p.operation[y][x];
	if (op === Op.None || op === Op.Given) {
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
	private lines: string[];
	private i = 0;
	constructor(text: string) {
		this.lines = text.split(/\r?\n/);
	}
	scan(): string | null {
		if (this.i >= this.lines.length) {
			return null;
		}
		return this.lines[this.i++];
	}
}

function fields(line: string): string[] {
	const t = line.trim();
	return t === "" ? [] : t.split(/\s+/);
}

// readMatrix ports read.go readMatrix.  For square matrices nRows == nCols;
// otherwise (V/H) nRows == nCols+1, so an N x (N-1) matrix is read.
function readMatrix(s: Scanner, label: string, square: boolean): string[][] | null {
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
	const m: string[][] = new Array(nRows);
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

function intMatrix(s: Scanner, label: string): number[][] | null {
	const m = readMatrix(s, label, true);
	if (m === null) {
		return null;
	}
	return m.map((row) =>
		row.map((t) => {
			const v = Number(t);
			if (!Number.isInteger(v)) {
				throw new MalformedPuzzleError();
			}
			return v;
		}),
	);
}

function boolMatrix(s: Scanner, label: string): boolean[][] | null {
	const m = readMatrix(s, label, false);
	if (m === null) {
		return null;
	}
	return m.map((row) =>
		row.map((t) => {
			switch (t) {
				case "0":
					return false;
				case "1":
					return true;
				default:
					throw new MalformedPuzzleError();
			}
		}),
	);
}

function parseOperation(s: string): Op {
	switch (s) {
		case "0":
			return Op.None;
		case "1":
			return Op.Given;
		case "+":
			return Op.Sum;
		case "-":
			return Op.Difference;
		case "*":
			return Op.Product;
		case "/":
			return Op.Quotient;
		default:
			throw new MalformedPuzzleError();
	}
}

function opMatrix(s: Scanner, label: string): Op[][] | null {
	const m = readMatrix(s, label, true);
	if (m === null) {
		return null;
	}
	return m.map((row) => row.map(parseOperation));
}

// parsePuzzleText ports read.go Read.  Throws MalformedPuzzleError on bad input.
function parsePuzzleText(text: string): Puzzle {
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

interface GameHooks {
	onChange: () => void;
	onWin: () => void;
	onLose: () => void;
}

interface Cell {
	x: number;
	y: number;
}

function clamp(v: number, lo: number, hi: number): number {
	return v < lo ? lo : v > hi ? hi : v;
}

class Game {
	readonly size: number;
	cells: number[][] = [];
	notes: boolean[][][] = [];
	noteStrings: string[][] = [];
	selected: Cell | null = null;
	notesMode = false;

	constructor(
		readonly puzzle: Puzzle,
		private hooks: GameHooks,
	) {
		this.size = puzzleSize(puzzle);
		this.initGame();
	}

	private initGame(): void {
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

	private constant(x: number, y: number): boolean {
		return isConstant(this.puzzle, x, y);
	}

	private assertNonConstant(x: number, y: number): void {
		if (this.constant(x, y)) {
			throw new Error(`modifying constant cell (${x}, ${y})`);
		}
	}

	private changed(): void {
		this.hooks.onChange();
	}

	addNotes(x: number, y: number): void {
		this.assertNonConstant(x, y);
		for (let i = 1; i <= this.size; i++) {
			if (this.cells[y][x] === 0 && !this.inRowOrColumn(x, y, i)) {
				this.notes[y][x][i - 1] = true;
			}
		}
		this.updateNoteLabel(x, y, true);
	}

	private updateCell(x: number, y: number, n: number): void {
		this.assertNonConstant(x, y);
		this.cells[y][x] = n;
		this.clearNotes(x, y);
		this.removeOtherNotes(x, y, n);
		this.changed();
		const [done, correct] = this.gameStatus();
		if (done) {
			if (correct) {
				this.hooks.onWin();
			} else {
				this.hooks.onLose();
			}
		}
	}

	private clearNotes(x: number, y: number): void {
		this.assertNonConstant(x, y);
		for (let i = 1; i <= this.size; i++) {
			this.notes[y][x][i - 1] = false;
		}
		this.updateNoteLabel(x, y, false);
	}

	private clearAll(x: number, y: number): void {
		this.assertNonConstant(x, y);
		this.cells[y][x] = 0;
		this.clearNotes(x, y);
	}

	private removeOtherNotes(x: number, y: number, n: number): void {
		this.assertNonConstant(x, y);
		for (let i = 1; i <= this.size; i++) {
			this.removeNote(i - 1, y, n);
			this.removeNote(x, i - 1, n);
		}
	}

	// Safe to call on constant cells (it no-ops), like read.go removeNote.
	private removeNote(x: number, y: number, n: number): void {
		if (!this.constant(x, y)) {
			this.notes[y][x][n - 1] = false;
			this.updateNoteLabel(x, y, true);
		}
	}

	private updateNote(x: number, y: number, n: number): void {
		this.assertNonConstant(x, y);
		if (this.cells[y][x] !== 0 || this.inRowOrColumn(x, y, n)) {
			return;
		}
		this.notes[y][x][n - 1] = !this.notes[y][x][n - 1];
		this.updateNoteLabel(x, y, false);
		this.changed();
	}

	private inRowOrColumn(x: number, y: number, n: number): boolean {
		for (let i = 1; i <= this.size; i++) {
			if (this.cells[y][i - 1] === n || this.cells[i - 1][x] === n) {
				return true;
			}
		}
		return false;
	}

	private updateNoteLabel(x: number, y: number, promote: boolean): void {
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
		} else {
			this.noteStrings[y][x] = noteStr;
		}
		this.changed();
	}

	private gameStatus(): [complete: boolean, correct: boolean] {
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

	restartGame(): void {
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

	select(x: number, y: number): void {
		this.selected = { x, y };
		this.changed();
	}

	moveSelection(dx: number, dy: number): void {
		if (!this.selected) {
			this.selected = { x: 0, y: 0 };
		} else {
			this.selected = {
				x: clamp(this.selected.x + dx, 0, this.size - 1),
				y: clamp(this.selected.y + dy, 0, this.size - 1),
			};
		}
		this.changed();
	}

	// enterDigit sets the selected cell, or toggles a note when forceNote (a
	// keyboard modifier) or notesMode (the on-screen toggle) is active.
	enterDigit(n: number, forceNote = false): void {
		const sel = this.selected;
		if (!sel || this.constant(sel.x, sel.y) || n < 1 || n > this.size) {
			return;
		}
		if (forceNote || this.notesMode) {
			this.updateNote(sel.x, sel.y, n);
		} else {
			this.updateCell(sel.x, sel.y, n);
		}
	}

	clearSelected(): void {
		const sel = this.selected;
		if (!sel || this.constant(sel.x, sel.y)) {
			return;
		}
		this.clearAll(sel.x, sel.y);
		this.changed();
	}

	addCandidates(): void {
		const sel = this.selected;
		if (!sel || this.constant(sel.x, sel.y)) {
			return;
		}
		this.addNotes(sel.x, sel.y);
	}

	toggleNotesMode(): boolean {
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

// Canvas colors aren't styleable through CSS, so the renderer carries its own
// palette.  It mirrors the light/dark page chrome in style.css and is selected
// from the system theme via prefers-color-scheme.
interface Palette {
	bg: string; // board background
	ink: string; // lines, digits, clues, notes
	selNormal: string; // selected-cell highlight (answer mode)
	selNotes: string; // selected-cell highlight (notes mode)
}

function palette(dark: boolean): Palette {
	return dark
		? { bg: "#1e1e1e", ink: "#e6e6e6", selNormal: "#2c4a66", selNotes: "#5c4a22" }
		: { bg: "#ffffff", ink: "#1a1a1a", selNormal: "#cfe8ff", selNotes: "#ffe7b3" };
}

// prefersDark reports the current system theme, defensively so the module can
// load in a non-browser context (e.g. the test harness) without a window.
function prefersDark(): boolean {
	return (
		typeof window !== "undefined" &&
		typeof window.matchMedia === "function" &&
		window.matchMedia("(prefers-color-scheme: dark)").matches
	);
}

class Renderer {
	private ctx: CanvasRenderingContext2D;
	private boardPx = 0;
	private colors: Palette = palette(prefersDark());
	cellSize = 0; // CSS pixels per cell

	constructor(
		private canvas: HTMLCanvasElement,
		private wrap: HTMLElement,
		private game: Game,
	) {
		const ctx = canvas.getContext("2d");
		if (!ctx) {
			throw new Error("2d canvas context unavailable");
		}
		this.ctx = ctx;
	}

	// setDark switches the palette for a system theme change.  The caller
	// schedules the redraw.
	setDark(dark: boolean): void {
		this.colors = palette(dark);
	}

	layout(): void {
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
	cellAt(px: number, py: number): Cell | null {
		const x = Math.floor(px / this.cellSize);
		const y = Math.floor(py / this.cellSize);
		if (x < 0 || x >= this.game.size || y < 0 || y >= this.game.size) {
			return null;
		}
		return { x, y };
	}

	draw(): void {
		const ctx = this.ctx;
		const g = this.game;
		ctx.clearRect(0, 0, this.boardPx, this.boardPx);
		ctx.fillStyle = this.colors.bg;
		ctx.fillRect(0, 0, this.boardPx, this.boardPx);
		// Selection highlight (no GTK equivalent; needed without per-cell focus).
		const sel = g.selected;
		if (sel) {
			ctx.fillStyle = g.notesMode ? this.colors.selNotes : this.colors.selNormal;
			ctx.fillRect(sel.x * this.cellSize, sel.y * this.cellSize, this.cellSize, this.cellSize);
		}
		for (let y = 0; y < g.size; y++) {
			for (let x = 0; x < g.size; x++) {
				this.drawCell(x, y);
			}
		}
	}

	private drawCell(x: number, y: number): void {
		const ctx = this.ctx;
		const g = this.game;
		const p = g.puzzle;
		const cs = this.cellSize;
		const ox = x * cs;
		const oy = y * cs;
		ctx.strokeStyle = this.colors.ink;
		ctx.fillStyle = this.colors.ink;
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
const KEYCODE: Record<string, number> = {
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

let game: Game | null = null;
let renderer: Renderer | null = null;

// Coalesce redraws across a cascade of state changes (<- per-cell QueueDraw).
let drawScheduled = false;
function scheduleDraw(): void {
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
	winnerWinner(): void {
		if (this.winning) {
			return;
		}
		this.winning = true;
		showModal("Correct!", [{ label: "OK", action: () => {} }]);
	},
	tryAgain(): void {
		if (this.losing) {
			return;
		}
		this.losing = true;
		showModal(
			"Try again?",
			[
				{ label: "Yes", action: () => game?.restartGame() },
				{ label: "No", action: () => {} },
			],
			() => {
				this.losing = false;
			},
		);
	},
};

const hooks: GameHooks = {
	onChange: scheduleDraw,
	onWin: () => ui.winnerWinner(),
	onLose: () => ui.tryAgain(),
};

function byId<T extends HTMLElement>(id: string): T {
	const e = document.getElementById(id);
	if (!e) {
		throw new Error(`missing element #${id}`);
	}
	return e as T;
}

function buildKeypad(size: number): void {
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

function startGame(puzzle: Puzzle, name: string): void {
	const canvas = byId<HTMLCanvasElement>("board");
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

function loadText(text: string, name: string): void {
	try {
		const puzzle = parsePuzzleText(text);
		startGame(puzzle, name);
	} catch (err) {
		const msg = err instanceof Error ? err.message : String(err);
		setStatus(`Could not load ${name || "puzzle"}: ${msg}`, true);
	}
}

function readFile(file: File): void {
	const reader = new FileReader();
	reader.onload = () => loadText(String(reader.result), file.name);
	reader.onerror = () => setStatus(`Could not read ${file.name}`, true);
	reader.readAsText(file);
}

function setStatus(text: string, isError: boolean): void {
	const status = byId("status");
	status.textContent = text;
	status.classList.toggle("error", isError);
}

function showModal(
	message: string,
	buttons: { label: string; action: () => void }[],
	onClose?: () => void,
): void {
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

function onKeyDown(e: KeyboardEvent): void {
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

function init(): void {
	const canvas = byId<HTMLCanvasElement>("board");
	const fileInput = byId<HTMLInputElement>("file-input");
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
		(e.currentTarget as HTMLElement).classList.toggle("active", game.toggleNotesMode());
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

	// Follow live system theme changes (CSS handles the chrome; the canvas
	// palette has to be repainted explicitly).
	const darkQuery = window.matchMedia("(prefers-color-scheme: dark)");
	darkQuery.addEventListener("change", (e) => {
		renderer?.setDark(e.matches);
		scheduleDraw();
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

// Guard auto-init so the module can be loaded for testing without a DOM.
if (typeof document !== "undefined") {
	if (document.readyState === "loading") {
		document.addEventListener("DOMContentLoaded", init);
	} else {
		init();
	}
}

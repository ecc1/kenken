// Dependency-free tests using the Node built-in test runner.
//
// app.js is a classic browser script with no exports, so we evaluate it in a
// vm context (with no DOM, so its guarded auto-init is skipped) and append a
// line that hands the top-level declarations back out for testing.

import { test } from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";
import vm from "node:vm";

const dir = dirname(fileURLToPath(import.meta.url));
const src = readFileSync(join(dir, "..", "app.js"), "utf8");

const ctx = { console };
vm.createContext(ctx);
vm.runInContext(
	`${src}\n;globalThis.__test = { palette, prefersDark, parsePuzzleText, puzzleSize, DEMO_PUZZLE };`,
	ctx,
);
const { palette, prefersDark, parsePuzzleText, puzzleSize, DEMO_PUZZLE } = ctx.__test;

// Relative luminance of a #rrggbb color (WCAG), used to assert contrast.
function luminance(hex) {
	const m = /^#([0-9a-f]{2})([0-9a-f]{2})([0-9a-f]{2})$/i.exec(hex);
	assert.ok(m, `not a #rrggbb color: ${hex}`);
	const lin = (c) => {
		const s = parseInt(c, 16) / 255;
		return s <= 0.03928 ? s / 12.92 : ((s + 0.055) / 1.055) ** 2.4;
	};
	return 0.2126 * lin(m[1]) + 0.7152 * lin(m[2]) + 0.0722 * lin(m[3]);
}

test("harness loaded the module without running auto-init", () => {
	assert.equal(typeof palette, "function");
	// Auto-init is DOM-guarded, so loading must not have thrown for a missing board.
	assert.equal(typeof parsePuzzleText, "function");
	assert.equal(puzzleSize(parsePuzzleText(DEMO_PUZZLE)), 6);
});

test("light palette has the expected values", () => {
	const p = palette(false);
	assert.equal(p.bg, "#ffffff");
	assert.equal(p.ink, "#1a1a1a");
	assert.equal(p.selNormal, "#cfe8ff");
	assert.equal(p.selNotes, "#ffe7b3");
});

test("dark palette differs from light on every role", () => {
	const d = palette(true);
	const l = palette(false);
	for (const k of ["bg", "ink", "selNormal", "selNotes"]) {
		assert.notEqual(d[k], l[k], `dark.${k} should differ from light.${k}`);
	}
});

test("dark theme inverts background/ink relative to light", () => {
	const d = palette(true);
	const l = palette(false);
	// Light: dark ink on light bg.  Dark: light ink on dark bg.
	assert.ok(luminance(l.bg) > luminance(l.ink), "light: bg brighter than ink");
	assert.ok(luminance(d.ink) > luminance(d.bg), "dark: ink brighter than bg");
	assert.ok(luminance(d.bg) < luminance(l.bg), "dark bg darker than light bg");
});

test("both palettes keep ink/bg readable (contrast)", () => {
	for (const dark of [false, true]) {
		const p = palette(dark);
		const hi = Math.max(luminance(p.bg), luminance(p.ink));
		const lo = Math.min(luminance(p.bg), luminance(p.ink));
		const ratio = (hi + 0.05) / (lo + 0.05);
		assert.ok(ratio >= 7, `${dark ? "dark" : "light"} ink/bg contrast ${ratio.toFixed(1)} < 7`);
	}
});

test("prefersDark is false without a window (defensive default)", () => {
	// The vm context has no window, exercising the typeof guard.
	assert.equal(prefersDark(), false);
});

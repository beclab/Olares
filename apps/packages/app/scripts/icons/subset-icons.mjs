#!/usr/bin/env node
/*
 * Material Symbols Rounded icon subsetting + missing-icon monitor.
 *
 * Source of truth is the full font shipped by the `material-symbols` npm
 * package. Both validation and subsetting are derived directly from that font
 * file (via HarfBuzz shaping + cmap), so there is no external codepoints map to
 * go stale -- a brand new icon works as soon as the `material-symbols` package
 * is recent enough.
 *
 * Modes:
 *   node scripts/icons/subset-icons.mjs           validate + write subset woff2
 *   node scripts/icons/subset-icons.mjs --check   validate only (CI / pre-commit)
 *
 * Validation fails the process (exit 1) when any referenced `sym_r_*` icon (or
 * an allowlisted dynamic icon) does not exist in the font, so missing icons are
 * caught before they ship instead of showing up as tofu online.
 */
import { createRequire } from 'module';
import { readFile, writeFile, mkdir } from 'fs/promises';
import { existsSync, readFileSync } from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import fg from 'fast-glob';
import {
	Blob as HbBlob,
	Face,
	Font,
	Buffer as HbBuffer,
	shape
} from 'harfbuzzjs';

const require = createRequire(import.meta.url);
const subsetFont = require('subset-font');
const fontverter = require('fontverter');

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const APP_ROOT = path.resolve(__dirname, '../..'); // packages/app
const SRC_DIR = path.join(APP_ROOT, 'src');
const ALLOWLIST_FILE = path.join(__dirname, 'icons.allowlist.txt');
const OUTPUT_FILE = path.join(
	SRC_DIR,
	'assets/fonts/MaterialSymbolsRounded.woff2'
);

const checkOnly = process.argv.includes('--check');

// Quasar references Material Symbols icons as `sym_r_<ligature>`; the ligature
// only contains lowercase letters, digits and underscores.
const ICON_RE = /sym_r_([a-z0-9_]+)/g;

// Glyphs required to form (input) the ligatures at render time: a-z, 0-9, `_`.
const LIGATURE_INPUT_CHARS = 'abcdefghijklmnopqrstuvwxyz0123456789_';

function resolveSourceFont() {
	const pkgJson = require.resolve('material-symbols/package.json');
	return path.join(path.dirname(pkgJson), 'material-symbols-rounded.woff2');
}

async function collectUsedIcons() {
	const files = await fg(['**/*.{vue,ts,tsx,js,jsx,html}'], {
		cwd: SRC_DIR,
		absolute: true
	});
	const locations = new Map(); // name -> string[] ("file:line")
	for (const file of files) {
		const content = await readFile(file, 'utf8');
		const lines = content.split('\n');
		for (let i = 0; i < lines.length; i++) {
			ICON_RE.lastIndex = 0;
			let m;
			while ((m = ICON_RE.exec(lines[i]))) {
				const name = m[1];
				if (!locations.has(name)) locations.set(name, []);
				const arr = locations.get(name);
				if (arr.length < 5)
					arr.push(`${path.relative(APP_ROOT, file)}:${i + 1}`);
			}
		}
	}
	return locations;
}

function readAllowlist() {
	if (!existsSync(ALLOWLIST_FILE)) return [];
	return readFileSync(ALLOWLIST_FILE, 'utf8')
		.split('\n')
		.map((l) => l.trim())
		.filter((l) => l && !l.startsWith('#'))
		.map((l) => l.replace(/^sym_r_/, ''));
}

// Build the HarfBuzz font + a glyph-id -> codepoint reverse map from the cmap.
function createFontContext(sfntBuffer) {
	const arrayBuffer = sfntBuffer.buffer.slice(
		sfntBuffer.byteOffset,
		sfntBuffer.byteOffset + sfntBuffer.byteLength
	);
	const blob = new HbBlob(arrayBuffer);
	const face = new Face(blob, 0);
	const font = new Font(face);

	const gidToCodepoint = new Map();
	const unicodes = face.collectUnicodes();
	for (const cp of unicodes) {
		const gid = font.nominalGlyph(cp);
		if (gid && !gidToCodepoint.has(gid)) gidToCodepoint.set(gid, cp);
	}
	return { font, gidToCodepoint };
}

// Shape a ligature name; a real icon collapses to exactly one non-.notdef glyph.
function shapeIcon(font, name) {
	const buf = new HbBuffer();
	buf.addText(name);
	buf.guessSegmentProperties();
	shape(font, buf);
	const infos = buf.getGlyphInfos();
	if (infos.length === 1 && infos[0].codepoint !== 0) {
		return infos[0].codepoint; // glyph id
	}
	return null;
}

async function main() {
	const sourceFont = resolveSourceFont();
	if (!existsSync(sourceFont)) {
		console.error(
			`[icons] source font not found: ${sourceFont}\nRun "npm install" so the "material-symbols" package is present.`
		);
		process.exit(1);
	}

	const sourceWoff2 = await readFile(sourceFont);
	const sfnt = await fontverter.convert(sourceWoff2, 'sfnt');

	const locations = await collectUsedIcons();
	for (const name of readAllowlist()) {
		if (!locations.has(name)) locations.set(name, ['(allowlist)']);
	}

	const names = [...locations.keys()].sort();
	const { font, gidToCodepoint } = createFontContext(sfnt);

	const missing = [];
	const iconCodepoints = new Set();
	for (const name of names) {
		const gid = shapeIcon(font, name);
		if (gid === null) {
			missing.push(name);
			continue;
		}
		const cp = gidToCodepoint.get(gid);
		if (cp === undefined) {
			// Icon glyph is only reachable via ligature (no PUA codepoint). Should
			// not happen for Material Symbols; treat as missing so it gets noticed.
			missing.push(name);
			continue;
		}
		iconCodepoints.add(cp);
	}

	if (missing.length) {
		console.error(
			`\n[icons] ${missing.length} icon(s) not found in Material Symbols Rounded:`
		);
		for (const name of missing) {
			console.error(`  - sym_r_${name}  (${locations.get(name).join(', ')})`);
		}
		console.error(
			'\nFix one of:\n' +
				'  - correct the icon name (typo / not a Material Symbols icon)\n' +
				'  - bump the "material-symbols" dependency if the font is outdated\n' +
				'  - add a valid dynamic icon name to scripts/icons/icons.allowlist.txt\n'
		);
		process.exit(1);
	}

	console.log(
		`[icons] validated ${names.length} icons against ${path.basename(
			sourceFont
		)} - all present.`
	);

	if (checkOnly) return;

	// Subset by the icons' own codepoints + the ligature input glyphs, with
	// no-layout-closure so we keep only the requested icons (not every ligature
	// reachable from a-z). Ligatures still form at render time because both the
	// input letter glyphs and the target icon glyphs are retained.
	const subsetText =
		[...iconCodepoints].map((cp) => String.fromCodePoint(cp)).join('') +
		LIGATURE_INPUT_CHARS;

	const subset = await subsetFont(sourceWoff2, subsetText, {
		targetFormat: 'woff2',
		noLayoutClosure: true
	});

	await mkdir(path.dirname(OUTPUT_FILE), { recursive: true });
	await writeFile(OUTPUT_FILE, subset);
	console.log(
		`[icons] wrote subset (${(subset.length / 1024).toFixed(1)} KB, ${
			iconCodepoints.size
		} icons) -> ${path.relative(APP_ROOT, OUTPUT_FILE)}`
	);
}

main().catch((err) => {
	console.error(err);
	process.exit(1);
});

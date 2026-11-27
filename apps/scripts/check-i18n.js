#!/usr/bin/env node
/* eslint-disable no-console */
/**
 * i18n hard-coded string checker
 *
 * Scans files for hard-coded user-facing English / Chinese text that has not
 * gone through the i18n layer (i.e. `t('...')`, `$t('...')`, `i18n.t('...')`).
 *
 * Usage:
 *   node scripts/check-i18n.js [--base <sha>] [--head <sha>] [file ...]
 *
 * When given explicit file paths, only those files are scanned. Otherwise the
 * script computes the changed file list from `git diff --name-only --diff-filter=ACMR <base>...<head>`.
 * Falls back to `git diff --cached` when neither flag is supplied (useful for
 * local pre-commit experiments).
 *
 * Inline escape hatches:
 *   - Add `// i18n-ignore` or `// i18n-disable-line` on the same line.
 *   - Add `<!-- i18n-ignore -->` on the same line inside Vue templates.
 *   - Add an entry to `.i18nignore` (gitignore-style globs) to skip whole files.
 *
 * Exits with code 1 when any violation is found. Exits 0 otherwise.
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const ROOT = process.cwd();

// ---------- CLI parsing ----------
const argv = process.argv.slice(2);
let baseRef = process.env.I18N_CHECK_BASE || '';
let headRef = process.env.I18N_CHECK_HEAD || '';
const explicitFiles = [];
let quiet = false;
let mode = 'auto'; // auto | files | diff | staged
let filesFrom = '';
for (let i = 0; i < argv.length; i += 1) {
	const a = argv[i];
	if (a === '--base') {
		baseRef = argv[++i];
		mode = 'diff';
	} else if (a === '--head') {
		headRef = argv[++i];
		mode = 'diff';
	} else if (a === '--staged') {
		mode = 'staged';
	} else if (a === '--files-from') {
		filesFrom = argv[++i];
		mode = 'files';
	} else if (a === '--quiet') {
		quiet = true;
	} else if (a === '--help' || a === '-h') {
		console.log(
			'Usage: node scripts/check-i18n.js [--base <sha>] [--head <sha>] [--staged] [--files-from <path>] [file ...]'
		);
		process.exit(0);
	} else {
		explicitFiles.push(a);
		mode = 'files';
	}
}

if (filesFrom) {
	try {
		const content = fs.readFileSync(filesFrom, 'utf8');
		for (const ln of content.split(/\r?\n/)) {
			const trimmed = ln.trim();
			if (trimmed) explicitFiles.push(trimmed);
		}
	} catch (e) {
		console.error(
			`[i18n-check] could not read --files-from ${filesFrom}: ${e.message}`
		);
		process.exit(2);
	}
}

// ---------- ignore list ----------
function loadIgnoreList() {
	const ignorePath = path.join(ROOT, '.i18nignore');
	if (!fs.existsSync(ignorePath)) return [];
	return fs
		.readFileSync(ignorePath, 'utf8')
		.split('\n')
		.map((l) => l.trim())
		.filter((l) => l && !l.startsWith('#'));
}

function globToRegex(glob) {
	// Minimal gitignore-style glob to RegExp: supports **, *, ?.
	let re = '^';
	for (let i = 0; i < glob.length; i += 1) {
		const c = glob[i];
		if (c === '*') {
			if (glob[i + 1] === '*') {
				re += '.*';
				i += 1;
				if (glob[i + 1] === '/') i += 1;
			} else {
				re += '[^/]*';
			}
		} else if (c === '?') {
			re += '[^/]';
		} else if ('.+^$()|{}[]\\'.includes(c)) {
			re += '\\' + c;
		} else {
			re += c;
		}
	}
	re += '$';
	return new RegExp(re);
}

const IGNORE_PATTERNS = loadIgnoreList().map((p) => ({
	raw: p,
	re: globToRegex(p)
}));

function isIgnored(relPath) {
	const normalized = relPath.split(path.sep).join('/');
	for (const { re } of IGNORE_PATTERNS) {
		if (re.test(normalized)) return true;
	}
	return false;
}

// ---------- file list ----------
function safeGit(cmd) {
	try {
		return execSync(cmd, { encoding: 'utf8' });
	} catch (e) {
		return '';
	}
}

function listChangedFiles() {
	if (mode === 'files') return explicitFiles;
	if (mode === 'staged') {
		const out = safeGit('git diff --cached --name-only --diff-filter=ACMR');
		return out.split('\n').filter(Boolean);
	}
	if (baseRef && headRef) {
		const out = safeGit(
			`git diff --name-only --diff-filter=ACMR ${baseRef}...${headRef}`
		);
		return out.split('\n').filter(Boolean);
	}
	// fallback to staged
	const out = safeGit('git diff --cached --name-only --diff-filter=ACMR');
	return out.split('\n').filter(Boolean);
}

// ---------- detection helpers ----------
const CHINESE_RE = /[\u4e00-\u9fff\u3400-\u4dbf]/;
// English text token: 2+ words containing only letters/whitespace/punctuation
// and at least one space. We avoid identifiers like 'fooBar'.
// The lowercase-leading branch allows mid-string tokens to start with a digit
// so that sentences like "page 12 of 30" or "test 18n check1" are recognised.
const ENGLISH_TEXT_RE =
	/(?:^|[^A-Za-z0-9_])([A-Z][A-Za-z0-9]+(?:[ \t\-,.:;!?'"()/&][A-Za-z0-9]+){1,}|[a-z][a-z0-9]*(?:[ \t][A-Za-z0-9][A-Za-z0-9]*){1,})/;

function hasUserVisibleText(text) {
	if (!text) return false;
	const stripped = text.trim();
	if (!stripped) return false;
	if (CHINESE_RE.test(stripped)) return true;
	// require at least one letter pair separated by whitespace
	if (ENGLISH_TEXT_RE.test(' ' + stripped)) return true;
	return false;
}

function lineHasIgnore(rawLine) {
	if (!rawLine) return false;
	return (
		rawLine.includes('i18n-ignore') || rawLine.includes('i18n-disable-line')
	);
}

function buildIgnoredLineSet(source) {
	// Collect a set of line numbers covered by `i18n-ignore-start` /
	// `i18n-ignore-end` (and the HTML-comment variants). Lines are 1-based.
	const ignored = new Set();
	const lines = source.split('\n');
	let active = false;
	for (let i = 0; i < lines.length; i += 1) {
		const ln = lines[i];
		if (!active && /i18n-ignore-start/.test(ln)) {
			active = true;
			ignored.add(i + 1);
			continue;
		}
		if (active) {
			ignored.add(i + 1);
			if (/i18n-ignore-end/.test(ln)) {
				active = false;
			}
		}
	}
	return ignored;
}

function offsetToLineCol(source, offset) {
	let line = 1;
	let col = 1;
	for (let i = 0; i < offset && i < source.length; i += 1) {
		if (source[i] === '\n') {
			line += 1;
			col = 1;
		} else {
			col += 1;
		}
	}
	return { line, col };
}

function getLine(source, lineNumber) {
	const lines = source.split('\n');
	return lines[lineNumber - 1] || '';
}

// ---------- Vue template scanner ----------
const VISIBLE_ATTRS = new Set([
	'label',
	'placeholder',
	'title',
	'tooltip',
	'hint',
	'error',
	'error-message',
	'message',
	'caption',
	'text',
	'aria-label',
	'alt',
	'content',
	'description',
	'subtitle'
]);

function extractBlock(source, name) {
	// Returns array of {start, end, content} for each top-level <name>...</name>
	// block, accounting for nested <name v-if>, <name v-for>, <name #slot> etc.
	const openRe = new RegExp(`<${name}(\\s[^>]*)?>`, 'g');
	const closeTag = `</${name}>`;
	const out = [];
	let cursor = 0;
	while (cursor < source.length) {
		openRe.lastIndex = cursor;
		const open = openRe.exec(source);
		if (!open) break;
		const contentStart = open.index + open[0].length;
		let depth = 1;
		let i = contentStart;
		while (i < source.length && depth > 0) {
			const nextOpen = source.indexOf(`<${name}`, i);
			const nextClose = source.indexOf(closeTag, i);
			if (nextClose === -1) {
				// Malformed source — abort to avoid an infinite loop.
				i = source.length;
				break;
			}
			// Only consider opens that are real tag starts (`<name>` or `<name `
			// or `<name/>` — i.e. followed by `>`, whitespace, or `/`).
			let nextOpenValid = -1;
			if (nextOpen !== -1 && nextOpen < nextClose) {
				const after = source[nextOpen + 1 + name.length];
				if (after === '>' || after === '/' || /\s/.test(after || '')) {
					nextOpenValid = nextOpen;
				}
			}
			if (nextOpenValid !== -1) {
				depth += 1;
				i = nextOpenValid + name.length + 1;
			} else {
				depth -= 1;
				i = nextClose + closeTag.length;
			}
		}
		const contentEnd = i - closeTag.length;
		out.push({
			start: contentStart,
			end: contentEnd,
			content: source.slice(contentStart, contentEnd)
		});
		cursor = i;
	}
	return out;
}

function tokenizeVueTemplate(working) {
	// Walk the template and emit tags (open/close/self-closing) using a
	// quote-aware state machine. This is necessary because attribute values
	// like `:disable="x > 0"` contain `>` characters that a naive regex would
	// consume too greedily.
	const tags = [];
	const len = working.length;
	let i = 0;
	while (i < len) {
		const c = working[i];
		if (c !== '<') {
			i += 1;
			continue;
		}
		const next = working[i + 1] || '';
		// only treat <letter or </letter as a tag start
		if (next === '/' && /[A-Za-z]/.test(working[i + 2] || '')) {
			const start = i;
			i += 2;
			while (i < len && working[i] !== '>') i += 1;
			tags.push({
				kind: 'close',
				start,
				end: Math.min(i + 1, len),
				attrsStr: '',
				attrsStart: i
			});
			i += 1;
			continue;
		}
		if (!/[A-Za-z]/.test(next)) {
			i += 1;
			continue;
		}
		const start = i;
		i += 1;
		// skip tag name
		while (i < len && /[A-Za-z0-9-]/.test(working[i])) i += 1;
		const attrsStart = i;
		while (i < len) {
			const ch = working[i];
			if (ch === '"' || ch === "'") {
				const q = ch;
				i += 1;
				while (i < len && working[i] !== q) i += 1;
				if (i < len) i += 1;
				continue;
			}
			if (ch === '>') break;
			i += 1;
		}
		const attrsEnd = i; // exclusive
		const attrsStr = working.slice(attrsStart, attrsEnd);
		const tagEnd = Math.min(i + 1, len);
		tags.push({ kind: 'open', start, end: tagEnd, attrsStart, attrsStr });
		i = tagEnd;
	}
	return tags;
}

function scanVueTemplate(
	templateContent,
	baseOffset,
	fullSource,
	violations,
	ignoredLines
) {
	// 1a. Scan mustache `{{ ... }}` expressions BEFORE masking them out.
	//     Their bodies are JS expressions, so the same string-literal rules
	//     used for <script> apply. This catches things like
	//     `<div>{{ 'Please save' }}</div>`.
	const mustacheRe = /\{\{([\s\S]*?)\}\}/g;
	let mm;
	while ((mm = mustacheRe.exec(templateContent))) {
		const exprContent = mm[1];
		const exprStartInTpl = mm.index + 2; // skip "{{"
		scanScript(
			exprContent,
			baseOffset + exprStartInTpl,
			fullSource,
			violations,
			ignoredLines
		);
	}

	// 1b. Mask out HTML comments and mustache expressions for the static
	//     tag / text-node walk below.
	const working = templateContent
		.replace(/<!--[\s\S]*?-->/g, (m) => ' '.repeat(m.length))
		.replace(/\{\{[\s\S]*?\}\}/g, (m) => ' '.repeat(m.length));

	// 2. Tokenize tags with a quote-aware scanner.
	const tags = tokenizeVueTemplate(working);

	// 3. Walk attributes inside each open tag.
	const STATIC_ATTR_RE =
		/(^|\s)([A-Za-z][A-Za-z0-9_-]*)\s*=\s*(?:"([^"]*)"|'([^']*)')/g;
	// `:foo="expr"`, `v-bind:foo="expr"`, `@evt="expr"`, `v-on:evt="expr"`,
	// `v-if="expr"`, `v-else-if="expr"`, `v-show="expr"`, `v-model[.x]="expr"`,
	// `v-html="expr"`, `v-text="expr"`.
	const BOUND_ATTR_RE =
		/(^|\s)(?::|@|v-bind:|v-on:|v-(?:if|else-if|show|model(?:\.[A-Za-z0-9_-]+)*|html|text)(?:\b|=))[A-Za-z0-9_:.-]*\s*=\s*(?:"([^"]*)"|'([^']*)')/g;

	for (const tag of tags) {
		if (tag.kind !== 'open' || !tag.attrsStr) continue;

		// 3a. Static visible attributes (label/placeholder/title/...): the
		//     attribute value is the literal user-visible text.
		STATIC_ATTR_RE.lastIndex = 0;
		let am;
		while ((am = STATIC_ATTR_RE.exec(tag.attrsStr))) {
			const attrName = am[2];
			const value = am[3] != null ? am[3] : am[4];
			if (!VISIBLE_ATTRS.has(attrName)) continue;
			if (!hasUserVisibleText(value)) continue;
			if (isInVocabulary(value)) continue;
			const valueOffsetInAttrs = am.index + am[0].indexOf(value);
			const absOffset = baseOffset + tag.attrsStart + valueOffsetInAttrs;
			const { line } = offsetToLineCol(fullSource, absOffset);
			if (ignoredLines && ignoredLines.has(line)) continue;
			const rawLine = getLine(fullSource, line);
			if (lineHasIgnore(rawLine)) continue;
			violations.push({
				line,
				kind: 'vue-attr',
				attr: attrName,
				text: value.trim()
			});
		}

		// 3b. Bound attributes (`:attr` / `@evt` / `v-*`): the value is a JS
		//     expression; reuse the script scanner so string literals inside
		//     are subject to the same rules (i18n call detection, vocabulary,
		//     ignore comments, etc.).
		BOUND_ATTR_RE.lastIndex = 0;
		let bm;
		while ((bm = BOUND_ATTR_RE.exec(tag.attrsStr))) {
			const expr = bm[2] != null ? bm[2] : bm[3];
			if (!expr) continue;
			const exprOffsetInAttrs = bm.index + bm[0].indexOf(expr);
			scanScript(
				expr,
				baseOffset + tag.attrsStart + exprOffsetInAttrs,
				fullSource,
				violations,
				ignoredLines
			);
		}
	}

	// 4. Scan text nodes between tags.
	const segments = [];
	let cursor = 0;
	for (const tag of tags) {
		if (tag.start > cursor) {
			segments.push({
				start: cursor,
				text: working.slice(cursor, tag.start)
			});
		}
		cursor = tag.end;
	}
	if (cursor < working.length) {
		segments.push({ start: cursor, text: working.slice(cursor) });
	}

	for (const seg of segments) {
		const text = seg.text;
		if (!hasUserVisibleText(text)) continue;
		if (isInVocabulary(text)) continue;
		const leadingWs = text.match(/^\s*/)[0].length;
		const absOffset = baseOffset + seg.start + leadingWs;
		const { line } = offsetToLineCol(fullSource, absOffset);
		if (ignoredLines && ignoredLines.has(line)) continue;
		const rawLine = getLine(fullSource, line);
		if (lineHasIgnore(rawLine)) continue;
		violations.push({
			line,
			kind: 'vue-text',
			text: text.trim().replace(/\s+/g, ' ').slice(0, 120)
		});
	}
}

// ---------- Script (JS/TS) scanner ----------
function stripScriptCommentsAndStrings(source) {
	// Mask comments with spaces but preserve string literals verbatim, so
	// downstream `findStringLiterals` still sees their original bytes and the
	// `// inside "https://..."` case doesn't get truncated.
	let result = '';
	let i = 0;
	const len = source.length;
	while (i < len) {
		const ch = source[i];
		const next = source[i + 1];

		// line comment
		if (ch === '/' && next === '/') {
			let j = i;
			while (j < len && source[j] !== '\n') {
				result += ' ';
				j += 1;
			}
			i = j;
			continue;
		}

		// block comment
		if (ch === '/' && next === '*') {
			let j = i;
			while (j < len && !(source[j] === '*' && source[j + 1] === '/')) {
				result += source[j] === '\n' ? '\n' : ' ';
				j += 1;
			}
			if (j < len) {
				result += '  ';
				j += 2;
			}
			i = j;
			continue;
		}

		// single / double quoted strings — copy verbatim (including escapes).
		if (ch === '"' || ch === "'") {
			const quote = ch;
			result += ch;
			i += 1;
			while (i < len) {
				const c = source[i];
				if (c === '\\' && i + 1 < len) {
					result += c + source[i + 1];
					i += 2;
					continue;
				}
				if (c === quote) {
					result += c;
					i += 1;
					break;
				}
				if (c === '\n') {
					// unterminated — bail out without consuming the newline so
					// subsequent lines still get scanned.
					break;
				}
				result += c;
				i += 1;
			}
			continue;
		}

		// template literal — copy verbatim, but recurse into ${...} so
		// comments / strings nested inside expressions stay correct.
		if (ch === '`') {
			result += ch;
			i += 1;
			while (i < len) {
				const c = source[i];
				if (c === '\\' && i + 1 < len) {
					result += c + source[i + 1];
					i += 2;
					continue;
				}
				if (c === '`') {
					result += c;
					i += 1;
					break;
				}
				if (c === '$' && source[i + 1] === '{') {
					// copy the `${` then walk the interpolation balancing braces.
					result += '${';
					i += 2;
					let depth = 1;
					while (i < len && depth > 0) {
						const ic = source[i];
						if (ic === '{') {
							depth += 1;
							result += ic;
							i += 1;
						} else if (ic === '}') {
							depth -= 1;
							result += ic;
							i += 1;
						} else {
							result += ic;
							i += 1;
						}
					}
					continue;
				}
				result += c;
				i += 1;
			}
			continue;
		}

		result += ch;
		i += 1;
	}
	return result;
}

function skipInterpolation(source, i, len) {
	// Called with `i` pointing to the character after `${`. Returns the index
	// just past the matching `}`. Handles nested strings, template literals,
	// and braces.
	let depth = 1;
	while (i < len && depth > 0) {
		const c = source[i];
		if (c === '\\') {
			i += 2;
			continue;
		}
		if (c === '{') {
			depth += 1;
			i += 1;
			continue;
		}
		if (c === '}') {
			depth -= 1;
			i += 1;
			continue;
		}
		if (c === "'" || c === '"') {
			const q = c;
			i += 1;
			while (i < len && source[i] !== q) {
				if (source[i] === '\\') {
					i += 2;
					continue;
				}
				if (source[i] === '\n') break;
				i += 1;
			}
			if (i < len) i += 1;
			continue;
		}
		if (c === '`') {
			i = skipTemplateLiteral(source, i + 1, len);
			continue;
		}
		i += 1;
	}
	return i;
}

function skipTemplateLiteral(source, i, len) {
	// `i` points to the character right after the opening backtick.
	while (i < len) {
		const c = source[i];
		if (c === '\\') {
			i += 2;
			continue;
		}
		if (c === '`') return i + 1;
		if (c === '$' && source[i + 1] === '{') {
			i = skipInterpolation(source, i + 2, len);
			continue;
		}
		i += 1;
	}
	return i;
}

function findStringLiterals(source) {
	// Return array of { start, end, value, quote } for each *plain* string
	// literal (single / double quoted). Template literals are skipped — they
	// are typically used for URLs / path / SQL building, and properly tracking
	// nested template literals + interpolations would require a fully blown
	// JS tokenizer.
	const literals = [];
	const len = source.length;
	let i = 0;
	while (i < len) {
		const ch = source[i];
		if (ch === '"' || ch === "'") {
			const quote = ch;
			const start = i;
			i += 1;
			let value = '';
			while (i < len) {
				const c = source[i];
				if (c === '\\') {
					value += c + (source[i + 1] || '');
					i += 2;
					continue;
				}
				if (c === quote) {
					i += 1;
					break;
				}
				if (c === '\n') {
					// unterminated single/double quoted string — bail.
					break;
				}
				value += c;
				i += 1;
			}
			literals.push({ start, end: i, value, quote });
			continue;
		}
		if (ch === '`') {
			i = skipTemplateLiteral(source, i + 1, len);
			continue;
		}
		i += 1;
	}
	return literals;
}

// Common vue-i18n / generic translation call sites whose argument is a key
// rather than a user-visible literal. Covers:
//   t('key'), $t('key'), tc('key', n), te('key'), tm('key'), d(...), n(...)
//   this.$t / this.$tc / this.$te / this.$tm
//   i18n.t / i18n.tc / i18n.te / i18n.tm / i18n.d / i18n.n
//   i18n.global.t / .tc / .te / .tm / .d / .n
//   i18next.t / .tc / .te / .tm / .d / .n  (electron main process)
//   trans('key'), translate('key')
const I18N_CALL_RE =
	/(?:^|[^A-Za-z0-9_$.])(?:(?:this\.)?\$?(?:t|tc|te|tm)|i18n\.(?:global\.)?(?:t|tc|te|tm|d|n)|i18next\.(?:t|tc|te|tm|d|n)|trans|translate)\s*\(\s*$/;
const IMPORT_FROM_RE = /(?:from|import|require)\s*\(?\s*$/;
const URL_LIKE_RE =
	/^(?:https?:\/\/|\/\/|\/|\.\.?\/|@\/|@didvault\/|~\/|file:|data:|blob:|mailto:|tel:)/;
const PATH_LIKE_RE =
	/^[A-Za-z0-9_\-./@]+\.(?:vue|ts|tsx|js|jsx|json|css|scss|svg|png|jpg|jpeg|gif|webp|woff2?|mp4|md)$/;
const SELECTOR_LIKE_RE = /^[#.]?[A-Za-z][A-Za-z0-9_-]*$/; // class / id token
// dev-only call sites whose string arguments are not user-facing.
const DEV_CALL_RE =
	/(?:^|[^A-Za-z0-9_$.])(?:console\.(?:log|info|warn|error|debug|trace|table|dir|group|groupCollapsed|groupEnd|assert|count|time|timeEnd|timeLog)|logger\.(?:log|info|warn|error|debug|trace)|debug|assert)\s*\(\s*$/;
// `throw new Error(...)` and friends — message is dev-facing.
const ERROR_CTOR_RE =
	/(?:^|[^A-Za-z0-9_$.])(?:throw\s+)?new\s+(?:Error|TypeError|RangeError|SyntaxError|ReferenceError|EvalError|URIError|AggregateError|DOMException)\s*\(\s*$/;
// moment / dayjs / date-fns format strings (e.g. "MM-DD HH:mm", "YYYY-MM-DD HH:mm:ss", "MMM Do YYYY").
const DATE_FORMAT_RE = /^[\sYMDhHmsSXxAaZzQwWeEoTLljk:.,\-/[\]'()]+$/;

function isLikelyDateFormat(value) {
	const v = value.trim();
	if (!v) return false;
	if (!DATE_FORMAT_RE.test(v)) return false;
	// require at least one date-format token (M, D, Y, H, m, s) to avoid false positives.
	if (!/[YMDHmsAa]/.test(v)) return false;
	return true;
}

function isLikelyCss(value) {
	const v = value.trim();
	if (!v) return false;
	// CSS calls (gradients, calc, rgba, hsl, var, url) or color hex codes.
	if (
		/^(?:linear-gradient|radial-gradient|conic-gradient|repeating-linear-gradient|calc|rgba?|hsla?|var|url|clamp|min|max|fit-content|attr|env|cubic-bezier|matrix|matrix3d|translate3?d?|rotate|scale|skew|perspective)\s*\(/.test(
			v
		)
	) {
		return true;
	}
	// looks like a CSS font stack: lots of font names separated by commas
	if (
		/,\s*(?:serif|sans-serif|monospace|cursive|fantasy|system-ui|-apple-system)$/i.test(
			v
		)
	) {
		return true;
	}
	// raw color hex
	if (/^#[0-9A-Fa-f]{3,8}$/.test(v)) return true;
	// CSS dimension-led shorthand values, e.g. "1px solid transparent",
	// "2px dashed #fff", "0 auto", "10px 20px". A value that starts with `0`
	// or a number+unit token is a CSS length/shorthand, never user-visible copy.
	if (
		/^(?:0|-?\d*\.?\d+(?:px|em|rem|%|vh|vw|vmin|vmax|pt|fr|deg|s|ms|ch|ex))\b(?:\s|$)/.test(
			v
		)
	) {
		return true;
	}
	return false;
}

function isLikelyHtmlSnippet(value) {
	const v = value.trim();
	if (!v) return false;
	// starts with an HTML tag like <div ...>, <span ...>, <pre>, <code>, <br/>, <a href...>.
	if (/^<\s*\/?\s*[a-zA-Z][\w-]*(\s[^>]*?)?\/?\s*>/.test(v)) return true;
	return false;
}

// Comma-separated list of HTTP method / header names — appears all over the
// codebase in CORS / axios config and is never user-visible.
const HTTP_METHODS = new Set([
	'GET',
	'POST',
	'PUT',
	'DELETE',
	'PATCH',
	'OPTIONS',
	'HEAD',
	'CONNECT',
	'TRACE'
]);

function isLikelyHttpCsv(value) {
	const v = value.trim();
	if (!v) return false;
	if (!/^[A-Za-z][A-Za-z0-9-]*(?:\s*,\s*[A-Za-z][A-Za-z0-9-]*)+$/.test(v)) {
		return false;
	}
	const tokens = v.split(/\s*,\s*/);
	return tokens.every((t) => {
		if (HTTP_METHODS.has(t)) return true;
		// Header name token: "X-Foo-Bar", "Content-Type", "ETag", etc.
		// We accept dash-separated PascalCase segments or all-caps acronyms.
		return /^(?:[A-Z][A-Za-z0-9]*|[A-Z]+)(?:-(?:[A-Z][A-Za-z0-9]*|[A-Z]+))*$/.test(
			t
		);
	});
}

function isLikelyNonText(value) {
	const v = value.trim();
	if (!v) return true;
	if (URL_LIKE_RE.test(v)) return true;
	if (PATH_LIKE_RE.test(v)) return true;
	if (SELECTOR_LIKE_RE.test(v)) return true;
	if (isLikelyDateFormat(v)) return true;
	if (isLikelyCss(v)) return true;
	if (isLikelyHtmlSnippet(v)) return true;
	if (isLikelyHttpCsv(v)) return true;
	// pure identifier-like (no spaces, only word characters and dashes/underscores)
	if (!/\s/.test(v) && /^[A-Za-z0-9_$-]+$/.test(v)) return true;
	// looks like a regex of CSS classes joined by spaces
	if (/^([A-Za-z0-9_-]+\s+){1,}[A-Za-z0-9_-]+$/.test(v)) {
		// All tokens are lowercase identifier-shaped. Real CSS class strings
		// (e.g. "flex items-center", "btn btn-primary disabled") either
		// contain a `-` / `_` somewhere, OR every token is short (<= 4 chars).
		// "test 18n check1" / "test18n check2" satisfy neither, so they are
		// more likely user-visible test text and should be reported.
		const tokens = v.split(/\s+/);
		const hasLowerOnly = tokens.every((t) => /^[a-z0-9-]+$/.test(t));
		if (hasLowerOnly) {
			const looksLikeClass =
				tokens.some((t) => t.includes('-') || t.includes('_')) ||
				tokens.every((t) => t.length <= 4);
			if (looksLikeClass) return true;
		}
	}
	return false;
}

function scanScript(
	scriptContent,
	baseOffset,
	fullSource,
	violations,
	ignoredLines
) {
	const masked = stripScriptCommentsAndStrings(scriptContent);
	const literals = findStringLiterals(masked);
	for (const lit of literals) {
		const value = lit.value;
		if (!hasUserVisibleText(value)) continue;
		if (isLikelyNonText(value)) continue;
		// look back a few characters to see whether this literal is inside
		// an i18n call or import/require/from clause
		const lookbackStart = Math.max(0, lit.start - 32);
		const lookback = masked.slice(lookbackStart, lit.start);
		if (I18N_CALL_RE.test(lookback)) continue;
		if (IMPORT_FROM_RE.test(lookback)) continue;
		if (DEV_CALL_RE.test(lookback)) continue;
		if (ERROR_CTOR_RE.test(lookback)) continue;
		if (isInVocabulary(value)) continue;
		const absOffset = baseOffset + lit.start;
		const { line } = offsetToLineCol(fullSource, absOffset);
		if (ignoredLines && ignoredLines.has(line)) continue;
		const rawLine = getLine(fullSource, line);
		if (lineHasIgnore(rawLine)) continue;
		violations.push({
			line,
			kind: 'script-string',
			text: value.trim().slice(0, 120)
		});
	}
}

// ---------- known i18n vocabulary ----------
// Collect every string literal that appears in the project's i18n dictionary
// files (packages/**/i18n/**). Any candidate violation whose text matches a
// dictionary entry is treated as already-translated, since it could only end
// up in the dictionary if someone wired it through the i18n layer (commonly
// via indirect calls like `t(displayText)` where `displayText` is a typed
// string-literal union variable that static scanning cannot follow).
let I18N_VOCAB_CACHE = null;

function collectVocabFiles() {
	const found = [];
	const queue = [];
	const seedRoots = [
		path.join(ROOT, 'packages/app/src/i18n'),
		path.join(ROOT, 'packages/locale'),
		path.join(ROOT, 'packages/app/src/locales')
	];
	for (const r of seedRoots) {
		if (fs.existsSync(r)) queue.push(r);
	}
	while (queue.length) {
		const dir = queue.shift();
		let entries;
		try {
			entries = fs.readdirSync(dir, { withFileTypes: true });
		} catch (e) {
			continue;
		}
		for (const ent of entries) {
			const full = path.join(dir, ent.name);
			if (ent.isDirectory()) {
				queue.push(full);
			} else if (/\.(ts|tsx|js|jsx|mjs|cjs|json)$/.test(ent.name)) {
				found.push(full);
			}
		}
	}
	return found;
}

function loadI18nVocabulary() {
	if (I18N_VOCAB_CACHE) return I18N_VOCAB_CACHE;
	const vocab = new Set();
	const files = collectVocabFiles();
	for (const file of files) {
		let content;
		try {
			content = fs.readFileSync(file, 'utf8');
		} catch (e) {
			continue;
		}
		if (file.endsWith('.json')) {
			// Crude JSON literal extraction: walk strings via findStringLiterals.
			const literals = findStringLiterals(content);
			for (const lit of literals) {
				const v = lit.value.trim();
				if (v.length >= 2) vocab.add(v);
			}
			continue;
		}
		const stripped = stripScriptCommentsAndStrings(content);
		const literals = findStringLiterals(stripped);
		for (const lit of literals) {
			const v = lit.value.trim();
			if (v.length >= 2) vocab.add(v);
		}
	}
	I18N_VOCAB_CACHE = vocab;
	return vocab;
}

function isInVocabulary(text) {
	if (!text) return false;
	const v = String(text).trim();
	if (!v) return false;
	return loadI18nVocabulary().has(v);
}

// ---------- per-file dispatcher ----------
function scanFile(absPath, relPath) {
	const violations = [];
	const ext = path.extname(relPath).toLowerCase();
	let source;
	try {
		source = fs.readFileSync(absPath, 'utf8');
	} catch (e) {
		return violations;
	}
	const ignoredLines = buildIgnoredLineSet(source);
	if (ext === '.vue') {
		// scan template blocks
		const templates = extractBlock(source, 'template');
		for (const block of templates) {
			scanVueTemplate(
				block.content,
				block.start,
				source,
				violations,
				ignoredLines
			);
		}
		// scan script blocks
		const scripts = extractBlock(source, 'script');
		for (const block of scripts) {
			scanScript(block.content, block.start, source, violations, ignoredLines);
		}
	} else if (['.ts', '.tsx', '.js', '.jsx', '.mjs', '.cjs'].includes(ext)) {
		scanScript(source, 0, source, violations, ignoredLines);
	}
	return violations;
}

// ---------- entry ----------
function isSupported(relPath) {
	return /\.(vue|ts|tsx|js|jsx|mjs|cjs)$/i.test(relPath);
}

function main() {
	const candidates = listChangedFiles()
		.map((f) => f.trim())
		.filter(Boolean)
		.filter(isSupported)
		.filter((f) => !isIgnored(f));

	if (!candidates.length) {
		if (!quiet) console.log('[i18n-check] no candidate files to scan.');
		process.exit(0);
	}

	let total = 0;
	const report = [];
	for (const rel of candidates) {
		const abs = path.resolve(ROOT, rel);
		if (!fs.existsSync(abs)) continue;
		const violations = scanFile(abs, rel);
		if (violations.length) {
			violations.sort((a, b) => a.line - b.line);
			total += violations.length;
			report.push({ file: rel, violations });
		}
	}

	if (!total) {
		if (!quiet)
			console.log('[i18n-check] passed. no hard-coded strings found.');
		process.exit(0);
	}

	console.error(
		`\n[i18n-check] Found ${total} hard-coded string${
			total > 1 ? 's' : ''
		} in ${report.length} file${report.length > 1 ? 's' : ''}:`
	);
	console.error('-'.repeat(72));
	for (const item of report) {
		console.error(`\n  ${item.file}`);
		for (const v of item.violations) {
			const label =
				v.kind === 'vue-attr'
					? `attr "${v.attr}"`
					: v.kind === 'vue-text'
					? 'template text'
					: 'script string';
			console.error(`    L${v.line}  [${label}]  ${JSON.stringify(v.text)}`);
		}
	}
	console.error(
		'\nFix suggestions:\n' +
			"  - Wrap user-visible strings with i18n: $t('key') in templates, t('key') in scripts.\n" +
			'  - Add the key/value pair to packages/app/src/i18n/{en-US,zh-CN}/index.ts (and the relevant locale module).\n' +
			'  - For unavoidable hard-coded text (logs, ids, urls, data dictionaries, etc.):\n' +
			'      * append // i18n-ignore (or <!-- i18n-ignore --> in templates) to a single line, or\n' +
			'      * wrap a block with // i18n-ignore-start ... // i18n-ignore-end (use <!-- ... --> in templates), or\n' +
			'      * add the path to .i18nignore.'
	);
	process.exit(1);
}

main();

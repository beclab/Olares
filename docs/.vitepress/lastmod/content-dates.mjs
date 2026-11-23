// Page "Last updated" from when the page's content last changed.
//
// VitePress shows `git log -1 <page>.md`, which is wrong in two ways here:
// - a commit that only edits metadata (keywords, description, ...) bumps the date;
// - a page that pulls in `<!--@include: ...-->` snippets never moves when a
//   snippet changes.
// So for every page this walks main's history (--first-parent: when a change
// landed, not when it was authored on a branch) and takes the newest of:
// - the last commit that changed the page itself, ignoring METADATA_KEYS;
// - for each snippet it includes (nested too), the last commit that changed the
//   part it includes, resolved with the same rules as VitePress 1.6's
//   processIncludes / findRegion (region, {start,end} range, frontmatter).
// Pure renames are not changes. The result feeds both the page footer and the
// sitemap <lastmod>, so the two always agree.

import { spawnSync } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';

// Frontmatter keys that never render on the page. Anything else (title is
// excluded: it only reaches <title>) counts as content, so a new key is
// treated as visible until it is listed here.
export const METADATA_KEYS = new Set([
  'description',
  'head',
  'title',
  'outline',
  'noindex',
  'search',
  'aside',
  'pageClass',
  'app_version',
  'doc_version',
  'doc_updated',
]);

const FRONTMATTER_RE = /^---\r?\n([\s\S]*?)\r?\n---[ \t]*(?:\r?\n|$)/;
const INCLUDES_RE = /<!--\s*@include:\s*(.*?)\s*-->/g;
const REGION_RE = /(#[\w-]+)/;
const RANGE_RE = /\{(\d*),(\d*)\}$/;
const ZERO_OID = /^0+$/;

// --- VitePress include semantics (vitepress/dist/node, processIncludes) ---

function testLine(line, regexp, regionName, end = false) {
  const [full, tag, name] = regexp.exec(line.trim()) || [];
  return full && tag && name === regionName && tag.match(end ? /^[Ee]nd ?[rR]egion$/ : /^[rR]egion$/);
}

function findRegion(lines, regionName) {
  const regionRegexps = [
    /^\/\/ ?#?((?:end)?region) ([\w*-]+)$/,
    /^\/\* ?#((?:end)?region) ([\w*-]+) ?\*\/$/,
    /^#pragma ((?:end)?region) ([\w*-]+)$/,
    /^<!-- #?((?:end)?region) ([\w*-]+) -->$/,
    /^#((?:End )Region) ([\w*-]+)$/,
    /^::#((?:end)region) ([\w*-]+)$/,
    /^# ?((?:end)?region) ([\w*-]+)$/,
  ];
  let regexp = null;
  let start = -1;
  for (const [lineId, line] of lines.entries()) {
    if (regexp === null) {
      for (const reg of regionRegexps) {
        if (testLine(line, reg, regionName)) {
          start = lineId + 1;
          regexp = reg;
          break;
        }
      }
    } else if (testLine(line, regexp, regionName, true)) {
      return { start, end: lineId };
    }
  }
  return null;
}

const stripFrontmatter = (text) => text.replace(FRONTMATTER_RE, '');

/** Parse one `@include:` argument the way VitePress does. */
export function parseInclude(arg, fromFile, srcDir) {
  if (!arg.length) return null;
  const range = arg.match(RANGE_RE);
  const region = arg.match(REGION_RE);
  let target = arg;
  if (region || range) target = target.slice(0, -((region?.[0].length || 0) + (range?.[0].length || 0)));
  const file =
    target[0] === '@'
      ? path.join(srcDir, target.slice(target[1] === '/' ? 2 : 1))
      : path.join(path.dirname(fromFile), target);
  return {
    file,
    region: region ? region[0].slice(1) : null,
    range: range ? [range[1], range[2]] : null,
  };
}

/** The part of `text` an include pulls in (before nested includes are expanded). */
export function includedPart(text, { file, region, range }) {
  if (text == null) return null;
  let content = text;
  if (region) {
    const lines = content.split(/\r?\n/);
    const r = findRegion(lines, region);
    content = lines.slice(r?.start, r?.end).join('\n');
  }
  if (range) {
    const lines = content.split(/\r?\n/);
    content = lines.slice(range[0] ? parseInt(range[0], 10) - 1 : undefined, range[1] ? parseInt(range[1], 10) : undefined).join('\n');
  }
  if (!region && !range && path.extname(file) === '.md') content = stripFrontmatter(content);
  return content;
}

/** A page as far as rendering is concerned: body plus non-metadata frontmatter. */
export function pageView(text, metadataKeys = METADATA_KEYS) {
  if (text == null) return null;
  const m = text.match(FRONTMATTER_RE);
  if (!m) return text;
  const kept = [];
  let skipping = false;
  for (const line of m[1].split(/\r?\n/)) {
    const key = /^([A-Za-z_][\w-]*)\s*:/.exec(line);
    // Continuation lines (indented, or `- ` list items under a key) belong to the
    // key above; a new top-level key ends the block.
    if (key) skipping = metadataKeys.has(key[1]);
    if (!skipping) kept.push(line);
  }
  return `${kept.join('\n')}\n---\n${text.slice(m[0].length)}`;
}

// --- git ---

function git(cwd, args, { input, env } = {}) {
  const res = spawnSync('git', args, {
    cwd,
    input,
    env: { ...process.env, ...env },
    maxBuffer: 1 << 30,
  });
  if (res.status !== 0) {
    throw new Error(`git ${args.slice(0, 3).join(' ')} failed: ${String(res.stderr).trim()}`);
  }
  return res.stdout;
}

/**
 * main's history under `prefix`, newest first, grouped by path.
 * Each entry: { seq, time, blob, status, from } with blob = content after the commit.
 */
export function readHistory(repoRoot, prefix) {
  const out = git(repoRoot, [
    '-c', 'core.quotePath=false',
    'log', '--first-parent', '--diff-merges=first-parent', '-M', '--raw', '--no-abbrev',
    '--format=@%ct', '--', prefix || '.',
  ]).toString('utf8');
  const byPath = new Map();
  let seq = -1;
  let time = 0;
  for (const line of out.split('\n')) {
    if (line.startsWith('@')) {
      seq += 1;
      time = Number(line.slice(1)) * 1000;
    } else if (line.startsWith(':')) {
      const [meta, ...paths] = line.split('\t');
      const [, , , blob, status] = meta.split(' ');
      const file = paths[paths.length - 1];
      const entry = { seq, time, blob, status, from: status.startsWith('R') ? paths[0] : null };
      if (!byPath.has(file)) byPath.set(file, []);
      byPath.get(file).push(entry);
    }
  }
  return byPath;
}

/** Versions of a path, newest first, following renames. */
export function* versions(history, file, olderThan = -1) {
  for (const e of history.get(file) || []) {
    if (e.seq <= olderThan) continue;
    yield { time: e.time, blob: ZERO_OID.test(e.blob) ? null : e.blob };
    if (e.from) {
      yield* versions(history, e.from, e.seq);
      return;
    }
    if (e.status === 'A') return;
  }
}

function readBlobs(repoRoot, oids) {
  const list = [...oids];
  // Anything still missing after prefetchBlobs must fail, not trickle in one round trip at a time.
  const out = git(repoRoot, ['cat-file', '--batch'], {
    input: `${list.join('\n')}\n`,
    env: { GIT_NO_LAZY_FETCH: '1' },
  });
  const blobs = new Map();
  let pos = 0;
  while (pos < out.length) {
    const nl = out.indexOf(10, pos);
    const [oid, type, size] = out.subarray(pos, nl).toString().split(' ');
    if (type === 'missing') throw new Error(`git object ${oid} is missing`);
    const start = nl + 1;
    blobs.set(oid, out.subarray(start, start + Number(size)).toString('utf8'));
    pos = start + Number(size) + 1;
  }
  return blobs;
}

/**
 * In a blobless clone (release-docs clones with --filter=blob:none) old file
 * contents are not local. Fetch them in one request: letting git fetch them
 * lazily means one round trip per blob, thousands of them.
 */
function promisorRemote(repoRoot) {
  const res = spawnSync('git', ['config', '--get-regexp', '^remote\\..*\\.promisor$'], { cwd: repoRoot });
  for (const line of String(res.stdout).split('\n')) {
    const m = /^remote\.(.+)\.promisor (.*)$/.exec(line.trim());
    if (m && m[2] === 'true') return m[1];
  }
  return null;
}

function prefetchBlobs(repoRoot, oids) {
  const remote = promisorRemote(repoRoot);
  if (!remote || !oids.size) return;
  git(
    repoRoot,
    ['-c', 'fetch.negotiationAlgorithm=noop', 'fetch', '--quiet', '--no-tags', '--no-write-fetch-head',
      '--recurse-submodules=no', '--filter=blob:none', remote, '--stdin'],
    { input: `${[...oids].join('\n')}\n` }
  );
}

// --- dates ---

/** When `view` of a file last changed: newest version whose view differs from the one before it. */
function lastChange(history, blobs, file, view) {
  let newer = null;
  for (const v of versions(history, file)) {
    const current = view(v.blob === null ? null : blobs.get(v.blob));
    if (newer && newer.view !== current) return newer.time;
    newer = { time: v.time, view: current };
  }
  return newer?.time;
}

function collectIncludes(text, fromFile, srcDir, seen, out) {
  if (text == null) return;
  for (const [, arg] of text.matchAll(INCLUDES_RE)) {
    const inc = parseInclude(arg, fromFile, srcDir);
    if (!inc) continue;
    const key = `${inc.file}#${inc.region}{${inc.range}}`;
    if (seen.has(key)) continue;
    seen.add(key);
    out.push(inc);
    let source;
    try {
      source = fs.readFileSync(inc.file, 'utf8');
    } catch {
      continue;
    }
    collectIncludes(includedPart(source, inc), inc.file, srcDir, seen, out);
  }
}

/**
 * Strict in CI, where the production build must not ship wrong dates silently.
 * Vercel sets CI too, but only serves previews and always clones shallow, so
 * it falls back to VitePress's dates instead of failing.
 */
export function defaultStrict(env = process.env) {
  return !!env.CI && !env.VERCEL;
}

/**
 * @param {{ srcDir: string, strict?: boolean, metadataKeys?: Set<string> }} options
 *   strict: throw instead of falling back to VitePress's dates.
 * @returns {{ forPage(relativePath: string): number | undefined }} ms timestamps
 */
export function createContentDates({ srcDir: givenSrcDir, strict = defaultStrict(), metadataKeys = METADATA_KEYS }) {
  let dates = null;

  function compute() {
    // git reports real paths; a symlinked srcDir (macOS /var -> /private/var) would not match them.
    const srcDir = fs.realpathSync(givenSrcDir);
    const repoRoot = git(srcDir, ['rev-parse', '--show-toplevel']).toString().trim();
    if (git(repoRoot, ['rev-parse', '--is-shallow-repository']).toString().trim() === 'true') {
      throw new Error('shallow clone: page dates need the full history (clone with --filter=blob:none instead)');
    }
    const prefix = path.relative(repoRoot, srcDir);
    const rel = (abs) => path.relative(repoRoot, abs).split(path.sep).join('/');
    const history = readHistory(repoRoot, prefix);

    const pages = git(repoRoot, ['-c', 'core.quotePath=false', 'ls-files', '-z', '--', prefix || '.'])
      .toString('utf8')
      .split('\0')
      .filter((f) => f.endsWith('.md'));
    const deps = new Map();
    for (const page of pages) {
      const abs = path.join(repoRoot, page);
      const out = [];
      collectIncludes(fs.readFileSync(abs, 'utf8'), abs, srcDir, new Set(), out);
      deps.set(page, out.map((inc) => ({ ...inc, file: rel(inc.file) })));
    }

    const needed = new Set();
    const addVersions = (file) => {
      for (const v of versions(history, file)) if (v.blob) needed.add(v.blob);
    };
    for (const [page, incs] of deps) {
      addVersions(page);
      for (const inc of incs) addVersions(inc.file);
    }
    prefetchBlobs(repoRoot, needed);
    const blobs = readBlobs(repoRoot, needed);

    const memo = new Map();
    const changed = (file, key, view) => {
      const k = `${file}\0${key}`;
      if (!memo.has(k)) memo.set(k, lastChange(history, blobs, file, view));
      return memo.get(k);
    };
    const result = new Map();
    for (const [page, incs] of deps) {
      const times = [changed(page, 'page', (t) => pageView(t, metadataKeys))];
      for (const inc of incs) {
        times.push(changed(inc.file, `${inc.region}{${inc.range}}`, (t) => includedPart(t, inc)));
      }
      const known = times.filter((t) => t !== undefined);
      if (known.length) result.set(path.posix.relative(prefix, page), Math.max(...known));
    }
    return result;
  }

  return {
    forPage(relativePath) {
      if (dates === null) {
        try {
          dates = compute();
        } catch (err) {
          if (strict) throw err;
          console.warn(`[last-updated] ${err.message}; falling back to VitePress's per-file git dates`);
          dates = new Map();
        }
      }
      return dates.get(relativePath);
    },
  };
}

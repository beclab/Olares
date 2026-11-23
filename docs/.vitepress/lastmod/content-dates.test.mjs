import test from 'node:test';
import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import { mkdtempSync, mkdirSync, rmSync, writeFileSync, renameSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { createContentDates, defaultStrict, includedPart, pageView } from './content-dates.mjs';

// A throwaway repo with docs/ under it; every commit gets a fixed date (day N).
function repo() {
  const root = mkdtempSync(path.join(tmpdir(), 'content-dates-'));
  const docs = path.join(root, 'docs');
  mkdirSync(path.join(docs, 'reusables'), { recursive: true });
  const git = (...args) => execFileSync('git', args, { cwd: root, env: { ...process.env, ...git.env } }).toString();
  git.env = {};
  git('init', '-q', '-b', 'main');
  git('config', 'user.email', 't@example.com');
  git('config', 'user.name', 't');
  const write = (file, text) => {
    mkdirSync(path.dirname(path.join(docs, file)), { recursive: true });
    writeFileSync(path.join(docs, file), text);
  };
  const commit = (day, msg = `day ${day}`) => {
    const date = `2026-01-${String(day).padStart(2, '0')}T00:00:00Z`;
    git.env = { GIT_AUTHOR_DATE: date, GIT_COMMITTER_DATE: date };
    git('add', '-A');
    git('commit', '-q', '--allow-empty', '-m', msg);
  };
  const dayOf = (page) => {
    const t = createContentDates({ srcDir: docs, strict: true }).forPage(page);
    return t && new Date(t).getUTCDate();
  };
  return { root, docs, git, write, commit, dayOf, done: () => rmSync(root, { recursive: true, force: true }) };
}

const fm = (fields, body) => `---\n${fields}\n---\n${body}`;

test('body changes count; metadata-only frontmatter changes do not', () => {
  const r = repo();
  r.write('a.md', fm('description: one\nhead:\n  - - meta\n    - name: keywords\n      content: x', '# A\n'));
  r.commit(1);
  r.write('a.md', fm('description: one\nhead:\n  - - meta\n    - name: keywords\n      content: x', '# A\n\nMore.\n'));
  r.commit(2);
  r.write('a.md', fm('description: two\ndoc_updated: "2026-01-03"\nhead:\n  - - meta\n    - name: keywords\n      content: y, z', '# A\n\nMore.\n'));
  r.commit(3);
  assert.equal(r.dayOf('a.md'), 2);
  r.done();
});

test('frontmatter keys that render on the page still count', () => {
  const r = repo();
  r.write('a.md', fm('connectionVersion: "1.12.6"', '# A\n'));
  r.commit(1);
  r.write('a.md', fm('connectionVersion: "1.12.7"', '# A\n'));
  r.commit(2);
  assert.equal(r.dayOf('a.md'), 2);
  r.done();
});

test('a snippet change moves the including page; another region of the same snippet does not', () => {
  const r = repo();
  const snippet = (a, b) => `<!-- #region a -->\n${a}\n<!-- #endregion a -->\n<!-- #region b -->\n${b}\n<!-- #endregion b -->\n`;
  r.write('reusables/s.md', snippet('A1', 'B1'));
  r.write('uses-a.md', '# Uses A\n<!--@include: ./reusables/s.md#a-->\n');
  r.write('uses-b.md', '# Uses B\n<!--@include: ./reusables/s.md#b-->\n');
  r.commit(1);
  r.write('reusables/s.md', snippet('A2', 'B1'));
  r.commit(2);
  assert.equal(r.dayOf('uses-a.md'), 2);
  assert.equal(r.dayOf('uses-b.md'), 1);
  r.done();
});

test('nested includes, line ranges, and whole-file includes without their frontmatter', () => {
  const r = repo();
  r.write('reusables/inner.md', 'inner 1\n');
  r.write('reusables/outer.md', fm('title: x', 'outer\n<!--@include: ./inner.md-->\n'));
  r.write('reusables/lines.md', 'l1\nl2\nl3\n');
  r.write('page.md', '# P\n<!--@include: @/reusables/outer.md-->\n<!--@include: ./reusables/lines.md{2,2}-->\n');
  r.commit(1);
  r.write('reusables/outer.md', fm('title: changed', 'outer\n<!--@include: ./inner.md-->\n'));
  r.write('reusables/lines.md', 'l1 edited\nl2\nl3\n');
  r.commit(2);
  assert.equal(r.dayOf('page.md'), 1, 'snippet frontmatter and lines outside the range are not included');
  r.write('reusables/inner.md', 'inner 2\n');
  r.commit(3);
  assert.equal(r.dayOf('page.md'), 3, 'nested snippet change');
  r.write('reusables/lines.md', 'l1 edited\nl2 edited\nl3\n');
  r.commit(4);
  assert.equal(r.dayOf('page.md'), 4, 'change inside the included range');
  r.done();
});

test('a pure rename is not a change', () => {
  const r = repo();
  r.write('old.md', '# Page\n\nSome text that makes rename detection certain.\n');
  r.commit(1);
  renameSync(path.join(r.docs, 'old.md'), path.join(r.docs, 'new.md'));
  r.commit(2);
  assert.equal(r.dayOf('new.md'), 1);
  r.done();
});

test('a change merged from a branch dates from the merge into main', () => {
  const r = repo();
  r.write('a.md', '# A\n');
  r.commit(1);
  r.git('checkout', '-q', '-b', 'feature');
  r.write('a.md', '# A\n\nFrom a branch.\n');
  r.commit(2);
  r.git('checkout', '-q', 'main');
  r.write('other.md', '# Other\n');
  r.commit(3);
  r.git.env = { GIT_AUTHOR_DATE: '2026-01-09T00:00:00Z', GIT_COMMITTER_DATE: '2026-01-09T00:00:00Z' };
  r.git('merge', '-q', '--no-ff', '-m', 'merge', 'feature');
  assert.equal(r.dayOf('a.md'), 9);
  r.done();
});

test('the latest of page and snippets wins', () => {
  const r = repo();
  r.write('reusables/s.md', 's1\n');
  r.write('page.md', '# P\n<!--@include: ./reusables/s.md-->\n');
  r.commit(1);
  r.write('reusables/s.md', 's2\n');
  r.commit(2);
  r.write('page.md', '# P edited\n<!--@include: ./reusables/s.md-->\n');
  r.commit(3);
  assert.equal(r.dayOf('page.md'), 3);
  r.done();
});

test('shallow clones fail in strict mode and fall back otherwise', () => {
  const r = repo();
  r.write('a.md', '# A\n');
  r.commit(1);
  r.write('a.md', '# A2\n');
  r.commit(2);
  const shallow = mkdtempSync(path.join(tmpdir(), 'content-dates-shallow-'));
  execFileSync('git', ['clone', '-q', '--depth', '1', `file://${r.root}`, shallow]);
  const srcDir = path.join(shallow, 'docs');
  assert.throws(() => createContentDates({ srcDir, strict: true }).forPage('a.md'), /shallow/);
  const warn = console.warn;
  console.warn = () => {};
  try {
    assert.equal(createContentDates({ srcDir, strict: false }).forPage('a.md'), undefined);
  } finally {
    console.warn = warn;
  }
  rmSync(shallow, { recursive: true, force: true });
  r.done();
});

test('strict by default in CI, but not on Vercel previews', () => {
  assert.equal(defaultStrict({ CI: 'true' }), true);
  assert.equal(defaultStrict({ CI: '1', VERCEL: '1' }), false);
  assert.equal(defaultStrict({}), false);
});

test('pageView keeps rendered keys and drops metadata blocks', () => {
  const view = pageView(fm('title: T\nhero:\n  name: Olares\nhead:\n  - - meta\n    - name: keywords\noutline: [2, 3]', 'Body'));
  assert.match(view, /hero:\n  name: Olares/);
  assert.doesNotMatch(view, /title|keywords|outline/);
  assert.match(view, /Body$/);
});

test('includedPart follows VitePress: missing region falls back to the whole file', () => {
  assert.equal(includedPart('x\ny\n', { file: 'a.md', region: 'nope', range: null }), 'x\ny\n');
  assert.equal(includedPart(fm('a: 1', 'body'), { file: 'a.md', region: null, range: null }), 'body');
  assert.equal(includedPart(fm('a: 1', 'body'), { file: 'a.txt', region: null, range: null }), fm('a: 1', 'body'));
});

import test from 'node:test';
import assert from 'node:assert/strict';
import { fileURLToPath } from 'node:url';
import { mkdtempSync, readFileSync, rmSync, statSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { createRenderer, createSocialImages, imageBaseUrl } from './generate.mjs';

const assets = fileURLToPath(new URL('.', import.meta.url));
const renderer = createRenderer(assets);
const meta = entries => Object.fromEntries(entries.map(([, a]) => [a.property || a.name, a.content]));

test('images use the deployed prefix, including archived versions and previews', () => {
  assert.equal(imageBaseUrl('/').href, 'https://www.olares.com/docs/');
  assert.equal(imageBaseUrl('/docs/1.12.4/').href, 'https://www.olares.com/docs/1.12.4/');
  assert.equal(imageBaseUrl('/docs/', 'https://preview.example.com/').href, 'https://preview.example.com/');
  assert.throws(() => imageBaseUrl('/', 'file:///tmp/'));
});

test('routes cannot collide, and changing the title invalidates image URLs', () => {
  const images = createSocialImages(assets);
  const page = { relativePath: 'a/b.md', title: 'Install Olares', frontmatter: {} };
  const original = meta(images.head(page, '/docs/'));
  const flattened = meta(images.head({ ...page, relativePath: 'a-b.md' }, '/docs/'));
  const updated = meta(images.head({ ...page, title: 'Activate Olares' }, '/docs/'));
  assert.notEqual(original['og:image'], flattened['og:image']);
  assert.notEqual(original['og:image'], updated['og:image']);
  assert.equal(original['og:image'], original['twitter:image']);
  assert.equal(original['og:image:type'], 'image/png');
  assert.equal(original['og:url'], 'https://www.olares.com/docs/a/b');
  const zh = meta(images.head({ ...page, relativePath: 'zh/index.md', frontmatter: { noindex: true, ogTitle: '安装与激活' } }, '/docs/'));
  assert.equal(zh['og:locale'], 'zh_CN');
  assert.equal(zh['og:image:alt'], '安装与激活');
  assert.equal(zh['og:url'], 'https://www.olares.com/docs/zh/');
  const withDescription = meta(images.head({ ...page, description: 'Install Olares on your own hardware.' }, '/docs/'));
  const changedDescription = meta(images.head({ ...page, description: 'Choose an installation method.' }, '/docs/'));
  assert.notEqual(withDescription['og:image'], changedDescription['og:image']);
  assert.equal(withDescription['og:description'], 'Install Olares on your own hardware.');
  const override = meta(images.head({ ...page, description: 'Full metadata description.', frontmatter: { ogDescription: 'Short image text.' } }, '/docs/'));
  assert.equal(override['og:description'], 'Full metadata description.');
  assert.notEqual(override['og:image'], original['og:image']);
});

test('description reserves three lines below long titles without reaching the footer', () => {
  for (const [lang, title, description] of [
    ['en', 'Use Ollama, vLLM, llama.cpp, and SGLang to run local AI models on your own hardware', 'Compare the supported inference engines, configure your models, and connect your applications to Olares.'],
    ['zh', '使用 Ollama、vLLM、llama.cpp 和 SGLang 运行本地大模型', '使用 Olares 上的 Engine Base 应用自托管大语言模型，并通过克隆基础应用运行不同的推理引擎。'],
  ]) {
    const titleLayout = renderer.layoutTitle(title, lang);
    const body = renderer.layoutDescription(description, titleLayout);
    assert.ok(body.top - titleLayout.bottom >= 36);
    assert.ok(body.bottom <= 540);
    assert.ok(body.lines.length <= 3);
    assert.ok(body.widths.every(width => width <= 1056));
    assert.equal(body.lines.join('').replace(/\s/g, ''), description.replace(/\s/g, ''));
    assert.equal(body.truncated, false);
    const overflow = renderer.layoutDescription(description.repeat(12), titleLayout);
    assert.equal(overflow.truncated, true);
    assert.equal(overflow.lines.length, 3);
    assert.ok(overflow.lines[2].endsWith('…'));
    assert.ok(overflow.widths.every(width => width <= 1056));
    assert.ok(overflow.bottom <= 540);
  }
  assert.deepEqual(renderer.layoutDescription('', renderer.layoutTitle('Olares', 'en')).lines, []);
});

test('Chinese, Latin, punctuation, and long identifiers fit without lost text', () => {
  for (const [lang, title] of [
    ['en', 'Access your Olares from anywhere with a secure connection'],
    ['zh', '通过安全连接，随时随地访问你的 Olares 和本地应用'],
    ['en', 'Configure very-long-application-identifier-without-spaces-on-olares'],
    ['en', 'Use <App> & "Settings" safely'],
  ]) {
    const layout = renderer.layoutTitle(title, lang);
    assert.ok(layout.lines.length <= 3);
    assert.ok(layout.widths.every(width => width <= 1056));
    assert.equal(layout.lines.join('').replace(/\s/g, ''), title.replace(/\s/g, ''));
  }
  assert.throws(() => renderer.layoutTitle('A very long document title '.repeat(30), 'en'), /ogTitle/);
});

test('rendered PNG is 1200 × 630; cache restores images into a fresh output', () => {
  const dir = mkdtempSync(path.join(tmpdir(), 'olares-og-'));
  try {
    const images = createSocialImages(assets);
    const page = { relativePath: 'zh/install.md', title: '安装 Olares', frontmatter: {} };
    const head = meta(images.head(page, '/docs/'));
    const filename = new URL(head['og:image']).pathname.split('/').at(-1);
    images.write(path.join(dir, 'out'), path.join(dir, 'cache'));
    const cached = path.join(dir, 'cache/og', filename);
    const modified = statSync(cached).mtimeMs;
    const png = readFileSync(path.join(dir, 'out/og', filename));
    assert.equal(png.subarray(1, 4).toString(), 'PNG');
    assert.equal(png.readUInt32BE(16), 1200);
    assert.equal(png.readUInt32BE(20), 630);
    images.head(page, '/docs/');
    images.write(path.join(dir, 'fresh-out'), path.join(dir, 'cache'));
    assert.equal(statSync(cached).mtimeMs, modified);
    assert.deepEqual(readFileSync(path.join(dir, 'fresh-out/og', filename)), png);
  } finally { rmSync(dir, { recursive: true, force: true }); }
});

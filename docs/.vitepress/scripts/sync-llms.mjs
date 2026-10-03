#!/usr/bin/env node
// sync-llms: publish the payment developer docs as agent-readable Markdown.
//
// Fans docs/developer/payment/*.md out to docs/public/developer/payment/:
//   1. md/<page>.md   — raw Markdown mirror of every page
//   2. llms.txt       — index of the section (llmstxt.org convention)
//   3. llms-full.txt  — every page concatenated, one read for the whole API
//
// English pages are the canonical LLM source. The output lives under public/,
// so it ships verbatim at <site>/developer/payment/llms.txt (docs base in
// production). Regenerated on every predev/prebuild — never hand-edit.

import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const docsRoot = path.resolve(__dirname, '..', '..');

const SRC = path.join(docsRoot, 'developer', 'payment');
const OUT = path.join(docsRoot, 'public', 'developer', 'payment');
const SITE = 'https://www.olares.com/docs/developer/payment';

// Sidebar order — the reading order of the full bundle.
const ORDER = [
  'index',
  'quickstart',
  'api-reference/index',
  'api-reference/payments',
  'api-reference/refunds',
  'api-reference/receiving',
  'api-reference/account',
  'webhooks',
  'ai-agents',
];

function titleOf(md, fallback) {
  const m = md.match(/^#\s+(.+)$/m);
  return m ? m[1].trim() : fallback;
}

// Page name → public HTML path fragment ('index' → '', 'api-reference/index' → 'api-reference/').
function htmlPath(name) {
  if (name === 'index') return '';
  if (name.endsWith('/index')) return name.slice(0, -'index'.length);
  return name;
}

const files = ORDER.map((name) => {
  const file = path.join(SRC, `${name}.md`);
  if (!fs.existsSync(file)) throw new Error(`sync-llms: missing page ${file}`);
  const md = fs.readFileSync(file, 'utf8');
  return { name, id: name.replaceAll('/', '-'), html: htmlPath(name), md, title: titleOf(md, name) };
});

fs.rmSync(OUT, { recursive: true, force: true });
fs.mkdirSync(path.join(OUT, 'md'), { recursive: true });

for (const f of files) {
  fs.writeFileSync(path.join(OUT, 'md', `${f.id}.md`), f.md);
}

const indexLines = [
  '# Olares Payment',
  '',
  '> Integrate stablecoin payments (USDC/USDT) into your own service: create a payment, send the hosted checkout link to the buyer, confirm via webhook or polling. Funds settle on-chain directly to the merchant wallet.',
  '',
  '## Pages',
  ...files.map((f) => `- [${f.title}](${SITE}/${f.html}) ([md](${SITE}/md/${f.id}.md))`),
  '',
  '## Bundles',
  `- [Full documentation in one file](${SITE}/llms-full.txt)`,
  '',
];
fs.writeFileSync(path.join(OUT, 'llms.txt'), indexLines.join('\n'));

const full = files
  .map((f) => [`# ${f.title}`, `Source: ${SITE}/${f.html}`, '', f.md].join('\n'))
  .join('\n---\n\n');
fs.writeFileSync(path.join(OUT, 'llms-full.txt'), full);

console.log(`[sync-llms] wrote ${files.length} pages → public/developer/payment/{llms.txt,llms-full.txt,md/}`);

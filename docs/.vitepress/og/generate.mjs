import fs from 'node:fs';
import path from 'node:path';
import { createHash } from 'node:crypto';
import { openSync } from 'fontkit';
import { Resvg } from '@resvg/resvg-js';

const hash = (value) => createHash('sha256').update(value).digest('hex');
const escape = (value) => String(value).replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&apos;'}[c]));

export function imageBaseUrl(base = '/', override) {
  const url = new URL(override || `https://www.olares.com${base === '/' ? '/docs/' : base}`);
  if (!['https:', 'http:'].includes(url.protocol) || url.search || url.hash) {
    throw new Error('DOCS_SITE_URL must be an absolute HTTP(S) site URL without a query or fragment');
  }
  url.pathname = url.pathname.replace(/\/?$/, '/');
  return url;
}

// Text is converted to outlines, so rendering never depends on system fonts.
export function createRenderer(assetDir) {
  const manrope = openSync(path.join(assetDir, 'fonts/Manrope.ttf'));
  const noto = openSync(path.join(assetDir, 'fonts/NotoSansSC.ttf'));
  const fontSets = new Map();
  const logo = fs.readFileSync(path.join(assetDir, 'logo.svg'), 'utf8')
    .replace(/<svg[^>]*>/, '').replace(/<\/svg>\s*$/, '');

  const background = fs.readFileSync(path.join(assetDir, 'background.svg'), 'utf8')
    .replace(/<svg[^>]*>/, '').replace(/<\/svg>\s*$/, '');

  function shape(text, size, weight = 500, spacing = 0) {
    const { latin, cjk } = typeof weight === 'number' ? { latin: weight, cjk: weight } : weight;
    const key = `${latin}/${cjk}`;
    if (!fontSets.has(key)) {
      fontSets.set(key, [manrope.getVariation({ wght: latin }), noto.getVariation({ wght: cjk })]);
    }
    const fonts = fontSets.get(key);
    const runs = [];
    for (const char of text) {
      const font = fonts.find(f => f.hasGlyphForCodePoint(char.codePointAt(0)));
      if (!font) throw new Error(`OG font has no glyph for ${JSON.stringify(char)} in ${JSON.stringify(text)}`);
      const previous = runs.at(-1);
      if (previous?.font === font) previous.text += char;
      else runs.push({ font, text: char });
    }
    let x = 0;
    const glyphs = [];
    for (const run of runs) {
      const scale = size / run.font.unitsPerEm;
      const layout = run.font.layout(run.text);
      layout.glyphs.forEach((glyph, i) => {
        const pos = layout.positions[i];
        glyphs.push({ glyph, x: x + pos.xOffset * scale, y: -pos.yOffset * scale, scale });
        x += pos.xAdvance * scale + spacing;
      });
    }
    return { glyphs, width: Math.max(0, x - (glyphs.length ? spacing : 0)) };
  }

  function textSvg(text, x, baseline, size, color, weight = 500, spacing = 0) {
    const shaped = shape(text, size, weight, spacing);
    return `<g fill="${color}" transform="translate(${x} ${baseline})">${shaped.glyphs.map(g => `<path transform="translate(${g.x} ${g.y}) scale(${g.scale} ${-g.scale})" d="${g.glyph.path.toSVG()}"/>`).join('')}</g>`;
  }

  function tokenize(text) {
    const tokens = [...new Intl.Segmenter('zh', { granularity: 'word' }).segment(text)].map(part => part.segment);
    for (let i = tokens.length - 1; i > 0; i--) {
      if (/^[，。！？、；：）》」』】]/u.test(tokens[i])) tokens.splice(i - 1, 2, tokens[i - 1] + tokens[i]);
    }
    return tokens;
  }

  function wrapText(text, maxWidth, measure) {
    const lines = [];
    let line = '';
    const graphemes = new Intl.Segmenter(undefined, { granularity: 'grapheme' });
    const parts = tokenize(text).flatMap(token => measure(token.trim()) > maxWidth
      ? [...graphemes.segment(token)].map(part => part.segment) : [token]);
    for (const token of parts) {
      const next = (line + token).trimStart();
      if (line.trim() && measure(next.trimEnd()) > maxWidth) {
        lines.push(line.trim());
        line = token.trimStart();
      } else line = next;
    }
    if (line.trim()) lines.push(line.trim());
    return lines;
  }

  function layoutTitle(rawTitle, lang) {
    const title = rawTitle.replace(/\s+/g, ' ').trim();
    const weight = { latin: 650, cjk: 550 };
    for (let size = 64; size >= 36; size -= 2) {
      const spacing = lang === 'zh' ? 0 : -0.6;
      const measured = new Map();
      const width = text => {
        if (!measured.has(text)) measured.set(text, shape(text, size, weight, spacing).width);
        return measured.get(text);
      };
      let lines = wrapText(title, 1056, width);
      const lineHeight = size * 1.23;
      const height = size + (lines.length - 1) * lineHeight;
      if (lines.length <= 3 && height <= 184) {
        return { lines, size, lineHeight, height, weight, spacing, top: 206, bottom: 206 + height, widths: lines.map(width) };
      }
    }
    throw new Error(`OG title is too long; add a shorter ogTitle in frontmatter: ${rawTitle}`);
  }

  function layoutDescription(rawDescription, titleLayout) {
    const text = rawDescription.replace(/\s+/g, ' ').trim();
    const size = 29, lineHeight = size * 1.4, weight = 400;
    const measured = new Map();
    const width = value => {
      if (!measured.has(value)) measured.set(value, shape(value, size, weight).width);
      return measured.get(value);
    };
    let lines = wrapText(text, 1056, width);
    const truncated = lines.length > 3;
    if (truncated) {
      lines = lines.slice(0, 3);
      // Only the image is shortened; the full description stays in metadata.
      const segments = [...new Intl.Segmenter(undefined, { granularity: 'grapheme' }).segment(lines[2])].map(part => part.segment);
      while (segments.length && width(segments.join('').trimEnd() + '…') > 1056) segments.pop();
      lines[2] = segments.join('').trimEnd() + '…';
    }
    const top = Math.max(330, titleLayout.bottom + 36);
    return { lines, size, weight, lineHeight, top, bottom: top + (lines.length ? size + (lines.length - 1) * lineHeight : 0), truncated, widths: lines.map(width) };
  }

  function render(title, lang = 'en', description = '') {
    const layout = layoutTitle(title, lang);
    const descriptionLayout = layoutDescription(description, layout);
    const kind = lang === 'zh' ? '文档' : 'Documentation';
    const kindWidth = shape(kind, 20).width;
    const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="1200" height="630" viewBox="0 0 1200 630"><title>${escape(title)}</title><desc>${escape(description)}</desc>
      ${background}
      <g transform="translate(72 64) scale(.184)">${logo}</g>
      ${textSvg(kind, 1128 - kindWidth, 94, 20, '#737373')}
      ${layout.lines.map((line, i) => textSvg(line, 72, layout.top + layout.size * .83 + i * layout.lineHeight, layout.size, '#171717', layout.weight, layout.spacing)).join('')}
      ${descriptionLayout.lines.map((line, i) => textSvg(line, 72, descriptionLayout.top + descriptionLayout.size * .83 + i * descriptionLayout.lineHeight, descriptionLayout.size, '#666666', descriptionLayout.weight)).join('')}
      ${textSvg('olares.com/docs', 72, 583, 18, '#808080', 400)}
    </svg>`;
    const png = new Resvg(svg, { font: { loadSystemFonts: false } }).render().asPng();
    return { svg, png, layout, descriptionLayout };
  }
  return { render, layoutTitle, layoutDescription };
}

export function createSocialImages(assetDir) {
  const inputs = ['generate.mjs', 'background.svg', 'logo.svg', 'fonts/Manrope.ttf', 'fonts/NotoSansSC.ttf'];
  const revision = hash(Buffer.concat(inputs.map(file => fs.readFileSync(path.join(assetDir, file)))));
  const jobs = new Map();
  let renderer;
  function head(pageData, base, siteUrl) {
    const relativePath = pageData.relativePath || '404.md';
    const lang = relativePath.startsWith('zh/') ? 'zh' : 'en';
    const siteName = lang === 'zh' ? 'Olares 文档' : 'Olares Docs';
    const title = String(pageData.frontmatter?.ogTitle || pageData.title || siteName).trim();
    const description = pageData.description?.trim() || (lang === 'zh' ? '让人们重新拥有自己的数据' : 'Let people own their data again');
    const imageDescription = String(pageData.frontmatter?.ogDescription ?? description).trim();
    const imageName = `${hash(JSON.stringify({ revision, relativePath, title, lang, imageDescription })).slice(0, 24)}.png`;
    jobs.set(imageName, { title, lang, description: imageDescription });
    const root = imageBaseUrl(base, siteUrl);
    const route = relativePath.replace(/(^|\/)index\.md$/, '$1').replace(/\.md$/, '');
    const image = new URL(`og/${imageName}`, root).href;
    const fullTitle = pageData.title ? `${pageData.title} | ${siteName}` : siteName;
    return [
      ...Object.entries({
        'og:url': new URL(route, root).href,
        'og:title': fullTitle, 'og:description': description,
        'og:image': image, 'og:image:width': '1200', 'og:image:height': '630',
        'og:image:type': 'image/png', 'og:image:alt': title,
        'og:type': 'website', 'og:site_name': siteName,
        'og:locale': lang === 'zh' ? 'zh_CN' : 'en_US',
      }).map(([property, content]) => ['meta', { property, content }]),
      ...Object.entries({
        'twitter:card': 'summary_large_image', 'twitter:title': fullTitle,
        'twitter:description': description, 'twitter:image': image, 'twitter:image:alt': title,
      }).map(([name, content]) => ['meta', { name, content }]),
    ];
  }
  function write(outDir, cacheDir) {
    const output = path.join(outDir, 'og');
    const cache = path.join(cacheDir, 'og');
    fs.mkdirSync(output, { recursive: true });
    fs.mkdirSync(cache, { recursive: true });
    let generated = 0;
    for (const [filename, job] of jobs) {
      const cached = path.join(cache, filename);
      if (!fs.existsSync(cached)) {
        renderer ??= createRenderer(assetDir);
        fs.writeFileSync(cached, renderer.render(job.title, job.lang, job.description).png);
        generated++;
      }
      fs.copyFileSync(cached, path.join(output, filename));
    }
    console.log(`OG images: ${jobs.size} published (${generated} generated, ${jobs.size - generated} cached)`);
    jobs.clear();
  }
  return { head, write };
}

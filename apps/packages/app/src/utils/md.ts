import DOMPurify from 'dompurify';
import type { Config } from 'dompurify';
import hljs from 'highlight.js';
import MarkdownIt from 'markdown-it';
import multimdTable from 'markdown-it-multimd-table';

const SANITIZE_HTML: Config = {
	ADD_ATTR: [
		'align',
		'target',
		'rel',
		'style',
		'width',
		'height',
		'colspan',
		'rowspan',
		'scope'
	],
	ALLOW_UNKNOWN_PROTOCOLS: true
};

const md = new MarkdownIt({
	html: true,
	linkify: true,
	typographer: false,
	breaks: false,
	highlight: (str, lang) => {
		const key = lang?.trim() ?? '';
		let highlighted: string;
		try {
			if (key && hljs.getLanguage(key)) {
				highlighted = hljs.highlight(str, { language: key }).value;
			} else if (key) {
				highlighted = hljs.highlight(str, { language: 'plaintext' }).value;
			} else {
				highlighted = hljs.highlightAuto(str).value;
			}
		} catch {
			highlighted = hljs.highlightAuto(str).value;
		}
		const langClass = key ? `language-${key}` : '';
		return `<pre><code class="hljs ${langClass}">${highlighted}</code></pre>`;
	}
}).use(multimdTable);

export function renderMarkdownToHtml(src: string): string {
	const html = md.render(src);
	return DOMPurify.sanitize(html, SANITIZE_HTML);
}

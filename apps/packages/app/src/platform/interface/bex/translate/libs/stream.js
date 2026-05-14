/* eslint-disable no-undef */
import {
	OPT_TRANS_OPENAI,
	OPT_TRANS_GEMINI,
	OPT_TRANS_OPENAI_2,
	OPT_TRANS_OPENAI_3,
	OPT_TRANS_OLLAMA,
	OPT_TRANS_OLLAMA_2,
	OPT_TRANS_OLLAMA_3,
	OPT_TRANS_CLOUDFLAREAI
} from '../config';
import { stripMarkdownCodeBlock } from './utils';

export const createSSEParser = () => {
	let buffer = '';

	return function* (chunk) {
		buffer += chunk;
		const lines = buffer.split('\n');
		buffer = lines.pop() || '';

		for (const line of lines) {
			if (!line.startsWith('data: ')) continue;
			const data = line.slice(6);
			if (data === '[DONE]') continue;
			yield data;
		}
	};
};

export const createAsyncQueue = () => {
	const queue = [];
	let resolve = null;
	let done = false;
	let error = null;

	return {
		push: (data) => {
			queue.push(data);
			if (resolve) {
				resolve();
				resolve = null;
			}
		},
		finish: () => {
			done = true;
			if (resolve) {
				resolve();
				resolve = null;
			}
		},
		error: (e) => {
			error = e;
			done = true;
			if (resolve) {
				resolve();
				resolve = null;
			}
		},
		async *iterate() {
			while (!done || queue.length > 0) {
				if (queue.length > 0) {
					yield queue.shift();
				} else if (!done) {
					await new Promise((r) => {
						resolve = r;
					});
				}
			}
			if (error) throw error;
		}
	};
};

export function getStreamDelta(json, apiType) {
	switch (apiType) {
		case OPT_TRANS_OPENAI:
		case OPT_TRANS_OPENAI_2:
		case OPT_TRANS_OPENAI_3:
		case OPT_TRANS_OLLAMA:
		case OPT_TRANS_OLLAMA_2:
		case OPT_TRANS_OLLAMA_3:
			return json.choices?.[0]?.delta?.content || '';
		case OPT_TRANS_GEMINI:
			return json.candidates?.[0]?.content?.parts?.[0]?.text || '';
		case OPT_TRANS_CLOUDFLAREAI:
			return json.response || '';
		default:
			return '';
	}
}

export function* parseStreamingSegments(content, processedIds) {
	if (!content) return;

	const xmlRegex =
		/<(t|item|seg)\s+id="(\d+)"(?:\s+sourceLanguage="([^"]*)")?[^>]*>([\s\S]*?)<\/\1>/gi;
	let match;
	let hasXml = false;
	while ((match = xmlRegex.exec(content)) !== null) {
		hasXml = true;
		const id = parseInt(match[2], 10);
		if (!processedIds.has(id)) {
			processedIds.add(id);
			const sourceLanguage = match[3] || '';
			const translation = [match[4].trim(), sourceLanguage];
			yield { id, translation };
		}
	}

	if (hasXml) return;

	const endsWithNewline = content.endsWith('\n');
	const lines = content.split('\n');
	const linesToProcess = endsWithNewline ? lines : lines.slice(0, -1);

	for (const line of linesToProcess) {
		const trimmedLine = line.trim();
		if (!trimmedLine) continue;

		const pipeMatch = trimmedLine.match(/^(\d+)\s*\|\s*(.*)/);
		if (pipeMatch) {
			const id = parseInt(pipeMatch[1], 10);
			if (!processedIds.has(id)) {
				processedIds.add(id);
				const translation = [
					pipeMatch[2].trim().replace(/<br\s*\/?>/gi, '\n'),
					''
				];
				yield { id, translation };
			}
		}
	}
}

export function createStreamingJsonParser() {
	const pending = [];
	let buffer = '';
	let inString = false;
	let escape = false;
	let depth = 0;
	let objectStart = -1;

	return {
		*write(delta) {
			buffer += delta;

			for (let i = 0; i < buffer.length; i++) {
				const char = buffer[i];

				if (escape) {
					escape = false;
					continue;
				}

				if (char === '\\') {
					escape = true;
					continue;
				}

				if (char === '"') {
					inString = !inString;
					continue;
				}

				if (inString) continue;

				if (char === '{') {
					if (depth === 0) {
						objectStart = i;
					}
					depth++;
				} else if (char === '}') {
					depth--;
					if (depth === 0 && objectStart !== -1) {
						const objectStr = buffer.substring(objectStart, i + 1);
						try {
							const obj = JSON.parse(objectStr);
							if (
								obj &&
								typeof obj === 'object' &&
								typeof obj.id === 'number' &&
								(typeof obj.text === 'string' ||
									typeof obj.translation === 'string')
							) {
								const id = obj.id;
								const translation = obj.text || obj.translation || '';
								const sourceLanguage = obj.sourceLanguage || obj.src || '';
								pending.push({
									id,
									translation: [translation, sourceLanguage]
								});
							}
						} catch (e) {
							// Ignore parsing errors
						}
						objectStart = -1;
					}
				}
			}

			// Clear processed buffer
			if (objectStart === -1 && depth === 0) {
				buffer = '';
			}

			while (pending.length > 0) {
				yield pending.shift();
			}
		},
		end() {
			if (buffer.trim()) {
				try {
					const obj = JSON.parse(buffer);
					if (
						obj &&
						typeof obj === 'object' &&
						typeof obj.id === 'number' &&
						(typeof obj.text === 'string' ||
							typeof obj.translation === 'string')
					) {
						const id = obj.id;
						const translation = obj.text || obj.translation || '';
						const sourceLanguage = obj.sourceLanguage || obj.src || '';
						pending.push({ id, translation: [translation, sourceLanguage] });
					}
				} catch (e) {
					// Ignore
				}
			}
		}
	};
}

export function detectStreamFormat(content) {
	const stripped = stripMarkdownCodeBlock(content, true).trim();

	const jsonStart = stripped.search(/[{\[]/);
	const xmlStart = stripped.search(/<(t|item|seg)\s/i);
	const lineStart = stripped.search(/^\d+\s*\|/m);

	if (jsonStart === -1 && xmlStart === -1 && lineStart === -1) {
		return { isJson: false, detected: false };
	}

	const positions = [
		{ type: 'json', pos: jsonStart },
		{ type: 'xml', pos: xmlStart },
		{ type: 'line', pos: lineStart }
	].filter((p) => p.pos !== -1);

	if (positions.length === 0) {
		return { isJson: false, detected: false };
	}

	const first = positions.reduce((a, b) => (a.pos < b.pos ? a : b));
	return { isJson: first.type === 'json', detected: true };
}

export const parseAIRes = (raw, useBatchFetch = true) => {
	if (!raw) {
		return [];
	}

	if (!useBatchFetch) {
		return [[raw]];
	}

	let content = stripMarkdownCodeBlock(raw).trim();

	try {
		const start = content.search(/(\{|\[)/);
		const end = content.lastIndexOf(content.includes('}') ? '}' : ']');

		if (start > -1 && end > -1) {
			const jsonStr = content.substring(start, end + 1);
			const parsed = JSON.parse(jsonStr);

			const list = Array.isArray(parsed)
				? parsed
				: parsed.translations || (parsed.result ? [parsed.result] : [parsed]);

			if (
				list.length > 0 &&
				(list[0].text !== undefined || list[0].translations)
			) {
				return list.map((item) => [
					String(item.text || ''),
					String(item.sourceLanguage || '')
				]);
			}
		}
	} catch (e) {
		//
	}

	const xmlTagPattern = /<(t|item|seg)\b/i;
	if (xmlTagPattern.test(content)) {
		try {
			const parser = new DOMParser();
			const doc = parser.parseFromString(content, 'text/html');
			const elements = doc.querySelectorAll('t, item, seg');

			if (elements.length > 0) {
				return Array.from(elements).map((el) => [
					el.innerHTML.trim(),
					el.getAttribute('sourceLanguage') || ''
				]);
			}
		} catch (e) {
			//
		}
	}

	return content.split('\n').map((line) => {
		const pipeMatch = line.match(/^\d+\s*\|\s*(.*)/);
		if (pipeMatch) {
			return [pipeMatch[1].trim(), ''];
		}

		const text = line.replace(/<br\s*\/?>/gi, '\n').trim();
		return [text, ''];
	});
};

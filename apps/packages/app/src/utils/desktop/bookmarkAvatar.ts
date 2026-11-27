// Text-avatar fallback for bookmarks without a custom icon. We render the
// avatar as a data-URI SVG so it can be consumed by every existing
// `<img :src="icon">` surface (launchpad, mobile grid, dock) without any
// per-surface template change. The letter is the first grapheme of the
// bookmark name (uppercased for latin), e.g. GitHub -> G, 百度 -> 百, on a
// deterministic color derived from the name so each bookmark stays distinct.

const PALETTE = [
	'#4F8CFF',
	'#FF7A59',
	'#34C759',
	'#AF52DE',
	'#FF9500',
	'#FF2D55',
	'#5AC8FA',
	'#1F8A70'
];

export function bookmarkAvatarLetter(name: string): string {
	const s = (name || '').trim();
	if (!s) return '#';
	const first = Array.from(s)[0];
	return /[a-z]/i.test(first) ? first.toUpperCase() : first;
}

export function bookmarkAvatarColor(name: string): string {
	const s = name || '';
	let h = 0;
	for (let i = 0; i < s.length; i++) {
		h = (h * 31 + s.charCodeAt(i)) >>> 0;
	}
	return PALETTE[h % PALETTE.length];
}

export function bookmarkAvatarDataUri(name: string): string {
	const letter = bookmarkAvatarLetter(name);
	const bg = bookmarkAvatarColor(name);
	const safe = letter
		.replace(/&/g, '&amp;')
		.replace(/</g, '&lt;')
		.replace(/>/g, '&gt;');
	const svg =
		`<svg xmlns="http://www.w3.org/2000/svg" width="128" height="128" viewBox="0 0 128 128">` +
		`<rect width="128" height="128" rx="28" fill="${bg}"/>` +
		`<text x="50%" y="50%" dy=".35em" text-anchor="middle" ` +
		`font-family="Roboto, Arial, sans-serif" font-size="64" font-weight="600" ` +
		`fill="#ffffff">${safe}</text></svg>`;
	return `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}`;
}

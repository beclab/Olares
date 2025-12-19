const loaded: Map<string, Promise<any>> = new Map<string, Promise<any>>();
export function loadScript(src: string, global?: string): Promise<any> {
	if (loaded.has(src)) {
		return loaded.get(src)!;
	}

	const s = document.createElement('script');
	s.src = src;
	s.type = 'text/javascript';
	const p = new Promise((resolve, reject) => {
		s.onload = () => resolve(global ? window[global] : undefined);
		s.onerror = (e: any) => reject(e);
		document.head.appendChild(s);
	});

	loaded.set(src, p);
	return p;
}

export function isTouch(): boolean {
	return window.matchMedia('(hover: none)').matches;
}

export function openPopup(
	url = '',
	{
		name = '_blank',
		width = 500,
		height = 800
	}: {
		url?: string;
		name?: string;
		width?: number;
		height?: number;
	} = {}
): Window | null {
	const { outerHeight, outerWidth, screenX, screenY } = window;
	const top = outerHeight / 2 + screenY - height / 2;
	const left = outerWidth / 2 + screenX - width / 2;
	return window.open(
		url,
		name,
		`toolbar=0,scrollbars=1,status=1,resizable=1,location=1,menuBar=0,width=${width},height=${height},top=${top},left=${left}`
	);
}

// Trailing-edge debounce that returns a callable with .cancel() / .flush().
// Plan §5.4: every layout PUT goes through this so high-frequency drag-end
// events collapse into a single network round-trip.

export interface Debounced<TArgs extends unknown[]> {
	(...args: TArgs): void;
	cancel(): void;
	flush(): void;
	pending(): boolean;
}

export function debounce<TArgs extends unknown[]>(
	fn: (...args: TArgs) => void | Promise<void>,
	waitMs: number
): Debounced<TArgs> {
	let timer: ReturnType<typeof setTimeout> | null = null;
	let lastArgs: TArgs | null = null;

	const invoke = () => {
		if (!lastArgs) return;
		const a = lastArgs;
		lastArgs = null;
		timer = null;
		try {
			void fn(...a);
		} catch (e) {
			// Layout-save errors are surfaced inside fn (toast or log). We swallow
			// here so a thrown handler can't break the debounce timer.
			console.error('[desktop-debounce] handler threw', e);
		}
	};

	const debounced = ((...args: TArgs) => {
		lastArgs = args;
		if (timer) clearTimeout(timer);
		timer = setTimeout(invoke, waitMs);
	}) as Debounced<TArgs>;

	debounced.cancel = () => {
		if (timer) clearTimeout(timer);
		timer = null;
		lastArgs = null;
	};

	debounced.flush = () => {
		if (timer) {
			clearTimeout(timer);
			invoke();
		}
	};

	debounced.pending = () => timer !== null;

	return debounced;
}

const DEVICE_ID_KEY = 'desktop_device_id';

export function getDeviceId(): string {
	try {
		let id = localStorage.getItem(DEVICE_ID_KEY);
		if (!id) {
			id =
				'd-' +
				(typeof crypto !== 'undefined' && (crypto as any).randomUUID
					? (crypto as any).randomUUID()
					: Math.random().toString(36).slice(2) + Date.now().toString(36));
			localStorage.setItem(DEVICE_ID_KEY, id);
		}
		return id;
	} catch (_e) {
		return 'd-anon';
	}
}

export function genClientOpId(): string {
	return (
		'op-' +
		(typeof crypto !== 'undefined' && (crypto as any).randomUUID
			? (crypto as any).randomUUID().slice(0, 16)
			: Math.random().toString(36).slice(2, 10) + Date.now().toString(36))
	);
}

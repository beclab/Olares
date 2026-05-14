export class PlatformAdapter {
	static isBrowser(): boolean {
		return typeof window !== 'undefined';
	}

	static addEventListener(event: string, listener: (...args: any[]) => void) {
		if (this.isBrowser()) {
			window.addEventListener(event, listener);
		} else {
			console.warn(`addEventListener is not supported in this environment`);
		}
	}

	static removeEventListener(
		event: string,
		listener: (...args: any[]) => void
	) {
		if (this.isBrowser()) {
			window.removeEventListener(event, listener);
		} else {
			console.warn(`removeEventListener is not supported in this environment`);
		}
	}
}

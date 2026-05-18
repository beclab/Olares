export async function checkNetwork(url: string): Promise<boolean> {
	try {
		const response = await fetch(url, {
			method: 'HEAD',
			cache: 'no-cache'
		});
		return true;
	} catch (error) {
		return false;
	}
}

interface NetworkMonitor {
	startMonitoring: (callback?: (isOnline: boolean) => void) => void;
	stopMonitoring: () => void;
}

export function networkMonitor(
	url = '/',
	intervalTime = 10000
): NetworkMonitor {
	let timer: number | null = null;
	let networkDetectCallback: ((isOnline: boolean) => void) | undefined;

	const startMonitoring = (callback?: (isOnline: boolean) => void): void => {
		stopMonitoring();

		networkDetectCallback = callback;

		timer = window.setInterval(async () => {
			const isOnline = await checkNetwork(url);
			if (networkDetectCallback) {
				networkDetectCallback(isOnline);
			}
		}, intervalTime);
	};

	const stopMonitoring = (): void => {
		if (!timer) return;

		clearInterval(timer);
		timer = null;
		networkDetectCallback = undefined;
	};

	return { startMonitoring, stopMonitoring };
}

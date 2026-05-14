import { TerminusDefaultDomain } from './basic';

export const initPing: any = (
	did: string,
	timeout?: number,
	pingUrl?: string
): Promise<boolean> => {
	const ping_url =
		pingUrl ||
		`https://healthz.local.${did}.` + TerminusDefaultDomain + `/ping`;

	return new Promise((resolve) => {
		try {
			const xhr = new XMLHttpRequest();
			xhr.timeout = timeout || 1000;

			xhr.onreadystatechange = () => {
				if (xhr.readyState === 4) {
					if (xhr.status === 200) {
						if (xhr.responseText === 'pong') {
							resolve(true);
						} else {
							resolve(false);
						}
					} else {
						resolve(false);
					}
				}
			};

			xhr.onerror = () => {
				resolve(false);
			};

			try {
				xhr.open('GET', ping_url, true);
				try {
					xhr.send(null);
				} catch (e2) {
					//
				}
			} catch (e) {
				resolve(false);
			}
		} catch (e) {
			resolve(false);
		}
	});
};

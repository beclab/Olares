const rule = new RegExp('^[a-z0-9]{4,}$');

function isAlphaNumeric(code) {
	if (
		!(code > 47 && code < 58) && // numeric (0-9)
		//!(code > 64 && code < 91) && // upper alpha (A-Z)
		!(code > 96 && code < 123)
	) {
		// lower alpha (a-z)
		return false;
	}

	return true;
}

export function getTerminusNameFromVC(
	type: string,
	name: string
): string | undefined {
	if (type == 'Google') {
		const s = name.split('@');
		if (s.length != 2) {
			return undefined;
		}
		// if (s[1] != 'gmail.com') {
		// 	return undefined;
		// }
		let n = '';
		const tempN = s[0].toLocaleLowerCase();
		for (let i = 0; i < tempN.length; i++) {
			if (isAlphaNumeric(tempN.charCodeAt(i))) {
				n += tempN[i];
			}
		}
		if (!rule.test(n)) {
			return undefined;
		}
		if (n.length < 4 || n.length > 63) {
			return undefined;
		}
		return n;
	} else if (type == 'Twitter') {
		const n = name.toLocaleLowerCase();
		if (!rule.test(n)) {
			return undefined;
		}
		if (n.length < 4 || n.length > 63) {
			return undefined;
		}
		return n;
	} else {
		return undefined;
	}
}

const keyword = [
	'user',
	'system',
	'space',
	'default',
	'os',
	'kubesphere',
	'kube',
	'kubekey',
	'kubernetes',
	'gpu',
	'tapr',
	'bfl',
	'bytetrade',
	'project',
	'pod'
];

export function isWordMatchTerminusKeyword(word: string): boolean {
	if (keyword.indexOf(word.toLocaleLowerCase()) < 0) {
		return false;
	}
	return true;
}

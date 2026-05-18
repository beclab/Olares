export function maskToDottedDecimal(mask: number): string {
	if (mask < 0 || mask > 32) {
		throw new Error('子网掩码位数必须在 0 到 32 之间');
	}

	let binaryMask = '';
	for (let i = 0; i < 32; i++) {
		binaryMask += i < mask ? '1' : '0';
	}

	const parts: number[] = [];
	for (let i = 0; i < 32; i += 8) {
		const part = binaryMask.substr(i, 8);
		parts.push(parseInt(part, 2));
	}

	return parts.join('.');
}

export const stringifyParamsForPost = (
	params: Record<string, unknown>
): Record<string, string> => {
	const data: Record<string, string> = {};
	for (const [key, value] of Object.entries(params)) {
		if (value === undefined || value === null) continue;
		data[key] = String(value);
	}
	return data;
};

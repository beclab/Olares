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

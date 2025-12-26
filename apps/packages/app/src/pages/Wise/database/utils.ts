export function convertIntegerPropertiesToBoolean<T>(
	obj: T,
	properties: string[]
): T {
	const newObj = { ...obj };
	properties.forEach((property: string) => {
		if (typeof newObj[property] === 'number') {
			newObj[property] = newObj[property] === 1;
		}
	});
	return newObj;
}

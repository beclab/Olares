export const formatFrequency = (value: number, fromUnit = 'Hz'): string => {
	if (value === 0) return '0 Hz';

	const units = ['Hz', 'kHz', 'MHz', 'GHz'];
	const unitIndex = units.indexOf(fromUnit);

	let frequency = value;
	let currentUnitIndex = unitIndex;

	while (frequency >= 1000 && currentUnitIndex < units.length - 1) {
		frequency /= 1000;
		currentUnitIndex++;
	}

	while (frequency < 1 && currentUnitIndex > 0) {
		frequency *= 1000;
		currentUnitIndex--;
	}

	return `${Math.round(frequency * 100) / 100}${units[currentUnitIndex]}`;
};

export const convertTemperature = (
	celsius: number,
	targetUnit: 'C' | 'F' | 'K'
) => {
	switch (targetUnit) {
		case 'F':
			return celsius * 1.8 + 32;
		case 'K':
			return celsius + 273.15;
		default:
			return celsius;
	}
};

import BigNumber from 'bignumber.js';

export function worthValue(v: string | number) {
	const t = new BigNumber(v);

	let k = 4;
	if (t.isGreaterThanOrEqualTo(new BigNumber('1000'))) {
		k = 0;
	}

	if (t.isGreaterThanOrEqualTo(new BigNumber('100'))) {
		k = 1;
	}

	if (t.isGreaterThanOrEqualTo(new BigNumber('10'))) {
		k = 2;
	}

	if (t.isGreaterThanOrEqualTo(new BigNumber('1'))) {
		k = 3;
	}

	return new BigNumber(t.toPrecision(k)).toFormat();
}

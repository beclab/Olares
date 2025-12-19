import moment from 'moment-mini';

export const getLocalTime = (time: string | number) => {
	let formatTime = time;

	if (time && typeof time == 'string') {
		if (time.indexOf(' +0000 UTC') !== -1) {
			formatTime = time.replace(' +0000 UTC', 'Z').replace(' ', 'T');
		}
	}

	return moment.utc(formatTime).local();
};

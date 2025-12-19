export const expiresStorage = {
	setItem(key: string, value: string, expires: number) {
		const obj = {
			value: value,
			expires: expires,
			startTime: new Date().getTime()
		};
		if (obj.expires) {
			localStorage.setItem(key, JSON.stringify(obj));
		} else {
			localStorage.setItem(key, JSON.stringify(obj.value));
		}
	},

	getItem(key: string) {
		const temp = localStorage.getItem(key)
			? JSON.parse(localStorage.getItem(key) || '')
			: null;
		if (temp && temp.expires) {
			const data = new Date().getTime();
			if (data - temp.startTime > temp.expires) {
				localStorage.removeItem(key);
				return;
			} else {
				return temp.value;
			}
		} else {
			return temp;
		}
	}
};

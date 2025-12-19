type Callback<T> = (args?: T) => void;

class Waiter<T = void> {
	private intervalId?: NodeJS.Timeout;
	private timeoutId?: NodeJS.Timeout;

	public waitForCondition(
		condition: () => boolean,
		callback: Callback<T>,
		interval = 1000,
		args?: T,
		immediate = true,
		timeoutMs?: number
	): void {
		if (immediate) {
			if (condition()) {
				callback(args);
			}
		}

		this.intervalId = setInterval(() => {
			if (condition()) {
				callback(args);
				this.clear();
			}
		}, interval);

		if (timeoutMs) {
			this.timeoutId = setTimeout(() => {
				callback(args);
				this.clear();
			}, timeoutMs);
		}
	}

	public waitForTime(
		seconds: number,
		callback: Callback<T>,
		args?: T,
		immediate = true
	): void {
		if (immediate) {
			callback(args);
		}

		this.timeoutId = setTimeout(() => {
			callback(args);
		}, seconds * 1000);
	}

	public clear(): void {
		if (this.intervalId) {
			clearInterval(this.intervalId);
			this.intervalId = undefined;
		}
		if (this.timeoutId) {
			clearTimeout(this.timeoutId);
			this.timeoutId = undefined;
		}
	}
}

export default Waiter;

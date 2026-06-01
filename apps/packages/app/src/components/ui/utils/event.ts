type Callback<T> = (data: T) => void;

export class PubSub<T> {
	private subscribers: { [event: string]: Callback<T>[] } = {};

	emit(event: string, callback: Callback<T>): void {
		if (!this.subscribers[event]) {
			this.subscribers[event] = [];
		}
		this.subscribers[event].push(callback);
	}

	on(event: string, data: T): void {
		if (this.subscribers[event]) {
			this.subscribers[event].forEach((callback) => callback(data));
		}
	}

	unsubscribe(event: string, callback: Callback<T>): void {
		if (!this.subscribers[event]) return;

		this.subscribers[event] = this.subscribers[event].filter(
			(subscriber) => subscriber !== callback
		);
	}
}

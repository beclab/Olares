export class Typewriter {
	private queue: string[] = [];
	private consuming = false;
	private timmer: any;
	constructor(private onConsume: (str: string) => void) {}
	dynamicSpeed() {
		const speed = 2000 / this.queue.length;
		if (speed > 200) {
			return 200;
		} else {
			return speed;
		}
	}
	add(str: string) {
		if (!str) return;
		this.queue.push(...str.split(''));
	}
	consume() {
		if (this.queue.length > 0) {
			const str = this.queue.shift();
			str && this.onConsume(str);
		}
	}
	next() {
		this.consume();
		this.timmer = setTimeout(() => {
			this.consume();
			if (this.consuming) {
				this.next();
			}
		}, this.dynamicSpeed());
	}
	start() {
		this.consuming = true;
		this.next();
	}
	done() {
		this.consuming = false;
		clearTimeout(this.timmer);
		this.onConsume(this.queue.join(''));
		this.queue = [];
	}
}

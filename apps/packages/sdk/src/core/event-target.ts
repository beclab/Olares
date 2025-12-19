export interface Event {
	type: string;
	detail?: any;
}

export type Listener = ((e: Event) => void) | { handle: (e: Event) => void };

export interface EventTarget {
	addEventListener(eventName: string, listener: Listener): any;
	removeEventListener(eventName: string, listener: Listener): any;
}

export class EventEmitter implements EventTarget {
	private _listeners = new Map<string, Set<Listener>>();

	addEventListener(eventName: string, listener: Listener) {
		if (!this._listeners.has(eventName)) {
			this._listeners.set(eventName, new Set<Listener>());
		}
		this._listeners.get(eventName)!.add(listener);
	}

	removeEventListener(eventName: string, listener: Listener) {
		const listeners = this._listeners.get(eventName);
		if (listeners) {
			listeners.delete(listener);
		}
	}

	dispatchEvent(e: Event) {
		const listeners = this._listeners.get(e.type);
		if (listeners) {
			for (const listener of listeners) {
				if (typeof listener === 'function') {
					listener(e);
				} else {
					listener.handle(e);
				}
			}
		}
	}

	dispatch(type: string, detail?: any) {
		this.dispatchEvent({ type, detail });
	}
}

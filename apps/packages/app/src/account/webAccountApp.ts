import { v4 as uuidv4 } from 'uuid';
import { getAppPlatform } from '../application/platform';
import { DeviceInfo } from './device';

function newAccountState() {
	return {
		device: new DeviceInfo(),
		settings: {
			autoLock: true,
			autoLockDelay: 5,
			extensionBadge: true
		}
	};
}

export type WebAccountState = ReturnType<typeof newAccountState>;

// The Web platform has always kept this state in memory. The local user
// records and recovery phrases are persisted separately by the user store.
const accountStates = new Map<string, WebAccountState>();

export class WebAccountApp {
	state = newAccountState();
	private loadedId?: string;
	private readonly subscribers = new Set<(state: WebAccountState) => void>();

	get settings() {
		return this.state.settings;
	}

	async load(did: string) {
		const previous = accountStates.get(did) ?? newAccountState();
		const device = new DeviceInfo(previous.device);
		const info = await getAppPlatform().getDeviceInfo();
		for (const key of Object.keys(device)) {
			if (key !== 'id' && Object.prototype.hasOwnProperty.call(info, key)) {
				device[key] = info[key];
			}
		}
		device.id ||= uuidv4();

		this.loadedId = did;
		Object.assign(this.state, previous, { device });
		await this.save();
		this.publish();
	}

	async save() {
		if (this.loadedId) {
			accountStates.set(this.loadedId, {
				...this.state,
				device: new DeviceInfo(this.state.device),
				settings: { ...this.state.settings }
			});
		}
	}

	async removeState(did: string) {
		accountStates.delete(did);
		if (did === this.loadedId) {
			this.loadedId = undefined;
			Object.assign(this.state, newAccountState());
			this.publish();
		}
	}

	async setSettings(settings: Partial<WebAccountState['settings']>) {
		Object.assign(this.state.settings, settings);
		await this.save();
		this.publish();
	}

	subscribe(subscriber: (state: WebAccountState) => void) {
		this.subscribers.add(subscriber);
	}

	unsubscribe(subscriber: (state: WebAccountState) => void) {
		this.subscribers.delete(subscriber);
	}

	private publish() {
		for (const subscriber of this.subscribers) {
			subscriber(this.state);
		}
	}
}

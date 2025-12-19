import { ImageDetails, ImageInfoUpdate } from 'src/constant/constants';

export class ImageUpdater {
	private readonly imageMap: Map<string, ImageDetails>;
	private readonly updateCallback: (data: Map<string, ImageDetails>) => void;
	private timeoutId: NodeJS.Timeout | null = null;
	private readonly updateDelay: number;
	private hasNewData = false;

	constructor(
		updateCallback: (data: Map<string, ImageDetails>) => void,
		updateDelay = 3000
	) {
		if (typeof updateCallback !== 'function') {
			throw new Error('updateCallback error');
		}

		this.updateCallback = updateCallback;
		this.updateDelay = updateDelay;
		this.imageMap = new Map();
	}

	handleSocketUpdate(update: ImageInfoUpdate): void {
		if (!update?.image_info) return;

		const newImage = update.image_info;
		this.imageMap.set(newImage.name, newImage);
		this.hasNewData = true;
		this.resetTimeout();
	}

	private resetTimeout(): void {
		if (this.timeoutId) clearTimeout(this.timeoutId);

		this.timeoutId = setTimeout(() => {
			if (this.hasNewData) {
				this.updateCallback(this.imageMap);
				this.hasNewData = false;
			}
			this.timeoutId = null;
		}, this.updateDelay);
	}

	forceUpdate(): void {
		this.updateCallback(this.imageMap);
		this.hasNewData = false;

		if (this.timeoutId) {
			clearTimeout(this.timeoutId);
			this.timeoutId = null;
		}
	}

	destroy(): void {
		if (this.timeoutId) clearTimeout(this.timeoutId);
		this.imageMap.clear();
	}

	getCurrentData(): Map<string, ImageDetails> {
		return this.imageMap;
	}
}

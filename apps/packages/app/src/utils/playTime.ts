export type VideoTimeType = {
	path: string;
	time: number;
};

const localItemName = 'playTime';

export function setVideoCurTime(params: VideoTimeType) {
	const playTime = localStorage.getItem(localItemName);
	// if (playTime == 'null') return false;

	let playTimeLocal: Map<string, number> = new Map();

	if (playTime) {
		playTimeLocal = JSON.parse(playTime);
	} else {
		playTimeLocal = new Map();
	}
	playTimeLocal[params.path] = params.time.toString();
	localStorage.setItem(localItemName, JSON.stringify(playTimeLocal));
}

export function getVideoCurTime(path: string): number {
	const playTime = localStorage.getItem(localItemName);
	if (!playTime || playTime == 'null') return 0;
	const playTimeLocal = JSON.parse(playTime);
	return Number(playTimeLocal[path]);
}

export function removeVideoCurTime(path: string) {
	const playTime = localStorage.getItem(localItemName);
	if (!playTime) return false;
	const playTimeLocal = JSON.parse(playTime);
	if (playTimeLocal[path]) {
		delete playTimeLocal[path];
	}

	localStorage.setItem(localItemName, JSON.stringify(playTimeLocal));
}

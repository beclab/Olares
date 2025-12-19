import {
	Entry,
	ENTRY_STATUS,
	Feed,
	FILE_TYPE,
	SearchFeed,
	SimpleEntry
} from './rss-types';
import { ref, onMounted, onBeforeUnmount } from 'vue';
import { DownloadRecord } from 'src/utils/interface/rss';
import { BackupPlan, RestorePlan, BackupSnapshot } from '../constant';
import { i18n } from 'src/boot/i18n';

export const PATTERN_UTC_TIME =
	/^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]+Z/;

export function useWinSize() {
	const size = ref({
		width: document.documentElement.clientWidth || document.body.clientWidth,
		height: document.documentElement.clientHeight || document.body.clientHeight
	});

	function onResize() {
		size.value = {
			width: document.body.clientWidth,
			height: document.body.clientHeight
		};
	}

	onMounted(() => {
		addEventListener('resize', onResize);
	});

	onBeforeUnmount(() => {
		removeEventListener('resize', onResize);
	});

	return size;
}

export const useIsMobile = () => {
	const isMobile = ref(false);
	if (
		navigator.userAgent.match(
			/(phone|pad|pod|iPhone|iPod|ios|iPad|Android|Mobile|BlackBerry|IEMobile|MQQBrowser|JUC|Fennec|wOSBrowser|BrowserNG|WebOS|Symbian|Windows Phone)/i
		)
	) {
		isMobile.value = true;
	}
	if (document.body.clientWidth < 800) {
		isMobile.value = true;
	}

	return isMobile.value;
};

export const utcToStamp = (utc_datetime: string) => {
	const timestamp = new Date(utc_datetime).getTime();
	return Math.floor(timestamp / 1000);
};

export const getPastTime = (stamp1: Date, stamp2: Date) => {
	const time = stamp1.getTime() - stamp2.getTime();
	const second = time / 1000;
	const minute = second / 60;
	if (minute < 1) {
		return 'just now';
	}
	if (minute < 60) {
		return `${minute.toFixed(0)} minutes ago`;
	}

	const hour = minute / 60;
	if (hour < 24) {
		return `${hour.toFixed(0)} hours ago`;
	}

	const day = hour / 24;
	if (day < 30) {
		return `${day.toFixed(0)} days ago`;
	}

	const month = day / 30;

	if (month < 12) {
		return `${month.toFixed(0)} months ago`;
	}

	const year = month / 12;
	return `${year.toFixed(0)} years ago`;
};

export function calculateTimeDifference(
	timestamp1: string,
	timestamp2: string,
	suffix = ' ago'
): string {
	const time1: Date = new Date(timestamp1);
	const time2: Date = new Date(timestamp2);

	const timeDifference: number = time2.getTime() - time1.getTime();

	const days: number = Math.floor(timeDifference / (1000 * 60 * 60 * 24));
	const hours: number = Math.floor(timeDifference / 3600000);
	const minutes: number = Math.floor((timeDifference % 3600000) / 60000);
	const seconds: number = Math.floor((timeDifference % 60000) / 1000);

	let timeDifferenceString = '';
	if (days > 0) {
		timeDifferenceString += `${days}d`;
	}
	if (hours > 0) {
		timeDifferenceString += `${hours}h`;
	}
	if (minutes > 0 || hours > 0) {
		timeDifferenceString += `${minutes}m`;
	}
	if (seconds > 0 || (days === 0 && hours === 0 && minutes === 0)) {
		timeDifferenceString += ` ${seconds}s`;
	}
	if (days > 30) {
		const months = Math.floor(days / 30);
		timeDifferenceString = `${months} month${months > 1 ? 's' : ''}${suffix}`;
	} else {
		timeDifferenceString += suffix;
	}

	return timeDifferenceString;
}

export const nextRunTime = (stamp1: Date, stamp2: Date) => {
	const time = stamp2.getTime() - stamp1.getTime();

	const second = time / 1000;
	const minute = second / 60;
	if (minute < 1) {
		return 'just now';
	}
	if (minute < 60) {
		return `in ${minute.toFixed(0)} minutes`;
	}

	const hour = minute / 60;
	if (hour < 24) {
		return `in ${hour.toFixed(0)} hours`;
	}

	const day = hour / 24;
	if (day < 30) {
		return `in ${day.toFixed(0)} days`;
	}

	const month = day / 30;

	if (month < 12) {
		return `in ${month.toFixed(0)} months`;
	}

	const year = month / 12;
	return `in ${year.toFixed(0)} years`;
};

export function getRequireImage(path: string): string {
	if (!path) {
		return '';
	}
	if (path.startsWith('http')) {
		return path;
	}
	return require(`../assets/${path}`);
}

export function arrayBufferToBase64(buffer: Uint8Array): string {
	let binary = '';
	buffer.forEach((byte) => {
		binary += String.fromCharCode(byte);
	});
	return btoa(binary);
}

export function extractHtml(entry: Entry): string {
	if (downloadableFileTypes(entry?.file_type)) {
		return '';
	}
	if (entry?.full_content) {
		let data = '';
		const startIndex = entry?.full_content.indexOf('<div id="js_content"');
		let content = entry?.full_content;
		if (startIndex !== -1) {
			content = entry?.full_content.substring(startIndex);
		}
		let plainText = content.replace(/<[^>]+>/g, '');
		plainText = plainText.replace(/&[^s]*;/g, '');
		data = plainText.substring(0, 300);
		if (data) {
			return data;
		}
	}

	return '';
}

export function getFeedIcon(feed: Feed | SearchFeed | undefined) {
	if (feed && feed.icon_content && feed.icon_type) {
		if (feed.icon_content.startsWith(feed.icon_type)) {
			return `data:${feed.icon_content}`;
		}
		return `data:${feed.icon_type};base64,${feed.icon_content}`;
	}
	return '';
}

//https://github.com/bhowell2/binary-insert-js
export type Comparator<T> = (a: T, b: T) => number;

/**
 * Takes in a __SORTED__ array and inserts the provided value into
 * the correct, sorted, position.
 * @param array the sorted array where the provided value needs to be inserted (in order)
 * @param insertValue value to be added to the array
 * @param comparator function that helps determine where to insert the value (
 */
export function binaryInsert<T>(
	array: T[],
	insertValue: T,
	comparator: Comparator<T>
) {
	/*
	 * These two conditional statements are not required, but will avoid the
	 * while loop below, potentially speeding up the insert by a decent amount.
	 * */
	if (array.length === 0 || comparator(array[0], insertValue) >= 0) {
		array.splice(0, 0, insertValue);
		return array;
	} else if (
		array.length > 0 &&
		comparator(array[array.length - 1], insertValue) <= 0
	) {
		array.splice(array.length, 0, insertValue);
		return array;
	}
	let left = 0,
		right = array.length;
	let leftLast = 0,
		rightLast = right;
	while (left < right) {
		const inPos = Math.floor((right + left) / 2);
		const compared = comparator(array[inPos], insertValue);
		if (compared < 0) {
			left = inPos;
		} else if (compared > 0) {
			right = inPos;
		} else {
			right = inPos;
			left = inPos;
		}
		// nothing has changed, must have found limits. insert between.
		if (leftLast === left && rightLast === right) {
			break;
		}
		leftLast = left;
		rightLast = right;
	}
	// use right, because Math.floor is used
	array.splice(right, 0, insertValue);
	return array;
}

export const CompareRecentlyEntry = (a: Entry, b: Entry) => {
	if (a.last_opened > b.last_opened) {
		return -1;
	} else if (a.last_opened < b.last_opened) {
		return 1;
	} else {
		return a.id < b.id ? -1 : 1;
	}
};

export const CompareFeed = (a: Feed, b: Feed) => {
	return a.id < b.id ? -1 : 1;
};

export const CompareLibrarySimpleEntry = (
	a: Entry | SimpleEntry,
	b: Entry | SimpleEntry
) => {
	if (a.createdAt > b.createdAt) {
		return -1;
	} else if (a.createdAt < b.createdAt) {
		return 1;
	} else {
		return a.id < b.id ? -1 : 1;
	}
};

export const CompareReaderLaterSimpleEntry = (
	a: SimpleEntry,
	b: SimpleEntry
) => {
	if (a.updatedAt > b.updatedAt) {
		return -1;
	} else if (a.updatedAt < b.updatedAt) {
		return 1;
	} else {
		return a.id < b.id ? -1 : 1;
	}
};

export const CompareBackup = (
	a: BackupPlan | RestorePlan | BackupSnapshot,
	b: BackupPlan | RestorePlan | BackupSnapshot
) => {
	if (a.createAt > b.createAt) {
		return -1;
	} else if (a.createAt < b.createAt) {
		return 1;
	} else {
		return a.id > b.id ? -1 : 1;
	}
};

export const CompareTrendSimpleEntry = (
	a: SimpleEntry | Entry,
	b: SimpleEntry | Entry
) => {
	if (a.published_at && b.published_at) {
		if (a.published_at > b.published_at) {
			return -1;
		} else if (a.published_at < b.published_at) {
			return 1;
		} else {
			return a.id < b.id ? -1 : 1;
		}
	} else {
		if (a.createdAt > b.createdAt) {
			return -1;
		} else if (a.createdAt < b.createdAt) {
			return 1;
		} else {
			return a.id < b.id ? -1 : 1;
		}
	}
};

export const CompareDownloadRecord = (a: DownloadRecord, b: DownloadRecord) => {
	const aTimestamp = new Date(a.created_time).getTime();
	const bTimestamp = new Date(b.created_time).getTime();
	if (aTimestamp > bTimestamp) {
		return -1;
	} else if (aTimestamp < bTimestamp) {
		return 1;
	} else {
		return a.id < b.id ? -1 : 1;
	}
};

export function downloadableFileTypes(fileType: string): boolean {
	return (
		fileType === FILE_TYPE.PDF ||
		fileType === FILE_TYPE.EBOOK ||
		fileType === FILE_TYPE.VIDEO ||
		fileType === FILE_TYPE.AUDIO
	);
}

export function downloadedFileTypes(entry: SimpleEntry): boolean {
	return (
		(downloadableFileTypes(entry.file_type) &&
			entry.status === ENTRY_STATUS.Completed) ||
		entry.file_type === FILE_TYPE.ARTICLE
	);
}

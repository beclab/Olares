import { FilePath } from 'src/stores/files';

const STORAGE_KEY = 'uploadHistory';
const DEFAULT_MAX_SAVE = 20;
const DEFAULT_FETCH_NUM = 5;

export const saveUploadHistory = (
	fileItem: FilePath,
	maxSave: number = DEFAULT_MAX_SAVE
) => {
	try {
		const rawHistory = localStorage.getItem(STORAGE_KEY);
		let historyList: Array<{
			createTime: number;
			data: Partial<FilePath>;
		}> = rawHistory ? JSON.parse(rawHistory) : [];

		const fileData = { ...fileItem };
		const newRecord = {
			createTime: Date.now(),
			data: fileData
		};

		historyList = historyList.filter(
			(item) =>
				!(
					item.data.path === fileItem.path &&
					item.data.driveType === fileItem.driveType
				)
		);

		historyList.unshift(newRecord);
		if (historyList.length > maxSave) {
			historyList = historyList.slice(0, maxSave);
		}

		localStorage.setItem(STORAGE_KEY, JSON.stringify(historyList));
	} catch (error) {
		console.error('保存上传历史失败：', error);
	}
};

export const getUploadHistory = (
	fetchNum: number = DEFAULT_FETCH_NUM
): FilePath[] => {
	try {
		const rawHistory = localStorage.getItem(STORAGE_KEY);
		if (!rawHistory) return [];

		const historyList: Array<{
			createTime: number;
			data: Partial<FilePath>;
		}> = JSON.parse(rawHistory);

		historyList.sort((a, b) => b.createTime - a.createTime);

		const targetList =
			fetchNum > 0 ? historyList.slice(0, fetchNum) : historyList;

		return targetList.map((item) => new FilePath(item.data));
	} catch (error) {
		console.error('读取上传历史失败：', error);
		return [];
	}
};

export const clearUploadHistory = () => {
	try {
		localStorage.removeItem(STORAGE_KEY);
	} catch (error) {
		console.error('清空上传历史失败：', error);
	}
};

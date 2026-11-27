import { Component, markRaw } from 'vue';
import {
	getFilePreviewType,
	PreviewType,
	txtReadonlyTypes
} from '@bytetrade/core';
import { dataAPIs } from '../../../api';
import { FileItem } from '../../../stores/files';

export interface FilePreviewComponents {
	text: Component;
	ebook: Component;
	pdf: Component;
	media: Component;
	unavailable: Component;
}

export interface ResolveFilePreviewOptions {
	usePdfComponent?: boolean;
	mediaTypes?: PreviewType[];
	darkTypes?: PreviewType[];
	textComponentResolver?: (item: FileItem) => Component;
}

export const defaultMediaPreviewTypes = [
	PreviewType.IMAGE,
	PreviewType.AUDIO,
	PreviewType.PDF,
	PreviewType.WORD,
	PreviewType.EXCEL
];

export const isImagePreviewItem = (item: any) =>
	!!item && !item.isDir && getFilePreviewType(item.name) === PreviewType.IMAGE;

export const getPreviewPermission = (
	permission: string | number | undefined
) => {
	if (typeof permission === 'string') {
		return permission;
	}
	if (typeof permission === 'number') {
		return permission >= 3 ? 'rw' : 'r';
	}
	return undefined;
};

export const resolveFilePreview = (
	item: FileItem,
	components: FilePreviewComponents,
	options: ResolveFilePreviewOptions = {}
) => {
	const previewType = getFilePreviewType(item.name);
	const dataAPI = dataAPIs(item.driveType);
	const mediaTypes = options.mediaTypes || defaultMediaPreviewTypes;
	const darkTypes = options.darkTypes || [];
	let component: Component = components.unavailable;
	let isDark = darkTypes.includes(previewType);
	let isEditEnable = false;

	if (previewType === PreviewType.TXT) {
		isEditEnable =
			dataAPI.fileEditEnable &&
			item.size <= dataAPI.fileEditLimitSize &&
			!txtReadonlyTypes.includes((item.extension || '').replace('.', ''));
		component = options.textComponentResolver?.(item) || components.text;
	} else if (previewType === PreviewType.EBOOK) {
		component = components.ebook;
	} else if (previewType === PreviewType.PDF && options.usePdfComponent) {
		component = components.pdf;
	} else if (mediaTypes.includes(previewType)) {
		component =
			previewType === PreviewType.AUDIO && !dataAPI.audioPlayEnable
				? components.unavailable
				: components.media;
	} else {
		component = components.unavailable;
		isDark = false;
	}

	return {
		component: markRaw(component),
		previewType,
		isEditEnable,
		isDark
	};
};

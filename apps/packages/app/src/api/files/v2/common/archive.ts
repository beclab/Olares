import {
	ArchiveConflict,
	ArchiveFormat,
	getArchiveErrorMessage,
	isArchivePasswordError
} from 'src/utils/interface/archive';
import { commonUrlPrefix } from './utils';
import { appendPath } from '../path';
import { CommonFetch } from '../../fetch';
import { notifyFailed } from 'src/utils/notifyRedefinedUtil';
import { Dialog } from 'quasar';
import ArchivePasswordDialog from 'src/components/files/archive/ArchivePasswordDialog.vue';
import { i18n } from 'src/boot/i18n';

const compress = async (
	params: {
		sources: string[];
		destination: string;
		format: ArchiveFormat;
		level?: number;
		conflict?: ArchiveConflict;
		preserveSymlinks?: boolean;
	},
	options: {
		volumeSizeMB?: number;
	},
	node: string,
	// user: string,
	password?: string
): Promise<{ task_id: string }> => {
	const url = appendPath(commonUrlPrefix('archive'), node || '', 'compress');
	const formatParams: {
		sources: string[];
		destination: string;
		format: ArchiveFormat;
		level: number;
		volumeSizeMB?: number;
		preserveSymlinks: boolean;
		conflict: ArchiveConflict;
	} = {
		sources: params.sources,
		destination: params.destination,
		format: params.format,
		level: params.level || 5,
		conflict: params.conflict || ArchiveConflict.Rename,
		preserveSymlinks: params.preserveSymlinks || false
	};

	if (
		params.format === ArchiveFormat.SevenZ ||
		params.format === ArchiveFormat.Zip
	) {
		const volumeSizeMB = Number(options.volumeSizeMB);
		if (Number.isInteger(volumeSizeMB) && volumeSizeMB > 0) {
			formatParams.volumeSizeMB = volumeSizeMB;
		}
	}

	const compressWithPassword = async (
		archivePassword?: string
	): Promise<{ task_id: string }> => {
		try {
			return await tryCompress(url, formatParams, archivePassword);
		} catch (error) {
			if (!isArchivePasswordError(error)) {
				notifyFailed(getArchiveErrorMessage(error) || error.message);
				throw error;
			}

			const nextPassword = await requestArchivePassword(
				params.destination,
				getArchiveErrorMessage(error),
				archivePassword
			);
			return compressWithPassword(nextPassword);
		}
	};

	return compressWithPassword(password);
};

const extract = async (
	params: {
		source: string;
		destination: string;
		conflict?: ArchiveConflict;
		preserveSymlinks?: boolean;
	},
	options: {
		format?: ArchiveFormat;
	},
	node: string,
	password?: string
): Promise<{ task_id: string }> => {
	const url = appendPath(commonUrlPrefix('archive'), node || '', 'extract');

	const formatParams: {
		source: string;
		destination: string;
		format?: ArchiveFormat;
		preserveSymlinks: boolean;
		conflict: ArchiveConflict;
	} = {
		source: params.source,
		destination: params.destination,
		format: options.format,
		conflict: params.conflict || ArchiveConflict.Rename,
		preserveSymlinks: params.preserveSymlinks || false
	};

	const extractWithPassword = async (
		archivePassword?: string
	): Promise<{ task_id: string }> => {
		try {
			return await tryExtract(url, formatParams, archivePassword);
		} catch (error) {
			if (!isArchivePasswordError(error)) {
				notifyFailed(getArchiveErrorMessage(error) || error.message);
				throw error;
			}

			const nextPassword = await requestArchivePassword(
				params.source,
				getArchiveErrorMessage(error),
				archivePassword
			);
			return extractWithPassword(nextPassword);
		}
	};

	return extractWithPassword(password);
};

const ensureTaskId = (result: any) => {
	if (!result?.task_id) {
		throw new Error('Archive task id is missing');
	}
	return { task_id: result.task_id };
};

const tryCompress = async (
	url: string,
	formatParams: any,
	password?: string
): Promise<{ task_id: string }> => {
	const result = (
		await CommonFetch.post(url, formatParams, {
			headers: {
				// Keep Content-Type explicit.
				'Content-Type': 'application/json',
				'X-Archive-Password': password
			},
			noToast: true
		})
	).data;
	return ensureTaskId(result);
};

const tryExtract = async (
	url: string,
	formatParams: any,
	password?: string
): Promise<{ task_id: string }> => {
	const result = (
		await CommonFetch.post(url, formatParams, {
			headers: {
				// Keep Content-Type explicit.
				'Content-Type': 'application/json',
				'X-Archive-Password': password
			},
			noToast: true
		})
	).data;
	return ensureTaskId(result);
};

const requestArchivePassword = (
	source: string,
	errorMessage?: string,
	defaultPassword?: string
) => {
	return new Promise<string>((resolve, reject) => {
		let settled = false;
		Dialog.create({
			component: ArchivePasswordDialog,
			componentProps: {
				archiveName: source.split('/').pop() || '',
				errorMessage: errorMessage || '',
				errorPassword: defaultPassword || '',
				title: i18n.global.t('files.archive.unlock_archive'),
				okText: defaultPassword
					? i18n.global.t('files.archive.try_again')
					: i18n.global.t('files.archive.confirm_password')
			}
		})
			.onOk((value: string) => {
				settled = true;
				resolve(value);
			})
			.onCancel(() => {
				settled = true;
				reject(new Error('archive_password_cancelled'));
			})
			.onDismiss(() => {
				if (!settled) {
					reject(new Error('archive_password_cancelled'));
				}
			});
	});
};

const getPreviewContentsStreamUrl = (
	params: {
		source: string;
		password?: string;
	},
	node: string
) => {
	const query = new URLSearchParams({
		source: params.source
	});
	return `${appendPath(
		commonUrlPrefix('archive'),
		node || '',
		'entries'
	)}?${query.toString()}`;
};

const getPreviewStreamUrl = (
	params: {
		source: string;
		path: string;
		password?: string;
	},
	node: string
) => {
	const query = new URLSearchParams({
		source: params.source,
		path: params.path
	});
	return `${appendPath(
		commonUrlPrefix('archive'),
		node || '',
		'entry'
	)}?${query.toString()}`;
};

export default {
	compress,
	extract,
	getPreviewContentsStreamUrl,
	getPreviewStreamUrl
};

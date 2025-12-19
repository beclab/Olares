import { useQuasar } from 'quasar';
import {
	SharePermission,
	ShareResult,
	ShareType
} from 'src/utils/interface/share';
import { notifyFailed, notifySuccess } from 'src/utils/notifyRedefinedUtil';
import { computed, ref } from 'vue';
import { i18n } from 'src/boot/i18n';
import { useFilesStore } from 'src/stores/files';
import zhCn from 'element-plus/dist/locale/zh-cn.mjs';
import en from 'element-plus/dist/locale/en.mjs';
import { generatePasword } from 'src/utils/format';
import share from '../../../../api/files/v2/common/share';
import { useDataStore } from 'src/stores/data';
import { getApplication } from 'src/application/base';

export enum DiskUnitMode {
	T = 'Ti',
	G = 'Gi',
	M = 'Mi'
}

export const diskUnitOptions = () => {
	return [
		{
			value: DiskUnitMode.M,
			label: DiskUnitMode.M
		},
		{
			value: DiskUnitMode.G,
			label: DiskUnitMode.G
		},
		{
			value: DiskUnitMode.T,
			label: DiskUnitMode.T
		}
	];
};

export function usePublicShare(origin_id: number) {
	const shareResult = ref<ShareResult | undefined>(undefined);

	const filesStore = useFilesStore();

	const $q = useQuasar();

	const publicPassword = ref(generatePasword(6));
	const setExpirationInDays = ref(false);
	const days = ref('7');

	const dateValue = ref<string>('');

	const uploadOnly = ref(false);

	const uploadLimiteOpen = ref(false);
	const uploadFileSizeLimit = ref('0');
	const uploadFileSizeUnit = ref(DiskUnitMode.G);

	const store = useDataStore();

	const copyShareLink = (e) => {
		e.stopPropagation();
		const copyTxt = getShareLink.value;
		getApplication()
			.copyToClipboard(copyTxt)
			.then(() => {
				notifySuccess(i18n.global.t('copy_success'));
			})
			.catch(() => {
				notifyFailed(i18n.global.t('copy_fail'));
			});
	};

	const copyLinkAndPassword = (e) => {
		e.stopPropagation();

		const copyTxt =
			i18n.global.t('Link') +
			':' +
			getShareLink.value +
			'<br>' +
			i18n.global.t('password') +
			':' +
			publicPassword.value;

		getApplication()
			.copyToClipboard(copyTxt.replace(/<br>/g, '\r\n'))
			.then(() => {
				notifySuccess(i18n.global.t('copy_success'));
			})
			.catch(() => {
				notifyFailed(i18n.global.t('copy_fail'));
			});
	};

	const getShareLink = computed(() => {
		if (shareResult.value) {
			return filesStore.getShareLinkAddress(shareResult.value.id);
		}
		return '';
	});

	const lang = computed(() =>
		i18n.global.locale.value.substring(0, 2) === 'zh' ? zhCn : en
	);

	const fileLimitSize = () => {
		let base = 1024 * 1024;
		if (uploadFileSizeUnit.value == DiskUnitMode.G) {
			base = base * 1024;
		} else if (uploadFileSizeUnit.value == DiskUnitMode.T) {
			base = base * base;
		}
		return base * Number(uploadFileSizeLimit.value);
	};

	const removeShare = async () => {
		if (!shareResult.value) {
			return;
		}
		try {
			$q.loading.show();
			await share.remove(shareResult.value!.id);
			$q.loading.hide();
			shareResult.value = undefined;
		} catch (error) {
			$q.loading.hide();
			notifyFailed(error.message);
		}
	};

	const disabledDate = (time: Date) => {
		return time.getTime() < new Date().setHours(0, 0, 0, 0);
	};

	const createPublicShare = async () => {
		try {
			const index = filesStore.selected[origin_id][0];
			const file = filesStore.getTargetFileItem(index, origin_id);

			if (!file) {
				return false;
			}

			const option = {
				name: decodeURI(file.name),
				share_type: ShareType.PUBLIC,
				permission: uploadOnly.value
					? SharePermission.UploadOnly
					: SharePermission.Edit,
				password: publicPassword.value,
				upload_size_limit: !uploadLimiteOpen.value ? 0 : fileLimitSize()
			} as any;

			if (setExpirationInDays.value) {
				option['expire_in'] = parseInt(days.value) * 24 * 3600 * 1000;
			} else {
				option['expire_time'] = new Date(dateValue.value).toISOString();
			}
			$q.loading.show();
			shareResult.value = await share.create(file, option);
			$q.loading.hide();
		} catch (error) {
			$q.loading.hide();
			return undefined;
		}
	};

	const datesLimitRule = (val: string) => {
		if (val.length === 0) {
			return i18n.global.t('errors.please_type_something');
		}
		const rule = /^[+-]?(\d+\.?\d*|\.\d+)$/;
		if (!rule.test(val)) {
			return i18n.global.t('errors.only_valid_numbers_can_be_entered');
		}
		return '';
	};

	const passwordLimitRule = (val: string) => {
		if (val.length === 0) {
			return i18n.global.t('errors.please_type_something');
		}
		if (val.length < 6) {
			return i18n.global.t('errors.share_public_password_error');
		}
		return '';
	};

	const onDisabled = computed(() => {
		if (!shareResult.value) {
			return (
				passwordLimitRule(publicPassword.value).length > 0 ||
				(setExpirationInDays.value && datesLimitRule(days.value).length > 0) ||
				(!setExpirationInDays.value && !dateValue.value)
			);
		}

		return false;
	});

	const filesSizeLimitRule = (val: string) => {
		const rule = /^[+-]?(\d+\.?\d*|\.\d+)$/;
		if (!rule.test(val)) {
			return i18n.global.t('errors.only_valid_numbers_can_be_entered');
		}
		return '';
	};

	const onCancel = () => {
		store.closeHovers();
	};

	return {
		copyShareLink,
		lang,
		shareResult,
		publicPassword,
		setExpirationInDays,
		days,
		dateValue,
		removeShare,
		disabledDate,
		getShareLink,
		createPublicShare,
		uploadOnly,
		onDisabled,
		passwordLimitRule,
		datesLimitRule,
		onCancel,
		copyLinkAndPassword,
		uploadLimiteOpen,
		uploadFileSizeLimit,
		uploadFileSizeUnit,
		filesSizeLimitRule
	};
}

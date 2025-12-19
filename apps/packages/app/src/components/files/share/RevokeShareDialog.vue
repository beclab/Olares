<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('files_popup_menu.Revoke sharing')"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		:okLoading="loading ? t('loading') : false"
		size="small"
		:platform="$q.platform.is.mobile ? 'mobile' : 'web'"
		@onSubmit="submit"
		@onHide="onCancel"
		@onCancel="onCancel"
	>
		<div
			class="dialog-desc"
			:style="{ textAlign: isMobile ? 'center' : 'left' }"
		>
			{{ t('files.Revoke Olares sharing?') }}
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import { ref } from 'vue';

import {
	notifyWaitingShow,
	notifyHide
} from '../../../utils/notifyRedefinedUtil';

import { useDataStore } from '../../../stores/data';
import { useFilesStore, FilesIdType } from '../../../stores/files';
import { useOperateinStore } from './../../../stores/operation';
import { dataAPIs } from '../../../api';
import { encodePath } from '../../../utils/url';

const props = defineProps({
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	}
});

const store = useDataStore();
const filesStore = useFilesStore();
const operateinStore = useOperateinStore();
const { t } = useI18n();

const $q = useQuasar();
const isMobile = ref(process.env.PLATFORM == 'MOBILE' || $q.platform.is.mobile);

const CustomRef = ref();
const loading = ref(false);

const submit = async () => {
	loading.value = true;
	notifyWaitingShow(t('Deleting, Please wait...'));
	operateinStore.executing = true;
	try {
		let selectFiles: any[] = [];

		if (store.preview.isShow) {
			const previewItem = JSON.parse(
				JSON.stringify(filesStore.previewItem[props.origin_id])
			);
			if (previewItem.front) {
				previewItem.path = encodePath(
					dataAPIs(previewItem.driveType).getPanelJumpPath(previewItem)
				);
			}

			selectFiles.push(previewItem);
		} else {
			selectFiles =
				filesStore.currentFileList[props.origin_id]?.items.filter((obj) =>
					filesStore.selected[props.origin_id].includes(obj.index)
				) || [];
		}
		const dataAPI = dataAPIs();
		await dataAPI.deleteItem(selectFiles);
		store.showConfirm && store.showConfirm();
		CustomRef.value.onDialogOK();
		loading.value = false;
		filesStore.resetSelected(props.origin_id);
		const currentPath = filesStore.currentPath[props.origin_id];
		await filesStore.refushCurrentRouter(
			currentPath.path + currentPath.param,
			filesStore.activeMenu(props.origin_id).driveType,
			props.origin_id
		);

		operateinStore.executing = false;
		notifyHide();
	} catch (error) {
		operateinStore.executing = false;
		notifyHide();
		loading.value = false;
	}
};

const onCancel = () => {
	store.closeHovers();
};
</script>

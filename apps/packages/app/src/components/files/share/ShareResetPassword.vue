<template>
	<bt-custom-dialog
		ref="CustomRef"
		:skip="false"
		:title="t('files_popup_menu.Reset Password')"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		size="medium"
		:okDisabled="publicPassword.length == 0"
		:platform="deviceStore.platform"
		@onCancel="onCancel"
		@onSubmit="onSubmit"
		@onHide="onCancel"
	>
		<GeneratePassword v-model="publicPassword" :copy="deviceStore.isMobile" />
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import GeneratePassword from './GeneratePassword.vue';
import { useI18n } from 'vue-i18n';
import { useDataStore } from 'src/stores/data';
import { useDeviceStore } from 'src/stores/settings/device';
import share from 'src/api/files/v2/common/share';
import { FilesIdType } from 'src/stores/files';
import { useFilesStore } from 'src/stores/files';
import { notifySuccess } from 'src/utils/notifyRedefinedUtil';
import { useQuasar } from 'quasar';

const props = defineProps({
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	}
});

const CustomRef = ref();

const publicPassword = ref('');
const { t } = useI18n();

const store = useDataStore();

const deviceStore = useDeviceStore();
const $q = useQuasar();

const onSubmit = async () => {
	const filesStore = useFilesStore();
	if (filesStore.selected[props.origin_id].length <= 0) {
		return;
	}
	const index = filesStore.selected[props.origin_id][0];
	const file = filesStore.getTargetFileItem(index, props.origin_id);
	if (file && file.isShareItem && file.id) {
		// return;
		try {
			$q.loading.show();
			await share.resetPassword(file.id, publicPassword.value);
			$q.loading.hide();
			notifySuccess(t('success'));
			store.closeHovers();
		} catch (error) {
			$q.loading.hide();
			console.log(error);
		}
	}
};

const onCancel = () => {
	store.closeHovers();
};
</script>

<style scoped lang="scss">
// :deep(.dialog-content) {
// 	margin: 0px 0px 10px !important;
// }
</style>

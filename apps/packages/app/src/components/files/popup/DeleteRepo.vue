<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('delete')"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		:ok-loading="submitLoading ? t('loading') : false"
		size="small"
		@onSubmit="submit"
	>
		<div
			class="dialog-desc text-ink-2"
			:style="{ textAlign: isMobile ? 'center' : 'left' }"
		>
			<div>
				{{
					t('files.Are you sure you want to delete {repo}', {
						repo: item?.repo_name
					})
				}}
			</div>
			<div v-if="shared_length && shared_length > 0" class="text-red">
				{{
					t('files.This library has been shared to {count} user(s)', {
						user: shared_length
					})
				}}
			</div>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { useQuasar } from 'quasar';
import { ref } from 'vue';
import { dataAPIs } from '../../../api';
import { useFilesStore } from '../../../stores/files';
import { useI18n } from 'vue-i18n';
import { DriveType } from '../../../utils/interface/files';
import { useOperateinStore } from './../../../stores/operation';

const props = defineProps({
	item: {
		type: Object,
		required: false
	},
	shared_length: {
		type: Number,
		require: false
	}
});

const $q = useQuasar();
const filesStore = useFilesStore();
const operateinStore = useOperateinStore();

const { t } = useI18n();

const isMobile = ref(process.env.PLATFORM == 'MOBILE' || $q.platform.is.mobile);
const CustomRef = ref();

const submitLoading = ref(false);
const dataAPI = dataAPIs(DriveType.Sync) as any;

const submit = async () => {
	submitLoading.value = true;
	try {
		await dataAPI.deleteRepo(props.item as any);
		submitLoading.value = false;
		CustomRef.value.onDialogOK();
		await filesStore.getMenu();
		await filesStore.setBrowserUrl(
			operateinStore.defaultPath,
			DriveType.Drive,
			true
		);
	} catch (error) {
		submitLoading.value = false;
	}
};
</script>

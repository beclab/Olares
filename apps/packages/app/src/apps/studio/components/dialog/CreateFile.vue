<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="title"
		:cancel="t('btn_cancel')"
		:ok="t('btn_confirm')"
		:okLoading="loading ? $t('loading') : false"
		size="small"
		@onSubmit="submit"
	>
		<div class="form-item">
			<div class="text-subtitle2 text-ink-3 q-mb-xs">
				{{ t('dialog_create_title') }}
			</div>
			<div class="form-item-value">
				<q-input
					ref="fileNameRef"
					dense
					borderless
					no-error-icon
					v-model="appName"
					autofocus
					class="form-item-input"
					input-class="text-ink-2"
					:placeholder="ruleConfig.file.placeholder"
					:rules="ruleConfig.file.rules"
				>
				</q-input>
			</div>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { ruleConfig } from './../../types/config';
import { useFileStore } from '../../stores/file';
import { OPERATE_ACTION } from '@apps/studio/src/types/core';

const props = defineProps({
	path: {
		type: String,
		required: true
	},
	label: {
		type: String,
		required: false
	},
	action: {
		type: String,
		required: false,
		default: OPERATE_ACTION.ADD_FILE
	}
});

const { t } = useI18n();

const fileStore = useFileStore();

const CustomRef = ref();
const fileNameRef = ref();
const loading = ref(false);
const appName = ref(props.label || '');
const title = ref();

const submit = async () => {
	console.log('click submit');
	fileNameRef.value.validate();
	if (fileNameRef.value.hasError) return;

	loading.value = true;

	try {
		switch (props.action) {
			case OPERATE_ACTION.ADD_FILE:
				await createFile();
				break;

			case OPERATE_ACTION.ADD_FOLDER:
				await createFolder();
				break;

			case OPERATE_ACTION.RENAME:
				await renameFile();
				break;
		}

		loading.value = false;

		CustomRef.value.onDialogOK(appName.value);
	} catch (error) {
		loading.value = false;
	}
};

const createFile = async () => {
	const currentPath = `${props.path}/${appName.value}`;
	await fileStore.putFile(currentPath);
};

const createFolder = async () => {
	const currentPath = `${props.path}/${appName.value}`;
	await fileStore.createFolder(currentPath);
};

const renameFile = async () => {
	const newPath = props.path.replace(props.label!, appName.value);
	const currentPath = `${props.path}?action=rename&destination=${newPath}`;

	await fileStore.pathFile(
		currentPath,
		{},
		{
			headers: { 'content-type': 'text/plain' }
		}
	);
};

onMounted(() => {
	switch (props.action) {
		case OPERATE_ACTION.ADD_FILE:
			title.value = t('dialog_create_file');
			break;

		case OPERATE_ACTION.ADD_FOLDER:
			title.value = t('dialog_create_folder');
			break;

		case OPERATE_ACTION.DELETE:
			title.value = t('message.rename_file');
			break;

		default:
			title.value = t('dialog_create_file');
			break;
	}
});
</script>

<style lang="scss" scoped>
.form-item {
	.form-item-key {
		width: 100px;
		height: 40px;
		line-height: 40px;
		text-align: center;
	}
	.form-item-value {
		flex: 1;
	}
}
</style>
../../types/config

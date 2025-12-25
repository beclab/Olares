<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('prompts.newDir')"
		:ok="t('buttons.create')"
		:okLoading="loading ? t('loading') : false"
		:cancel="t('buttons.cancel')"
		:size="$q.platform.is.mobile ? 'small' : 'medium'"
		:platform="$q.platform.is.mobile ? 'mobile' : 'web'"
		@onSubmit="submit"
		@onHide="handleClose"
		@onCancel="handleClose"
	>
		<div class="card-content">
			<p class="text-ink-3">{{ t('prompts.newFileMessage') }}</p>
			<input
				class="input input--block text-ink-1"
				v-focus
				ref="inputRef"
				type="text"
				@keyup.enter="submit"
				v-model.trim="name"
			/>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref, nextTick } from 'vue';
import { useDataStore } from '../../../stores/data';
import { dataAPIs } from '../../../api';
import { useFilesStore, FilesIdType } from '../../../stores/files';
import { notifyWarning } from '../../../utils/notifyRedefinedUtil';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';

import { useI18n } from 'vue-i18n';

const props = defineProps({
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	}
});

const store = useDataStore();
const filesStore = useFilesStore();
const name = ref<string>('');
const inputRef = ref();
const { t } = useI18n();
const loading = ref(false);

const CustomRef = ref();

const submit = async () => {
	if (!name.value) {
		notifyWarning('The input content cannot be empty!');
		return false;
	}

	if (name.value.includes('\\') || name.value.includes('/')) {
		BtNotify.show({
			type: NotifyDefinedType.WARNING,
			message: t('files.backslash_create')
		});

		return false;
	}

	// const dataAPI = dataAPIs();
	loading.value = true;

	const currentPath = filesStore.currentPath[props.origin_id];

	const dataAPI = dataAPIs(currentPath.driveType, props.origin_id);

	try {
		await dataAPI.createDir(name.value, currentPath.path);
		await filesStore.refushCurrentRouter(
			currentPath.path + currentPath.param,
			filesStore.activeMenu(props.origin_id).driveType,
			props.origin_id
		);

		loading.value = false;
		store.closeHovers();
		CustomRef.value.onDialogOK();
	} catch (error) {
		loading.value = false;
		console.log(error);
	}
};

const handleClose = () => {
	store.closeHovers();
};

nextTick(() => {
	setTimeout(() => {
		inputRef.value && inputRef.value.focus();
	}, 100);
});
</script>

<style lang="scss" scoped>
.card-content {
	padding: 0 0;
	.input {
		border-radius: 5px;
		border: 1px solid $input-stroke;
		background-color: transparent;
		&:focus {
			border: 1px solid $yellow-disabled;
		}
	}
}
</style>

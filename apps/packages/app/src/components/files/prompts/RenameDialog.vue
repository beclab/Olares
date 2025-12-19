<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('prompts.rename')"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		:okLoading="loading ? t('loading') : false"
		:size="$q.platform.is.mobile ? 'small' : 'medium'"
		:platform="$q.platform.is.mobile ? 'mobile' : 'web'"
		@onSubmit="submit"
		@onHide="onCancel"
		@onCancel="onCancel"
	>
		<div class="card-content">
			<div class="text-body3 text-ink-3 q-mb-xs">
				{{ t('prompts.renameMessage') }}
			</div>
			<input
				class="input input--block text-ink-1"
				v-focus
				type="text"
				v-model.trim="name"
				@keyup.enter="submit"
				ref="renameRef"
			/>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref, onMounted, nextTick } from 'vue';
import { useDataStore } from '../../../stores/data';
import { useI18n } from 'vue-i18n';
import { useFilesStore, FilesIdType } from '../../../stores/files';
import {
	notifyHide,
	notifyWaitingShow,
	notifyWarning
} from '../../../utils/notifyRedefinedUtil';
import { dataAPIs } from '../../../api';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';

const props = defineProps({
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	}
});
const name = ref();
const store = useDataStore();
const filesStore = useFilesStore();
const { t } = useI18n();
const loading = ref(false);
const renameRef = ref();

const CustomRef = ref();

onMounted(() => {
	nextTick(() => {
		setTimeout(() => {
			renameRef.value && renameRef.value.focus();
		}, 100);
	});
	name.value = oldName();
});

const oldName = () => {
	const index = filesStore.selected[props.origin_id][0];
	return filesStore.getTargetFileItem(index, props.origin_id)?.name;
};

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
	const item = filesStore.getTargetFileItem(
		filesStore.selected[props.origin_id][0],
		props.origin_id
	);
	if (!item) {
		return;
	}

	const dataAPI = dataAPIs();

	loading.value = true;

	try {
		await dataAPI.renameItem(item, name.value);
		store.closeHovers();
		CustomRef.value.onDialogOK();
		loading.value = false;
		filesStore.resetSelected(props.origin_id);
		const currentPath = filesStore.currentPath[props.origin_id];
		await filesStore.refushCurrentRouter(
			currentPath.path + currentPath.param,
			filesStore.activeMenu(props.origin_id).driveType,
			props.origin_id
		);

		// notifyHide();
	} catch (error) {
		loading.value = false;
		// notifyHide();
	}
};

const onCancel = () => {
	store.closeHovers();
};
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

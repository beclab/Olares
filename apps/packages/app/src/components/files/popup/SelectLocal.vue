<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('prompts.syncRepo')"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		size="medium"
		@onSubmit="submit"
	>
		<div class="card-content">
			<div class="text-body3 text-ink-3 q-mb-xs">
				{{ t('download_location') }}
			</div>
			<div class="row items-center justify-center">
				<q-input outlined v-model="showPath" dense style="flex: 1" />
				<div class="viewBtn text-subtitle3" @click="selectSyncPath">
					{{ t('select') }}
				</div>
			</div>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { useQuasar } from 'quasar';
import { useFilesStore } from '../../../stores/files';

import { useI18n } from 'vue-i18n';
import { useUserStore } from '../../../stores/user';

const props = defineProps({
	item: {
		type: Object,
		required: false
	}
});

const filesStore = useFilesStore();
const savePath = ref<string>('');
const showPath = ref<string>('');
const userStore = useUserStore();
const CustomRef = ref();

const $q = useQuasar();

const { t } = useI18n();

filesStore.resetSelected();

const appendPath = $q.platform.is.win ? '\\' : '/';

const submit = async () => {
	if ($q.platform.is.electron) {
		window.electron.api.files.repoAddSync({
			worktree: savePath.value,
			repo_id: props.item?.repo_id,
			name: props.item?.repo_name,
			password: '',
			readonly: props.item?.permission == 'r'
		});
	}
	CustomRef.value.onDialogOK();
};

const selectSyncPath = async () => {
	if ($q.platform.is.electron) {
		const path = await window.electron.api.files.selectSyncSavePath();
		const morePath = path + appendPath + props.item?.repo_name;
		showPath.value = await formatPath(morePath);
		savePath.value = path;
	}
};

onMounted(async () => {
	if ($q.platform.is.electron) {
		const path = await window.electron.api.files.defaultSyncSavePath();
		savePath.value = path + appendPath + userStore.current_user!.local_name;
		const morePath = savePath.value + appendPath + props.item?.repo_name;
		showPath.value = await formatPath(morePath);
	}
});

const formatPath = async (path: string) => {
	if (!$q.platform.is.electron) {
		return path;
	}
	let index = 0;
	let isExist = true;
	let returnPath = '';
	while (isExist) {
		returnPath = index == 0 ? path : path + '-' + index;
		isExist = await window.electron.api.files.repoSyncPathIsExist(returnPath);
		if (isExist) {
			index = index + 1;
		}
	}
	return returnPath;
};
</script>

<style lang="scss" scoped>
.card-content {
	padding: 0 0px;

	.input {
		border-radius: 5px;

		&:focus {
			border: 1px solid $yellow-disabled;
		}
	}

	.viewBtn {
		background: $yellow-1;
		border-radius: 8px;
		width: 76px;
		height: 32px;
		line-height: 32px;

		text-align: center;
		margin-left: 20px;
		cursor: pointer;
		color: $ink-1;

		border: 1px solid $yellow;

		&:hover {
			background: $yellow-13;
		}
	}
}
</style>

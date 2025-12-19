<template>
	<q-dialog
		ref="dialogRef"
		seamless
		maximized
		position="bottom"
		v-model="show"
		@hide="hide"
	>
		<q-card class="files-shard-root" style="height: 100vh">
			<files-shard-page
				v-if="currentPage === CurrentPageType.FilesShardPage"
				:origin_id="origin_id"
				:origins="origins"
				@open="open"
				@close="close"
			/>
			<files-page
				v-if="currentPage === CurrentPageType.FilesPage"
				:origin_id="origin_id"
				:selectType="selectType"
				@back="back"
				@close="close"
				@on-submit="onSubmit"
			/>
			<files-repo-page
				v-if="currentPage === CurrentPageType.FilesRepoPage"
				:origin_id="origin_id"
				:selectType="selectType"
				@open-sync-page="openSyncPage"
				@back="back"
				@close="close"
			/>
		</q-card>
	</q-dialog>
</template>

<script lang="ts" setup>
enum CurrentPageType {
	FilesShardPage,
	FilesPage,
	FilesRepoPage
}

import { ref, PropType } from 'vue';
import { useDialogPluginComponent } from 'quasar';
import { useFilesStore, PickType, MenuItemType } from '../../stores/files';
import { common, filesIsV2 } from './../../api';
import { DriveType } from '../../utils/interface/files';

import FilesShardPage from '../../pages/Mobile/file/FilesShardPage.vue';
import FilesPage from '../../pages/Mobile/file/FilesPage.vue';
import FilesRepoPage from '../../pages/Mobile/file/FilesRepoPage.vue';

const props = defineProps({
	origins: {
		type: Array as PropType<DriveType[]>,
		required: false,
		default: () => {
			if (filesIsV2()) {
				return [DriveType.Drive, DriveType.External];
			}
			return [DriveType.Drive, DriveType.External, DriveType.Sync];
		}
	},
	origin_id: {
		type: Number,
		required: true
	},
	selectType: {
		type: String as PropType<PickType>,
		required: false,
		default: PickType.FOLDER
	}
});

const emits = defineEmits(['onSubmit']);

const { dialogRef, onDialogCancel, onDialogOK } = useDialogPluginComponent();

const filesStore = useFilesStore();

const show = ref(true);
const currentPage = ref(CurrentPageType.FilesShardPage);

const open = async (item: MenuItemType) => {
	if (item.driveType === DriveType.Sync) {
		currentPage.value = CurrentPageType.FilesRepoPage;
	} else {
		currentPage.value = CurrentPageType.FilesPage;

		const path = await filesStore.formatRepotoPath(item);
		filesStore.setBrowserUrl(path, item.driveType, true, props.origin_id);
	}
};

const back = (value: string) => {
	if (value === 'init') {
		currentPage.value = CurrentPageType.FilesShardPage;
		return false;
	}
	console.log(
		'filesStore currentPath',
		filesStore.currentPath[props.origin_id]
	);
	const driveType = common().formatUrltoDriveType(
		filesStore.currentPath[props.origin_id].path
	);
	if (driveType === DriveType.Sync) {
		currentPage.value = CurrentPageType.FilesRepoPage;
	} else {
		currentPage.value = CurrentPageType.FilesShardPage;
	}
};

const close = () => {
	currentPage.value = CurrentPageType.FilesShardPage;
	onDialogCancel();
};

const openSyncPage = () => {
	currentPage.value = CurrentPageType.FilesPage;
};

const hide = () => {
	onDialogCancel();
	filesStore.removeIdState(props.origin_id);
};

const onSubmit = (value: any) => {
	onDialogOK(value);
	emits('onSubmit', value);
};
</script>

<style lang="scss" scoped></style>

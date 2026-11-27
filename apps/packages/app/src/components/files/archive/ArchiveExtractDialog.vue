<template>
	<bt-custom-dialog
		ref="dialogRef"
		:title="t('files.archive.extract_archive')"
		:ok="t('files.archive.extract')"
		:cancel="t('cancel')"
		size="medium"
		:platform="deviceStore.isMobile ? 'mobile' : 'web'"
		:okDisabled="onDisabled"
		:okLoading="loading ? t('loading') : false"
		@onSubmit="submit"
	>
		<div class="archive-dialog">
			<div class="archive-banner row items-center no-wrap">
				<div class="row items-center no-wrap archive-banner-main">
					<q-icon name="sym_r_folder_open" size="24px" class="q-mr-sm" />
					<div class="text-body1 text-ink-1 text-ellipsis">
						{{
							t('files.archive.extracting_file', {
								name: archiveItem?.name || ''
							})
						}}
					</div>
				</div>
				<div v-if="archiveSummary" class="text-body1 text-ink-3">
					{{ archiveSummary }}
				</div>
			</div>
			<div>
				<div class="archive-name-title q-mb-xs">
					{{ t('files.archive.extract_to_folder') }}
				</div>
				<TransfetSelectTo
					ref="selectToDialog"
					class="q-mt-xs"
					:selectHeight="42"
					:defaultPathNumber="origin_id"
					@setSelectPath="setSelectPath"
					@setSelectCancel="setSelectCancel"
					:origins="[driveType as DriveType]"
				/>
			</div>
			<div>
				<div class="archive-name-title q-mb-xs">
					{{ t('files.archive.on_conflict') }}
				</div>
				<bt-select
					v-model="conflictStrategy"
					:options="localizedConflictStrategyOptions"
					:border="true"
				/>
			</div>
			<bt-check-box-component
				class="col"
				v-model="openFolderWhenDone"
				:label="t('files.archive.open_folder_when_done')"
			/>
		</div>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { computed, PropType, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import {
	ArchiveConflict,
	archiveConflictOptions
} from '../../../utils/interface/archive';
import {
	useFilesStore,
	FileItem,
	FilesIdType,
	FilePath
} from '../../../stores/files';
import { DriveType } from '../../../utils/interface/files';
import Taskmanager from '../../../services/olaresTask';
import BtSelect from 'src/components/base/BtSelect.vue';
import BtCheckBoxComponent from 'src/components/settings/base/BtCheckBoxComponent.vue';
import TransfetSelectTo from 'src/pages/Electron/Transfer/TransfetSelectTo.vue';
import { format as formatUtils } from '../../../utils/format';
import { useDeviceStore } from 'src/stores/settings/device';
const props = defineProps({
	origin_id: {
		type: Number,
		default: FilesIdType.PAGEID
	},
	driveType: {
		type: String,
		default: DriveType.Drive
	},
	defaultPath: {
		type: String,
		default: ''
	},
	item: {
		type: Object as PropType<FileItem | undefined>,
		default: undefined
	},
	password: {
		type: String,
		default: ''
	},
	origins: {
		type: Array as PropType<DriveType[]>,
		required: false,
		default: () => {
			return [
				DriveType.Drive,
				DriveType.External,
				DriveType.Cache,
				DriveType.Data,
				DriveType.Common,
				DriveType.GoogleDrive
			];
		}
	}
});

function filePathFromPath(path: string): FilePath {
	return new FilePath({
		path,
		isDir: true,
		driveType: props.driveType as DriveType,
		param: ''
	});
}

const { t } = useI18n();
const filesStore = useFilesStore();
const deviceStore = useDeviceStore();
const dialogRef = ref();
const openFolderWhenDone = ref(true);
const conflictStrategy = ref(ArchiveConflict.Rename);
const loading = ref(false);
const fileSavePathRef = ref<FilePath | undefined>(
	props.defaultPath
		? filePathFromPath(props.defaultPath)
		: filesStore.currentPath[props.origin_id]
);
const localizedConflictStrategyOptions = computed(() =>
	archiveConflictOptions.map((item) => ({
		...item,
		label: t(item.label)
	}))
);

const archiveItem = computed(
	() =>
		props.item ||
		(filesStore.getTargetFileItem(
			filesStore.selected[props.origin_id][0],
			props.origin_id
		) as FileItem | undefined)
);

const archiveSummary = computed(() => {
	const size = archiveItem.value?.size || 0;
	return size ? formatUtils.humanStorageSize(size) : '';
});

const submit = async () => {
	if (loading.value) {
		return;
	}
	const item = archiveItem.value;
	if (!item || !fileSavePathRef.value) {
		return;
	}

	loading.value = true;
	try {
		const currentPath = filesStore.currentPath[props.origin_id];
		const task = await Taskmanager.archiveTask.createExtractTask({
			items: [item],
			parentPath: fileSavePathRef.value.path || currentPath.path,
			driveType: props.driveType as DriveType,
			param: currentPath.param,
			targetFolderName: '',
			password: props.password,
			conflictStrategy: conflictStrategy.value,
			preserveSymbolicLinks: true,
			openFolderWhenDone: openFolderWhenDone.value,
			node:
				filesStore.currentFileList[props.origin_id]?.node ||
				filesStore.currentNode[props.origin_id]?.name
		});
		if (task) {
			await Taskmanager.addTasks([task]);
		}
		dialogRef.value.onDialogOK();
	} finally {
		loading.value = false;
	}
};

const setSelectPath = (fileSavePath: FilePath) => {
	fileSavePathRef.value = fileSavePath;
};

const setSelectCancel = () => {
	fileSavePathRef.value = props.defaultPath
		? filePathFromPath(props.defaultPath)
		: filesStore.currentPath[props.origin_id];
};

const onDisabled = computed(() => {
	return loading.value || !archiveItem.value || !fileSavePathRef.value;
});
</script>

<style scoped lang="scss">
.archive-dialog {
	width: 520px;
	display: flex;
	flex-direction: column;
	gap: 20px;
}

@media (max-width: 599px) {
	.archive-dialog {
		width: calc(100vw - 40px);
		max-width: 100%;
	}
}

.archive-banner {
	min-height: 48px;
	padding: 12px;
	border-radius: 8px;
	background: $light-blue-alpha;
}

.archive-banner-main {
	flex: 1;
	min-width: 0;
}

.archive-grid {
	display: grid;
	grid-template-columns: repeat(2, minmax(0, 1fr));
	gap: 12px;
}
</style>

<style lang="scss">
.archive-name-title {
	color: $ink-3;
}
</style>

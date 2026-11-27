<template>
	<div class="files-list-root">
		<terminus-title-bar
			v-if="selectIds === null"
			:title="filesStore.getCurrentRepo(route.path, origin_id)"
			:is-dark="isDark"
			:hook-back-action="true"
			@on-return-action="onReturnAction"
		>
			<template v-slot:right>
				<div class="row items-center" v-if="origin_id">
					<q-btn
						class="text-ink-1 btn-size-sm btn-no-text btn-no-border"
						icon="sym_r_close"
						text-color="ink-2"
						@click="closeDialog"
					>
					</q-btn>
				</div>
				<div class="row items-center" v-else>
					<q-btn
						v-if="showFloatingButton"
						class="text-ink-1 btn-size-sm btn-no-text btn-no-border"
						icon="sym_r_checklist"
						text-color="ink-2"
						@click="intoCheckedMode"
					>
					</q-btn>
					<q-btn
						v-if="isShareRoot"
						class="text-ink-1 btn-size-sm btn-no-text btn-no-border q-mr-xs"
						icon="sym_r_tune"
						text-color="ink-2"
						@click="showFilter"
					/>
					<q-btn
						class="text-ink-1 btn-size-sm btn-no-text btn-no-border"
						icon="sym_r_more_horiz"
						text-color="ink-2"
					>
						<PopupMenu
							:origin_id="origin_id"
							:show-paste="showPasteButton"
							@handleEvent="handleEvent"
						/>
					</q-btn>
				</div>
			</template>
		</terminus-title-bar>
		<terminus-select-header
			v-else
			:selectIds="selectIds"
			@handle-close="handleClose"
			@handle-select-all="handleSelectAll"
			@handle-remove="removeFiles"
		>
			<template #right-actions="{ disabled }">
				<q-btn
					class="text-ink-1 btn-size-sm btn-no-text btn-no-border"
					icon="sym_r_more_horiz"
					text-color="ink-2"
					:disable="disabled"
					@click="showSelectionActions"
				/>
			</template>
		</terminus-select-header>

		<div class="content" :class="filesStore.isShard ? 'shard' : ''">
			<div
				v-if="store.loading"
				style="width: 100%; height: 100%"
				class="row items-center justify-center"
			>
				<q-spinner-dots color="primary" size="3em" />
			</div>
			<errors v-else-if="error" :errorCode="error.status" />
			<listing-files
				:origin_id="origin_id"
				:selectType="selectType"
				:lock-touch="isShareRoot"
				ref="FilesListRef"
				@show-select-mode="showSelectMode"
				v-else
			/>
		</div>

		<listing-files-footer
			v-if="showSharedButtons"
			:origin_id="origin_id"
			:selectType="selectType"
			@onSubmit="onSubmit"
		/>
		<add-files v-if="showFloatingButton" @addFile="addFile" />
	</div>
</template>

<script lang="ts" setup>
import { ref, onMounted, watch, computed, PropType } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import Errors from '../../Files/Errors.vue';
import ListingFiles from './ListingFiles.vue';
import TerminusSelectHeader from '../../../components/common/TerminusSelectHeader.vue';
import TerminusTitleBar from '../../../components/common/TerminusTitleBar.vue';
import DirOperationDialog from './DirOperationDialog.vue';
import ListingFilesFooter from './ListingFilesFooter.vue';
import { useQuasar } from 'quasar';
import { useFilesStore, FilesIdType, PickType } from '../../../stores/files';
import { useDataStore } from '../../../stores/data';
import { common, isArchiveEnable } from '../../../api';
import { FilesSortType, OPERATE_ACTION } from '../../../utils/contact';
import { hideHeaderOpt } from '../../../utils/file';
import { notifySuccess } from '../../../utils/notifyRedefinedUtil';
import { useOperateinStore } from '../../../stores/operation';
import PopupMenu from './PopupMenu.vue';
import AddFiles from './AddFiles.vue';
import { useI18n } from 'vue-i18n';
import FilterMobileView from '../../../components/files/share/FilterMobileView.vue';
import { DriveType, FileViewMode } from 'src/utils/interface/files';
import ArchiveCompressDialog from '../../../components/files/archive/ArchiveCompressDialog.vue';
import SelectionActionsDialog from './SelectionActionsDialog.vue';
import {
	isArchiveDriveTypeSupported,
	isArchiveFile,
	isUnsupportedArchiveExtract
} from '../../../utils/interface/archive';

interface SelectionActionItem {
	key: string;
	icon: string;
	label: string;
	disabled?: boolean;
	handler: () => void | Promise<void>;
}

const props = defineProps({
	origin_id: {
		type: Number,
		required: false
	},
	selectType: {
		type: String as PropType<PickType>,
		required: false,
		default: ''
	}
});

const emits = defineEmits(['back', 'close', 'onSubmit']);

const route = useRoute();
const router = useRouter();
const store = useDataStore();
const filesStore = useFilesStore();
const operateinStore = useOperateinStore();
const error = ref<any>(null);
const origin_id = ref(props.origin_id || FilesIdType.PAGEID);

const $q = useQuasar();
const { t } = useI18n();
const isDark = ref(false);

const selectIds = ref<string[] | null>(null);
const FilesListRef = ref();

const isLarePass = process.env.APPLICATION === 'LAREPASS';

const activeDriveType = computed(
	() => filesStore.activeMenu(origin_id.value).driveType
);

const selectedItems = computed(() =>
	(selectIds.value || [])
		.map((index) =>
			filesStore.getTargetFileItem(Number(index), origin_id.value)
		)
		.filter(Boolean)
);

const canCompressSelection = computed(
	() =>
		(selectIds.value?.length || 0) > 1 &&
		isArchiveEnable() &&
		isArchiveDriveTypeSupported(activeDriveType.value) &&
		!selectedItems.value.some((item) => isArchiveFile(item))
);

const canExtractSelection = computed(() => {
	if ((selectIds.value?.length || 0) !== 1) {
		return false;
	}
	const item = selectedItems.value[0];
	return (
		isArchiveEnable() &&
		!!item &&
		isArchiveDriveTypeSupported(activeDriveType.value) &&
		isArchiveFile(item) &&
		!isUnsupportedArchiveExtract(item)
	);
});

const closeDialog = () => {
	emits('close');
};
const onSubmit = (value) => {
	emits('onSubmit', value);
};

const showSharedButtons = computed(() => {
	if (!hideHeaderOpt(route.path)) {
		return false;
	}
	return filesStore.isShard || props.selectType.length > 0;
});

const showFloatingButton = computed(() => {
	if (filesStore.isShard) {
		return false;
	}

	if (selectIds.value) {
		return false;
	}

	if (props.selectType) {
		return false;
	}

	if (isShareRoot.value) {
		return false;
	}

	if (!hideHeaderOpt(route.path)) {
		return false;
	}

	return true;
});

const showPasteButton = computed(() => {
	if (selectIds.value) {
		return false;
	}

	if (props.selectType) {
		return false;
	}

	if (isShareRoot.value) {
		return false;
	}

	if (!hideHeaderOpt(route.path)) {
		return false;
	}

	if (isExternalNodeListOrRoot.value) {
		return false;
	}

	return operateinStore.copyFiles.length > 0;
});

const isShareRoot = computed(() => {
	return common().isShareRootPage(route.path);
});

const isExternalNodeListOrRoot = computed(() => {
	const normalizedPath = route.path.replace(/\/+$/, '');
	if (
		normalizedPath !== '/Files/External' &&
		!normalizedPath.startsWith('/Files/External/')
	) {
		return false;
	}

	const pathParts = normalizedPath.split('/').filter(Boolean);
	return pathParts.length <= 3;
});

// Copy / cut are available when something is selected on an operable location
// (not share root, not the external connection list/root).
const canClipboardSelection = computed(() => {
	if (!selectIds.value?.length) {
		return false;
	}
	if (!hideHeaderOpt(route.path)) {
		return false;
	}
	return !isShareRoot.value && !isExternalNodeListOrRoot.value;
});

const showSelectMode = (value) => {
	selectIds.value = value;
};

const selectedFileIndexes = () => {
	const selectedIds = Array.isArray(selectIds.value) ? selectIds.value : [];
	return selectedIds.map((id) => Number(id.slice(0, id.indexOf('_'))));
};

const syncSelectedFiles = () => {
	filesStore.selected[origin_id.value] = selectedFileIndexes();
};

const handleClose = () => {
	if (FilesListRef.value) {
		FilesListRef.value.handleClose();
	}
};

const handleSelectAll = () => {
	if (FilesListRef.value) {
		FilesListRef.value.handleSelectAll();
	}
};

const intoCheckedMode = () => {
	if (filesStore.isShard) {
		return false;
	}
	if (FilesListRef.value) {
		FilesListRef.value.intoCheckedMode();
	}
};

const showFilter = () => {
	$q.dialog({
		component: FilterMobileView,
		componentProps: {}
	}).onOk((value) => {});
};

const onReturnAction = () => {
	if (props.origin_id) {
		if (filesStore.backStack[props.origin_id].length === 1) {
			emits('back');
		}
		filesStore.back(props.origin_id);
	} else {
		if (isLarePass) {
			router.go(-1);
			filesStore.back(props.origin_id);
		} else {
			if (filesStore.backStack[0].length > 1) {
				router.back();
				filesStore.back(props.origin_id);
			} else {
				if (
					filesStore.activeMenu(origin_id.value).driveType == DriveType.Sync
				) {
					if (
						window.history &&
						window.history.state &&
						!window.history.state.back
					) {
						router.replace('/Files/');
						return;
					}
					router.back();
				} else {
					router.replace('/Files/');
				}
			}
		}
	}
};

const addFile = () => {
	$q.dialog({
		component: DirOperationDialog,
		componentProps: {}
	}).onOk((value) => {
		console.log('addFileaddFile', value);
		if (value === 'uploadFiles') {
			// uploadFiles();
			filesStore.selectUploadFiles();
		} else if (value === 'createFolder') {
			createFolder();
		} else if (value === 'uploadImgOrVideo') {
			console.log(value);
			filesStore.selectUploadFiles(true);
		}
	});
};

const createFolder = () => {
	operateinStore.handleFileOperate(
		origin_id.value,
		null,
		route,
		OPERATE_ACTION.CREATE_FOLDER,
		filesStore.activeMenu(origin_id.value).driveType,
		async () => {}
	);
};

const pasteFiles = () => {
	operateinStore.handleFileOperate(
		origin_id.value,
		null,
		route,
		OPERATE_ACTION.PASTE,
		filesStore.activeMenu(origin_id.value).driveType,
		async () => {}
	);
};

const cancelPaste = () => {
	operateinStore.handleFileOperate(
		origin_id.value,
		null,
		route,
		OPERATE_ACTION.CANCEL,
		filesStore.activeMenu(origin_id.value).driveType,
		async () => {}
	);
};

const showSelectionActions = () => {
	if (!selectIds.value?.length) {
		return;
	}
	const actions = selectionActions.value
		.filter((item) => !item.disabled)
		.map(({ handler, ...action }) => action);
	$q.dialog({
		component: SelectionActionsDialog,
		componentProps: {
			actions
		}
	}).onOk((key: string) => {
		const action = selectionActions.value.find((item) => item.key === key);
		if (!action || action.disabled) {
			return;
		}
		action.handler();
	});
};

const compressFiles = () => {
	if (!canCompressSelection.value) {
		return;
	}

	syncSelectedFiles();
	$q.dialog({
		component: ArchiveCompressDialog,
		componentProps: {
			origin_id: origin_id.value,
			driveType: activeDriveType.value
		}
	}).onOk(() => {
		if (FilesListRef.value) {
			FilesListRef.value.handleRemove();
		}
	});
};

const extractFiles = () => {
	if (!canExtractSelection.value) {
		return;
	}

	syncSelectedFiles();
	operateinStore.handleFileOperate(
		origin_id.value,
		null,
		route,
		OPERATE_ACTION.EXTRACT,
		filesStore.activeMenu(origin_id.value).driveType,
		async () => {}
	);
};

const clipboardSelection = (action: OPERATE_ACTION, successText: string) => {
	if (!canClipboardSelection.value) {
		return;
	}

	syncSelectedFiles();
	operateinStore.handleFileOperate(
		origin_id.value,
		null,
		route,
		action,
		filesStore.activeMenu(origin_id.value).driveType,
		async () => {
			notifySuccess(t(successText));
			if (FilesListRef.value) {
				FilesListRef.value.handleClose();
			}
		}
	);
};

const copyFiles = () => {
	clipboardSelection(OPERATE_ACTION.COPY, 'files.copied_to_clipboard');
};

const cutFiles = () => {
	clipboardSelection(OPERATE_ACTION.CUT, 'files.cut_to_clipboard');
};

const selectionActions = computed<SelectionActionItem[]>(() => [
	{
		key: 'delete',
		icon: 'sym_r_delete',
		label: t('delete'),
		disabled: !selectIds.value?.length,
		handler: removeFiles
	},
	{
		key: 'copy',
		icon: 'sym_r_content_copy',
		label: t('copy'),
		disabled: !canClipboardSelection.value,
		handler: copyFiles
	},
	{
		key: 'cut',
		icon: 'sym_r_move_up',
		label: t('files.cut'),
		disabled: !canClipboardSelection.value,
		handler: cutFiles
	},
	{
		key: 'compress',
		icon: 'sym_r_archive',
		label: t('files_popup_menu.compress_to'),
		disabled: !canCompressSelection.value,
		handler: compressFiles
	},
	{
		key: 'extract',
		icon: 'sym_r_folder_open',
		label: t('files_popup_menu.extract_to'),
		disabled: !canExtractSelection.value,
		handler: extractFiles
	}
]);

const removeFiles = async () => {
	syncSelectedFiles();
	await operateinStore.handleFileOperate(
		origin_id.value,
		null,
		route,
		OPERATE_ACTION.DELETE,
		filesStore.activeMenu(origin_id.value).driveType,
		async () => {}
	);
};

// const enterHistory = () => {
// 	router.push('/transfer/history');
// };

const handleEvent = (value: string) => {
	switch (value) {
		case 'paste':
			pasteFiles();
			break;

		case 'cancel':
			cancelPaste();
			break;

		case 'mosaic':
			switchView(value);
			break;

		case 'list':
			switchView(value);
			break;

		case 'name':
			fileSort(FilesSortType.NAME);
			break;

		case 'type':
			fileSort(FilesSortType.TYPE);
			break;

		case 'file_modified':
			fileSort(FilesSortType.Modified);
			break;

		case 'size':
			fileSort(FilesSortType.SIZE);
			break;

		default:
			break;
	}
};

const switchView = async (type: FileViewMode) => {
	if (type === store.user.viewMode) {
		return false;
	}
	const data = {
		id: store.user.id,
		viewMode: type
	};
	// users.update(data, ['viewMode']).catch();
	store.updateUser(data);
};

const fileSort = (sort: FilesSortType) => {
	if (filesStore.activeSort[origin_id.value].by == sort) {
		filesStore.updateActiveSort(
			sort,
			!filesStore.activeSort[origin_id.value].asc,
			origin_id.value
		);
	} else {
		filesStore.updateActiveSort(sort, true, origin_id.value);
	}
};

watch(
	() => operateinStore.executing,
	(newVal, oldVal) => {
		if (newVal === oldVal) {
			return false;
		}

		if (!newVal) {
			if (FilesListRef.value && selectIds.value) {
				FilesListRef.value.handleRemove();
			}
		}
	}
);

onMounted(() => {
	filesStore.selected[origin_id.value] = [];
	filesStore.refreshPathItems(origin_id.value);
});
</script>

<style lang="scss" scoped>
.files-list-root {
	width: 100%;
	height: 100%;

	.content {
		width: 100%;
		height: calc(100% - 56px);
		&.shard {
			padding-bottom: 64px;
		}
	}
}
</style>

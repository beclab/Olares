<template>
	<q-dialog
		position="bottom"
		ref="dialogRef"
		@show="onShow"
		transition-show="jump-up"
		transition-hide="jump-down"
		transition-duration="300"
	>
		<terminus-dialog-display-content :dialog-ref="dialogRef">
			<template v-slot:title>
				<div class="operate-dialog-title-module row justify-start items-center">
					<terminus-file-icon
						class="q-mr-md"
						:name="selectItem?.name"
						:type="selectItem?.type"
						:is-dir="selectItem?.isDir"
						:modified="selectItem?.modified"
						:path="selectItem?.path"
						:driveType="selectItem?.driveType"
						style="width: 30px; height: 30px"
					/>

					<div class="column justify-center" style="width: calc(100% - 50px)">
						<div class="text-subtitle2 title">{{ name }}</div>
						<div class="text-body3 detail">
							{{ formatFileModified(modified) }}
						</div>
					</div>
				</div>
			</template>
			<template v-slot:content>
				<!-- {{ selectItem }} -->
				<q-list style="margin: 8px">
					<file-operation-item
						v-if="
							!disableHomeMenu &&
							canWrite &&
							!selectItem?.isShareItem &&
							!isExternalRoot &&
							!isSyncAndRepo
						"
						:origin_id="origin_id"
						icon="sym_r_move_up"
						:label="t('files.cut')"
						:action="OPERATE_ACTION.CUT"
						@on-item-click="onItemClick"
					/>
					<file-operation-item
						v-if="!selectItem?.isShareItem && !isExternalRoot && !isSyncAndRepo"
						:origin_id="origin_id"
						icon="sym_r_content_copy"
						:label="t('copy')"
						:action="OPERATE_ACTION.COPY"
						@on-item-click="onItemClick"
					/>
					<file-operation-item
						v-if="
							selectItem &&
							!selectItem.isDir &&
							!selectItem?.isShareItem &&
							!isExternalRoot
						"
						:origin_id="origin_id"
						icon="sym_r_browser_updated"
						:label="t('buttons.download')"
						:action="OPERATE_ACTION.DOWNLOAD"
						@on-item-click="onItemClick"
					/>
					<file-operation-item
						v-if="showArchivePreview"
						:origin_id="origin_id"
						icon="sym_r_visibility"
						:label="t('files_popup_menu.preview_contents')"
						:action="OPERATE_ACTION.PREVIEW_CONTENTS"
						@hide-menu="hideDialog"
					/>
					<file-operation-item
						v-if="showArchiveExtract"
						:origin_id="origin_id"
						icon="sym_r_folder_open"
						:label="t('files_popup_menu.extract_to')"
						:action="OPERATE_ACTION.EXTRACT"
						:mobile-submenu="true"
						@hide-menu="hideDialog"
					/>
					<file-operation-item
						v-if="showArchiveCompress"
						:origin_id="origin_id"
						icon="sym_r_archive"
						:label="t('files_popup_menu.compress_to')"
						:action="OPERATE_ACTION.COMPRESS"
						:mobile-submenu="true"
						@hide-menu="hideDialog"
					/>
					<file-operation-item
						v-if="showShareMenu"
						:origin_id="origin_id"
						icon="sym_r_share_windows"
						:label="t('files_popup_menu.share_to')"
						:action="OPERATE_ACTION.SHARE"
						:mobile-submenu="true"
						@hide-menu="hideDialog"
					/>
					<file-operation-item
						v-if="showBackup"
						:origin_id="origin_id"
						icon="sym_r_backup"
						:label="t('files_popup_menu.backup')"
						:action="OPERATE_ACTION.BACKUP"
						@hide-menu="hideDialog"
					/>
					<file-operation-item
						v-if="
							!disableHomeMenu &&
							canWrite &&
							!selectItem?.isShareItem &&
							!isExternalRoot
						"
						:origin_id="origin_id"
						icon="sym_r_delete"
						:label="t('delete')"
						:action="
							isSyncAndRepo ? OPERATE_ACTION.REPO_DELETE : OPERATE_ACTION.DELETE
						"
					/>
					<file-operation-item
						v-if="
							!disableHomeMenu &&
							canWrite &&
							!selectItem?.isShareItem &&
							!isExternalRoot
						"
						:origin_id="origin_id"
						icon="sym_r_edit_square"
						:label="t('buttons.rename')"
						:action="
							isSyncAndRepo ? OPERATE_ACTION.REPO_RENAME : OPERATE_ACTION.RENAME
						"
					/>
					<file-operation-item
						v-if="
							!!selectItem?.isShareItem &&
							selectItem.share_type == ShareType.PUBLIC
						"
						:origin_id="origin_id"
						icon="sym_r_visibility_lock"
						:label="t('files_popup_menu.Reset Password')"
						:action="OPERATE_ACTION.RESET_PASSWORD"
					/>

					<file-operation-item
						v-if="
							!!selectItem?.isShareItem &&
							((selectItem.share_type == ShareType.INTERNAL &&
								selectItem.permission == SharePermission.ADMIN) ||
								selectItem.share_type == ShareType.SMB)
						"
						:origin_id="origin_id"
						icon="sym_r_admin_panel_settings"
						:label="t('files_popup_menu.Edit permissions')"
						:action="OPERATE_ACTION.EDIT_PERMISSIONS"
					/>
					<file-operation-item
						v-if="!!selectItem?.isShareItem"
						:origin_id="origin_id"
						icon="sym_r_do_not_disturb_on"
						:label="t('files_popup_menu.Revoke sharing')"
						:action="OPERATE_ACTION.REVOKE_SHARING"
					/>

					<file-operation-item
						v-if="isExternalRoot"
						:origin_id="origin_id"
						icon="sym_r_delete"
						:label="t('files_popup_menu.unmount')"
						:action="OPERATE_ACTION.UNMOUNT"
					/>

					<file-operation-item
						:origin_id="origin_id"
						icon="sym_r_contract"
						:label="t('files.attributes')"
						:action="OPERATE_ACTION.ATTRIBUTES"
					/>
					<!--
						{
										name: 'files_popup_menu.Share to Internal',
										icon: 'sym_r_folder_supervised',
										action: OPERATE_ACTION.SHARE_IN_INTERNAL,
										condition: (event: EventType) =>
											(event.type === DriveType.Drive || event.type === DriveType.Sync) &&
											event.isSelected &&
											event.isDir
									}, -->
				</q-list>
			</template>
		</terminus-dialog-display-content>
	</q-dialog>
</template>

<script setup lang="ts">
import { useMenuStore } from '../../../stores/files-menu';

import TerminusDialogDisplayContent from '../../../components/common/TerminusDialogDisplayContent.vue';
import FileOperationItem from '../../../components/files/files/FileOperationItem.vue';
import { stopScrollMove } from '../../../utils/utils';
import TerminusFileIcon from '../../../components/common/TerminusFileIcon.vue';
import { useDialogPluginComponent } from 'quasar';
import { ref, PropType, computed } from 'vue';
import { useRoute } from 'vue-router';
import { formatFileModified } from '../../../utils/file';
import { notifySuccess } from '../../../utils/notifyRedefinedUtil';
import { OPERATE_ACTION } from '../../../utils/contact';
import { getParams } from '../../../utils/utils';
import { useI18n } from 'vue-i18n';
import {
	useFilesStore,
	FilesIdType,
	ExternalType
} from './../../../stores/files';
import { common, isShareEnable } from './../../../api';
import { useOperateinStore } from './../../../stores/operation';
import { backupOriginsRef } from './../../../constant';
import { DriveType } from '../../../utils/interface/files';
import { ShareType, SharePermission } from 'src/utils/interface/share';
import {
	isArchiveDriveTypeSupported,
	isArchiveFile,
	isArchivePreviewSupported,
	isUnsupportedArchiveExtract
} from '../../../utils/interface/archive';
import { isArchiveEnable } from '../../../api';

const { dialogRef, onDialogCancel } = useDialogPluginComponent();

const props = defineProps({
	driveType: {
		type: String as PropType<DriveType>,
		default: '',
		required: true
	},
	name: {
		type: String,
		default: '',
		required: true
	},
	modified: {
		type: String,
		default: '',
		required: true
	},
	type: {
		type: String,
		default: '',
		required: true
	},
	index: {
		type: Number,
		default: -1,
		required: true
	},
	isDir: {
		type: Boolean,
		required: false,
		default: false
	},
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	},
	permission: {
		type: String,
		required: false,
		default: 'rw'
	}
});

const { t } = useI18n();

const menuStore = useMenuStore();
const filesStore = useFilesStore();
const copied = ref(false);
const route = useRoute();
const isSyncAndRepo = ref(false);
const operateinStore = useOperateinStore();
const isExternalRoot = ref(false);

const canWrite = computed(() => ['rw', 'admin'].includes(props.permission));

const onShow = async () => {
	const driveType = common().formatUrltoDriveType(route.fullPath);

	const id = route.query.id;
	isSyncAndRepo.value = driveType == DriveType.Sync && !id;

	const item = filesStore.getTargetFileItem(props.index, props.origin_id);

	if (item) {
		menuStore.shareRepoInfo = JSON.parse(JSON.stringify(item));

		if (common().displayConnectServer(route.fullPath) && item.externalType) {
			isExternalRoot.value = [
				ExternalType.HDD,
				ExternalType.SMB,
				ExternalType.NFS,
				ExternalType.USB
			].includes(item.externalType);
		} else {
			isExternalRoot.value = false;
		}

		if (getParams(menuStore.shareRepoInfo.path, 'id')) {
			menuStore.shareRepoInfo.id = getParams(
				menuStore.shareRepoInfo.path,
				'id'
			);
			menuStore.shareRepoInfo.repo_name = props.name;
			const start =
				menuStore.shareRepoInfo.path.indexOf(props.name) + props.name.length;
			const end = menuStore.shareRepoInfo.path.indexOf('?', start);
			const resultPath = menuStore.shareRepoInfo.path.substring(start, end);
			menuStore.shareRepoInfo.path = resultPath;
		}

		stopScrollMove();
		filesStore.resetSelected(props.origin_id);
		filesStore.addSelected(props.index, props.origin_id);
	}
};

const onItemClick = async (action: OPERATE_ACTION) => {
	if (action === OPERATE_ACTION.COPY) {
		copied.value = true;
		notifySuccess(t('files.copied_to_clipboard'));
	} else if (action === OPERATE_ACTION.CUT) {
		copied.value = true;
		notifySuccess(t('files.cut_to_clipboard'));
	} else if (action === OPERATE_ACTION.PASTE) {
		copied.value = false;
	}
};

const hideDialog = () => {
	onDialogCancel();
};

const disableHomeMenu = computed(() => {
	const item = filesStore.getTargetFileItem(props.index, props.origin_id);
	if (!item) {
		return false;
	}
	return operateinStore.isDisableMenuItem(item.name || '', route.path);
});

const isArchiveSelected = computed(() => {
	const item = selectItem.value;
	return item ? isArchiveFile(item) : false;
});

const showArchivePreview = computed(() => {
	return (
		isArchiveEnable() &&
		!!selectItem.value &&
		!selectItem.value.isDir &&
		isArchivePreviewSupported(selectItem.value) &&
		isArchiveDriveTypeSupported(props.driveType)
	);
});

const showArchiveExtract = computed(() => {
	return (
		isArchiveEnable() &&
		!!selectItem.value &&
		isArchiveSelected.value &&
		canWrite.value &&
		isArchiveDriveTypeSupported(props.driveType) &&
		!isExternalRoot.value &&
		!selectItem.value?.isShareItem
	);
});

const showArchiveCompress = computed(() => {
	return (
		isArchiveEnable() &&
		!!selectItem.value &&
		canWrite.value &&
		!isArchiveSelected.value &&
		isArchiveDriveTypeSupported(props.driveType) &&
		!isExternalRoot.value &&
		!selectItem.value?.isShareItem &&
		!isUnsupportedArchiveExtract(selectItem.value)
	);
});

// Mirrors the PC context-menu `canBackup` rule: a single directory living in a
// backup-capable drive. Hidden on the 1.11 build, matching the desktop guard.
const showBackup = computed(() => {
	if (process.env.VERSIONTAG === '1.11') {
		return false;
	}
	const item = selectItem.value;
	return (
		!!item &&
		!!item.isDir &&
		!item.isShareItem &&
		!isExternalRoot.value &&
		backupOriginsRef.value.includes(props.driveType)
	);
});

const showShareMenu = computed(() => {
	return (
		isShareEnable() &&
		[
			DriveType.Drive,
			DriveType.Data,
			DriveType.Common,
			DriveType.Sync,
			DriveType.External,
			DriveType.Cache
		].includes(props.driveType) &&
		!!selectItem.value?.isDir &&
		!selectItem.value?.isShareItem
	);
});

const selectItem = computed(() => {
	const item = filesStore.currentFileList[props.origin_id]?.items.find(
		(item) => item.index === props.index
	);
	return item;
});
</script>

<style lang="scss" scoped>
.operate-dialog-title-module {
	width: 100%;
	height: 100%;

	.title {
		text-align: left;
		color: $ink-1;
		width: 100%;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.detail {
		text-align: left;
		color: $ink-3;
		max-width: 100%;
		width: 100%;

		text-overflow: ellipsis;
		white-space: nowrap;
		overflow: hidden;
	}
}
</style>

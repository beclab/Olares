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
							permission === 'rw' &&
							!selectItem?.isShareItem
						"
						icon="sym_r_move_up"
						:label="t('files.cut')"
						:action="OPERATE_ACTION.CUT"
						@on-item-click="onItemClick"
					/>
					<file-operation-item
						v-if="permission === 'rw' && !selectItem?.isShareItem"
						icon="sym_r_content_copy"
						:label="t('copy')"
						:action="OPERATE_ACTION.COPY"
						@on-item-click="onItemClick"
					/>
					<file-operation-item
						v-if="selectItem && !selectItem.isDir && !selectItem?.isShareItem"
						icon="sym_r_browser_updated"
						:label="t('buttons.download')"
						:action="OPERATE_ACTION.DOWNLOAD"
						@on-item-click="onItemClick"
					/>
					<file-operation-item
						v-if="
							(driveType === DriveType.Sync ||
								driveType === DriveType.Drive ||
								driveType === DriveType.External ||
								driveType === DriveType.Cache ||
								driveType === DriveType.Data) &&
							selectItem?.isDir &&
							!selectItem?.isShareItem
						"
						icon="sym_r_folder_supervised"
						:label="t('files_popup_menu.Share to Internal')"
						:action="OPERATE_ACTION.SHARE_IN_INTERNAL"
						@on-item-click="onItemClick"
					/>
					<file-operation-item
						v-if="
							(driveType === DriveType.Drive ||
								driveType === DriveType.Data ||
								driveType === DriveType.Cache ||
								driveType === DriveType.External) &&
							selectItem?.isDir &&
							!selectItem?.isShareItem
						"
						icon="sym_r_smb_share"
						:label="t('files_popup_menu.Share to SMB')"
						:action="OPERATE_ACTION.SHARE_IN_SMB"
						@on-item-click="onItemClick"
					/>
					<file-operation-item
						v-if="
							(driveType === DriveType.Drive || driveType === DriveType.Data) &&
							selectItem?.isDir &&
							!selectItem?.isShareItem
						"
						icon="sym_r_link"
						:label="t('files_popup_menu.Share to Public')"
						:action="OPERATE_ACTION.SHARE_IN_PUBLIC"
						@on-item-click="onItemClick"
					/>
					<file-operation-item
						v-if="driveType === DriveType.Sync && !selectItem?.isShareItem"
						icon="sym_r_share_windows"
						:label="t('buttons.share')"
						:action="OPERATE_ACTION.SHARE"
					/>
					<file-operation-item
						v-if="
							!disableHomeMenu &&
							permission === 'rw' &&
							!selectItem?.isShareItem
						"
						icon="sym_r_delete"
						:label="t('delete')"
						:action="
							isSyncAndRepo ? OPERATE_ACTION.REPO_DELETE : OPERATE_ACTION.DELETE
						"
					/>
					<file-operation-item
						v-if="
							!disableHomeMenu &&
							permission === 'rw' &&
							!selectItem?.isShareItem
						"
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
						icon="sym_r_visibility_lock"
						:label="t('files_popup_menu.Reset Password')"
						:action="OPERATE_ACTION.RESET_PASSWORD"
					/>

					<!-- {
										name: 'files_popup_menu.Edit permissions',
										icon: 'sym_r_admin_panel_settings',
										action: OPERATE_ACTION.EDIT_PERMISSIONS,
										condition: (event: EventType) =>
											filesIsV2() && event.isEditPermissionsEnable && event.isSelected
									}, -->
					<file-operation-item
						v-if="
							!!selectItem?.isShareItem &&
							((selectItem.share_type == ShareType.INTERNAL &&
								selectItem.permission == SharePermission.ADMIN) ||
								selectItem.share_type == ShareType.SMB)
						"
						icon="sym_r_admin_panel_settings"
						:label="t('files_popup_menu.Edit permissions')"
						:action="OPERATE_ACTION.EDIT_PERMISSIONS"
					/>
					<file-operation-item
						v-if="!!selectItem?.isShareItem"
						icon="sym_r_do_not_disturb_on"
						:label="t('files_popup_menu.Revoke sharing')"
						:action="OPERATE_ACTION.REVOKE_SHARING"
					/>

					<file-operation-item
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
											filesIsV2() &&
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
import { useFilesStore, FilesIdType } from './../../../stores/files';
import { common } from './../../../api';
import { useOperateinStore } from './../../../stores/operation';
import { DriveType } from '../../../utils/interface/files';
import { ShareType, SharePermission } from 'src/utils/interface/share';

const { dialogRef } = useDialogPluginComponent();

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

const onShow = async () => {
	const driveType = common().formatUrltoDriveType(route.fullPath);
	const id = route.query.id;
	isSyncAndRepo.value = driveType == DriveType.Sync && !id;

	const index = filesStore.currentFileList[props.origin_id]?.items.findIndex(
		(item) => item.index === props.index
	);

	if (!index) {
		return;
	}

	if (index !== -1) {
		menuStore.shareRepoInfo = JSON.parse(
			JSON.stringify(filesStore.currentFileList[props.origin_id]?.items)
		)[index];

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
		filesStore.resetSelected();
		filesStore.addSelected(index);
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

const disableHomeMenu = computed(() => {
	const index = filesStore.currentFileList[props.origin_id]?.items.findIndex(
		(item) => item.index === props.index
	);
	if (!index || index < 0) {
		return false;
	}

	return operateinStore.isDisableMenuItem(
		filesStore.currentFileList[props.origin_id]?.items[index].name || '',
		route.path
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

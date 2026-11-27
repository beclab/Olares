<template>
	<q-item
		v-if="!isSubmenuAction"
		clickable
		v-close-popup
		class="file-operation-item q-pb-md"
		@click="handle($event, action!)"
	>
		<div class="file-operation-div row justify-start items-center text-ink-1">
			<q-icon :name="icon" size="20px" />
			<div class="file-operation-text text-body3">{{ label }}</div>
		</div>
	</q-item>
	<template v-else>
		<q-item
			clickable
			class="file-operation-item q-pb-md"
			@click.stop.prevent="handleSubmenuClick"
			@mouseenter="handleSubmenuMouseEnter"
			@mouseleave="handleSubmenuMouseLeave"
		>
			<div class="file-operation-div row justify-start items-center text-ink-1">
				<q-icon :name="icon" size="20px" />
				<div class="file-operation-text text-body3">{{ label }}</div>
				<q-space />
				<q-icon name="sym_r_keyboard_arrow_right" size="20px" />
			</div>
			<q-menu
				v-if="!mobileSubmenu"
				v-model="submenuShowing"
				no-parent-event
				anchor="top end"
				self="top start"
				:offset="[8, 0]"
				class="popup-menu bg-background-2"
			>
				<q-list
					dense
					padding
					@mouseenter="openSubmenu"
					@mouseleave="scheduleCloseSubmenu"
				>
					<q-item
						v-for="item in submenuOptions"
						:key="item.value || item.label"
						clickable
						v-close-popup
						class="file-operation-item"
						@click="handleSubmenuItem($event, item.value)"
					>
						<div
							class="file-operation-div row justify-start items-center text-ink-1"
						>
							<q-icon :name="item.icon" size="16px" />
							<div class="file-operation-text text-body3">
								{{ getSubmenuLabel(item.label) }}
							</div>
						</div>
					</q-item>
				</q-list>
			</q-menu>
		</q-item>
		<q-dialog
			v-if="mobileSubmenu"
			v-model="mobileSubmenuShowing"
			position="bottom"
			full-width
			transition-show="jump-up"
			transition-hide="jump-down"
			transition-duration="300"
		>
			<div class="mobile-submenu bg-background-2">
				<div class="mobile-submenu-header row items-center text-ink-1">
					<terminus-file-icon
						v-if="selectedItem"
						class="q-mr-md"
						:name="selectedItem.name"
						:type="selectedItem.type"
						:is-dir="selectedItem.isDir"
						:modified="selectedItem.modified"
						:path="selectedItem.path"
						:driveType="selectedItem.driveType"
						style="width: 32px; height: 32px"
					/>
					<div class="mobile-submenu-title-wrapper column justify-center">
						<div class="text-subtitle2 text-ellipsis">
							{{ mobileSubmenuTitle }}
						</div>
						<div class="text-body3 text-ink-3 text-ellipsis">
							{{ mobileSubmenuSubtitle }}
						</div>
					</div>
					<q-btn
						flat
						dense
						round
						icon="sym_r_close"
						size="sm"
						color="ink-2"
						@click="mobileSubmenuShowing = false"
					/>
				</div>
				<q-list>
					<q-item
						v-for="item in submenuOptions"
						:key="item.value || item.label"
						clickable
						class="file-operation-item mobile-submenu-item"
						@click="handleMobileSubmenuItem($event, item.value)"
					>
						<div
							class="file-operation-div row justify-start items-center text-ink-1"
						>
							<q-icon :name="item.icon" size="20px" />
							<div class="file-operation-text text-body3">
								{{ getSubmenuLabel(item.label) }}
							</div>
						</div>
					</q-item>
				</q-list>
			</div>
		</q-dialog>
	</template>
</template>

<script lang="ts" setup>
import { useOperateinStore } from './../../../stores/operation';
import { useQuasar } from 'quasar';
import { computed, PropType, ref } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { OPERATE_ACTION, shareMenuOptions } from '../../../utils/contact';
import { DriveType } from '../../../utils/interface/files';
import {
	ArchiveFormat,
	archiveExtractOptions,
	archiveFormatOptions,
	getArchiveDefaultExtractName,
	isArchivePreviewSupported
} from '../../../utils/interface/archive';
import { getParams } from '../../../utils/utils';
import { useFilesStore, FilesIdType } from '../../../stores/files';
import { useMenuStore } from '../../../stores/files-menu';
import DeleteRepo from './../popup/DeleteRepo.vue';
import ReName from './../popup/ReName.vue';
import ArchiveCompressDialog from '../archive/ArchiveCompressDialog.vue';
import ArchiveExtractDialog from '../archive/ArchiveExtractDialog.vue';
import ArchivePreviewDialog from '../archive/ArchivePreviewDialog.vue';
import TerminusFileIcon from '../../common/TerminusFileIcon.vue';
import { formatFileModified } from '../../../utils/file';

const props = defineProps({
	icon: String,
	label: String,
	action: Object as PropType<OPERATE_ACTION>,
	origin_id: {
		type: Number,
		required: true,
		default: FilesIdType.PAGEID
	},
	mobileSubmenu: {
		type: Boolean,
		default: false
	}
});

const $q = useQuasar();
const route = useRoute();
const operateinStore = useOperateinStore();
const filesStore = useFilesStore();
const menuStore = useMenuStore();
const { t } = useI18n();
const submenuShowing = ref(false);
const mobileSubmenuShowing = ref(false);
let submenuCloseTimer: ReturnType<typeof setTimeout> | undefined;

const isSubmenuAction =
	props.action === OPERATE_ACTION.COMPRESS ||
	props.action === OPERATE_ACTION.EXTRACT ||
	props.action === OPERATE_ACTION.SHARE;

const submenuOptions = computed(() =>
	props.action === OPERATE_ACTION.SHARE
		? shareMenuOptions.filter((item) => canShowShareOption(item.value))
		: props.action === OPERATE_ACTION.EXTRACT
		? archiveExtractOptions.map((item) => ({
				...item,
				label: t(item.label, { name: getExtractFolderName() })
		  }))
		: archiveFormatOptions.map((item) => ({
				...item,
				label: t(item.label)
		  }))
);

const selectedItem = computed(() => getSelectedItem());

const getSelectedItem = () =>
	filesStore.getTargetFileItem(
		filesStore.selected[props.origin_id][0],
		props.origin_id
	);

const mobileSubmenuTitle = computed(() =>
	(props.label || '').replace(/[.…]+$/, '')
);

const mobileSubmenuSubtitle = computed(() => {
	const item = selectedItem.value;
	if (!item) {
		return '';
	}
	const modified = item.modified ? formatFileModified(item.modified) : '';
	return modified ? `${item.name} • ${modified}` : item.name;
});

const getExtractFolderName = () =>
	getArchiveDefaultExtractName(getSelectedItem()) || 'project_backup';

const getSubmenuLabel = (label: string) =>
	label.startsWith('files_popup_menu.') ? t(label) : label;

const canShowShareOption = (action: OPERATE_ACTION) => {
	const selectedItem = getSelectedItem();
	if (!selectedItem?.isDir || selectedItem.isShareItem) {
		return false;
	}

	const driveType = selectedItem.driveType;
	switch (action) {
		case OPERATE_ACTION.SHARE_IN_INTERNAL:
			return [
				DriveType.Drive,
				DriveType.Data,
				DriveType.Sync,
				DriveType.External,
				DriveType.Cache
			].includes(driveType);
		case OPERATE_ACTION.SHARE_IN_SMB:
			return [
				DriveType.Drive,
				DriveType.Data,
				DriveType.External,
				DriveType.Cache
			].includes(driveType);
		case OPERATE_ACTION.SHARE_IN_PUBLIC:
			return [DriveType.Drive, DriveType.Data, DriveType.Common].includes(
				driveType
			);
		default:
			return false;
	}
};

const openSubmenu = () => {
	if (submenuCloseTimer) {
		clearTimeout(submenuCloseTimer);
		submenuCloseTimer = undefined;
	}
	submenuShowing.value = true;
};

const scheduleCloseSubmenu = () => {
	if (submenuCloseTimer) {
		clearTimeout(submenuCloseTimer);
	}
	submenuCloseTimer = setTimeout(() => {
		submenuShowing.value = false;
		submenuCloseTimer = undefined;
	}, 150);
};

const handleSubmenuClick = () => {
	if (props.mobileSubmenu) {
		mobileSubmenuShowing.value = true;
		return;
	}
	openSubmenu();
};

const handleSubmenuMouseEnter = () => {
	if (!props.mobileSubmenu) {
		openSubmenu();
	}
};

const handleSubmenuMouseLeave = () => {
	if (!props.mobileSubmenu) {
		scheduleCloseSubmenu();
	}
};

const emit = defineEmits(['onItemClick', 'hideMenu']);

const handle = (
	e: any,
	action: OPERATE_ACTION,
	options?: {
		archiveFormat?: ArchiveFormat;
		archiveExtractTargetName?: string;
		extractToCurrentFolder?: boolean;
	}
) => {
	if (submenuCloseTimer) {
		clearTimeout(submenuCloseTimer);
		submenuCloseTimer = undefined;
	}
	submenuShowing.value = false;
	if (action === OPERATE_ACTION.REPO_DELETE) {
		return deleteRepo();
	}

	if (action === OPERATE_ACTION.REPO_RENAME) {
		return renameRepo();
	}

	if (action === OPERATE_ACTION.PREVIEW_CONTENTS) {
		const item = getSelectedItem();
		if (!isArchivePreviewSupported(item)) {
			return;
		}
		emit('hideMenu');
		return $q.dialog({
			component: ArchivePreviewDialog,
			componentProps: {
				item,
				origin_id: props.origin_id,
				driveType: filesStore.activeMenu(props.origin_id).driveType
			}
		});
	}

	emit('hideMenu');
	operateinStore.handleFileOperate(
		props.origin_id,
		e,
		route,
		action,
		filesStore.activeMenu(props.origin_id).driveType,
		async (action: OPERATE_ACTION, data: any) => {
			emit('onItemClick', action, data);
		},
		options
	);
};

const handleSubmenuItem = (e: any, value?: string | OPERATE_ACTION) => {
	if (props.action === OPERATE_ACTION.SHARE) {
		return handle(e, value as OPERATE_ACTION);
	}

	if (props.action === OPERATE_ACTION.COMPRESS) {
		if (value) {
			return handle(e, OPERATE_ACTION.COMPRESS, {
				archiveFormat: value as ArchiveFormat
			});
		}
		emit('hideMenu');
		return $q.dialog({
			component: ArchiveCompressDialog,
			componentProps: {
				origin_id: props.origin_id,
				driveType: filesStore.activeMenu(props.origin_id).driveType
			}
		});
	}

	if (value === 'current') {
		return handle(e, OPERATE_ACTION.EXTRACT, {
			extractToCurrentFolder: true
		});
	}

	if (value === 'new_folder') {
		return handle(e, OPERATE_ACTION.EXTRACT, {
			archiveExtractTargetName: getExtractFolderName()
		});
	}

	emit('hideMenu');
	return $q.dialog({
		component: ArchiveExtractDialog,
		componentProps: {
			origin_id: props.origin_id,
			driveType: filesStore.activeMenu(props.origin_id).driveType
		}
	});
};

const handleMobileSubmenuItem = (e: any, value?: string | OPERATE_ACTION) => {
	mobileSubmenuShowing.value = false;
	return handleSubmenuItem(e, value);
};

const deleteRepo = async () => {
	const foucsItem = filesStore.getTargetFileItem(
		filesStore.selected[props.origin_id][0],
		props.origin_id
	);
	const repo_id = getParams(foucsItem?.path || '', 'id');
	const res = await menuStore.fetchShareInfo(repo_id);

	const shared_user_emails_length = res.shared_user_emails.length || 0;

	$q.dialog({
		component: DeleteRepo,
		componentProps: {
			item: { repo_id, ...foucsItem },
			shared_length: shared_user_emails_length
		}
	});
};

const renameRepo = async () => {
	const foucsItem = filesStore.getTargetFileItem(
		filesStore.selected[props.origin_id][0],
		props.origin_id
	);
	const repo_id = getParams(foucsItem?.path || '', 'id');

	$q.dialog({
		component: ReName,
		componentProps: {
			item: { repo_id, ...foucsItem }
		}
	});
};
</script>

<style scoped lang="scss">
.file-operation-item {
	// width: 135px;
	height: 36px;
	margin: 0;
	padding: 0;
	border-radius: 4px;

	.file-operation-div {
		width: 100%;
		height: 100%;
		padding: 8px;

		.file-operation-text {
			margin-left: 8px;
		}
	}
}

.mobile-submenu {
	width: 100%;
	max-width: 100%;
	border-radius: 12px 12px 0 0;
	overflow: hidden;
}

.mobile-submenu-header {
	height: 64px;
	padding: 0 20px;
	backdrop-filter: blur(30px);
	flex-wrap: nowrap;

	:deep(.q-btn) {
		flex-shrink: 0;
	}
}

.mobile-submenu-title-wrapper {
	min-width: 0;
	flex: 1;

	> div {
		width: 100%;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
}

.mobile-submenu-item {
	height: 48px;

	.file-operation-div {
		padding: 12px 20px;
	}
}
</style>

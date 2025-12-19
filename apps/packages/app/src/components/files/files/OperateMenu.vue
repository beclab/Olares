<template>
	<q-list
		dense
		class="menu-list text-ink-2 bg-background-2"
		v-if="visible && filteredContextmenuMenu.length > 0"
		:style="{ top: top + 'px', left: left + 'px' }"
	>
		<template v-for="(item, index) in filteredContextmenuMenu" :key="index">
			<file-operation-item
				:origin_id="origin_id"
				:icon="item.icon"
				:label="$t(item.name)"
				:action="item.action"
				@hide-menu="hideMenu"
			/>
		</template>
	</q-list>
</template>

<script setup lang="ts">
import { defineEmits, defineProps, ref, watch, computed, reactive } from 'vue';
import { useRoute } from 'vue-router';
import { useMenuStore } from '../../../stores/files-menu';
import { SYNC_STATE } from '../../../utils/contact';
import { useOperateinStore } from './../../../stores/operation';
import { common } from './../../../api';
import { backupOriginsRef } from './../../../constant';

import FileOperationItem from './FileOperationItem.vue';
import {
	useFilesStore,
	FilesIdType,
	ExternalType
} from '../../../stores/files';
import { ShareType, SharePermission } from 'src/utils/interface/share';
import { EventType } from '../../../stores/operation';
import { DriveType } from '../../../utils/interface/files';

const props = defineProps({
	clientX: Number,
	clientY: {
		type: Number,
		default: 0,
		required: true
	},
	offsetRight: Number,
	offsetBottom: Number,
	menuVisible: Boolean,
	menuList: Object,
	origin_id: {
		type: Number,
		required: true,
		default: FilesIdType.PAGEID
	}
});

const Route = useRoute();
const menuStore = useMenuStore();
const operateinStore = useOperateinStore();
const filesStore = useFilesStore();

const emit = defineEmits(['changeVisible']);

const visible = ref(false);
const top = ref<number>(props.clientY!);
const left = ref<number>(props.clientX!);

const eventType = reactive<EventType>({
	type: undefined,
	isSelected: false, //	true: Right click on a item; false: Right click on a blank area on the files
	hasCopied: false,
	showRename: true,
	isHomePage: false,
	selectCount: 0,
	rw: true,
	isExternal: false,
	canBackup: false,
	isDir: false,
	isShareItem: false,
	isOwner: false,
	isEditPermissionsEnable: false,
	isShareRoot: false,
	isPublicShare: false
	// isShareEditEnable: false
});

const filteredContextmenuMenu = computed(() => {
	const result = operateinStore.contextmenu.filter((item) =>
		item.condition(eventType)
	);
	return result;
});

watch(
	() => [
		Route.path,
		filesStore.selected[props.origin_id],
		Route.query.type as any,
		Route.query.p as any
	],
	(newVal) => {
		eventType.isExternal = false;
		eventType.canBackup = false;
		eventType.selectCount = newVal[1].length;
		eventType.isShareRoot = false;

		const fileId = newVal[1][0];
		const selectFiles =
			filesStore.currentFileList[props.origin_id]?.items?.[fileId];

		const permission = filesStore.currentFileList[props.origin_id]?.permission;
		eventType.rw = true;
		if (newVal[3] === 'shared' && newVal[4] === 'r') {
			eventType.rw = false;
		} else if (permission && permission == SharePermission.View) {
			eventType.rw = false;
		}

		if (!selectFiles) {
			if (common().displayConnectServer(newVal[0] as any)) {
				eventType.isExternal = true;
			}
			if (common().isShareRootPage(newVal[0] as any)) {
				eventType.isShareRoot = true;
			}
			return;
		}

		if (common().displayConnectServer(newVal[0] as any)) {
			eventType.isExternal = [
				ExternalType.HDD,
				ExternalType.SMB,
				ExternalType.USB
			].includes(selectFiles.externalType);
		} else {
			eventType.isExternal = false;
		}

		if (process.env.VERSIONTAG === '1.11') return;

		const driveType = common().formatUrltoDriveType(newVal[0] as string);
		if (
			driveType &&
			backupOriginsRef.value.includes(driveType) &&
			selectFiles.isDir
		) {
			eventType.canBackup = true;
		}
	},
	{ deep: true, immediate: true }
);

watch(
	() => [props.menuList],
	(newVal) => {
		if (newVal[0]) {
			eventType.isSelected = true;
			filterItem(newVal[0]);
		} else {
			eventType.isSelected = false;
		}

		const menuHeight = filteredContextmenuMenu.value.length * 36;
		if (props.offsetBottom && props.offsetBottom < menuHeight) {
			top.value = props.clientY - menuHeight;
		} else {
			top.value = props.clientY || 0;
		}
	}
);

watch(
	() => operateinStore.copyFiles,
	(newVaule) => {
		if (newVaule && newVaule.length > 0) {
			eventType.hasCopied = true;
		} else {
			eventType.hasCopied = false;
		}
	},
	{
		deep: true
	}
);

const filterItem = (item: any) => {
	eventType.isDir = false;
	eventType.isShareItem = item.isShareItem || false;
	eventType.isPublicShare = false;

	if (filesStore.selectedCount(props.origin_id) > 1) {
		eventType.showRename = false;
		eventType.isEditPermissionsEnable = false;
		eventType.isOwner = false;
	} else {
		eventType.showRename = true;
		eventType.isEditPermissionsEnable =
			item.isShareItem &&
			((item.share_type == ShareType.INTERNAL &&
				item.permission == SharePermission.ADMIN) ||
				item.share_type == ShareType.SMB);
		eventType.isDir = item.isDir || false;
		eventType.isOwner = item.owner === filesStore.users?.owner;
		eventType.isPublicShare = item.share_type == ShareType.PUBLIC;
	}
	eventType.type = item.driveType;

	const hasSelected = filesStore.currentFileList[
		props.origin_id
	]?.items?.filter((_, index) => {
		return filesStore.selected[props.origin_id].includes(index);
	});

	const hasSameValue = hasSelected?.find((item) =>
		operateinStore.isDisableMenuItem(item.name, Route.path)
	);

	if (hasSameValue) {
		eventType.isHomePage = true;
	} else {
		eventType.isHomePage = false;
	}
};

watch(
	() => props.clientX,
	(newVaule) => {
		if (props.offsetRight && props.offsetRight < 144) {
			left.value = (newVaule as number) - 144;
		} else {
			left.value = newVaule || 0;
		}
	}
);
watch(
	() => props.clientY,
	(newVaule) => {
		const menuHeight = filteredContextmenuMenu.value.length * 36 + 10;
		if (props.offsetBottom && props.offsetBottom < menuHeight) {
			top.value = (newVaule as number) - menuHeight;
		} else {
			top.value = newVaule || 0;
		}
	}
);

watch(
	() => props.menuVisible,
	(newVaule) => {
		visible.value = newVaule;
	}
);

const hideMenu = () => {
	emit('changeVisible');
};
</script>

<style scoped lang="scss">
.menu-list {
	border-radius: 8px;
	cursor: pointer;
	padding: 8px;
	position: fixed;
	z-index: 10;
	box-shadow: 0px 4px 10px 0px rgba(0, 0, 0, 0.2);

	.menu-item {
		width: 100%;
		border-radius: 4px;
		min-height: 36px !important;
		padding-left: 8px !important;
		padding-right: 2px !important;
		padding-top: 2px !important;
		padding-bottom: 2px !important;

		&:hover {
			background-color: $background-hover;
		}
	}
}
</style>

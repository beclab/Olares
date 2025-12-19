<template>
	<q-menu
		@update:model-value="showPopupProxy"
		class="menu-list bg-background-2"
	>
		<q-list dense padding style="width: 120px">
			<q-item
				class="row items-center justify-start text-ink-2 popup-item"
				clickable
				v-close-popup
				v-for="item in filteredContextmenuMenu"
				:key="item.action"
				@click="handleEvent(item.action, $event)"
			>
				<q-icon :name="item.icon" size="20px" class="q-mr-sm" />
				<q-item-section class="menuName"> {{ $t(item.name) }}</q-item-section>
			</q-item>
		</q-list>
	</q-menu>
</template>

<script setup lang="ts">
import { defineProps, computed, reactive } from 'vue';
import { useRoute } from 'vue-router';
import { useQuasar } from 'quasar';
import { useMenuStore } from '../../../stores/files-menu';
import { OPERATE_ACTION, SYNC_STATE } from '../../../utils/contact';
import { useOperateinStore } from './../../../stores/operation';

import { useFilesStore, FilesIdType } from '../../../stores/files';
import { EventType } from '../../../stores/operation';
import { useDataStore } from '../../../stores/data';
import { DriveType } from '../../../utils/interface/files';
// import FileOperationItem from '../files/FileOperationItem.vue';

const props = defineProps({
	menuList: Object,
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	}
});

const Route = useRoute();
// const dataStore = useDataStore();
const menuStore = useMenuStore();
const operateinStore = useOperateinStore();
const filesStore = useFilesStore();
const dataStore = useDataStore();

const $q = useQuasar();

const resetData = () => {
	if (props.menuList && props.menuList.length > 0) {
		eventType.isSelected = true;
		eventType.selectCount = props.menuList.length;
		filterItem(props.menuList[0]);
	} else {
		eventType.isSelected = false;
		eventType.selectCount = 0;
	}
};

const showPopupProxy = (value: boolean) => {
	if (value) {
		resetData();
	}
};

const eventType = reactive<EventType>({
	type: undefined,
	isSelected: false, //	true: Right click on a item; false: Right click on a blank area on the files
	hasCopied: false,
	showRename: true,
	isHomePage: false,
	selectCount: 0
});

const filteredContextmenuMenu = computed(() => {
	return operateinStore.contextmenu.filter((item) => item.condition(eventType));
});

const filterItem = (item: any) => {
	if (props.menuList && props.menuList) {
		eventType.showRename = true;
	} else {
		eventType.showRename = false;
	}

	if (item.type === DriveType.Sync) {
		const repo_id = Route.query.id?.toString();
		if (repo_id) {
			const status = menuStore.syncReposLastStatusMap[repo_id]
				? menuStore.syncReposLastStatusMap[repo_id].status
				: 0;

			let statusFalg =
				status > SYNC_STATE.DISABLE && status != SYNC_STATE.UNKNOWN;
			if (filesStore.selectedCount(props.origin_id) === 1 && statusFalg) {
				eventType.type = DriveType.Sync;
			} else {
				eventType.type = undefined;
			}
		} else {
			eventType.type = undefined;
		}
	} else {
		if (item.driveType && item.driveType !== DriveType.Sync) {
			eventType.type = item.driveType;
		}
	}

	const cur_menuList = props.menuList ? (props.menuList as any) : [];
	const hasSameValue = cur_menuList.find((item) =>
		operateinStore.isDisableMenuItem(item.name, Route.path)
	);

	if (hasSameValue) {
		eventType.isHomePage = true;
	} else {
		eventType.isHomePage = false;
	}
};

const handleEvent = async (action: OPERATE_ACTION, e: any) => {
	dataStore.showPadPopup = true;
	filesStore.selected[props.origin_id] = (props.menuList as any).map(
		(e) => e.index
	);
	operateinStore.handleFileOperate(
		props.origin_id,
		e,
		Route,
		action,
		filesStore.activeMenu(props.origin_id).driveType,
		async (action: OPERATE_ACTION, data: any) => {
			dataStore.showPadPopup = false;
			const url = Route.fullPath;
			const splitUrl = url.split('?');
			await filesStore.setFilePath(
				{
					path: splitUrl[0],
					isDir: true,
					driveType: filesStore.activeMenu(props.origin_id).driveType,
					param: splitUrl[1] ? `?${splitUrl[1]}` : ''
				},
				false,
				false
			);

			if (action == OPERATE_ACTION.PASTE) {
				operateinStore.resetCopyFiles();
			} else if (action == OPERATE_ACTION.OPEN_LOCAL_SYNC_FOLDER) {
				const repo_id = Route.query.id as string;
				const isElectron = $q.platform.is.electron;
				if (isElectron) {
					window.electron.api.files.openLocalRepo(repo_id, data);
				}
			} else if (
				action == OPERATE_ACTION.COPY ||
				action == OPERATE_ACTION.CUT ||
				action == OPERATE_ACTION.DOWNLOAD
			) {
				filesStore.resetSelected(props.origin_id);
			}
		}
	);
};
</script>

<style scoped lang="scss">
.menu-list {
	border-radius: 10px;
	cursor: pointer;
	padding: 5px 8px;
	position: fixed;
	z-index: 10;
	box-shadow: 0px 4px 10px 0px rgba(0, 0, 0, 0.2);
	width: 150px;

	.menu-item {
		width: 100%;
		border-radius: 5px;
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

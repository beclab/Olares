<template>
	<q-scroll-area
		class="bottom-menu-list"
		:thumb-style="contentStyle as any"
		:visible="true"
		content-style="height: 100%;"
	>
		<q-card-section
			horizontal
			class="row items-center text-ink-3"
			style="height: 48px"
		>
			<template v-for="(item, index) in filteredContextmenuMenu" :key="index">
				<div class="row items-center justify-center" style="width: 150px">
					<file-operation-item
						:icon="item.icon"
						:label="$t(item.name)"
						:action="item.action"
						@hide-menu="hideMenu"
					/>
				</div>
			</template>
		</q-card-section>
	</q-scroll-area>
</template>

<script setup lang="ts">
import { defineEmits, defineProps, ref, watch, computed, reactive } from 'vue';
import { useRoute } from 'vue-router';
import { useMenuStore } from '../../../stores/files-menu';
import { SYNC_STATE } from '../../../utils/contact';
import { useOperateinStore } from './../../../stores/operation';
import FileOperationItem from './FileOperationItem.vue';

import { useFilesStore, FilesIdType } from '../../../stores/files';
import { EventType } from '../../../stores/operation';
import { DriveType } from '../../../utils/interface/files';

const props = defineProps({
	menuList: Object,
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	}
});

const Route = useRoute();
const menuStore = useMenuStore();
const operateinStore = useOperateinStore();
const filesStore = useFilesStore();

const emit = defineEmits(['changeVisible']);

const eventType = reactive<EventType>({
	type: undefined,
	isSelected: false,
	hasCopied: false,
	showRename: true,
	isHomePage: false,
	selectCount: 0,
	rw: true,
	isExternal: false
});

const filteredContextmenuMenu = computed(() => {
	return operateinStore.contextmenu.filter((item) => item.condition(eventType));
});

watch(
	() => filesStore.selected[props.origin_id],
	() => {
		resetData();
	},
	{
		deep: true
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
	if (filesStore.selectedCount(props.origin_id) > 1) {
		eventType.showRename = false;
	} else {
		eventType.showRename = true;
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

const hideMenu = () => {
	emit('changeVisible');
};

const contentStyle = ref({
	height: 0
});

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
resetData();
</script>

<style scoped lang="scss">
.bottom-menu-list {
	width: 100%;
	height: 48px;
	border-top: 1px solid $separator;
	.menu-item {
		min-width: 120px;
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

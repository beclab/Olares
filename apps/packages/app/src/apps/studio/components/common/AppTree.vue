<template>
	<q-tree
		v-if="chartNodes && chartNodes.length > 0 && chartNodes[0].children"
		ref="treeRef"
		dense
		default-expand-all
		:nodes="chartNodes"
		icon="sym_r_keyboard_arrow_right"
		node-key="path"
		label-key="label"
		:selected="selected"
		v-model:expanded="expanded"
		@update:selected="onSelected"
		@update:expanded="onExpanded"
		no-selection-unset
		no-results-label=" "
		no-nodes-label=" "
		text-color="ink-2"
		class="my-tree-wrapper"
		@lazy-load="onLazyLoad"
	>
		<template v-slot:default-header="prop">
			<div
				:id="prop.node.key"
				class="my-tree-header-container"
				@mouseenter="mouseenter(prop.node.path)"
				@mouseleave="mouseleave(prop.node.path)"
			>
				<div
					class="row no-wrap items-center justify-between my-tree-header-wrapper"
					:id="prop.node.key"
				>
					<div class="file-item">
						<img
							v-if="prop.node.label === 'OlaresManifest.yaml'"
							class="q-mx-xs"
							src="../../assets/olares-icon.svg"
						/>
						<img
							v-else-if="!prop.node.isDir"
							class="q-mx-xs"
							src="../../assets/icon-yaml.svg"
						/>

						<span class="file-name">{{ prop.node.label }}</span>
					</div>

					<q-icon
						class="horiz q-mr-xs"
						:class="{ 'hide-horiz': mouseItemKey === prop.node.path }"
						rounded
						clickable
						name="sym_r_more_horiz"
						size="18px"
						:color="selected === mouseItemKey ? 'teal-pressed' : 'ink-2'"
						@click.stop
					>
						<PopupMenu
							:items="
								prop.node.isDir ? [...fileMenu, ...operateMenu] : operateMenu
							"
							:path="prop.node.path"
							:label="prop.node.label"
							@handleEvent="handleEvent"
						/>
					</q-icon>
				</div>
			</div>
		</template>
	</q-tree>
</template>

<script lang="ts" setup>
import { PropType, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import axios from 'axios';
import { useQuasar } from 'quasar';
import {
	ApplicationInfo,
	FilesSelectType,
	OPERATE_ACTION
} from '@apps/studio/src/types/core';
import { useDevelopingApps } from '../../stores/app';
import { useFileStore } from '../../stores/file';

import PopupMenu from '../common/PopupMenu.vue';
import CreateFile from '../../components/dialog/CreateFile.vue';
import DialogConfirm from '../../components/dialog/DialogConfirm.vue';

const props = defineProps({
	app: {
		type: Object as PropType<ApplicationInfo>,
		required: true
	},
	selected: {
		type: String,
		required: true,
		default: ''
	},
	chartNodes: {
		type: Object as PropType<FilesSelectType[]>,
		required: true
	}
});

const emits = defineEmits(['onSelected', 'updatePathNode']);

const { t } = useI18n();
const $q = useQuasar();
const store = useDevelopingApps();
const fileStore = useFileStore();

const expanded = ref<string[]>([]);
const mouseItemKey = ref();

const fileMenu = ref([
	{
		label: t(`enums.${OPERATE_ACTION.ADD_FOLDER}`),
		name: OPERATE_ACTION.ADD_FOLDER,
		icon: 'sym_r_create_new_folder'
	},
	{
		label: t(`enums.${OPERATE_ACTION.ADD_FILE}`),
		name: OPERATE_ACTION.ADD_FILE,
		icon: 'sym_r_note_add'
	}
]);

const operateMenu = ref([
	{
		label: t(`enums.${OPERATE_ACTION.RENAME}`),
		name: OPERATE_ACTION.RENAME,
		icon: 'sym_r_edit_square'
	},
	{
		label: t(`enums.${OPERATE_ACTION.DELETE}`),
		name: OPERATE_ACTION.DELETE,
		icon: 'sym_r_delete'
	}
]);

const onLazyLoad = async ({ node, done }) => {
	const res = await loadChildren(node);
	done(res);
};

const loadChildren = async (node: any) => {
	const res: any = await axios.get(store.url + '/api/files/' + node.path);

	const setChildren = (n: any, path: any, children: any) => {
		for (let i in n) {
			if (n[i].path == path && n[i].isDir) {
				n[i].children = children;
				return;
			}

			if (n[i].isDir && n[i].children.length > 0) {
				setChildren(n[i].children, path, children);
			}
		}
	};

	const children = getChildren(res.items);
	return children;
};

const getChildren = (items: any) => {
	let children: FilesSelectType[] = [];

	for (let n in items) {
		const data = items[n];
		const selectData: FilesSelectType = {
			label: data.name,
			icon: data.isDir ? 'folder' : 'article',
			path: data.path,
			expandable: data.isDir,
			selectable: !data.isDir,
			children: data.isDir ? [{}] : null,
			isDir: data.isDir,
			lazy: data.isDir ? true : false
		};

		if (data.items) {
			selectData.children = getChildren(data.items);
		}

		children.push(selectData);
	}

	return children;
};

const onSelected = async (value) => {
	emits('onSelected', value);
};

const onExpanded = (value) => {
	expanded.value = value;
};

const mouseenter = (key: string) => {
	mouseItemKey.value = key;
};

const mouseleave = (key: string) => {
	mouseItemKey.value = null;
};

const handleEvent = (action: OPERATE_ACTION, path: string, label: string) => {
	switch (action) {
		case OPERATE_ACTION.ADD_FOLDER:
			createDialog(path, action);
			break;

		case OPERATE_ACTION.ADD_FILE:
			createDialog(path, action);
			break;

		case OPERATE_ACTION.RENAME:
			renameDialog(path, label, action);
			break;

		case OPERATE_ACTION.DELETE:
			deleteFile(path);
			break;
	}
};

const createDialog = (path: string, action: OPERATE_ACTION) => {
	$q.dialog({
		component: CreateFile,
		componentProps: {
			path,
			action
		}
	}).onOk((val) => {
		expanded.value?.push(path);
		emits('updatePathNode', {
			path,
			action,
			name: val
		});
	});
};

const renameDialog = (path: string, label: string, action: OPERATE_ACTION) => {
	$q.dialog({
		component: CreateFile,
		componentProps: {
			path,
			action,
			label
		}
	}).onOk(() => {
		const parentPath = path.split('/').slice(0, -1).join('/');
		emits('updatePathNode', {
			path: parentPath,
			action
		});
	});
};

const deleteFile = async (path: string) => {
	const target_name = path.substring(path.lastIndexOf('/') + 1);

	$q.dialog({
		component: DialogConfirm,
		componentProps: {
			title: t('btn_delete'),
			message: t('message.deleteTip', { name: target_name })
		}
	}).onOk(async () => {
		_deleteFile(path);
	});
};

const _deleteFile = async (path: string) => {
	const parentPath = path.split('/').slice(0, -1).join('/');
	await fileStore.deleteFile(path);
	emits('updatePathNode', {
		path: parentPath,
		action: OPERATE_ACTION.DELETE
	});
};
</script>

<style lang="scss" scoped>
.file-item {
	width: calc(100% - 24px);
	height: 24px;

	display: flex;
	align-items: center;
	justify-content: flex-start;

	img {
		width: 16px;
		float: left;
	}

	.file-name {
		width: 100%;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
}

.horiz {
	opacity: 0;
}
.hide-horiz {
	opacity: 1;
}

.my-tree-wrapper {
	width: 100%;

	$my_tree_width_base: 4px;
	$my_tree_left: -($my_tree_width_base + 4px);
	&::v-deep(.q-tree) {
		padding: 0 20px;
	}

	&::v-deep(.q-tree__node-header) {
		padding: 0;
		border-radius: 4px;
		.file-name {
			font-size: 14px !important;
			color: $ink-1 !important;
			font-weight: 500;
		}
	}

	&::v-deep(.q-tree__node-collapsible .q-tree__node-header) {
		.file-name {
			font-size: 12px !important;
			color: $ink-2 !important;
			font-weight: 400;
		}
	}

	::v-deep(.q-tree__node--selected) {
		background-color: $teal-soft;
		.file-item {
			.file-name {
				color: $teal-pressed !important;
			}
		}
	}

	::v-deep(.q-tree__arrow) {
		font-size: 14px;
		border: 1px solid rgba(255, 255, 255, 0);
		color: $ink-1;
	}
}
::v-deep(.selected-node) {
	position: relative;
	&::after {
		content: '';
		position: absolute;
		left: -4px;
		right: 4px;
		top: 0;
		bottom: 0;
		background-color: rgba(34, 111, 255, 0.12);
		border-radius: 4px;
	}
}
.my-tree-header-container {
	width: 100%;
	height: 24px;
	line-height: 24px;
	position: relative;
	border-radius: 4px;
	.my-tree-badge-wrapper {
		margin-left: 4px;
	}
	.my-tree-avatar {
		position: relative;
		& ~ .my-tree-badge-wrapper {
			position: absolute;
			bottom: 0;
			right: 0;
		}
	}
}

$my_tree_width_base: 8px;
$my_tree_left: -($my_tree_width_base + 4px);

::v-deep(.q-tree__node--child) {
	padding-left: 0px;
}
::v-deep(.q-tree__node-header:before) {
	height: 24px;
	border-bottom: none;
}

::v-deep(.q-tree__node--parent .q-tree__node-header:before) {
	border-left: 1px solid $separator-2 !important;
}

::v-deep(.q-tree__node--child > .q-tree__node-header:before) {
	left: -8px;
	top: 0px;
	border-left: 1px solid $separator-2 !important;
}

::v-deep(.q-tree--dense .q-tree__node-header:before) {
	display: none !important;
}

::v-deep(
		.q-tree__node--parent > .q-tree__node-collapsible > .q-tree__node-body:after
	) {
	border-color: #e0e0e0;
}
::v-deep(.q-tree__node:after) {
	border-color: $separator-2;
}

::v-deep(.my-menu-before-scroll .q-scrollarea__container) {
	overflow-x: hidden;
}
::v-deep(.my-menu-before-scroll .q-scrollarea__content) {
	min-width: initial;
	width: 100%;
}
::v-deep(.my-menu-before-scroll .q-tree__node-collapsible) {
	overflow-x: hidden;
}
</style>
./PopupMenu.vue../dialog/CreateFile.vue../dialog/DialogConfirm.vue

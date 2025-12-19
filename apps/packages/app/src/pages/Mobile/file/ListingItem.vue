<template>
	<div
		class="mobile-item"
		:class="itemClass"
		role="button"
		tabindex="0"
		:draggable="isDraggable"
		@dragstart="dragStart"
		@dragover="dragOver"
		@drop="drop"
		@click="itemClick"
		:aria-label="item.name"
		:aria-selected="isSelected"
	>
		<terminus-file-icon
			:name="item.name"
			:type="item.type"
			:path="item.path"
			:modified="item.modified"
			:is-dir="item.isDir"
			:driveType="item.driveType"
			:icon-size="viewMode === 'mosaic' ? 64 : 32"
		/>

		<q-icon
			v-if="viewMode === 'mosaic'"
			name="sym_r_more_horiz"
			size="20px"
			@click.stop="fileOperation"
			style="
				position: absolute;
				right: 0px;
				top: 0px;
				cursor: pointer;
				border-radius: 12px;
			"
		>
		</q-icon>

		<div v-if="viewMode === 'mosaic'" style="margin-top: 0.3rem">
			<span class="name text-ink-1">{{
				translateFolderName(route.path, item.name)
			}}</span>
			<div class="modified text-ink-3">
				{{
					formatFileModified(item.modified, 'YYYY-MM-DD') + ' · ' + humanSize()
				}}
			</div>
		</div>
		<div v-else style="width: calc(100% - 32px)" class="q-pl-md">
			<div style="width: 100%" class="row items-center justify-between">
				<div class="column justify-center" style="width: calc(100% - 40px)">
					<div class="name text-ink-1">
						{{ translateFolderName(route.path, item.name) }}
					</div>
					<div class="modified text-ink-3" style="max-width: 100%">
						{{ formatFileModified(item.modified) }}
						<span class="q-ml-xs" v-if="!item.isDir">{{ humanSize() }}</span>
					</div>
				</div>
				<div
					v-if="showMoreHoriz"
					class="column justify-center items-center"
					style="height: 32px; width: 32px"
					@click.stop="fileOperation"
				>
					<q-icon name="sym_r_more_horiz" size="20px" class="grey-8"> </q-icon>
				</div>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { computed, onMounted, onUnmounted, PropType } from 'vue';
import { useDataStore } from '../../../stores/data';
import { useQuasar } from 'quasar';
import { format } from '../../../utils/format';

import { useRoute } from 'vue-router';
import TerminusFileIcon from '../../../components/common/TerminusFileIcon.vue';
import FileOperationDialog from './FileOperationDialog.vue';
import {
	useFilesStore,
	FileItem,
	FilesIdType,
	PickType
} from '../../../stores/files';
import { formatFileModified, hideHeaderOpt } from '../../../utils/file';
import { useOperateinStore } from '../../../stores/operation';
import { OPERATE_ACTION } from '../../../utils/contact';
import { getParams } from '../../../utils/utils';
// import { useI18n } from 'vue-i18n';
import { translateFolderName } from '../../../utils/file';
import { busOff, busOn } from 'src/utils/bus';

const props = defineProps({
	item: {
		type: Object as PropType<FileItem>,
		required: true
	},
	viewMode: {
		type: String,
		required: true
	},
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	},
	selectType: {
		type: String as PropType<PickType>,
		required: false,
		default: ''
	},
	from: {
		type: String,
		default: '',
		required: false
	}
});

const { humanStorageSize } = format;

const emits = defineEmits(['openSyncPage']);

// const { t } = useI18n();
const store = useDataStore();
const filesStore = useFilesStore();
const route = useRoute();
const operateinStore = useOperateinStore();

const itemClass = computed(() => {
	if (filesStore.isShard && !props.item.isDir) {
		return 'disabled';
	}
	if (props.selectType === PickType.FOLDER && !props.item.isDir) {
		return 'disabled';
	}
	return '';
});

const showMoreHoriz = computed(() => {
	if (filesStore.isShard) {
		return false;
	}

	if (!hideHeaderOpt(route.path)) {
		return false;
	}

	return true;
});

const isSelected = computed(function () {
	if (filesStore.selected && filesStore.selected[props.origin_id]) {
		return (
			filesStore.selected[props.origin_id].indexOf(props.item.index) !== -1
		);
	}
	return false;
});

const isDraggable = computed(function () {
	return store.user?.perm?.rename;
});

const canDrop = computed(function () {
	if (!props.item.isDir) return false;

	for (let i of filesStore.selected[props.origin_id]) {
		if (
			filesStore.currentFileList[props.origin_id]?.items[i].url ===
			props.item.url
		) {
			return false;
		}
	}

	return true;
});

const humanSize = () => {
	return props.item.type == 'invalid_link'
		? 'invalid link'
		: humanStorageSize(props.item.size);
};

const dragStart = () => {
	if (filesStore.selectedCount(props.origin_id) === 0) {
		filesStore.addSelected(props.item.index);
		return;
	}

	if (!isSelected.value) {
		filesStore.resetSelected();
		filesStore.addSelected(props.item.index);
	}
};

const dragOver = (event: any) => {
	if (!canDrop.value) return;

	event.preventDefault();
	let el = event.target;

	for (let i = 0; i < 5; i++) {
		if (!el.classList.contains('item')) {
			el = el.parentElement;
		}
	}

	el.style.opacity = 1;
};

const drop = async (event: any) => {
	let canMove = true;
	for (const item of filesStore.selected[props.origin_id]) {
		if (
			operateinStore.isDisableMenuItem(
				filesStore.currentFileList[props.origin_id]?.items[item].name || '',
				route.path
			)
		) {
			canMove = false;
			break;
		}
	}

	if (!canMove) return false;

	if (!canDrop.value) return;
	event.preventDefault();

	if (filesStore.selectedCount(props.origin_id) === 0) return;

	operateinStore.handleFileOperate(
		props.origin_id,
		event,
		{
			...route,
			path: props.item.path
		},
		OPERATE_ACTION.MOVE,
		props.item.driveType,
		async () => {}
	);
};

const itemClick = () => {
	open();
};

const $q = useQuasar();

const fileOperation = () => {
	filesStore.addSelected(props.item.index);
	const path = props.item.path;
	const p = getParams(path, 'p');
	let permission = 'rw';
	if (p) {
		permission = p;
	}

	$q.dialog({
		component: FileOperationDialog,
		componentProps: {
			index: props.item.index,
			name: props.item.name,
			modified: props.item.modified,
			type: props.item.type,
			isDir: props.item.isDir,
			driveType: props.item.driveType,
			origin_id: props.origin_id,
			permission
		}
	});
};

const busFileItemAction = (index: number) => {
	if (props.item.index != index) {
		return;
	}
	fileOperation();
};

onMounted(() => {
	busOn('fileItemOpenOperation', busFileItemAction);
});

onUnmounted(() => {
	busOff('fileItemOpenOperation', busFileItemAction);
});

const open = async () => {
	if (props.from === 'sync') {
		emits('openSyncPage');
	}
	if (!props.item.isDir && !filesStore.isShard) {
		if (props.selectType) return false;
		if (store.preview.isShow) {
			return;
		}

		filesStore.openPreviewDialog(props.item, props.origin_id);
	} else {
		const splitUrl = props.item.path.split('?');
		filesStore.setCurrentNode(
			props.item.node || props.item.fileExtend,
			props.origin_id
		);

		await filesStore.setFilePath(
			{
				path: splitUrl[0],
				isDir: props.item.isDir,
				driveType: props.item.driveType,
				param: splitUrl[1] ? `?${splitUrl[1]}` : ''
			},
			false,
			true,
			props.origin_id
		);
	}
};
</script>

<style lang="scss" scoped>
.disabled {
	pointer-events: none;
	opacity: 0.5;
}

#listing.list {
	.mobile-item {
		display: flex;
		align-items: center;
		justify-content: space-between;
		border-bottom: 1px solid $separator;

		.name {
			width: 100%;
			overflow: hidden;
			text-overflow: ellipsis;
			white-space: nowrap;
		}
	}
}
</style>

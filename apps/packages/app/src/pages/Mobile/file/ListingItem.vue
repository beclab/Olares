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
		@touchstart="onTouchStart"
		@touchend="onTouchEnd"
		@touchcancel="onTouchEnd"
		@touchmove="onTouchMove"
		:aria-label="item.name"
		:aria-selected="isSelected"
	>
		<div
			v-if="viewMode === 'mosaic' && deviceStore.isInEditor && canSelect"
			class="mosaic-select-box"
			@click.stop="toggleSelect"
		>
			<img v-if="isSelected" :src="checkboxActiveImage" />
			<img v-else-if="!$q.dark.isActive" :src="checkboxNormalImage" />
			<img v-else :src="checkboxNormalDarkImage" />
		</div>

		<terminus-file-icon
			:name="item.name"
			:type="item.type"
			:path="item.path"
			:modified="item.modified"
			:is-dir="item.isDir"
			:driveType="item.driveType"
			:icon-size="
				viewMode === 'mosaic' ? (deviceStore.isInEditor ? 48 : 64) : 32
			"
		/>

		<q-icon
			v-if="viewMode === 'mosaic' && showMoreHoriz && !deviceStore.isInEditor"
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

		<div
			v-if="viewMode === 'mosaic'"
			class="mosaic-meta"
			style="margin-top: 0.3rem"
		>
			<span class="name text-ink-1">{{
				translateFolderName(route.path, item.name)
			}}</span>
			<div class="modified text-ink-3" v-if="item.modified && item.size > 0">
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
import { useDeviceStore } from '../../../stores/device';
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
import { useI18n } from 'vue-i18n';
import { translateFolderName } from '../../../utils/file';
import { busOff, busOn } from 'src/utils/bus';
import { FileViewMode } from 'src/utils/interface/files';

const props = defineProps({
	item: {
		type: Object as PropType<FileItem>,
		required: true
	},
	viewMode: {
		type: String as PropType<FileViewMode>,
		required: false,
		default: 'list'
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

const emits = defineEmits(['openSyncPage', 'enterSelectMode', 'selectChanged']);

const { t } = useI18n();
const store = useDataStore();
const filesStore = useFilesStore();
const deviceStore = useDeviceStore();
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

// Some entries (e.g. system folders under drive/Home, external/AI/common
// folder whitelists) can't be operated on and must not be selectable. Mirrors
// the `selectedEnable` rule used by the list-view multi-select.
const canSelect = computed(() => {
	if (itemClass.value) {
		return false;
	}
	return !operateinStore.isDisableMenuItem(props.item.name, route.path);
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
		const item = filesStore.getTargetFileItem(i, props.origin_id);
		if (!item) {
			continue;
		}
		if (item.url === props.item.url) {
			return false;
		}
	}

	return true;
});

const humanSize = () => {
	return props.item.type == 'invalid_link'
		? t('files.invalid_link')
		: props.item.size
		? humanStorageSize(props.item.size)
		: '--';
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
		const fileItem = filesStore.getTargetFileItem(item, props.origin_id);
		if (operateinStore.isDisableMenuItem(fileItem?.name || '', route.path)) {
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

const $q = useQuasar();

const checkboxActiveImage = `${
	$q.platform.is.electron ? '.' : ''
}/img/checkbox/check_box.svg`;
const checkboxNormalImage = `${
	$q.platform.is.electron ? '.' : ''
}/img/checkbox/uncheck_box_light.svg`;
const checkboxNormalDarkImage = `${
	$q.platform.is.electron ? '.' : ''
}/img/checkbox/uncheck_box_dark.svg`;

// Long-press to enter multi-select mode (mosaic view). `longPressed` swallows
// the synthetic click that fires after the press is released, so the item just
// selected by the long-press is not immediately toggled back off.
const LONG_PRESS_DURATION = 500;
// Tolerate small finger drift during a press; cancelling on any movement makes
// the long-press almost impossible to trigger on real touch devices.
const LONG_PRESS_MOVE_TOLERANCE = 10;
let pressTimer: ReturnType<typeof setTimeout> | null = null;
let longPressed = false;
let touchStartX = 0;
let touchStartY = 0;

const isSelectModeActive = computed(
	() => props.viewMode === 'mosaic' && deviceStore.isInEditor
);

const toggleSelect = () => {
	if (!canSelect.value) {
		return;
	}
	if (isSelected.value) {
		filesStore.removeSelected(props.item.index, props.origin_id);
	} else {
		filesStore.addSelected(props.item.index, props.origin_id);
	}
	emits('selectChanged');
};

const clearPressTimer = () => {
	if (pressTimer) {
		clearTimeout(pressTimer);
		pressTimer = null;
	}
};

const onTouchStart = (event: TouchEvent) => {
	if (
		props.viewMode !== 'mosaic' ||
		props.selectType ||
		deviceStore.isInEditor ||
		!canSelect.value
	) {
		return;
	}
	const touch = event.touches[0];
	if (touch) {
		touchStartX = touch.clientX;
		touchStartY = touch.clientY;
	}
	longPressed = false;
	clearPressTimer();
	pressTimer = setTimeout(() => {
		longPressed = true;
		emits('enterSelectMode', props.item.index);
	}, LONG_PRESS_DURATION);
};

const onTouchMove = (event: TouchEvent) => {
	const touch = event.touches[0];
	if (!touch) {
		return;
	}
	if (
		Math.abs(touch.clientX - touchStartX) > LONG_PRESS_MOVE_TOLERANCE ||
		Math.abs(touch.clientY - touchStartY) > LONG_PRESS_MOVE_TOLERANCE
	) {
		clearPressTimer();
	}
};

const onTouchEnd = () => {
	clearPressTimer();
};

const itemClick = () => {
	// Ignore the click that immediately follows a long-press selection.
	if (longPressed) {
		longPressed = false;
		return;
	}
	if (isSelectModeActive.value) {
		toggleSelect();
		return;
	}
	open();
};

const fileOperation = () => {
	filesStore.addSelected(props.item.index, props.origin_id);
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

.mosaic-select-box {
	position: absolute;
	left: 4px;
	top: 4px;
	z-index: 2;
	width: 16px;
	height: 16px;
	display: flex;
	align-items: center;
	justify-content: center;

	img {
		width: 16px;
		height: 16px;
	}
}

#listing-mobile.mosaic .mobile-item[aria-selected='true'] {
	background: $background-hover;
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

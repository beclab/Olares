<template>
	<div
		class="item text-ink-1"
		:class="itemClass"
		role="button"
		tabindex="0"
		:draggable="isDraggable"
		@dragstart="dragStart"
		@dragover="dragOver"
		@drop="drop"
		@click.stop="click"
		@dblclick.stop="open"
		:data-dir="item.isDir"
		:data-type="item.type"
		:aria-label="item.name"
		:aria-selected="isSelected"
	>
		<p
			v-if="isPad"
			class="select-common"
			:class="isSelected && !store.showPadPopup ? 'selected' : 'unselect'"
			:style="
				viewMode === 'mosaic'
					? 'position: absolute;left: 10px;top: 10px;cursor: pointer;'
					: ''
			"
			@click.stop="
				isSelected
					? filesStore.removeSelected(item.index)
					: filesStore.addSelected(item.index)
			"
		>
			<q-icon class="icon text-ink-on-brand" v-if="isSelected" name="check" />
		</p>
		<terminus-file-icon
			:name="item.name"
			:type="item.type"
			:path="item.path"
			:modified="item.modified"
			:is-dir="item.isDir"
			:driveType="item.driveType"
			:icon-size="viewMode === 'mosaic' ? 64 : 32"
			:class="isPad && viewMode === 'mosaic' ? 'q-mt-md' : ''"
		/>

		<q-icon
			v-if="
				isPad &&
				viewMode === 'mosaic' &&
				filesStore.selected[origin_id].length == 0
			"
			name="sym_r_more_horiz"
			class="text-ink-3"
			size="20px"
			style="
				position: absolute;
				right: 10px;
				top: 10px;
				cursor: pointer;
				border-radius: 12px;
			"
			@click.stop=""
		>
			<PadFileItemPopupMenu :menu-list="[item]" />
		</q-icon>

		<div
			v-if="viewMode === 'mosaic'"
			class="q-px-sm"
			style="margin-top: 0.5rem"
		>
			<span>{{ translateFolderName(route.path, item.name) }}</span>
		</div>
		<div v-else>
			<template v-if="item.isShareItem == true">
				<p
					:class="{
						'share-name': true
					}"
				>
					{{ translateFolderName(route.path, item.name) }}
				</p>
				<p class="expiration-date">
					{{
						item.share_type == ShareType.INTERNAL ||
						item.share_type == ShareType.SMB
							? '-'
							: formatFileModified(item.expire_time || '')
					}}
				</p>
				<p class="permission">
					{{
						sharePermissionStr(
							item.shared_by_me ? SharePermission.ADMIN : item.permission
						)
					}}
				</p>

				<p class="share-scope">
					{{ item.share_type ? shareTypeStr(item.share_type) : '-' }}
				</p>
				<p class="owner">
					{{ item.owner }}
				</p>
			</template>
			<template v-else>
				<p
					:class="{
						name: !isPad || filesStore.selected[origin_id].length == 0,
						action_name: isPad && filesStore.selected[origin_id].length > 0
					}"
				>
					{{ translateFolderName(route.path, item.name) }}
				</p>
				<p class="modified">
					<time :datetime="item.modified">
						{{ humanTime() }}
					</time>
				</p>
				<p class="type">{{ item.externalType || item.type || 'Folder' }}</p>
				<p class="size" :data-order="humanSize()">
					{{ humanSize() }}
				</p>
			</template>
		</div>
	</div>
</template>

<script setup lang="ts">
import {
	ref,
	computed,
	watch,
	onMounted,
	defineProps,
	PropType,
	defineEmits
} from 'vue';
import { useDataStore } from '../../stores/data';
import { useOperateinStore } from '../../stores/operation';
import moment from 'moment';
import { formatDateFromNow } from 'src/utils/format';
import { useRoute } from 'vue-router';
import TerminusFileIcon from '../common/TerminusFileIcon.vue';
import { notifyWarning } from '../../utils/notifyRedefinedUtil';

import { OPERATE_ACTION } from '../../utils/contact';
import { formatFileModified, translateFolderName } from '../../utils/file';
import { useI18n } from 'vue-i18n';
import {
	useFilesStore,
	FileItem,
	FilesIdType,
	PickType,
	sharePermissionStr,
	shareTypeStr
} from '../../stores/files';

import { getAppPlatform } from '../../application/platform';
import PadFileItemPopupMenu from './popup/PadFileItemPopupMenu.vue';
import { format } from '../../utils/format';
import { ShareType, SharePermission } from 'src/utils/interface/share';

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
	}
});

const emits = defineEmits(['resetOpacity', 'closeMenu']);

const { humanStorageSize } = format;

const filesStore = useFilesStore();
const store = useDataStore();
const route = useRoute();
const operateinStore = useOperateinStore();
const fileName = ref('');
const { t } = useI18n();

const isPad = ref(getAppPlatform() && getAppPlatform().isPad);

onMounted(() => {
	fileName.value = props.item.name;
});

const isCurSelectType = computed(() => {
	if (
		(props.selectType === PickType.FILE && props.item.isDir) ||
		(props.selectType === PickType.FOLDER && !props.item.isDir)
	) {
		return true;
	}
	return false;
});

const itemClass = computed(() => {
	let className = '';
	if (isPad.value && props.viewMode === 'mosaic') {
		className += 'bg-background-3 ';
	}
	if (props.selectType) {
		className += 'noBorder text-subtitle3 ';
		if (isCurSelectType.value) {
			className += 'opacity ';
		}
	}
	return className;
});

const singleClick = computed(function () {
	return store.user.singleClick;
});

const isSelected = computed(function () {
	return (
		filesStore.selected[props.origin_id] &&
		filesStore.selected[props.origin_id].indexOf(props.item.index) !== -1
	);
});

const isDraggable = computed(function () {
	return store.user?.perm?.rename;
});

const canDrop = computed(function () {
	if (!props.item.isDir) return false;

	for (let i of filesStore.selected[props.origin_id]) {
		if (
			filesStore.currentFileList[props.origin_id]?.items[i].path ===
			props.item.path
		) {
			return false;
		}
	}

	return true;
});

const humanSize = () => {
	if (props.item.type == 'invalid_link') {
		return t('files.invalid_link');
	}

	if (props.item.isDir) {
		return '-';
	}

	return humanStorageSize(props.item.size);
};

const humanTime = () => {
	if (!props.item.modified) return '-';
	if (store.user.dateFormat) {
		return moment(props.item.modified).format('L LT');
	}

	return formatDateFromNow(props.item.modified);
};

const dragStart = () => {
	if (filesStore.selectedCount(props.origin_id) === 0) {
		filesStore.addSelected(props.item.index);
		return;
	}

	if (!isSelected.value) {
		filesStore.resetSelected(props.origin_id);
		filesStore.addSelected(props.item.index, props.origin_id);
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
	event.preventDefault();

	let canMove = true;
	for (const item of filesStore.selected[props.origin_id]) {
		if (
			operateinStore.isDisableMenuItem(
				filesStore.currentFileList[props.origin_id]?.items[item].name || '',
				route.path
			)
		) {
			canMove = false;
			notifyWarning(t('files.the_files_contains_unmovable_items'));
			break;
		}
	}

	if (!canMove) return false;

	emits('resetOpacity');
	if (!canDrop.value) return;

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

const click = (event: any) => {
	if (isCurSelectType.value) return false;

	if (isPad.value) {
		open();
		return;
	}
	emits('closeMenu');
	if (filesStore.selectedCount(props.origin_id) !== 0) event.preventDefault();

	if (filesStore.selected[props.origin_id].indexOf(props.item.index) !== -1) {
		filesStore.removeSelected(props.item.index, props.origin_id);
		return;
	}

	if (event.shiftKey && filesStore.selected[props.origin_id].length > 0) {
		let fi = 0;
		let la = 0;
		const fileRes = filesStore.currentFileList[props.origin_id];
		if (!fileRes) {
			return;
		}

		const curFilesIndex = fileRes.items.findIndex(
			(item) => item.index === props.item.index
		);

		const firstFilesIndex = fileRes.items.findIndex(
			(item) => item.index === filesStore.selected[props.origin_id][0]
		);

		if (curFilesIndex > firstFilesIndex) {
			fi = firstFilesIndex + 1;
			la = curFilesIndex;
		} else {
			fi = curFilesIndex;
			la = firstFilesIndex - 1;
		}

		for (; fi <= la; fi++) {
			const selectFileIndex = fileRes.items[fi].index;

			if (filesStore.selected[props.origin_id].indexOf(selectFileIndex) == -1) {
				filesStore.addSelected(selectFileIndex, props.origin_id);
			}
		}

		return;
	}

	if (!singleClick.value && !event.ctrlKey && !event.metaKey) {
		filesStore.resetSelected(props.origin_id);
	}

	filesStore.addSelected(props.item.index, props.origin_id);
};

const open = async () => {
	emits('closeMenu');
	const splitUrl = props.item.path.split('?');

	if (!props.item.isDir) {
		if (props.selectType) return false;
		if (store.preview.isShow) {
			return;
		}
		filesStore.openPreviewDialog(props.item, props.origin_id);
	} else {
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

const isMutilSelected = ref(false);

// const isSelected = computed(function () {
// 	return store.selected.indexOf(props.index) !== -1;
// });

if (isPad.value) {
	watch(
		() => filesStore.selected,
		() => {
			console.log('store.mutilSelected ===>');
			console.log(filesStore.selected);
			isMutilSelected.value =
				filesStore.selected[props.origin_id].indexOf(props.item.index) !== -1;
		},
		{
			deep: true
		}
	);
}
</script>

<style lang="scss" scoped>
#listing.list .item {
	border-bottom: 1px solid $separator;
}

#listing.list .item.noBorder {
	border: 0;
	padding: 4px 14px;
	cursor: pointer;
}
#listing.list .item.opacity {
	opacity: 0.5;
}

#listing .item[aria-selected='true'] {
	background: $background-hover;
}

#listing .item[aria-selected-border='true'] {
	border: 1px solid $light-blue-default;
}

#listing .item:hover {
	background: $background-hover;
}
</style>

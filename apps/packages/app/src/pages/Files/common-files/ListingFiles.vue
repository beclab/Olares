<template>
	<div
		ref="pageRef"
		tabindex="0"
		@contextmenu.prevent="rightClick($event, undefined)"
		@click.stop="leftClick"
		class="files-body"
	>
		<div
			class="empty column items-center justify-center full-height"
			v-if="
				(filesStore.currentDirItems(origin_id)?.length || 0) +
					(filesStore.currentFileItems(origin_id)?.length || 0) ==
				0
			"
		>
			<img src="../../../assets/nodata.svg" alt="empty" />
			<span
				class="text-body2 text-ink-1"
				v-if="filesStore.loading[origin_id]"
				>{{ $t('files.loading') }}</span
			>
			<span class="text-body2 text-ink-1" v-else>{{ $t('files.lonely') }}</span>
		</div>
		<bt-scroll-area
			v-else
			id="listing"
			ref="listing"
			:class="
				isHovered
					? `${store.user.viewMode} file-icons hovered`
					: `${store.user.viewMode} file-icons`
			"
			@mouseover="handleMouseOver"
			@mouseleave="handleMouseLeave"
		>
			<!-- :by-me="currentDriveType === DriveType.ShareByMe" -->
			<files-share-table-header :origin_id="origin_id" v-if="isSharePageRoot" />
			<files-table-header :origin_id="origin_id" v-else />

			<div class="common-div" :style="isPad ? '' : 'padding: 0 20px 34px 20px'">
				<ListingItem
					v-for="item in filesStore.currentDirItems(origin_id)"
					:key="base64(item.name) + item.id"
					:item="item"
					:origin_id="origin_id"
					v-bind:viewMode="store.user.viewMode"
					@contextmenu.stop="rightClick($event, item)"
					@closeMenu="changeVisible"
					@resetOpacity="resetOpacity"
					@showMenu="
						(event) => {
							rightClick(event, item);
						}
					"
				>
				</ListingItem>

				<ListingItem
					v-for="item in filesStore.currentFileItems(origin_id)"
					:key="base64(item.name) + item.id"
					:item="item"
					:origin_id="origin_id"
					v-bind:viewMode="store.user.viewMode"
					@contextmenu.stop="rightClick($event, item)"
					@closeMenu="changeVisible"
					@resetOpacity="resetOpacity"
					@showMenu="
						(event) => {
							rightClick(event, item);
						}
					"
				>
				</ListingItem>
			</div>
			<!-- </BtScrollArea> -->
		</bt-scroll-area>

		<OperateMenuVue
			:origin_id="origin_id"
			:clientX="clientX"
			:clientY="clientY"
			:offsetRight="offsetRight"
			:offsetBottom="offsetBottom"
			:menuVisible="menuVisible"
			:menuList="menuList"
			@changeVisible="changeVisible"
		/>
	</div>

	<files-uploader
		:origin_id="origin_id"
		:auto-bind-resumable="
			application.filesUploadConfig
				? application.filesUploadConfig.autoBindResumable
				: false
		"
		@files-update="filesUpdate"
	/>
</template>

<script lang="ts" setup>
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue';
import { useRoute } from 'vue-router';
import throttle from 'lodash.throttle';
import { useI18n } from 'vue-i18n';
import { useDataStore } from '../../../stores/data';
import { scanFiles } from '../../../utils/upload';
import { stopScrollMove, startScrollMove } from '../../../utils/utils';
import { disabledClick, hideHeaderOpt } from '../../../utils/file';
import ListingItem from '../../../components/files/ListingItem.vue';
import OperateMenuVue from '../../../components/files/files/OperateMenu.vue';
import { OPERATE_ACTION } from '../../../utils/contact';

import { useOperateinStore } from './../../../stores/operation';
import { getAppPlatform } from '../../../application/platform';
import FilesUploader from './FilesUploader.vue';

import { useFilesStore, FilesIdType, FilePath } from './../../../stores/files';

import { DriveType } from '../../../utils/interface/files';
import { useTransfer2Store } from '../../../stores/transfer2';

import { common, commonV2 } from './../../../api';
import { stringToBase64 } from '@didvault/sdk/src/core';

import FilesTableHeader from '../../../components/files/FilesTableHeader.vue';
import FilesShareTableHeader from '../../../components/files/header/FilesShareTableHeader.vue';
import { busEmit } from '../../../utils/bus';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { getApplication } from 'src/application/base';

const props = defineProps({
	origin_id: {
		type: Number,
		required: true,
		default: FilesIdType.PAGEID
	}
});

const { t } = useI18n();
const store = useDataStore();
const route = useRoute();
const operateinStore = useOperateinStore();

const dragCounter = ref<number>(0);
const itemWeight = ref<number>(0);
const listing = ref<any>(null);
const clientX = ref<number>(0);
const clientY = ref<number>(0);
const offsetRight = ref<number>(0);
const offsetBottom = ref<number>(0);
const menuVisible = ref(false);
const menuList = ref<undefined | object>({});
const width = ref<number>(window.innerWidth);
let showLimit = 50;
const isHovered = ref();
// const repoId = ref();

const filesStore = useFilesStore();
const transferStore = useTransfer2Store();
const isPad = ref(getAppPlatform() && getAppPlatform().isPad);

const application = getApplication();

const pageRef = ref();

const isExternal = ref(common().displayConnectServer(route.fullPath));

const isSharePageRoot = ref(false);

const currentDriveType = ref(DriveType.Drive);

watch(
	() => route.fullPath,
	(newVal) => {
		currentDriveType.value =
			common().formatUrltoDriveType(newVal) || DriveType.Drive;

		isExternal.value = common().displayConnectServer(route.fullPath);
		isSharePageRoot.value = commonV2.isShareRootPage(newVal);
		menuVisible.value = false;
	},
	{
		immediate: true
	}
);

watch(
	() => filesStore.currentFileList[props.origin_id],
	() => {
		showLimit = 50;

		nextTick(() => {
			setItemWeight();
			fillWindow(true);
		});
	},
	{
		deep: true
	}
);

const selectFileCount = ref(0);
if (isPad.value) {
	watch(
		() => filesStore.selected[props.origin_id],
		async () => {
			selectFileCount.value = filesStore.selected[props.origin_id].length;
		},
		{
			deep: true
		}
	);
}
const handleMouseOver = () => {
	isHovered.value = true;
};

const handleMouseLeave = () => {
	isHovered.value = false;
};

onMounted(() => {
	// How much every listing item affects the window height
	setItemWeight();

	// Fill and fit the window with listing items
	fillWindow(true);

	// Add the needed event listeners to the window and document.
	if (pageRef.value) {
		pageRef.value.addEventListener('keydown', keyEvent);
		pageRef.value.addEventListener('scroll', scrollEvent);
		pageRef.value.addEventListener('resize', windowsResize);
		pageRef.value.addEventListener('click', handleClick);

		if (!store.user?.perm?.create) return;
		pageRef.value.addEventListener('dragover', preventDefault);
		pageRef.value.addEventListener('dragenter', dragEnter);
		pageRef.value.addEventListener('dragleave', dragLeave);
		pageRef.value.addEventListener('drop', drop);
	}
});

onUnmounted(() => {
	if (pageRef.value) {
		pageRef.value.removeEventListener('keydown', keyEvent);
		pageRef.value.removeEventListener('scroll', scrollEvent);
		pageRef.value.removeEventListener('resize', windowsResize);
		pageRef.value.removeEventListener('click', handleClick);

		if (store.user && !store.user?.perm?.create) return;
		pageRef.value.removeEventListener('dragover', preventDefault);
		pageRef.value.removeEventListener('dragenter', dragEnter);
		pageRef.value.removeEventListener('dragleave', dragLeave);
		pageRef.value.removeEventListener('drop', drop);
	}
});

const handleClick = () => {
	menuVisible.value = false;
};

const base64 = (name: string) => {
	return stringToBase64(name);
};

const keyEvent = (event: any) => {
	// No prompts are shown
	if (store.show !== null) {
		return;
	}

	// Esc!
	if (event.keyCode === 27) {
		// Reset files selection.
		filesStore.resetSelected(props.origin_id);
	}

	// Del!
	if (event.keyCode === 46) {
		if (
			!store.user.perm.delete ||
			filesStore.selectedCount(props.origin_id) == 0
		)
			return;

		// Show delete prompt.
		store.showHover('delete');
	}

	// F2!
	if (event.keyCode === 113) {
		if (
			!store.user.perm.rename ||
			filesStore.selectedCount(props.origin_id) !== 1
		)
			return;

		// Show rename prompt.
		store.showHover('rename');
	}

	// Ctrl is pressed
	if (!event.ctrlKey && !event.metaKey) {
		return;
	}

	let key = String.fromCharCode(event.which).toLowerCase();

	switch (key) {
		case 'f':
			event.preventDefault();
			store.showHover('search');
			break;
		case 'c':
			optionAction(event, OPERATE_ACTION.COPY);
			break;
		case 'x':
			optionAction(event, OPERATE_ACTION.CUT);
			break;
		case 'v':
			optionAction(event, OPERATE_ACTION.PASTE);
			break;
		case 'a':
			event.preventDefault();
			if (filesStore.currentFileList[props.origin_id])
				for (let file of filesStore.currentFileList[props.origin_id]!.items) {
					filesStore.addSelected(file.index, props.origin_id);
				}
			break;
		case 's':
			event.preventDefault();
			document.getElementById('download-button')?.click();
			break;
	}
};

const preventDefault = (event: any) => {
	// Wrapper around prevent default.
	event.preventDefault();
};

const optionAction = (event: any, type: OPERATE_ACTION) => {
	const path = route.path;
	if (!hideHeaderOpt(path)) {
		return;
	}

	const driveType = common().formatUrltoDriveType(path);

	driveType &&
		operateinStore.handleFileOperate(
			props.origin_id,
			event,
			route,
			type,
			driveType,
			async () => {}
		);
};

const scrollEvent = throttle(() => {
	const totalItems =
		(filesStore.currentDirItems(props.origin_id)?.length || 0) +
		(filesStore.currentFileItems(props.origin_id)?.length || 0);

	// All items are displayed
	if (showLimit >= totalItems) return;

	const currentPos = window.innerHeight + window.scrollY;

	// Trigger at the 75% of the window height
	const triggerPos = document.body.offsetHeight - window.innerHeight * 0.25;

	if (currentPos > triggerPos) {
		// Quantity of items needed to fill 2x of the window height
		const showQuantity = Math.ceil((window.innerHeight * 2) / itemWeight.value);

		// Increase the number of displayed items
		showLimit += showQuantity;
	}
}, 100);

const dragEnter = () => {
	if (isExternal.value) {
		return;
	}

	dragCounter.value++;

	// When the user starts dragging an item, put every
	// file on the listing with 50% opacity.
	let items: any = document.getElementsByClassName('item');

	Array.from(items).forEach((file: any) => {
		file.style.opacity = 0.5;
	});
};

const dragLeave = () => {
	dragCounter.value--;
	if (dragCounter.value == 0) {
		resetOpacity();
	}
};

const drop = async (e: any) => {
	e.preventDefault();

	if (isExternal.value) {
		return;
	}

	dragCounter.value = 0;
	resetOpacity();

	let dt = e.dataTransfer;

	if (dt.files.length <= 0) return;
	if (!transferStore.isUploadProgressDialogShow) {
		transferStore.isUploadProgressDialogShow = true;
	}

	if (getAppPlatform().isDesktop) {
		busEmit('electronSelectUploadFiles', dt.files);
		return;
	}

	let files = await scanFiles(dt);

	const dataTransfer = new DataTransfer();

	files.forEach((file) => {
		if (file instanceof File) {
			dataTransfer.items.add(file);
		}
	});

	const fileInput: any = document.getElementById('uploader-input');
	fileInput.value = '';
	fileInput.removeAttribute('webkitdirectory');
	fileInput.files = dataTransfer.files;
	const event = new Event('change', { bubbles: true });
	fileInput.dispatchEvent(event);
};

const resetOpacity = () => {
	let items = document.getElementsByClassName('item');

	Array.from(items).forEach((file: any) => {
		file.style.opacity = 1;
	});
};

const windowsResize = throttle(() => {
	width.value = window.innerWidth;

	// Listing element is not displayed
	if (listing.value == null) return;

	// How much every listing item affects the window height
	setItemWeight();

	// Fill but not fit the window
	fillWindow();
}, 100);

const setItemWeight = () => {
	// Listing element is not displayed
	if (listing.value == null) return;

	let itemQuantity =
		(filesStore.currentDirItems(props.origin_id)?.length || 0) +
		(filesStore.currentFileItems(props.origin_id)?.length || 0);
	if (itemQuantity > showLimit) itemQuantity = showLimit;

	// How much every listing item affects the window height
	itemWeight.value = listing.value.offsetHeight / itemQuantity;
};

const fillWindow = (fit = false) => {
	const totalItems =
		(filesStore.currentDirItems(props.origin_id)?.length || 0) +
		(filesStore.currentFileItems(props.origin_id)?.length || 0);

	// More items are displayed than the total
	if (showLimit >= totalItems && !fit) return;

	const windowHeight = window.innerHeight;

	// Quantity of items needed to fill 2x of the window height
	const showQuantity = Math.ceil(
		(windowHeight + windowHeight * 2) / itemWeight.value
	);

	// Less items to display than current
	if (showLimit > showQuantity && !fit) return;

	// Set the number of displayed items
	showLimit = showQuantity > totalItems ? totalItems : showQuantity;
};

const rightClick = (e: any, item?: any) => {
	if (!disabledClick(route.path)) {
		return false;
	}

	if (e.target.offsetParent!.classList[1] === 'header') {
		return false;
	}

	offsetRight.value = document.body.clientWidth - e.clientX;
	offsetBottom.value = document.body.clientHeight - e.clientY;

	e.preventDefault();

	if (
		process.env.APPLICATION !== 'FILES' &&
		process.env.APPLICATION !== 'SHARE'
	) {
		clientX.value = e.clientX - 170;
		clientY.value = e.clientY - 12;
	} else {
		clientX.value = e.clientX;
		clientY.value = e.clientY;
	}

	menuList.value = item;
	menuVisible.value = true;
	stopScrollMove(e);
	if (filesStore.selected[props.origin_id].length === 0) {
		item && filesStore.addSelected(item.index, props.origin_id);
		return;
	}
	if (item && filesStore.selected[props.origin_id].indexOf(item.index) === -1) {
		filesStore.resetSelected(props.origin_id);
		filesStore.addSelected(item.index, props.origin_id);
	}
};

const leftClick = (e: any) => {
	if (menuVisible.value) {
		menuVisible.value = false;
		startScrollMove(e);
	}
	filesStore.resetSelected(props.origin_id);
};

const changeVisible = (e: any) => {
	menuVisible.value = false;
	startScrollMove(e);
};

const filesUpdate = (event: any) => {
	// const targetRef = filterFiles(event);
	// const splitUrl = route.fullPath.split('?');
	// const fileSavePath = new FilePath({
	// 	path: splitUrl[0],
	// 	param: splitUrl[1] ? `?${splitUrl[1]}` : '',
	// 	isDir: true,
	// 	driveType: DriveType.Sync
	// });
	// filesStore.uploadSelectFile(targetRef, fileSavePath);
	// application.filesUploadConfig?.filesUpdate
	if (!application.filesUploadConfig?.filesUpdate) {
		return;
	}

	application.filesUploadConfig.filesUpdate(props.origin_id, event);
};

const filterFiles = (event) => {
	const files = event.target.files;

	if (
		Array.from(files).find(
			(file: any) =>
				file!.name.indexOf('\\') !== -1 || file!.name.indexOf('/') !== -1
		)
	) {
		BtNotify.show({
			type: NotifyDefinedType.WARNING,
			message: t('files.backslash_upload')
		});
	}

	const filteredFiles = Array.from(files).filter((file: any) => {
		return file!.name.indexOf('\\') === -1 && file!.name.indexOf('/') === -1;
	});

	const dataTransfer = new DataTransfer();

	filteredFiles.forEach((file: any) => dataTransfer.items.add(file));

	event.target.files = dataTransfer.files;

	return event;
};
</script>

<style scoped lang="scss">
.empty {
	img {
		width: 240px;
		height: 240px;
		margin-bottom: 20px;
	}
}

#listing {
	position: relative;
	&.hovered::-webkit-scrollbar-thumb {
		background-color: rgba(0, 0, 0, 0.2);
	}
}

.files-body {
	width: 100%;
	// height: calc(100vh - 40px);
	height: 100%;

	.connect-server-btn {
		float: right;
		padding: 8px 12px;
		margin-right: 20px;
	}
}

.iconClockwise {
	animation: rotate 0.2s linear forwards;
}

@keyframes rotate {
	from {
		transform: rotate(0deg);
	}
	to {
		transform: rotate(180deg);
	}
}

.iconAnticlockwise {
	animation: reverse-rotate 0.2s linear forwards;
}

@keyframes reverse-rotate {
	0% {
		transform: rotate(180deg);
	}
	100% {
		transform: rotate(0deg);
	}
}
</style>

<template>
	<div
		@click.stop="leftClick"
		:style="{ width: '100%', height: '100%' }"
		class="q-py-xs"
	>
		<!-- <listing-files-header v-if="!filesStore.isShard" /> -->

		<div
			id="listing"
			ref="listing"
			:class="store.user.viewMode + ' file-icons'"
		>
			<div class="listing-inter" style="padding: 0 20px">
				<template v-if="store.user.viewMode === 'list'">
					<terminus-select-all
						ref="terminusSelect"
						:lockTouch="lockTouch"
						:lockEvent="false"
						:items="filesStore.currentFileIds(origin_id) || []"
						@show-select-mode="showSelectMode"
					>
						<template v-slot="{ file }">
							<listing-item
								:key="file"
								:selectType="selectType"
								:from="from"
								:origin_id="origin_id"
								:item="filesStore.currentFileListMap(origin_id)[file.id]"
								v-bind:viewMode="store.user.viewMode"
								@open-sync-page="openSyncPage"
							/>
						</template>
					</terminus-select-all>
				</template>

				<template
					v-else-if="
						(filesStore.currentFileList[origin_id]?.items?.length || 0) > 0
					"
				>
					<ListingItem
						v-for="item in filesStore.currentFileList[origin_id]?.items"
						:key="base64(item.name)"
						:item="item"
						:selectType="selectType"
						:from="from"
						:origin_id="origin_id"
						v-bind:viewMode="store.user.viewMode"
						@open-sync-page="openSyncPage"
					>
					</ListingItem>
				</template>
				<file-transfer-no-data v-else style="width: calc(100vw - 40px)" />
			</div>
			<div style="height: 20px; width: 100%"></div>
		</div>
	</div>

	<files-uploader :origin_id="origin_id" />

	<prompts-component :origin_id="origin_id" />
</template>

<script lang="ts" setup>
import { ref, onMounted, onUnmounted, watch, nextTick, PropType } from 'vue';
import throttle from 'lodash.throttle';
import { stringToBase64 } from '@didvault/sdk/src/core';
import { useDataStore } from '../../../stores/data';
import { useFilesStore, FilesIdType, PickType } from '../../../stores/files';

import ListingItem from './ListingItem.vue';
import FilesUploader from '../../Files/common-files/FilesUploader.vue';
import PromptsComponent from '../../../components/files/prompts/PromptsComponent.vue';
import TerminusSelectAll from './../../../components/common/TerminusSelectAll.vue';
import FileTransferNoData from './../../../components/common/FileTransferNoData.vue';

const store = useDataStore();

const itemWeight = ref<number>(0);
const listing = ref<any>(null);
const filesStore = useFilesStore();

let showLimit = 50;

const terminusSelect = ref();
const selectedIds = ref();

const props = defineProps({
	lockTouch: {
		type: Boolean,
		required: false,
		default: false
	},
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	},
	from: {
		type: String,
		default: '',
		required: false
	},
	selectType: {
		type: String as PropType<PickType>,
		required: false,
		default: ''
	}
});

const emits = defineEmits(['showSelectMode', 'openSyncPage']);

const openSyncPage = () => {
	emits('openSyncPage');
};

const showSelectMode = (value) => {
	selectedIds.value = value;
	emits('showSelectMode', value);
};

const handleClose = () => {
	if (terminusSelect.value) {
		terminusSelect.value.handleClose();
	}
};

const handleSelectAll = () => {
	if (terminusSelect.value) {
		terminusSelect.value.toggleSelectAll();
	}
};

const intoCheckedMode = () => {
	if (terminusSelect.value) {
		terminusSelect.value.intoCheckedMode();
	}
};

const handleRemove = () => {
	console.log('handleRemove', selectedIds.value);
	if (terminusSelect.value) {
		terminusSelect.value.handleRemove();
	}
};

defineExpose({
	handleClose,
	handleSelectAll,
	handleRemove,
	intoCheckedMode
});

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

onMounted(() => {
	// How much every listing item affects the window height
	setItemWeight();

	// Fill and fit the window with listing items
	fillWindow(true);

	// Add the needed event listeners to the window and document.
	window.addEventListener('scroll', scrollEvent);

	if (!store.user?.perm?.create) return;
	document.addEventListener('dragover', preventDefault);
});

onUnmounted(() => {
	// Remove event listeners before destroying this page.
	window.removeEventListener('scroll', scrollEvent);

	if (store.user && !store.user?.perm?.create) return;
	document.removeEventListener('dragover', preventDefault);
});

const base64 = (name: string) => {
	return stringToBase64(name);
};

const preventDefault = (event: any) => {
	// Wrapper around prevent default.
	event.preventDefault();
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

const leftClick = () => {
	filesStore.resetSelected();
};
</script>

<style scoped lang="scss">
#listing {
	width: 100%;
	height: 100%;
	position: relative;
	overflow-y: scroll;
	overflow-x: hidden;
	&::-webkit-scrollbar {
		width: 0;
		height: 0;
	}
}
</style>

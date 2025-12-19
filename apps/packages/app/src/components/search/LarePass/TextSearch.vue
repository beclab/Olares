<template>
	<div class="searchCard">
		<img
			class="cursor-pointer"
			src="app-icon/text-back.svg"
			alt="delete"
			@click="goBack"
			style="width: 20px; height: 20px"
		/>
		<div class="appIcon">
			<img class="icon" :src="item?.icon" alt="search" />
			<span class="q-ml-sm text-ink-2 text-body3">{{ item?.title }}</span>
		</div>
		<input
			class="input text-ink-1"
			type="text"
			ref="filesRef"
			v-model.trim="searchFiles"
			:placeholder="t('search_file_placeholder')"
		/>
		<q-icon
			v-if="searchFiles"
			class="cursor-pointer"
			name="sym_r_close"
			size="24px"
			style="color: #adadad"
			@click="handleClear"
		/>
	</div>
	<div class="search-files bg-background-2">
		<div
			class="fileDetails"
			ref="scrollAreaBG"
			v-if="fileData && fileData.length > 0"
		>
			<bt-scroll-area
				class="full-width full-height q-pa-md my-scroll-area-wrapper"
				ref="scrollArea"
			>
				<q-infinite-scroll
					@load="getMore"
					:offset="20"
					:disable="!hasMore"
					v-if="props.item?.title === 'Drive' && seachOSVersionLargeThan12_3()"
				>
					<SearchItemsComponent
						:item="item"
						:activeItem="activeItem"
						:fileData="fileData"
						@open="open"
						@clickItem="clickItem"
					/>
					<template v-slot:loading>
						<div class="row justify-center q-my-md">
							<q-spinner-dots color="primary" size="40px" />
						</div>
					</template>
				</q-infinite-scroll>
				<SearchItemsComponent
					v-else
					:item="item"
					:activeItem="activeItem"
					:fileData="fileData"
					@open="open"
					@clickItem="clickItem"
				/>
			</bt-scroll-area>
		</div>
		<div v-else class="no_data column items-center justify-center">
			<img
				class="icon"
				src="../../../assets/desktop/nodata.svg"
				alt="no_data"
			/>
			<div class="text-grey-8">{{ t('file_no_data') }}</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';

import { useRouter } from 'vue-router';
import { useFilesStore } from '../../../stores/files';
import { DriveType } from '../../../utils/interface/files';
import { TextSearchItem } from 'src/utils/interface/search';
import { LayoutMenu } from 'src/utils/constants';
import { useMenuStore } from 'src/stores/menu';
import { dataAPIs } from 'src/api/files/v2';
import { useAppAbilitiesStore } from 'src/stores/appAbilities';

import { seachOSVersionLargeThan12_3 } from 'src/api/common/search';
import SearchItemsComponent from '../SearchItemsComponent.vue';
import { useSearch } from '../search';
import { getApplication } from 'src/application/base';

const props = defineProps({
	showSearchDialog: {
		type: Boolean,
		require: false,
		default: false
	},
	handSearchFiles: {
		type: String,
		require: false
	},
	item: {
		type: Object,
		require: false
	},
	commandList: {
		type: Object,
		require: false
	}
});

const emits = defineEmits(['goBack', 'hideSearchDialog']);

const filesStore = useFilesStore();

const menuStore = useMenuStore();
const router = useRouter();

const open = async (item: TextSearchItem) => {
	if (props.item?.title === 'Drive' || props.item?.title === 'Sync') {
		const curLayoutMenu = LayoutMenu[0];
		const dataAPI = dataAPIs(item.driveType);
		await dataAPI.transferItemBackToFiles(item as any, router);
		menuStore.pushTerminusMenuCache(curLayoutMenu.identify);
		filesStore.resetSelected();
		filesStore.setBrowserUrl(item.path, item.driveType);
	} else if (props.item?.title === 'Wise') {
		const appAbilitiesStore = useAppAbilitiesStore();
		const baseurl = appAbilitiesStore.getAppDomain('wise');
		const url = baseurl + '/' + item.meta?.file_type + '/' + item.meta?.id;
		const openUrl = url.startsWith('https') ? url : 'https://' + url;
		let enterUrl = new URL(openUrl);
		enterUrl.searchParams.append('preview_name', item.title);
		getApplication().openUrl(enterUrl.toString());
	}
	setTimeout(() => {
		emits('hideSearchDialog');
	}, 300);
};

const clickTimer: any = ref(null);
const clickDelay = ref(300);

const clickItem = (file: any, index: number) => {
	if (clickTimer.value) {
		clearTimeout(clickTimer.value);
		clickTimer.value = null;
		chooseFile(index);
		open(file);
	} else {
		clickTimer.value = setTimeout(() => {
			if (activeItem.value === file.id) {
				open(file);
			} else {
				chooseFile(index);
			}
			clickTimer.value = null;
		}, clickDelay.value);
	}
};

const {
	fileData,
	t,
	searchFiles,
	filesRef,
	activeItem,
	scrollArea,
	scrollAreaBG,
	hasMore,
	chooseFile,
	goBack,
	handleClear,
	getMore
} = useSearch(
	props.item?.serviceType as any,
	open,
	emits,
	props.handSearchFiles
);
</script>

<style>
hi {
	color: #3377ff !important;
}
</style>

<style lang="scss" scoped>
.search-files {
	height: calc(100% - 46px);
	border-top: 1px solid $separator-color;
	overflow: hidden;
	display: flex;

	.fileDetails {
		width: 100%;
	}

	.no_data {
		width: 100%;
	}
}
.my-scroll-area-wrapper {
	min-width: 400px;
	&.my-scroll-area-no-header {
		height: 100vh;
		margin-top: 0px;
	}
	background: $background-6;

	& > ::v-deep(.q-scrollarea__container) {
		min-width: 0px;
		overflow-x: hidden;
	}

	& > ::v-deep(.q-scrollarea__container > .q-scrollarea__content) {
		min-width: 0px;
		width: 100% !important;
	}
}
</style>

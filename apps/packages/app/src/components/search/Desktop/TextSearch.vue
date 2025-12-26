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
					v-if="props.item?.title === 'Drive'"
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

import { TextSearchItem } from 'src/utils/interface/search';
import SearchItemsComponent from '../../../components/search/SearchItemsComponent.vue';

import { useSearch } from '../search';

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

const open = (item: TextSearchItem) => {
	const commandList = JSON.parse(JSON.stringify(props.commandList));
	if (props.item?.title === 'Drive' || props.item?.title === 'Sync') {
		const filesApp = commandList.find(
			(el: { name: string }) => el.name && el.name === 'files'
		);
		filesApp.url = filesApp.url + item.path;

		const openUrl = filesApp.url.startsWith('https')
			? filesApp.url
			: 'https://' + filesApp.url;
		window.open(openUrl);
	} else if (props.item?.title === 'Wise') {
		const filesApp = commandList.find(
			(el: { name: string; fatherName: string }) =>
				(el.name && el.name === 'wise') ||
				(el.fatherName && el.fatherName === 'wise')
		);

		filesApp.url =
			filesApp.url + '/' + item.meta?.file_type + '/' + item.meta?.id;

		const openUrl = filesApp.url.startsWith('https')
			? filesApp.url
			: 'https://' + filesApp.url;

		let enterUrl = new URL(openUrl);
		enterUrl.searchParams.append('preview_name', item.title);

		window.open(enterUrl);
	}
};

const emits = defineEmits(['goBack']);

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

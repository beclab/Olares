<template>
	<VueDragResize
		:isActive="true"
		:isResizable="false"
		:w="800"
		:h="556"
		:x="screenWidth / 2 - 400"
		:y="screenHeight / 2 - 250"
		dragHandle=".drag"
		:parentLimitation="true"
	>
		<q-card
			ref="dialogRef"
			@update:model-value="UpdateShowSearchDialog"
			class="dialog_card"
		>
			<div
				class="drag"
				style="height: 16px; width: 100%; cursor: pointer"
			></div>

			<div class="content">
				<home-component
					v-if="searchType === SearchType.HomePage"
					:commandList="commandList"
					@openCommand="openCommand"
				/>

				<text-search
					v-else-if="searchType === SearchType.TextSearch"
					:item="filesItem"
					:handSearchFiles="handSearchFiles"
					:commandList="commandList"
					@goBack="goBack"
				/>

				<files-component
					v-else-if="searchType === SearchType.FilesPage"
					:handSearchFiles="handSearchFiles"
					@handleAction="handleAction"
					@goBack="goBack"
				/>

				<terminus-component
					v-else-if="
						searchType === SearchType.AshiaPage ||
						searchType === SearchType.AshiaDocPage
					"
					:filesItem="filesItem"
					:isDoc="searchType === SearchType.AshiaDocPage"
					@openCommand="openCommand"
				/>

				<empty-page v-else @openCommand="openCommand" />
			</div>

			<footer-component />
		</q-card>
	</VueDragResize>
</template>

<script lang="ts" setup>
import { ref, onMounted, onBeforeUnmount } from 'vue';
import { useDialogPluginComponent } from 'quasar';
import { AppClickInfo, SearchType } from '../../../utils/interface/search';
import FilesComponent from '../FilesComponent.vue';
import TextSearch from './TextSearch.vue';
import HomeComponent from '../HomeComponent.vue';
import TerminusComponent from '../TerminusComponent.vue';
import FooterComponent from '../FooterComponent.vue';
import EmptyPage from '../EmptyPage.vue';
import { useSearchStore } from 'src/stores/search';
import { searchCancel } from 'src/api/common/search';

const emits = defineEmits([
	...useDialogPluginComponent.emits,
	'appClick',
	'hide'
]);

const searchType = ref(SearchType.HomePage);
const commandList = ref();
const filesItem = ref();
const handSearchFiles = ref();
const screenWidth = ref(document.body.clientWidth);
const screenHeight = ref(document.body.clientHeight);

const searchStore = useSearchStore();

const UpdateShowSearchDialog = () => {
	emits('hide', false);
};

const openCommand = (item?: any) => {
	if (!item) {
		searchType.value = SearchType.HomePage;
		return false;
	}

	if (item && item.type === 'Command') {
		searchType.value = SearchType.TextSearch;
		filesItem.value = item;
		handSearchFiles.value = item.searchFiles;
	} else {
		openWindow(item);
	}
};

const openWindow = async (item: any) => {
	emits('appClick', {
		appid: item.id,
		data: {}
	} as AppClickInfo);
};

const handleAction = (item: any) => {
	searchType.value = SearchType.AshiaPage;
	filesItem.value = item;
};

const goBack = () => {
	openCommand();
};

const keydownEnter = (event: any) => {
	if (event.keyCode === 27) {
		if (searchType.value === SearchType.HomePage) {
			UpdateShowSearchDialog();
		} else {
			searchType.value = SearchType.HomePage;
		}
	}
};

onMounted(() => {
	commandList.value = searchStore.getCommand();
	window.addEventListener('keydown', keydownEnter);
});

onBeforeUnmount(() => {
	searchStore.initChat();
	window.removeEventListener('keydown', keydownEnter);
	searchStore.cancelSearch();
});
</script>

<style lang="scss">
.dialog_card {
	width: 100%;
	height: 556px;
	border-radius: 12px;

	.content {
		height: 492px;
	}

	.searchCard {
		width: calc(100% - 32px);
		height: 40px;
		line-height: 40px;
		border-radius: 12px !important;
		margin: 0 auto 8px;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 16px;
		position: relative;

		.icon {
			width: 18px;
			height: 18px;
		}

		.btn {
			padding: 8px 12px;
			font-weight: 400;
			font-size: 14px;
			line-height: 14px;
			color: #1a130f;
			background: rgba(26, 19, 15, 0.06);
			border-radius: 4px;
			cursor: pointer;

			&:hover {
				background: rgba(26, 19, 15, 0.1);
			}
		}

		.input {
			flex: 1;
			height: 100%;
			border: none;
			outline: none;
			margin: 0 12px;
			font-weight: 400;
			font-size: 16px;
			line-height: 20px;
			color: #857c77;
			background: rgba(255, 255, 255, 0);

			&::placeholder {
				color: #bdbdbd;
			}
		}

		.appIcon {
			height: 32px;
			display: flex;
			align-items: center;
			justify-content: space-around;
			background: $background-4;
			border-radius: 4px;
			padding: 8px 12px;
			margin-left: 18px;
		}
	}
}
</style>

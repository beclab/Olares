<template>
	<q-card
		ref="dialogRef"
		@update:model-value="UpdateShowSearchDialog"
		class="dialog_card"
	>
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
				@hideSearchDialog="UpdateShowSearchDialog"
			/>
		</div>
		<footer-component />
	</q-card>
</template>

<script lang="ts" setup>
import { ref, onMounted, onBeforeUnmount } from 'vue';
import { useDialogPluginComponent } from 'quasar';
import TextSearch from './TextSearch.vue';
import HomeComponent from '../HomeComponent.vue';
import FooterComponent from '../FooterComponent.vue';
import { useSearchStore } from '../../../stores/search';
import { SearchType } from '../../../utils/interface/search';
import { getApplication } from 'src/application/base';

const emits = defineEmits([...useDialogPluginComponent.emits, 'hide']);

const searchType = ref(SearchType.HomePage);
const commandList = ref();
const filesItem = ref();
const handSearchFiles = ref();

const searchStore = useSearchStore();

const UpdateShowSearchDialog = () => {
	emits('hide', false);
};

const openCommand = async (item?: any) => {
	if (!item) {
		searchType.value = SearchType.HomePage;
		return false;
	}

	if (item && item.type === 'Command') {
		searchType.value = SearchType.TextSearch;
		filesItem.value = item;
		handSearchFiles.value = item.searchFiles;
	} else {
		getApplication().openUrl(
			item.url.startsWith('http') ? item.url : 'https://' + item.url
		);
		UpdateShowSearchDialog();
	}
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
	window.removeEventListener('keydown', keydownEnter);
	searchStore.cancelSearch();
});
</script>

<style lang="scss">
.dialog_card {
	width: 800px;
	height: 556px;
	border-radius: 12px !important;
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	margin: auto;
	z-index: 1000;
	background-color: $background-1;

	.content {
		height: 492px;
	}

	.searchCard {
		width: calc(100% - 32px);
		height: 40px;
		line-height: 40px;
		border-radius: 12px !important;
		margin: 8px auto 8px;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 0 16px;

		.icon {
			width: 24px;
			height: 24px;
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
			margin-right: 16px;
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
			height: 42px;
			display: flex;
			align-items: center;
			justify-content: space-around;
			background: rgba(26, 19, 15, 0.06);
			border-radius: 4px;
			padding: 8px 12px;
		}
	}
}
</style>

<template>
	<div class="container">
		<header>
			<files-header :origin_id="origin_id" />
		</header>

		<main>
			<errors v-if="error" :errorCode="error.status" />
			<listing-files :origin_id="origin_id" v-else />
		</main>
		<footer
			v-if="
				isPad &&
				!store.showPadPopup &&
				(filesStore.selectedCount(origin_id) > 0 ||
					operateinStore.copyFiles.length > 0)
			"
		>
			<BottomOperateMenu
				:origin_id="origin_id"
				:menuList="filesStore.selected[origin_id]"
			/>
		</footer>

		<footer>
			<files-footer :origin_id="origin_id" />
		</footer>
	</div>
</template>

<script lang="ts" setup>
import { ref, onMounted, onUnmounted, watch } from 'vue';
import { useDataStore } from '../../stores/data';
import { useFilesStore, FilesIdType } from '../../stores/files';
import { useRoute, onBeforeRouteUpdate } from 'vue-router';
import { UserStatusActive } from '../../utils/checkTerminusState';
import { useTermipassStore } from '../../stores/termipass';
import FilesHeader from './common-files/FilesHeader.vue';
import FilesFooter from './common-files/FilesFooter.vue';
import ListingFiles from './common-files/ListingFiles.vue';
import Errors from './Errors.vue';
import BottomOperateMenu from '../../components/files/files/BottomOperateMenu.vue';
import { getAppPlatform } from '../../application/platform';
import { useOperateinStore } from '../../stores/operation';
import { common } from './../../api';

defineProps({
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	}
});

const store = useDataStore();
const route = useRoute();
const error = ref<any>(null);
const termipassStore = useTermipassStore();
const filesStore = useFilesStore();
const isPad = ref(getAppPlatform() && getAppPlatform().isPad);
const operateinStore = useOperateinStore();

watch(
	() => termipassStore.totalStatus?.isError,
	(newVal) => {
		checkUserStatus(newVal);
	}
);

const checkUserStatus = (status: any) => {
	if (status === UserStatusActive.error) {
		error.value = status;
		return false;
	}
	if (error.value != null && error.value != undefined) {
		setTimeout(
			async () => {
				let url = route.fullPath;
				const driveType = common().formatUrltoDriveType(url);
				if (driveType == undefined) {
					return;
				}
				error.value = undefined;
				const splitUrl = url.split('?');
				await filesStore.setFilePath(
					{
						path: splitUrl[0],
						isDir: true,
						driveType: driveType,
						param: splitUrl[1] ? `?${splitUrl[1]}` : ''
					},
					false,
					false
				);
			},
			termipassStore.isLocal ? 5000 : 2000
		);
	}
	return true;
};

const keyEvent = (event: any) => {
	if (event.keyCode === 112) {
		event.preventDefault();
		store.showHover('help');
	}
};

const isPreview = ref(false);

onBeforeRouteUpdate((_to, from, next) => {
	if (from.query.type === 'preview') {
		isPreview.value = true;
	} else {
		isPreview.value = false;
	}
	next();
});

onMounted(async () => {
	window.addEventListener('keydown', keyEvent);
});

onUnmounted(() => {
	window.removeEventListener('keydown', keyEvent);
	if (store.showShell) {
		store.toggleShell();
	}
});
</script>

<style lang="scss" scoped>
.container {
	width: 100%;
	height: 100%;
	// position: absolute;
	// left: 0;
	// top: 0;
	display: flex;
	flex-direction: column;
	background: $background-1;
	overflow: hidden;

	header,
	footer {
		flex: 0 0 auto;
	}

	main {
		flex: 1 1 auto;
		overflow-y: auto;
		// border-left: solid 1px $separator;
	}
}
</style>

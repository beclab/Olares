<template>
	<q-layout view="lhr lpr lfr" class="layout layout-app">
		<div class="mainlayout">
			<q-page-container style="width: 100%; height: 100%">
				<div
					v-if="isRouterView"
					:class="
						tabbarShow ? 'container-show-tabbar' : 'container-hide-tabbar'
					"
				>
					<router-view />
				</div>
				<div
					v-else
					:class="
						tabbarShow ? 'container-show-tabbar' : 'container-hide-tabbar'
					"
				>
					<FilesPage />
				</div>
				<tabbar-component
					v-if="tabbarShow"
					class="files-mobile-tabbar"
					:current="currentTabIndex"
					:items="tabItems"
				/>
			</q-page-container>
		</div>
	</q-layout>
</template>

<script setup lang="ts">
import { computed, onMounted, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import FilesPage from 'src/pages/Mobile/file/FilesPage.vue';
import TabbarComponent from 'src/components/common/TerminusTabbarComponent.vue';
import { useFilesStore, FilesIdType } from './../../stores/files';
import {
	initFilesLayoutFromRoute,
	isFilesMobileShellRoutePath,
	isRouteSyncedToFilesStore
} from './initFilesLayoutFromRoute';
import { common } from 'src/api';
import * as filesUtil from 'src/api/files/v2/common/utils';
import { useTransfer2Store } from 'src/stores/transfer2';
import { TabbarItem } from 'src/application/interface';
import { TransferFront, TransferStatus } from 'src/utils/interface/transfer';

const route = useRoute();
const router = useRouter();

const filesStore = useFilesStore();
const transferStore = useTransfer2Store();
// Align with LayoutPc: init PAGEID state before child components mount.
filesStore.initIdState(FilesIdType.PAGEID);

const isRouterView = computed(() => isFilesMobileShellRoutePath(route.path));
const tabbarShow = computed(() => !filesStore.isShard);
const currentTabIndex = computed(() =>
	route.path.startsWith('/transfer') ? 1 : 0
);
const transferBadge = computed(() => {
	const visibleFronts = [
		TransferFront.upload,
		TransferFront.download,
		TransferFront.copy,
		TransferFront.move,
		TransferFront.compress,
		TransferFront.extract
	];
	const total = transferStore.transfers.filter((id) => {
		const item = transferStore.transferMap[id];
		return (
			item &&
			visibleFronts.includes(item.front) &&
			item.status !== TransferStatus.Completed &&
			item.status !== TransferStatus.Canceled
		);
	}).length;
	if (total <= 0) {
		return '';
	}
	return total > 99 ? '99+' : String(total);
});

const handleFileTabChanged = () => {
	const lastStack =
		filesStore.backStack[FilesIdType.PAGEID]?.[
			filesStore.backStack[FilesIdType.PAGEID].length - 1
		];
	if (lastStack) {
		const params = new URLSearchParams(lastStack.param);
		router.replace({
			path: lastStack.path,
			query: Object.fromEntries(params)
		});
		return true;
	}
	if (filesStore.mobileRepo) {
		router.replace(filesStore.mobileRepo);
		return true;
	}
	return false;
};

const tabItems = computed<TabbarItem[]>(() => [
	{
		name: 'files.files',
		identify: 'file',
		normalImage: 'tab_files_normal',
		activeImage: 'tab_files_active',
		darkActiveImage: 'tab_files_dark_active',
		to: '/Files/',
		tabChanged: handleFileTabChanged
	},
	{
		name: 'transmission.title',
		identify: 'transfer',
		normalImage: 'tab_transfer_normal',
		activeImage: 'tab_transfer_active',
		to: '/transfer/history',
		badge: transferBadge.value
	}
]);

// Keep the embedded <FilesPage /> in sync when the route changes via system
// back/forward (popstate). In-app navigation updates the store before the
// route, so it is already in sync and skipped here; only route-driven changes
// (where the store still points at the old location) trigger a re-sync, and we
// avoid pushing a new back-stack entry or re-triggering router navigation.
watch(
	() => route.fullPath,
	(fullPath) => {
		if (isFilesMobileShellRoutePath(route.path)) {
			return;
		}
		if (!common().formatUrltoDriveType(route.path)) {
			return;
		}
		if (isRouteSyncedToFilesStore(fullPath, filesStore)) {
			return;
		}
		void initFilesLayoutFromRoute(route, filesStore, {
			fromHistory: true
		});
	}
);

onMounted(() => {
	filesUtil.fetchNodeList();
	if (!isRouterView.value) {
		void initFilesLayoutFromRoute(route, filesStore);
	}
});
</script>

<style scoped lang="scss">
.mainlayout {
	height: 100%;
	width: 100%;
}

.container-hide-tabbar {
	height: 100dvh;
	width: 100%;
}

.container-show-tabbar {
	height: calc(100dvh - 65px);
	width: 100%;
}

.files-mobile-tabbar {
	position: absolute;
	bottom: 1px;
	width: 100%;
}
</style>

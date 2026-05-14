<template>
	<q-layout view="lhr lpr lfr" class="layout layout-app">
		<div class="mainlayout">
			<q-page-container style="width: 100%; height: 100%">
				<div class="container-hide-tabbar" v-if="isRouterView">
					<router-view />
				</div>
				<div class="container-hide-tabbar" v-else>
					<FilesPage />
				</div>
			</q-page-container>
		</div>
	</q-layout>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import FilesPage from 'src/pages/Mobile/file/FilesPage.vue';
import { useFilesStore, FilesIdType } from './../../stores/files';
import {
	initFilesLayoutFromRoute,
	isFilesMobileShellRoutePath
} from './initFilesLayoutFromRoute';

const route = useRoute();

const filesStore = useFilesStore();
// Align with LayoutPc: init PAGEID state before child components mount.
filesStore.initIdState(FilesIdType.PAGEID);

const isRouterView = computed(() => isFilesMobileShellRoutePath(route.path));

onMounted(() => {
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
	height: 100vh;
	width: 100%;
}
</style>

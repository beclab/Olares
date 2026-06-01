<template>
	<q-layout
		view="lHh LpR lFr"
		:container="application == 'FILES' ? false : true"
		:class="application == 'FILES' ? '' : ''"
	>
		<FilesDrawer :origin_id="origin_id" />

		<q-page-container>
			<q-page
				class="files-content"
				:class="
					($q.platform.is.win || $q.platform.is.linux) &&
					$q.platform.is.electron
						? 'files-content-win'
						: $q.platform.is.ipad
						? 'files-content-pad'
						: $q.platform.is.android
						? 'files-content-android-pad'
						: 'files-content-common'
				"
			>
				<FilesPage :origin_id="origin_id" />
			</q-page>
		</q-page-container>

		<prompts-component :origin_id="origin_id" />
	</q-layout>
</template>

<script setup lang="ts">
import { nextTick, onMounted, ref } from 'vue';
import { useQuasar } from 'quasar';
import { useRoute } from 'vue-router';
import { bytetrade } from '@bytetrade/core';

import { useFilesStore, FilesIdType } from './../../stores/files';
import FilesPage from '../../pages/Files/FilesPage.vue';

import PromptsComponent from '../../components/files/prompts/PromptsComponent.vue';
import FilesDrawer from '../TermipassLayout/FilesDrawer.vue';
import { initFilesLayoutFromRoute } from './initFilesLayoutFromRoute';

const $q = useQuasar();
const route = useRoute();
const filesStore = useFilesStore();
// const socketStore = useVaultSocketStore();

const application = ref(process.env.APPLICATION);
const origin_id = ref(FilesIdType.PAGEID);
filesStore.initIdState(origin_id.value);

// socketStore.restart();

onMounted(async () => {
	void initFilesLayoutFromRoute(route, filesStore);

	nextTick(() => {
		bytetrade.observeUrlChange.childPostMessage({
			type: 'Files'
		});
	});
});
</script>

<style lang="scss">
.files-content {
	width: 100%;
}

.files-content-common {
	height: calc(100vh - 73px) !important;
}

.files-content-android-pad {
	height: calc(100vh - 88px) !important;
}
.files-content-ipad {
	height: calc(100vh - 116px) !important;
}
.files-content-win {
	height: calc(100vh - 145px) !important;
}
</style>

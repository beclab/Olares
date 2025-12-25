<template>
	<PageCard :title="$t('collect')">
		<template #extra>
			<CookieContent />
		</template>
		<div v-if="appAbilitiesStore.wise.running">
			<PageContent />
			<div class="q-mt-md q-pt-lg">
				<div class="text-h6 text-ink-1">{{ $t('buttons.download') }}</div>
				<div class="q-mt-sm">
					<DownloadContent />
				</div>
			</div>
		</div>
		<EmptyUninstallWise v-else></EmptyUninstallWise>
	</PageCard>
</template>

<script setup lang="ts">
import PageContent from './PageContent.vue';
import DownloadContent from './DownloadContent.vue';
import CookieContent from './CookieContent2.vue';
import { useCollectStore } from '../../../stores/collect';
import { onMounted, watch } from 'vue';
import {
	bexFrontBusOff,
	bexFrontBusOn
} from 'src/platform/interface/bex/utils';
import { onUnmounted } from 'vue';
import { queryDownloadFile } from '../../../api/wise/download';
import axios, { CancelTokenSource } from 'axios';
import PageCard from 'src/pages/Plugin/components/PageCard.vue';
import { getCurrentTabInfo } from 'src/utils/bex/tabs';
import EmptyUninstallWise from 'src/pages/Plugin/components/EmptyUninstallWise.vue';
import { useAppAbilitiesStore } from 'src/stores/appAbilities';
const appAbilitiesStore = useAppAbilitiesStore();

const CancelToken = axios.CancelToken;
let pageListSource: undefined | CancelTokenSource = undefined;

const collectStore = useCollectStore();
async function getInfos() {
	collectStore.setFileList([]);
	setData();
}

async function setData() {
	const tab = await getCurrentTabInfo();
	setDownload(tab);
}

async function setDownload(tab) {
	pageListSource && pageListSource.cancel();
	pageListSource = CancelToken.source();
	const info = await queryDownloadFile(tab.url, pageListSource.token);
	const downloadData = info ? [info] : [];
	collectStore.setFileList(downloadData);
}

watch(
	() => appAbilitiesStore.wise.running,
	(newValue) => {
		if (newValue) {
			getInfos();
		}
	}
);

onMounted(() => {
	getInfos();
	bexFrontBusOn('COLLECTION_TAB_UPDATE', getInfos);
});

onUnmounted(() => {
	bexFrontBusOff('COLLECTION_TAB_UPDATE', getInfos);
});
</script>

<style scoped lang="scss">
.collect-root {
	width: 100%;
	height: 100%;
	// background-color: #fff;

	// &__header {
	// 	width: 100%;
	// 	height: 56px;
	// }

	&__slider {
		width: 100%;
		height: 44px;
		border-radius: 32px;
		padding: 4px 5px;
		background-color: $background-3;
		.slider-item {
			height: 100%;
		}

		.slider-item-select {
			background-color: $background-1;
			border-radius: 18px;
		}
	}

	.tab-common {
		padding: 0;
	}
}
</style>

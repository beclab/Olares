<template>
	<PageCard :title="$t('bex.subscribe')" :loading="loading">
		<template #extra>
			<RouteToWiseFeed />
		</template>
		<div v-if="appAbilitiesStore.wise.running">
			<RssContent />
			<EmptyData
				v-if="collectStore.rssList.length <= 0"
				:title="$t('no_data')"
				class="absolute-center"
				@click="setData"
			></EmptyData>
		</div>
		<EmptyUninstallWise v-else></EmptyUninstallWise>
	</PageCard>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import RssContent from 'src/pages/Mobile/collect/RssContent.vue';
import { useCollectStore } from '../../../stores/collect';
import { onMounted } from 'vue';
import { onUnmounted } from 'vue';
import { searchFeed } from '../../../api/wise/feed';
import { RssStatus } from 'src/pages/Mobile/collect/utils';
import PageCard from 'src/pages/Plugin/components/PageCard.vue';
import { getTabUrl } from 'src/utils/bex/tabs';
import EmptyData from 'src/pages/Plugin/components/EmptyData.vue';
import RouteToWiseFeed from 'src/pages/Plugin/containers/RouteToWiseFeed.vue';
import EmptyUninstallWise from 'src/pages/Plugin/components/EmptyUninstallWise.vue';
import { ROUTE_CONST } from 'src/router/route-const';
import { useAppAbilitiesStore } from 'src/stores/appAbilities';
import { browser } from 'src/platform/interface/bex/browser/target';

const collectStore = useCollectStore();
const appAbilitiesStore = useAppAbilitiesStore();
const loading = ref(false);

async function getInfos() {
	collectStore.setRssList([]);
	setData();
}

async function setData() {
	const debugUrl =
		process.env.DEV_PLATFORM_BEX && process.env.RSS_DEBUG_URL
			? process.env.RSS_DEBUG_URL
			: undefined;
	const url = await getTabUrl(debugUrl);
	getFeed(url);
}

async function getFeed(url: string) {
	if (collectStore.rssList.length > 0) {
		return;
	}
	loading.value = true;
	try {
		const data = await searchFeed(url);
		if (data && data?.length > 0) {
			const list = data.map((item) => ({
				status: item.is_subscribed ? RssStatus.added : RssStatus.none,
				title: item.title,
				url: item.feed_url,
				image: item.icon_content,
				feed: item
			}));

			collectStore.setRssList(list);
		}
	} catch (error) {
		console.error(error);
	}
	loading.value = false;
}

watch(
	() => appAbilitiesStore.wise.running,
	(newValue) => {
		if (newValue) {
			setData();
		}
	}
);

onMounted(() => {
	setData();
	browser.tabs.onActivated.addListener(getInfos);
	browser.tabs.onUpdated.addListener(getInfos);
});

onUnmounted(() => {
	browser.tabs.onActivated.removeListener(getInfos);
	browser.tabs.onUpdated.removeListener(getInfos);
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

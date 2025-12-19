<template>
	<div class="column flex-gap-y-lg">
		<CookieMessage />
		<SiteCardContainer
			:title="$t('collectToLibrary')"
			:info="$t('createLibraryRecordAndAutoDownloadAttachment')"
			v-if="entry.file_type || collectSiteStore.data.is_entry_available"
		>
			<div class="column no-wrap flex-gap-y-md">
				<app-message
					v-if="collectSiteStore.data.is_entry_available"
					:message="collectSiteStore.data.is_entry_available"
				/>
				<CollectSiteCard v-if="entry.file_type" :data="entry" />
			</div>
		</SiteCardContainer>
		<div class="column flex-gap-y-lg full-width">
			<SiteCardContainer
				:title="$t('downloadFile')"
				:info="$t('directDownloadFilesToOlaresWithoutLibraryRecords')"
				v-if="
					downloadList.length > 0 || collectSiteStore.data.is_download_available
				"
			>
				<div class="column no-wrap flex-gap-y-md">
					<app-message
						v-if="collectSiteStore.data.is_download_available"
						:message="collectSiteStore.data.is_download_available"
					/>
					<template v-if="downloadList.length > 0">
						<DownloadSiteCard
							v-for="(item, index) in downloadList"
							:key="index"
							:data="item"
						/>
					</template>
				</div>
			</SiteCardContainer>
			<SiteCardContainer
				:title="$t('subscribeFeed')"
				:info="$t('subscribeRssFeedAndAutoDownloadAttachment')"
				v-if="feedList.length > 0 || collectSiteStore.data.is_feed_available"
			>
				<div class="column no-wrap flex-gap-y-md">
					<app-message
						v-if="collectSiteStore.data.is_feed_available"
						:message="collectSiteStore.data.is_feed_available"
					/>
					<template v-if="feedList.length > 0">
						<div v-for="(item, index) in feedList" :key="index">
							<FeedSiteCard :feed="item" />
						</div>
					</template>
				</div>
			</SiteCardContainer>
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import CollectSiteCard from 'src/containers/collection/CollectSiteCard.vue';
import FeedSiteCard from 'src/containers/collection/FeedSiteCard.vue';
import DownloadSiteCard from 'src/containers/collection/DownloadSiteCard.vue';
import { CollectInfo } from 'src/types/commonApi';
import SiteCardContainer from 'src/components/collection/SiteCardContainer.vue';
import CookieMessage from 'src/containers/collection/CookieMessage.vue';
import { useCollectSiteStore } from 'src/stores/collect-site';
import { COOKIE_LEVEL } from 'src/utils/rss-types';
import AppMessage from 'src/containers/collection/AppMessage.vue';

const data = ref<CollectInfo>();

const collectSiteStore = useCollectSiteStore();
collectSiteStore.init();

const uploadCookieButtonShow = computed(
	() =>
		!collectSiteStore.cookie.cookieExist &&
		collectSiteStore.cookie.cookieRequire === COOKIE_LEVEL.REQUIRED
);

const entry = computed(() => ({
	...collectSiteStore.entry,
	disabled: uploadCookieButtonShow.value
}));

const feedList = computed(() =>
	collectSiteStore.feed.map((item) => ({
		...item,
		disabled: uploadCookieButtonShow.value
	}))
);

const downloadList = computed(() =>
	collectSiteStore.download.map((item) => ({
		...item,
		disabled: uploadCookieButtonShow.value
	}))
);

onMounted(() => {
	// const url =
	// 	'https://www.bilibili.com/video/BV13WKUzXEPg/?spm_id_from=333.1007.tianma.1-1-1.click';
	// collectSiteStore.search(url);
});
</script>

<style></style>

<template>
	<div class="column flex-gap-y-lg">
		<CookieMessage v-if="collectSiteStore.errorCode" />
		<template v-else>
			<SiteCardContainer
				:title="$t('collectToLibrary')"
				:info="$t('createLibraryRecordAndAutoDownloadAttachment')"
				v-if="entry.file_type || collectSiteStore.data.is_entry_available"
			>
				<div class="column no-wrap flex-gap-y-md">
					<app-message
						v-if="collectSiteStore.data.is_entry_available"
						:message="collectSiteStore.data.is_entry_available"
						type="collect"
					/>
					<CollectSiteCard v-if="entry.file_type" :data="entry" />
				</div>
			</SiteCardContainer>
			<div class="column flex-gap-y-lg full-width">
				<SiteCardContainer
					:title="$t('downloadFile')"
					:info="$t('directDownloadFilesToOlaresWithoutLibraryRecords')"
					v-if="
						downloadList.length > 0 ||
						collectSiteStore.data.is_download_available
					"
				>
					<div class="column no-wrap flex-gap-y-md">
						<app-message
							v-if="collectSiteStore.data.is_download_available"
							:message="collectSiteStore.data.is_download_available"
							type="download"
						/>
						<template v-if="downloadList.length > 0">
							<DownloadSiteCard :data-list="downloadList" />
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
							type="feed"
						/>
						<template v-if="feedList.length > 0">
							<div v-for="(item, index) in feedList" :key="index">
								<FeedSiteCard :feed="item" />
							</div>
						</template>
					</div>
				</SiteCardContainer>
			</div>
		</template>
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
</script>

<style></style>

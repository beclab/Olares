<template>
	<div class="column flex-gap-y-lg">
		<CookieMessage />
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
			<EmptyData
				v-if="
					!collectSiteStore.loading &&
					searchUrl &&
					downloadList.length == 0 &&
					!collectSiteStore.data.is_download_available
				"
				:title="$t('no_data')"
				btn-hidden
				size="sm"
			></EmptyData>
		</div>
	</div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue';
import DownloadSiteCard from '../../../containers/collection/DownloadSiteCard.vue';
import SiteCardContainer from '../../../components/collection/SiteCardContainer.vue';
import CookieMessage from '../../../containers/collection/CookieMessage.vue';
import { useCollectSiteStore } from '../../../stores/collect-site';
import { COOKIE_LEVEL } from '../../../utils/rss-types';
import AppMessage from '../../../containers/collection/AppMessage.vue';
import EmptyData from 'src/pages/Plugin/components/EmptyData.vue';

const props = defineProps({
	searchUrl: {
		type: String,
		required: false,
		default: ''
	}
});

const collectSiteStore = useCollectSiteStore();
collectSiteStore.init();

const uploadCookieButtonShow = computed(
	() =>
		!collectSiteStore.cookie.cookieExist &&
		collectSiteStore.cookie.cookieRequire === COOKIE_LEVEL.REQUIRED
);

const downloadList = computed(
	() => []
	// collectSiteStore.download.map((item) => ({
	// 	...item,
	// 	disabled: uploadCookieButtonShow.value
	// }))
);

onMounted(() => {
	// const url =
	// 	'https://www.bilibili.com/video/BV13WKUzXEPg/?spm_id_from=333.1007.tianma.1-1-1.click';
	// collectSiteStore.search(url);
});
</script>

<style></style>

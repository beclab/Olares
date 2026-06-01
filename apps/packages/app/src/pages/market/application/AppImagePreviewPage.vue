<template>
	<page-container :is-app-details="true" :title-height="56">
		<template v-slot:title>
			<app-title-bar
				:app-title="appTitle"
				:app-version="appVersion"
				:app-icon="appIcon"
				:show-icon="showIcon(appEntry?.cfgType)"
				:show-header-bar="true"
			/>
		</template>
		<template v-slot:page>
			<div
				class="app-image-preview-page column justify-center items-center"
				:style="{
					'--imagePaddingTop': deviceStore.isMobile ? '40%' : '20px',
					'--imagePaddingX': deviceStore.isMobile ? '20px' : '0'
				}"
			>
				<app-store-swiper
					style="margin-bottom: 20px"
					ref="swiper"
					:ratio="1.6"
					:max-height="imageHeight"
					v-if="appEntry && appEntry.promoteImage"
					:data-array="
						appEntry && appEntry.promoteImage ? appEntry.promoteImage : []
					"
					:slides-per-view="1"
				>
					<template v-slot:swiper="{ item }">
						<q-img class="promote-img" ratio="1.6" :src="item">
							<template v-slot:loading>
								<q-skeleton class="promote-img" style="height: 100%" />
							</template>
						</q-img>
					</template>
				</app-store-swiper>
			</div>
		</template>
	</page-container>
</template>

<script lang="ts" setup>
import PageContainer from '../../../components/base/PageContainer.vue';
import AppStoreSwiper from '../../../components/base/AppStoreSwiper.vue';
import AppTitleBar from '../../../components/appintro/AppTitleBar.vue';
import SimpleWaiter from '../../../utils/simpleWaiter';
import { useDeviceStore } from '../../../stores/settings/device';
import { useAppStore } from '../../../stores/market/appStore';
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { getI18nValue } from '../../../constant/constants';
import { showIcon } from '../../../constant/config';
import { useRoute } from 'vue-router';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';

const swiper = ref();
const $q = useQuasar();
const route = useRoute();
const { locale } = useI18n();
const imageHeight = ref(0);
const initialSlide = ref(0);
const deviceStore = useDeviceStore();
const appStore = useAppStore();
let resizeTimer: NodeJS.Timeout | null = null;
const simpleWaiter = new SimpleWaiter();
const sourceId = route.params.sourceId as string;
const appName = route.params.appName as string;

const appAggregation = computed(() => {
	return appStore.getAppAggregationInfo(appName, sourceId);
});

onMounted(() => {
	window.addEventListener('resize', resize);
	simpleWaiter.waitForCondition(
		() => appEntry.value && swiper.value,
		() => {
			swiper.value.slideTo(Number(route.params.index));
			console.log(initialSlide.value);

			updateImageHeight();
		}
	);
});

onUnmounted(() => {
	window.removeEventListener('resize', resize);
});

const simpleInfo = computed(
	() => appAggregation.value?.app_simple_latest?.app_simple_info
);
const appEntry = computed(
	() => appAggregation.value?.app_full_info?.app_info?.app_entry
);

const appIcon = computed(() => simpleInfo.value?.app_icon ?? '');
const appTitle = computed(
	() => getI18nValue(simpleInfo.value?.app_title, locale.value) ?? ''
);
const appVersion = computed(() => simpleInfo.value?.app_version ?? '');

const resize = () => {
	if (resizeTimer) {
		clearTimeout(resizeTimer);
	}
	resizeTimer = setTimeout(function () {
		updateImageHeight();
	}, 200);
};

const updateImageHeight = () => {
	imageHeight.value = $q.screen.height - 56 - 40;
};
</script>
<style lang="scss" scoped>
.app-image-preview-page {
	width: 100%;
	height: 100%;
	padding-top: var(--imagePaddingTop);
	padding-right: var(--imagePaddingX);
	padding-left: var(--imagePaddingX);

	.promote-img {
		border-radius: 20px;
		width: 100%;
	}
}
</style>

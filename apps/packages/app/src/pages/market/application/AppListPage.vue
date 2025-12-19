<template>
	<page-container
		:vertical-position="56"
		v-model="showShadow"
		:hide-gradient="true"
		:title-height="56"
	>
		<template v-slot:title>
			<title-bar
				:show-back="true"
				:show="true"
				:title="categoryTitle"
				:show-title="showShadow"
				:shadow="showShadow"
				@onReturn="router.back()"
			/>
		</template>
		<template v-slot:page>
			<empty-view
				v-if="applications.length === 0"
				:show-title="false"
				class="empty_view"
			/>
			<div
				class="list-page"
				:style="{
					'--paddingX': deviceStore.isMobile ? '20px' : '44px',
					'--paddingBottom': deviceStore.isMobile ? '20px' : '56px'
				}"
				v-else
			>
				<div
					:class="
						deviceStore.isMobile
							? 'app-store-application-mobile'
							: 'app-store-application'
					"
				>
					<template v-for="item in applications" :key="item.name">
						<base-app-card
							:app-name="item"
							:source-id="settingStore.marketSourceId"
						/>
					</template>
					<app-card-hide-border />
				</div>
			</div>
		</template>
	</page-container>
</template>
<script lang="ts" setup>
import AppCardHideBorder from '../../../components/appcard/AppCardHideBorder.vue';
import PageContainer from '../../../components/base/PageContainer.vue';
import BaseAppCard from '../../../components/appcard/BaseAppCard.vue';
import EmptyView from '../../../components/base/EmptyView.vue';
import TitleBar from '../../../components/base/TitleBar.vue';
import { useSettingStore } from '../../../stores/market/setting';
import { useDeviceStore } from '../../../stores/settings/device';
import { useCenterStore } from '../../../stores/market/center';
import { useMenuStore } from '../../../stores/market/menu';
import { CONTENT_TYPE } from '../../../constant/constants';
import { useRouter, useRoute } from 'vue-router';
import { ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';

const centerStore = useCenterStore();
const deviceStore = useDeviceStore();
const settingStore = useSettingStore();
const showShadow = ref(false);
const menuStore = useMenuStore();
const router = useRouter();
const route = useRoute();
const { t } = useI18n();
const category = route.params.categories as string;
const type = route.params.type as string;

const categoryName = computed(() => {
	if (category === 'All') {
		return 'Olares';
	}
	return menuStore.getCategoryName(category);
});

const categoryTitle = computed(() => {
	switch (type) {
		case CONTENT_TYPE.RECOMMENDS:
			return t('recommend_app_in', {
				category: categoryName.value
			});
		case CONTENT_TYPE.TOP:
			return t('top_app_in', { category: categoryName.value });
		case CONTENT_TYPE.LATEST:
			return t('latest_app_in', { category: categoryName.value });
		default:
			return t('app_list');
	}
});

const applications = computed(() => {
	const pageData = centerStore.pagesMap.get(category);
	if (pageData) {
		const item = pageData.find((item) => item.type === type);
		return item ? item.content : [];
	} else {
		return [];
	}
});
</script>
<style lang="scss" scoped>
.list-page {
	width: 100%;
	height: calc(100% - var(--paddingBottom));
	padding: 0 var(--paddingX) var(--paddingBottom);
}

.empty_view {
	width: 100%;
	height: calc(100% - var(--paddingBottom));
}
</style>

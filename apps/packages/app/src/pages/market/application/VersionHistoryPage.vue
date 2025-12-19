<template>
	<page-container v-model="showShadow" :title-height="56">
		<template v-slot:title>
			<title-bar
				:show="true"
				:show-title="showShadow"
				:title="t('detail.version_history')"
				@onReturn="router.back()"
				:shadow="showShadow"
			/>
		</template>
		<template v-slot:page>
			<div
				class="version-history-scroll"
				:style="{ '--paddingX': deviceStore.isMobile ? '20px' : '44px' }"
			>
				<app-store-body
					:label="deviceStore.isMobile ? '' : t('detail.version_history')"
					:loading="versionList.length === 0"
					:title-separator="true"
				>
					<template v-slot:loading>
						<template v-for="item in 20" :key="item">
							<version-record-item :skeleton="true" />
						</template>
					</template>
					<template v-slot:body>
						<empty-view
							v-if="versionList.length === 0"
							:label="t('detail.no_version_history_desc')"
							class="empty_view"
						/>
						<div v-else>
							<template v-for="item in versionList" :key="item.mergedAt">
								<version-record-item :record="item" />
							</template>
						</div>
					</template>
				</app-store-body>
			</div>
		</template>
	</page-container>
</template>

<script lang="ts" setup>
import VersionRecordItem from '../../../components/appcard/VersionRecordItem.vue';
import PageContainer from '../../../components/base/PageContainer.vue';
import AppStoreBody from '../../../components/base/AppStoreBody.vue';
import EmptyView from '../../../components/base/EmptyView.vue';
import TitleBar from '../../../components/base/TitleBar.vue';
import { useDeviceStore } from '../../../stores/settings/device';
import { useCenterStore } from '../../../stores/market/center';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { computed, ref } from 'vue';

const { t } = useI18n();
const router = useRouter();
const route = useRoute();
const showShadow = ref(false);
const centerStore = useCenterStore();
const deviceStore = useDeviceStore();

const sourceId = route.params.sourceId as string;
const appName = route.params.appName as string;

const appAggregation = computed(() => {
	return centerStore.getAppAggregationInfo(appName, sourceId);
});

const versionList = computed(() => {
	return (
		appAggregation.value?.app_full_info?.app_info?.app_entry?.versionHistory ??
		[]
	);
});
</script>

<style scoped lang="scss">
.version-history-scroll {
	width: 100%;
	height: calc(100dvh - 56px);
	padding: 0 var(--paddingX);

	.empty_view {
		width: 100%;
		height: 500px;
	}
}
</style>

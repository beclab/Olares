<template>
	<page-container :title-height="deviceStore.isMobile ? 0 : 56">
		<template v-slot:page>
			<div
				class="update-scroll"
				:style="{ '--paddingX': deviceStore.isMobile ? '0px' : '44px' }"
			>
				<app-store-body
					:title="t('my.available_updates')"
					:title-separator="true"
					:padding-exclude-body="deviceStore.isMobile ? 12 : 0"
					:right="centerStore.updateList.length === 0 ? '' : t('my.update_all')"
					@on-right-click="onUpdateAll"
				>
					<template v-slot:left>
						<q-icon
							v-if="deviceStore.isMobile"
							size="20px"
							style="margin: 6px"
							name="sym_r_menu_open"
							class="cursor-pointer"
							@click="menuStore.setDrawerOpen(true)"
						/>
					</template>
					<template v-slot:body>
						<div :style="{ padding: `0 ${deviceStore.isMobile ? 20 : 0}px` }">
							<empty-view
								v-if="showList.length === 0"
								:label="t('my.everything_up_to_date')"
								class="empty_view"
							/>
							<div
								:class="
									deviceStore.isMobile
										? 'app-store-workflow-mobile'
										: 'app-store-workflow'
								"
								style="margin-top: 32px"
							>
								<template v-for="item in showList" :key="item">
									<function-app-card
										:app-name="item.split('_')[1]"
										:source-id="item.split('_')[0]"
										:is-update="true"
									/>
								</template>
								<app-card-hide-border />
							</div>
						</div>
					</template>
				</app-store-body>
			</div>
		</template>
	</page-container>
</template>

<script setup lang="ts">
import AppCardHideBorder from '../../../components/appcard/AppCardHideBorder.vue';
import FunctionAppCard from '../../../components/appcard/FunctionAppCard.vue';
import PageContainer from '../../../components/base/PageContainer.vue';
import AppStoreBody from '../../../components/base/AppStoreBody.vue';
import EmptyView from '../../../components/base/EmptyView.vue';
import { AppService } from '../../../stores/market/appService';
import { useDeviceStore } from '../../../stores/settings/device';
import { useCenterStore } from '../../../stores/market/center';
import { useMenuStore } from '../../../stores/market/menu';
import { useI18n } from 'vue-i18n';
import { computed } from 'vue';

const centerStore = useCenterStore();
const deviceStore = useDeviceStore();
const menuStore = useMenuStore();
const { t } = useI18n();

const showList = computed(() => {
	return centerStore.updateList.concat(centerStore.upgradingList);
});

const onUpdateAll = () => {
	centerStore.updateList.forEach((value) => {
		const array = value.split('_');
		if (array.length === 2) {
			const status = centerStore.getAppStatus(array[1], array[0]);
			const simpleLatest = centerStore.getAppSimpleInfo(array[1], array[0]);
			if (status && simpleLatest) {
				AppService.upgradeApp(status, {
					app_name: array[1],
					source: array[0],
					version: simpleLatest.app_simple_info.app_version
				});
			}
		}
	});
};
</script>

<style lang="scss">
.update-scroll {
	width: 100%;
	height: 100%;
	padding: 0 var(--paddingX);

	.empty_view {
		width: 100%;
		height: calc(100dvh - 112px - 64px);
	}
}
</style>

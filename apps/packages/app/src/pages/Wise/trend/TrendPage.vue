<template>
	<div class="wise-page-root bg-color-white column justify-start">
		<title-bar>
			<template v-slot:before>
				<bt-breadcrumbs
					:title="t('main.for_you')"
					icon="sym_r_volunteer_activism"
				>
					<template v-slot:end>
						<q-tabs
							dense
							v-model="configStore.menuChoice.tab"
							class="wise-page-tabs q-ml-md"
							active-color="orange-default"
							indicator-color="orange-default"
							align="left"
							:breakpoint="0"
						>
							<template
								v-for="(algorithm, index) in rssStore.support_algorithm"
								:key="`al` + index"
							>
								<q-tab class="wise-page-tab" :name="algorithm.id">
									<template v-slot:default>
										<tab-item
											:hide-border="true"
											:title="algorithm.title"
											:selected="configStore.menuChoice.tab === algorithm.id"
											@click="configStore.setMenuTab(algorithm.id)"
										/>
									</template>
								</q-tab>
							</template>
						</q-tabs>
					</template>
				</bt-breadcrumbs>
			</template>
			<template v-slot:after>
				<title-right-layout />
			</template>
		</title-bar>
		<div class="wise-page-content column justify-start">
			<q-tab-panels
				v-if="rssStore.support_algorithm.length > 0"
				v-model="configStore.menuChoice.tab"
				animated
				class="wise-page-tab-panels"
				keep-alive
			>
				<template
					v-for="(algorithm, index) in rssStore.support_algorithm"
					:key="`al` + index"
				>
					<q-tab-panel :name="algorithm.id" class="wise-page-tab-panel">
						<trend-source-page :algorithm="algorithm.id" />
					</q-tab-panel>
				</template>
			</q-tab-panels>
			<empty-view class="trend-router" v-else />
		</div>
	</div>
</template>

<script setup lang="ts">
import { useConfigStore } from '../../../stores/rss-config';
import { useRssStore } from '../../../stores/rss';
import TitleRightLayout from '../../../components/base/TitleRightLayout.vue';
import TrendSourcePage from './TrendSourcePage.vue';
import EmptyView from '../../../components/rss/EmptyView.vue';
import TabItem from '../../../components/rss/TabItem.vue';
import TitleBar from '../../../components/rss/TitleBar.vue';
import BtBreadcrumbs from '../../../components/base/BtBreadcrumbs.vue';
import { useI18n } from 'vue-i18n';
import { onActivated } from 'vue-demi';
import HotkeyManager from '../../../directives/hotkeyManager';
import { MenuType } from '../../../utils/rss-menu';

const configStore = useConfigStore();
const rssStore = useRssStore();
const { t } = useI18n();

onActivated(() => {
	HotkeyManager.setScope(MenuType.Trend);
});
</script>

<style lang="scss" scoped>
::v-deep(.q-scrollarea__thumb--h) {
	height: 0 !important;
}
</style>

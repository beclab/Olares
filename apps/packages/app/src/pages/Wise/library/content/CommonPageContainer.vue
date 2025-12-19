<template>
	<div class="wise-page-root bg-color-white column justify-start">
		<title-bar>
			<template v-slot:before>
				<bt-breadcrumbs
					:title="MenuInfoMap[menuType].title"
					:icon="MenuInfoMap[menuType].icon"
					:margin="margin"
				>
					<template v-slot:end>
						<q-tabs
							v-model="configStore.menuChoice.tab"
							dense
							class="wise-page-tabs q-ml-md"
							active-color="orange-default"
							indicator-color="orange-default"
							align="left"
							:breakpoint="0"
						>
							<template v-for="item in tabs" :key="item">
								<q-tab class="wise-page-tab" :name="item">
									<template v-slot:default>
										<tab-item
											:hide-border="true"
											:selected="configStore.menuChoice.tab === item"
											:title="TabInfoMap[item]?.title"
											@click="configStore.setMenuTab(item)"
										/>
									</template>
								</q-tab>
							</template>
						</q-tabs>
					</template>
				</bt-breadcrumbs>
			</template>

			<template v-slot:after>
				<title-right-layout>
					<slot name="title-right" />
				</title-right-layout>
			</template>
		</title-bar>
		<div class="wise-page-content column justify-start">
			<q-tab-panels
				v-model="configStore.menuChoice.tab"
				animated
				class="wise-page-tab-panels"
				keep-alive
			>
				<template v-for="(item, key) in tabs" :key="item">
					<q-tab-panel :name="item" class="wise-page-tab-panel">
						<slot :name="'tab-content-' + (key + 1)" :tab="item" />
					</q-tab-panel>
				</template>
			</q-tab-panels>
		</div>
	</div>
</template>

<script lang="ts" setup>
import TitleRightLayout from '../../../../components/base/TitleRightLayout.vue';
import { useConfigStore } from '../../../../stores/rss-config';
import { MenuInfoMap, TabInfoMap } from '../../../../utils/rss-menu';
import TitleBar from '../../../../components/rss/TitleBar.vue';
import BtBreadcrumbs from '../../../../components/base/BtBreadcrumbs.vue';
import TabItem from '../../../../components/rss/TabItem.vue';
import { computed } from 'vue';

const props = defineProps({
	menuType: {
		type: String
	},
	margin: {
		type: String,
		require: false
	}
});
const configStore = useConfigStore();
const tabs = computed(() => {
	return configStore.userTabs.get(props.menuType);
});
</script>

<style lang="scss" scoped></style>

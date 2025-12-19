<template>
	<q-card flat :bordered="false" square class="my-overview-container">
		<q-tabs
			v-model="tab"
			align="left"
			class="text-grey"
			active-color="primary"
			indicator-color="primary"
		>
			<q-tab
				v-for="item in tabList"
				:key="item.name"
				:name="item.name"
				style="min-height: 44px"
				no-caps
			>
				<span class="my-tab-wrapper">
					{{ item.label }}
				</span>
			</q-tab>
		</q-tabs>
	</q-card>

	<bt-scroll-area style="height: calc(100vh - 30px); margin: -8px">
		<q-tab-panels
			v-model="tab"
			animated
			style="padding: 0; background: transparent"
		>
			<q-tab-panel name="1">
				<Overview></Overview>
			</q-tab-panel>

			<q-tab-panel name="2">
				<PhysicalResource></PhysicalResource>
			</q-tab-panel>

			<q-tab-panel name="3">
				<ApplicationResource />
			</q-tab-panel>

			<q-tab-panel name="4">
				<ApplicationUsage></ApplicationUsage>
			</q-tab-panel>
		</q-tab-panels>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { InfoCardItemProps } from '@apps/control-hub/components/InfoCard/InfoCardItem.vue';
import ApplicationResource from './ApplicationResource/IndexPage.vue';
import Overview from './OverviewTab/IndexPage.vue';
import PhysicalResource from './PhysicalResource/IndexPage.vue';
import ApplicationUsage from './ApplicationUsage/IndexPage.vue';

const tab = ref('1');
const tabList = [
	{
		name: '1',
		label: 'Overview'
	},
	{
		name: '2',
		label: 'Physical Resource'
	},
	{
		name: '3',
		label: 'Application Resource'
	},
	{
		name: '4',
		label: 'Application Usage Ranking'
	}
];
</script>

<style lang="scss" scoped>
.my-overview-container {
	// padding: 0 8px;
	.my-tab-wrapper {
		font-family: 'Roboto';
		font-style: normal;
		font-weight: 500;
		font-size: 12px;
		line-height: 16px;
	}
}
.refresh-icon {
	color: $input-text-color;
}
.export-icon {
	color: $input-text-color;
}
</style>

<template>
	<bt-scroll-area class="full-height">
		<MyCard title="Realtime" flat class="q-ma-sm">
			<template #extra>
				<QSectionStyle>
					<q-select
						v-model="selected"
						:options="websites"
						option-label="name"
						option-value="id"
						outlined
						dense
					/>
				</QSectionStyle>
			</template>

			<div class="q-gutter-y-md">
				<WebsiteChart
					v-for="item in websites"
					:key="item.id"
					:name="item.name"
					:websiteId="item.id"
				>
				</WebsiteChart>
			</div>
			<PageView :id="selected.value"></PageView>
		</MyCard>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import { getWebsites } from '@apps/dashboard/src/network';
import { ref } from 'vue';
import PageView from './PageView.vue';
import WorldMap from './WorldMap.vue';
import MyCard from '@apps/control-panel-common/src/components/MyCard.vue';
import QSectionStyle from '@apps/dashboard/src/components/QSectionStyle.vue';
import WebsiteChart from './WebsiteChart.vue';

const selected = ref({ label: '', value: '' });
const websites = ref();

const fetchData = () => {
	getWebsites().then((res) => {
		websites.value = res.data;
		selected.value = websites.value[0];
	});
};

fetchData();
</script>

<style></style>

<template>
	<div>
		<q-tabs dense v-model="filter" class="text-teal">
			<q-tab
				v-for="item in options"
				:key="item.key"
				:name="item.key"
				:label="item.label"
				noCaps
				class="text-grey-9"
				active-color="primary"
				indicator-color="primary"
				narrow-indicator
			/>
		</q-tabs>
		<bt-scroll-area style="height: 300px">
			<template v-if="logs.length > 0">
				<div v-for="item in logs" :key="item.id">
					<div>
						<q-badge color="positive" />
						<span class="q-ml-sm">{{ getTime(item) }}</span>
						<q-icon name="visibility" class="q-ml-lg" />
						<span class="q-ml-md">{{ item.urlPath }}</span>
					</div>
				</div>
			</template>
			<Empty v-else size="small" center></Empty>
		</bt-scroll-area>
	</div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import firstBy from 'thenby';
import { dateFormat } from '@apps/dashboard/src/utils/analytics-date';
import Empty from '@apps/control-panel-common/src/components/Empty.vue';

interface Props {
	data: any;
}
const TYPE_ALL = 'all';
const TYPE_PAGEVIEW = 'pageview';
const TYPE_SESSION = 'session';
const TYPE_EVENT = 'event';

const options = [
	{
		label: TYPE_ALL,
		key: TYPE_ALL
	},
	{
		label: TYPE_PAGEVIEW,
		key: TYPE_PAGEVIEW
	},
	{
		label: TYPE_SESSION,
		key: TYPE_SESSION
	},
	{
		label: TYPE_EVENT,
		key: TYPE_EVENT
	}
];

const props = withDefaults(defineProps<Props>(), {});
const filter = ref(TYPE_ALL);
const logs = computed(() => {
	if (!props.data) {
		return [];
	}

	const { pageviews, visitors, events } = props.data;
	const logs = [...pageviews, ...visitors, ...events].sort(
		firstBy('createdAt', -1)
	);

	if (filter.value !== TYPE_ALL) {
		return logs.filter(({ __type }) => __type === filter.value);
	}

	return logs;
});

const getTime = ({ createdAt }) => dateFormat(new Date(createdAt), 'pp');
</script>

<style></style>

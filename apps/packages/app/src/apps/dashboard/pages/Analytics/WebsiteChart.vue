<template>
	<MyCard
		:title="name"
		class="website-card-wrapper highlight"
		@click.prevent="onClick"
	>
		<template #avatar>
			<MyImg :src="icon" :loading="loading" />
		</template>
		<template #extra>
			<QButtonStyle>
				<q-btn
					icon="sym_r_open_in_new"
					class="rounded-borders-sm q-mr-sm"
					dense
					outline
					color="ink-2"
					@click.stop="openUrl(url)"
				>
					<q-tooltip
						><span class="text-body3 text-ink-tooltip">{{
							$t('OPEN_APP')
						}}</span></q-tooltip
					>
				</q-btn>
			</QButtonStyle>
			<QButtonStyle>
				<q-btn
					icon="sym_r_assignment"
					class="rounded-borders-sm"
					dense
					outline
					color="ink-2"
					@click.stop="onClick"
				>
					<q-tooltip
						><span class="text-body3 text-ink-tooltip">{{
							$t('VIEW_DETAIL')
						}}</span></q-tooltip
					>
				</q-btn>
			</QButtonStyle>
		</template>
		<div class="q-gutter-y-xl">
			<PageviewsChart
				:data="data"
				:website-id="websiteId"
				style="height: 190px"
				:loading="loading"
			></PageviewsChart>
			<MetricsBar
				:websiteId="websiteId"
				:key="updateKey"
				ref="metricsBarRef"
				:loading="loading"
			></MetricsBar>
		</div>
	</MyCard>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import MetricsBar from './MetricsBar.vue';
import { useRouter } from 'vue-router';
import QButtonStyle from '@apps/control-panel-common/src/components/QButtonStyle.vue';
import MyCard from '@apps/dashboard/src/components/MyCard.vue';
import PageviewsChart from '@apps/dashboard/src/pages/Analytics/PageviewsChart.vue';
import { getPageviews } from '@apps/dashboard/src/network';
import { DEFAULT_DATE_OPTIONS_VALUE } from './DatefilterState';
import { getTimezone, parseDateRange } from './date';
import MyImg from '@apps/control-panel-common/src/components/MyImg.vue';
interface Props {
	websiteId: string;
	name: string;
	icon: string;
	loading?: boolean;
	url: string;
}

const router = useRouter();
const props = withDefaults(defineProps<Props>(), {});
const metricsBarRef = ref();
const updateKey = ref();
const data = ref([]);

const onClick = () => {
	router.push({
		path: `/analytics/details/${props.websiteId}`,
		query: {
			name: props.name,
			icon: props.icon
		}
	});
};

const fetchData = () => {
	const { startDate, endDate, unit } = parseDateRange(
		DEFAULT_DATE_OPTIONS_VALUE
	);

	const timezone = getTimezone();

	const params = {
		startAt: +startDate,
		endAt: +endDate,
		unit,
		timezone
	};
	getPageviews(props.websiteId, params).then((res) => {
		data.value = res.data;
	});
};

const openUrl = (url: string) => {
	window.open(url);
};

watch(
	() => props.websiteId,
	() => {
		if (props.websiteId) {
			fetchData();
		}
	},
	{
		immediate: true
	}
);
</script>

<style lang="scss" scoped>
.website-card-wrapper {
	position: relative;
	&:active {
		background: $background-hover;
	}
	&:hover {
		&.highlight::after {
			content: '';
			position: absolute;
			top: 0;
			left: 0;
			right: 0;
			bottom: 0;
			border: 2px solid $light-blue-6;
			border-radius: 20px;
			pointer-events: none;
		}
	}
}
.title-img {
	box-shadow: 0px 2px 4px 0px rgba(0, 0, 0, 0.1);
}
</style>
./WebsiteDetails.vue

<template>
	<div class="row items-center justify-between">
		<div class="row items-center">
			<div class="parent-container" :class="iconCard">
				<q-img :src="icon" :ratio="1" :width="iconWidth" />
			</div>
			<div class="text-h4 q-ml-lg text-ink-1">{{ value }}</div>
		</div>
		<div class="q-px-lg q-py-sm rounded-borders-lg" :class="trendClass">
			<q-icon :name="trendIcon" size="24px" />
			<!-- <q-icon name="sym_r_trending_down" /> -->
			<span class="text-subtitle1 q-ml-sm">{{ trend }}{{ change || '' }}</span>
		</div>
	</div>
	<div class="text-subtitle1 text-ink-2 q-mt-lg">{{ subtile }}</div>
</template>

<script setup lang="ts">
import supervisorAccountIcon from '@apps/dashboard/assets/supervisor_account.svg';
import supervisorAccountIconDark from '@apps/dashboard/assets/supervisor_account-dark.svg';
import visibilityIcon from '@apps/dashboard/assets/visibility.svg';
import visibilityIconDark from '@apps/dashboard/assets/visibility-dark.svg';
import { computed } from 'vue';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
interface Props {
	type: 'view' | 'visit';
	value: number;
	change: number;
}
const props = withDefaults(defineProps<Props>(), {
	type: 'view',
	value: 0
});

const $q = useQuasar();

const icon = computed(() => {
	if (props.type === 'view') {
		return $q.dark.isActive ? visibilityIconDark : visibilityIcon;
	} else {
		return $q.dark.isActive ? supervisorAccountIconDark : supervisorAccountIcon;
	}
});

const iconWidth = computed(() => (props.type === 'view' ? '28px' : '32px'));

const subtile = computed(() =>
	props.type === 'view'
		? t('VIEWS_COUNT_VIEW', { count: 24 })
		: t('VISITORS_COUNT_VISITOR', { count: 24 })
);
const trend = computed(() => (props.change > 0 ? '+' : ''));
const trendIcon = computed(() =>
	props.change > 0
		? 'sym_r_trending_up'
		: props.change < 0
		? 'sym_r_trending_down'
		: ''
);
const trendClass = computed(() =>
	props.change > 0
		? 'text-positive bg-green-soft'
		: props.change < 0
		? 'text-negative bg-red-soft'
		: ''
);

const iconCard = computed(() =>
	props.type === 'view'
		? 'bg-light-green-disabled'
		: props.type === 'visit'
		? 'bg-light-blue-disabled'
		: ''
);
</script>

<style lang="scss" scoped>
.parent-container {
	width: 48px;
	height: 48px;
	display: flex;
	align-items: center;
	justify-content: center;
	border-radius: 12px !important;
	overflow: hidden;
}
</style>

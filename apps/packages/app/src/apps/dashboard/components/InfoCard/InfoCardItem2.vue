<template>
	<div class="info-card-container">
		<div
			class="row items-center no-wrap info-card-wrapper q-pa-lg"
			:class="[props.active ? 'info-card-wrapper-active' : '']"
			:style="{ background: props.active ? backgroundMode : undefined }"
		>
			<div
				class="img-wrapper row items-center justify-center"
				:class="[props.active ? 'img-active' : '']"
			>
				<q-img :src="props.img" width="24px" ratio="1" no-spinner />
			</div>
			<div class="row no-wrap items-center justify-between q-ml-lg">
				<div>
					<div
						class="row items-center info-ratio no-wrap text-h5 text-ink-1 value-wrapper"
					>
						<q-skeleton v-if="loading" type="text" width="88px" />
						<template v-else>
							<span>{{ _used }}</span>
							<span>/</span>
							<span class="ratio-foolter">{{ _total || '-' }}</span>
						</template>
					</div>
					<div class="text-subtitle1 text-ink-3 q-mt-xs">
						<q-skeleton v-if="loading" type="text" width="64px" />
						<template v-else>
							<span>{{ _capitalize(name) }}&nbsp;</span>
							<span>{{ _unit }}</span>
						</template>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import {
	getSuitableUnit,
	getValueByUnit
} from '@apps/dashboard/src/utils/monitoring';
import { computed, toRefs } from 'vue';
import { _capitalize } from '@apps/dashboard/src/utils/index';
import { isNaN, round } from 'lodash';
import { useQuasar } from 'quasar';
const $q = useQuasar();

export interface InfoCardItemProps {
	used: string;
	total: string;
	name?: string;
	active?: boolean;
	unitType?: any;
	unit?: string;
	img?: string;
	img_active?: string;
	loading?: boolean;
}

const props = withDefaults(defineProps<InfoCardItemProps>(), {
	active: false,
	title: ''
});

const { total, used, unitType, unit } = toRefs(props);
const _unit = computed(
	() =>
		getSuitableUnit(total.value || used.value, unitType.value as any) ||
		unit?.value
);
const _used = computed(() => getValueByUnit(used.value, _unit.value));
const _total = computed(() => getValueByUnit(total.value, _unit.value));

const backgroundMode = computed(() =>
	$q.dark.isActive
		? 'linear-gradient(125deg, #1f1f1f 4.57%, #262e37 87.85%)'
		: 'linear-gradient(125deg, #FFF 4.57%, #EBF5FF 87.85%)'
);
</script>

<style lang="scss" scoped>
.info-card-container {
	cursor: default;
	.img-wrapper {
		width: 48px;
		height: 48px;
		border-radius: 50%;
		background: $background-3;
	}
	.info-card-wrapper {
		position: relative;
		border-radius: 12px;
		border: 1px solid $separator;
		&.info-card-wrapper-active {
			border: 1px solid $light-blue-6;
		}
		.img-active {
			background: $blue-alpha;
		}
		.value-wrapper {
			white-space: nowrap;
		}
	}
}
</style>

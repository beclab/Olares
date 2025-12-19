<template>
	<div class="column q-pa-lg into-progress-wrapper">
		<div class="row text-subtitle2">
			<span class="text-ink-1 text-body1">
				{{ title }}
			</span>

			<q-btn dense>
				<q-icon name="sym_r_help" size="16px" color="ink-3"></q-icon>
				<q-tooltip>{{ tip }}</q-tooltip>
			</q-btn>

			<q-space />

			<!--			<div-->
			<!--				class="row justify-end items-center text-info text-body2 cursor-pointer"-->
			<!--			>-->
			<!--				{{ t('view_details') }}-->
			<!--			</div>-->
		</div>
		<q-linear-progress
			color="positive"
			:track-color="overage == 0 ? 'background-4' : 'orange-default'"
			class="q-mt-sm"
			:animationSpeed="0"
			rounded
			size="8px"
			:value="computedUsed / computedTotal"
		/>
		<div class="row items-center q-mt-md">
			<div class="row items-center">
				<div class="legend-item bg-positive"></div>
				<div class="text-body3 text-ink-3 q-ml-xs">
					{{
						t('Within plan Storage', {
							Storage: calculateSize(used || 0)
						})
					}}
				</div>
			</div>
			<div class="q-ml-lg row items-center" v-if="overage > 0">
				<div class="legend-item bg-orange-default"></div>
				<div class="text-body3 text-ink-3 q-ml-xs">
					{{
						t('Within plan Storage', {
							Storage: calculateSize(overage || 0)
						})
					}}
				</div>
			</div>
			<div class="row items-center q-ml-lg">
				<div class="legend-item bg-background-4"></div>
				<div class="text-body3 text-ink-3 q-ml-xs">
					{{
						t('Plan cap Storage', {
							Storage: calculateSize(total || 0)
						})
					}}
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { getSuitableValue } from 'src/utils/settings/monitoring';
import { computed } from 'vue';

const props = defineProps({
	title: {
		type: String,
		required: true
	},
	tip: {
		type: String,
		required: true
	},
	total: {
		type: Number,
		required: true
	},
	used: {
		type: Number,
		required: true
	},
	overage: {
		type: Number,
		required: true
	}
});

const { t } = useI18n();

const calculateSize = (size: number) => {
	return getSuitableValue(size.toString(), 'disk');
};

const computedUsed = computed(() => {
	if (props.overage <= 0) {
		return props.used;
	}
	return props.total;
});

const computedTotal = computed(() => {
	if (props.overage <= 0) {
		return props.total;
	}
	return props.total + props.overage;
});
</script>

<style scoped lang="scss">
.legend-item {
	width: 8px;
	height: 8px;
	border-radius: 4px;
}

.into-progress-wrapper {
	::v-deep(.q-linear-progress__track) {
		opacity: 1;
	}
}
</style>

<template>
	<span
		class="compute-badge text-caption"
		:class="[`compute-badge--${tone}`, { 'compute-badge--status': dot }]"
	>
		<span v-if="dot" class="compute-badge__dot" />
		{{ label }}
		<slot />
	</span>
</template>

<script setup lang="ts">
import { PropType } from 'vue';
import { ComputeTone } from 'src/constant/compute';

defineProps({
	label: {
		type: String,
		required: true
	},
	/** Visual tone; `neutral` is the muted grey tag used for type chips. */
	tone: {
		type: String as PropType<ComputeTone>,
		default: 'neutral'
	},
	/** Show a leading status dot (used by node status badges). */
	dot: {
		type: Boolean,
		default: false
	}
});
</script>

<style scoped lang="scss">
.compute-badge {
	display: inline-flex;
	align-items: center;
	padding: 2px 4px;
	border-radius: 4px;
	font-weight: 500;
	line-height: 16px;
	white-space: nowrap;
	text-transform: uppercase;

	&__dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		margin-right: 6px;
		background: currentColor;
	}

	&--neutral {
		color: $ink-3;
		background: $background-3;
	}

	&--info {
		color: $blue-default;
		background: $blue-alpha;
	}

	&--brand {
		color: $ink-on-brand;
		background: $blue-default;
	}

	&--ok {
		color: $green-default;
		background: transparent;
	}

	&--success {
		color: $green-default;
		background: $teal-alpha;
	}

	&--warn {
		color: $warning;
		background: $yellow-alpha;
	}

	&--bad {
		color: $negative;
		background: $red-soft;
	}

	// Status badges (with leading dot) are text-only, no filled background.
	&--status {
		padding: 0;
		background: transparent;
		font-weight: 400;
	}
}
</style>

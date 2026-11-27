<template>
	<div
		v-if="message"
		class="compute-validation-notice row justify-start items-start"
		:class="`compute-validation-notice--${tone}`"
	>
		<q-icon
			name="sym_r_error"
			size="16px"
			:color="tone === 'bad' ? 'negative' : 'orange-default'"
			class="q-mr-sm"
		/>
		<div class="compute-validation-notice__text text-body3" style="flex: 1">
			{{ message }}
		</div>
	</div>
</template>

<script setup lang="ts">
import { PropType } from 'vue';
import { ComputeTone } from 'src/constant/compute';

defineProps({
	message: {
		type: String,
		default: ''
	},
	tone: {
		type: String as PropType<ComputeTone>,
		default: 'bad'
	}
});
</script>

<style scoped lang="scss">
.compute-validation-notice {
	border-radius: 8px;
	padding: 8px 12px;

	&__text {
		white-space: pre-line;
	}

	&--bad {
		color: $negative;
		background: $red-soft;
	}

	&--warn {
		color: $orange-default;
		background: $orange-soft;
	}
}
</style>

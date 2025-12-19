<template>
	<div
		class="bt-grid-root"
		:class="deviceStore.isMobile ? '' : 'bt-grid-border'"
		:style="{
			'--padding-y': paddingY + 'px'
		}"
	>
		<slot name="title" />
		<bt-separator />
		<div
			class="row bt-grid-grid"
			:class="deviceStore.isMobile ? 'q-mt-lg' : 'q-mt-md'"
			:style="{
				'--repeat_count': repeatCount
			}"
		>
			<slot name="grid" />
		</div>
		<q-linear-progress
			class="line-progress"
			v-if="showProgress"
			:value="progress"
			size="4px"
			color="info"
		/>
	</div>
</template>

<script lang="ts" setup>
import { useDeviceStore } from 'src/stores/settings/device';
import BtSeparator from '../base/BtSeparator.vue';

const deviceStore = useDeviceStore();

defineProps({
	repeatCount: {
		type: Number,
		required: false,
		default: 3
	},
	paddingY: {
		type: Number,
		required: false,
		default: 16
	},
	showProgress: {
		type: Boolean,
		default: false
	},
	progress: {
		type: Number,
		default: 0
	}
});
</script>

<style scoped lang="scss">
.bt-grid-root {
	width: 100%;
	height: auto;
	border-radius: 12px;
	margin-top: 20px;
	padding: var(--padding-y) 20px;
	position: relative;
	overflow: hidden;

	.bt-grid-grid {
		width: 100%;
		display: grid;
		grid-column-gap: 12px;
		grid-row-gap: 20px;
		grid-template-columns: repeat(var(--repeat_count), minmax(0, 1fr));
	}

	.line-progress {
		width: 100%;
		position: absolute;
		bottom: 0;
		left: 0;
		right: 0;
	}
}

.bt-grid-border {
	border: 1px solid $separator;
}
</style>

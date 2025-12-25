<template>
	<div class="row items-center">
		<div class="row items-center justify-center node-bg">
			<div
				class="status-node"
				:style="{
					'--node-size': nodeSize + 'px'
				}"
				:class="
					displayStatus(status).color
						? `bg-${displayStatus(status).color}`
						: 'bg-background-5'
				"
			></div>
		</div>
		<div
			class="q-ml-xs text-body3"
			:class="
				displayStatus(status).color
					? `text-${displayStatus(status).color}`
					: 'text-ink-3'
			"
		>
			{{ displayStatus(status).status }}
		</div>
	</div>
</template>

<script setup lang="ts">
import { PropType } from 'vue';
import {
	displayStatus,
	TerminusStatus
} from '../../services/abstractions/mdns/service';

defineProps({
	status: {
		type: Object as PropType<TerminusStatus>,
		required: true
	},
	nodeSize: {
		type: Number,
		required: false,
		default: 12
	}
});
</script>

<style scoped lang="scss">
.node-bg {
	width: 20px;
	height: 20px;
	.status-node {
		width: var(--node-size, 12px);
		height: var(--node-size, 12px);
		border-radius: 50%;
	}
}
</style>

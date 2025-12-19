<template>
	<div class="row items-center olares_status">
		<div
			class="node"
			:class="
				status && displayStatus(status).color
					? `bg-${displayStatus(status).color}`
					: 'bg-background-5'
			"
		></div>
		<div
			class="q-ml-xs text-body3"
			:class="
				status && displayStatus(status).color
					? `text-${displayStatus(status).color}`
					: 'text-ink-3'
			"
		>
			{{ (status && displayStatus(status).status) || statusStr }}
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
		required: false
	},
	statusStr: {
		type: String,
		required: false,
		default: ''
	}
});
</script>

<style scoped lang="scss">
.olares_status {
	height: 20px;
	.node {
		width: 8px;
		height: 8px;
		border-radius: 50%;
	}
}
</style>

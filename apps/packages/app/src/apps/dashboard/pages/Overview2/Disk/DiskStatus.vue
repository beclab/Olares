<template>
	<div
		class="gpu-status-container row inline justify-center items-center q-py-xs q-px-md"
		:class="[
			!isBoolean(status)
				? 'text-status-pending'
				: status
				? 'text-green-default'
				: 'text-red-default'
		]"
	>
		{{ label }}
	</div>
</template>

<script setup lang="ts">
import { isBoolean } from 'lodash';

interface Props {
	status: boolean;
	label: string;
}

defineProps<Props>();
</script>

<template>
	<q-avatar :size="`${size}px`">
		<TerminusAvatar :info="terminusStore.terminusInfo" :size="size" />
	</q-avatar>
</template>

<script lang="ts" setup>
import { useTerminusStore } from 'src/stores/terminus';

defineProps({
	size: {
		type: Number,
		require: true
	}
});

const terminusStore = useTerminusStore();
</script>

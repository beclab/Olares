<template>
	<q-avatar :size="`${size}px`">
		<TerminusAvatar :info="adminStore.terminus" :size="size" />
	</q-avatar>
</template>

<script lang="ts" setup>
import { useAdminStore } from 'src/stores/settings/admin';

defineProps({
	size: {
		type: Number,
		require: true
	}
});

const adminStore = useAdminStore();
</script>

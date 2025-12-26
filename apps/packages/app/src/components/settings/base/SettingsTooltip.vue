<template>
	<q-icon
		v-if="description"
		size="16px"
		class="q-ml-sm"
		:class="classes"
		name="sym_r_help"
	>
		<q-tooltip max-width="280px" anchor="bottom middle" self="top left">
			<span class="text-body3 text-ink-tooltip">{{ description }}</span>
		</q-tooltip>
	</q-icon>
</template>

<script setup lang="ts">
defineProps({
	description: {
		type: String,
		required: false,
		default: ''
	},
	classes: {
		required: false,
		default: 'text-ink-3',
		type: String
	}
});
</script>

<style scoped lang="scss"></style>

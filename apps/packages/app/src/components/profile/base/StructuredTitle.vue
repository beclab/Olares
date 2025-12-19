<template>
	<div class="column">
		<div class="text-h5 text-ink-1 q-mt-xl">{{ primaryTitle }}</div>
		<div v-if="primaryDescription" class="text-body2 text-ink-3 q-mt-xs">
			{{ primaryDescription }}
		</div>
		<div v-if="hasPrimarySlot" class="q-mt-md">
			<slot name="primary" />
		</div>
		<div v-for="(item, index) in secondaryItems" :key="index">
			<div v-if="item.visible" class="column q-mt-lg">
				<div v-if="item.title" class="text-subtitle1 text-ink-1">
					{{ item.title }}
				</div>
				<div v-if="item.description" class="text-body2 text-ink-3 q-mt-xs">
					{{ item.description }}
				</div>
				<div class="q-mt-md">
					<slot :name="'secondary-' + index" :item="item" />
				</div>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { defineProps, useSlots, computed } from 'vue';

defineProps({
	primaryTitle: {
		type: String,
		required: true
	},
	primaryDescription: {
		type: String,
		default: ''
	},
	secondaryItems: {
		type: Array as () => Array<{
			title: string;
			description?: string;
			visible?: boolean;
		}>,
		default: () => []
	}
});

const slots = useSlots();
const hasPrimarySlot = computed(() => !!slots.primary);
</script>

<style lang="scss" scoped></style>

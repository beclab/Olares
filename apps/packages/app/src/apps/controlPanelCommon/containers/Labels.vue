<template>
	<div v-if="hasData" class="q-gutter-md">
		<div
			v-for="(value, key) in data"
			:key="key"
			class="labels-content rounded-borders q-pr-sm"
		>
			<q-chip
				v-if="key"
				square
				size="12px"
				color="blue-default"
				text-color="ink-on-brand"
				class="q-ma-none"
				style="border-radius: 4px"
			>
				{{ key }}
			</q-chip>
			<span class="q-ml-sm">
				{{ value }}
			</span>
		</div>
	</div>
	<Empty v-else-if="!noEmpty" size="small"></Empty>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import Empty from '../components/Empty.vue';
interface Props {
	data: { [key: string]: string };
	noEmpty: boolean;
}
const props = withDefaults(defineProps<Props>(), {
	noEmpty: false
});
const hasData = computed(() => {
	return props.data && Object.keys(props.data).length > 0;
});
</script>
<style lang="scss" scoped>
.labels-content {
	display: inline-block;
	border: 1px solid $separator;
}
</style>

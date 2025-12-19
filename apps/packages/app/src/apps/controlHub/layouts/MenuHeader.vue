<template>
	<MyContentHeader :icon="currentItem?.icon" :class="paddingX">
		<template #title>
			<div class="row items-center">
				<slot v-if="$slots.title" name="title"></slot>
				<span v-else class="text-h6 text-ink-1">{{ currentItem?.label }}</span>
				<slot v-if="$slots['title-after']" name="title-after"></slot>
			</div>
		</template>
	</MyContentHeader>
</template>

<script setup lang="ts">
import MyContentHeader from '@apps/control-panel-common/src/components/MyContentHeader.vue';
import { currentItem } from '@apps/control-hub/layouts/breadcrumbs';
import { computed } from 'vue';

interface Props {
	size?: 'sm' | 'md' | 'lg' | 'xl' | 'xxl';
}

const props = withDefaults(defineProps<Props>(), {
	size: 'lg'
});

const paddingX = computed(() => `q-px-${props.size}`);
</script>

<style></style>

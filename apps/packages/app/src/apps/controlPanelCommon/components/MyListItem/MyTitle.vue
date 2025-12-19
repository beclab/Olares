<template>
	<div
		class="my-title-container full-width text-ink-1"
		:class="[sizeClass, ellipsis ? 'ellipsis' : '']"
		:style="{ fontWeight: bold ? '500' : 'normal' }"
	>
		<template v-if="$slots.default">
			<slot></slot>
		</template>
		<template v-else>
			{{ title }}
		</template>
	</div>
</template>

<script setup lang="ts">
import { computed, toRefs } from 'vue';

interface Props {
	title?: string;
	bolder?: boolean;
	size?: 'sm' | 'md' | 'lg';
	ellipsis?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
	bolder: true,
	size: 'md',
	ellipsis: true
});

const { bolder: bold } = toRefs(props);

const sizeClass = computed(() => `${props.size}-wrapper`);
</script>
<style lang="scss" scoped>
.my-title-container {
	flex: 1;
	&.xs-wrapper {
		font-weight: 400;
		line-height: 16px;
	}
	&.sm-wrapper {
		font-weight: 500;
		line-height: 16px;
	}
	&.md-wrapper {
		font-weight: 500;
		line-height: 20px;
	}
	&.lg-wrapper {
		font-size: 16px;
		font-weight: 500;
		line-height: 24px;
	}
}
</style>

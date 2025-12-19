<template>
	<div :class="className" @click="handleClick">
		<div>
			<template v-if="$slots.title">
				<slot name="title" />
			</template>
			<template v-else-if="!title">-</template>
			<template v-else>{{ title }}</template>
		</div>
		<div class="text-grey">
			<template v-if="$slots.description">
				<slot name="description" />
			</template>
			<div>
				{{ description }}
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { defineProps, Slot } from 'vue';
import { isFunction } from 'lodash';

interface CardProps {
	icon?: string | Slot;
	title?: string | Slot;
	description?: string;
	className?: string;
	ellipsis?: boolean;
	extra?: string;
	onClick?: () => void;
}

const props = defineProps<CardProps>();

const handleClick = () => {
	if (isFunction(props.onClick)) {
		props.onClick();
	}
};
</script>

<style scoped lang="scss"></style>

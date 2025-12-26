<template>
	<q-card flat :bordered="bordered" class="rounded-borders-lg q-pa-xl">
		<div class="row justify-between items-center q-pb-xl" v-if="headerShow">
			<div class="row items-center">
				<div v-if="$slots.avatar" class="q-mr-md">
					<slot name="avatar"></slot>
				</div>
				<div v-else>{{ icon }}</div>

				<div v-if="$slots.title"><slot name="title"></slot></div>
				<div v-else class="text-h4 text-ink-1">{{ title }}</div>
			</div>
			<div><slot name="extra"></slot></div>
		</div>
		<slot></slot>
	</q-card>
</template>

<script setup lang="ts">
import { computed, useSlots } from 'vue';

interface Props {
	title?: string;
	icon?: string;
	extra?: string;
	bordered?: boolean;
}
const props = withDefaults(defineProps<Props>(), {
	bordered: true
});
const slots = useSlots();
const headerShow = computed(
	() =>
		props.icon ||
		props.extra ||
		props.title ||
		slots.extra ||
		slots.title ||
		slots.icon
);
</script>

<style></style>

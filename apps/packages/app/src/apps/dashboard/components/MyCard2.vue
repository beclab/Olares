<template>
	<q-card flat class="transparent">
		<div class="row justify-between items-center q-py-lg" v-if="headerShow">
			<div class="row items-center">
				<div v-if="$slots.avatar" class="q-mr-md">
					<slot name="avatar"></slot>
				</div>
				<div v-else>{{ icon }}</div>

				<div v-if="$slots.title"><slot name="title"></slot></div>
				<div v-else class="text-h5 text-ink-1">{{ title }}</div>
			</div>
			<div class="row items-center q-gutter-md">
				<div><slot name="extra"></slot></div>
				<div
					class="text-h6 text-light-blue-default cursor-pointer"
					@click.stop="linkHandler"
				>
					{{ link }}
				</div>
			</div>
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
	link?: string;
}
const props = withDefaults(defineProps<Props>(), {});
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

const emits = defineEmits(['linkHandler']);
const linkHandler = () => {
	emits('linkHandler');
};
</script>

<style lang="scss" scoped></style>

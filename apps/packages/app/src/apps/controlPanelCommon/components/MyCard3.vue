<template>
	<q-card flat class="transparent">
		<div class="row justify-between items-center q-py-lg" v-if="headerShow">
			<div class="row items-center">
				<div v-if="$slots.avatar" class="q-mr-md">
					<slot name="avatar"></slot>
				</div>
				<div v-else>{{ icon }}</div>

				<div v-if="$slots.title" class="text-ink-1">
					<slot name="title"></slot>
				</div>
				<div v-else class="text-subtitle1 text-ink-1">{{ title }}</div>
			</div>
			<div class="card3-extra">
				<slot name="extra"></slot>
				<div
					class="text-h6 text-primary cursor-pointer"
					@click.stop="linkHandler"
					v-if="link"
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

<style lang="scss" scoped>
.card3-extra {
	display: flex;
	gap: 12px;
}
</style>

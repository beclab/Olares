<template>
	<q-card
		:flat="flat"
		:bordered="bordered"
		:dark="dark"
		:tag="tag"
		class="my-card-container q-pa-none"
	>
		<q-card-section v-if="titleShow" class="my-card-section-wrapper q-pa-none">
			<div
				class="row no-wrap justify-between items-center q-py-lg q-pl-lg q-pr-md"
				style="gap: 20px"
				:class="[contentClass]"
			>
				<div class="text-subtitle2 text-ink-1 ellipsis">
					<slot v-if="$slots.title" name="title"></slot>
					<span v-else>{{ title }}</span>
				</div>
				<div class="my-card-extra row justify-end text-ink-2" style="flex: 1">
					<slot name="extra" v-if="slots.extra"></slot>
					<template v-if="extra">
						{{ extra }}
					</template>
				</div>
			</div>
		</q-card-section>
		<q-card-section class="my-card-content q-px-lg q-pb-lg q-pt-none">
			<slot></slot>
		</q-card-section>
	</q-card>
</template>

<script setup lang="ts">
import MyTitle from './MyListItem/MyTitle.vue';
import { computed, toRefs, useSlots } from 'vue';

interface Props {
	title?: string;
	extra?: string;
	/**
	 * Notify the component that the background is a dark color
	 */
	dark?: boolean | undefined;
	/**
	 * Removes border-radius so borders are squared
	 */
	square?: boolean | undefined;
	/**
	 * Applies a 'flat' design (no default shadow)
	 */
	flat?: boolean | undefined;
	/**
	 * Applies a default border to the component
	 */
	bordered?: boolean | undefined;
	/**
	 * HTML tag to use
	 * Default value: div
	 */
	tag?: string | undefined;
	/**
	 * no card content gap padding
	 */
	noContentGap?: boolean | undefined;
	contentClass?: string;
}

const slots = useSlots();

const props = withDefaults(defineProps<Props>(), {
	title: '',
	extra: '',
	square: false,
	flat: true
});

const { title, extra, dark, square, flat, bordered, tag } = toRefs(props);
const titleShow = computed(
	() => title.value || extra || slots.extra || slots.title
);
</script>

<style lang="scss" scoped>
.my-card-container {
	position: relative;
	background: $background-1;
	border-radius: 12px;

	.my-card-section-wrapper {
		background: rgba(0, 0, 0, 0);
	}

	.my-card-title {
		font-size: 14px;
		font-style: normal;
		font-weight: 500;
		line-height: 20px;
		color: $input-text-color;
	}

	.my-card-extra {
		font-weight: 400;
		font-size: 12px;
		color: $ink-2;
	}
}
</style>

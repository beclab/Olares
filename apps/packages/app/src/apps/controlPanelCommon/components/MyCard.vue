<template>
	<q-card
		:flat="flat"
		:bordered="bordered"
		:dark="dark"
		:square="square"
		:tag="tag"
		class="my-card-container"
	>
		<q-card-section
			v-if="titleShow"
			class="my-card-section-wrapper q-py-sm q-px-lg"
		>
			<div class="row no-wrap justify-between items-center">
				<MyTitle size="lg">
					<slot name="title"></slot>
					<template v-if="title">
						{{ title }}
					</template>
				</MyTitle>
				<div class="col-auto my-card-extra">
					<slot name="extra" v-if="slots.extra"></slot>
					<template v-if="extra">
						{{ extra }}
					</template>
				</div>
			</div>
		</q-card-section>
		<q-card-section class="q-pa-lg">
			<slot></slot>
		</q-card-section>
	</q-card>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { toRefs, useSlots } from 'vue';
import MyTitle from './MyListItem/MyTitle.vue';

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
}

const slots = useSlots();

const props = withDefaults(defineProps<Props>(), {
	title: '',
	extra: ''
});

const { title, extra, dark, square, flat, bordered, tag } = toRefs(props);
const titleShow = computed(
	() => title.value || extra || slots.extra || slots.title
);
</script>

<style lang="scss" scoped>
.my-card-container {
	position: relative;
	.my-card-extra {
		color: #37f;
		font-size: 12px;
		font-weight: 500;
		line-height: 16px;
	}
}
</style>

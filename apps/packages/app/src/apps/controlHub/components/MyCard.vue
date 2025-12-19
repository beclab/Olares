<template>
	<q-card
		:flat="flat"
		:bordered="bordered"
		:dark="dark"
		:square="square"
		:tag="tag"
		class="my-card-container q-mx-md"
	>
		<q-card-section
			v-if="titleShow"
			class="my-card-section-wrapper q-px-none q-py-xs"
		>
			<div
				class="row no-wrap justify-between items-center"
				style="height: 32px"
			>
				<MyTitle size="sm" style="flex: 1">
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
		<q-card-section class="my-card-content q-px-lg q-pb-lg q-pt-none">
			<slot></slot>
		</q-card-section>
	</q-card>
</template>

<script setup lang="ts">
import MyTitle from '@apps/control-panel-common/src/components/MyListItem/MyTitle.vue';
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
	background: rgba(0, 0, 0, 0);
	.my-card-section-wrapper {
	}

	.my-card-title {
		font-weight: 500;
		font-size: 16px;
		line-height: 24px;
		color: $input-text-color;
	}

	.my-card-extra {
		font-weight: 400;
		font-size: 12px;
	}
	.my-card-content {
		border-radius: 12px;
	}
}
</style>

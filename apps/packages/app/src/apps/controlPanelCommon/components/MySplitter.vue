<template>
	<div class="my-splitter-container">
		<q-splitter
			v-model="splitterModel"
			unit="px"
			:limits="[120, Infinity]"
			v-bind="$attrs"
			class="my-splitter-wrapper"
		>
			<template v-slot:before>
				<slot name="before" v-bind:value="splitterModel"></slot>
			</template>
			<template v-slot:after>
				<slot name="after" v-bind:value="splitterModel"></slot>
			</template>
		</q-splitter>
	</div>
</template>

<script setup lang="ts">
import { ref, toRefs } from 'vue';

interface Props {
	limits?: [number, number];
	defaultValue?: number;
}
const props = withDefaults(defineProps<Props>(), {
	limits: [0, Infinity] as any,
	defaultValue: 240
});
const { limits } = toRefs(props);
const splitterModel = ref(props.defaultValue);
</script>

<style lang="scss" scoped>
$my_tree_width_base: 8px;
$my_tree_left: -($my_tree_width_base + 8px);
.my-splitter-container {
	&::v-deep(.q-tree) {
		padding: 0 20px;
	}
	&::v-deep(.q-tree--dense .q-tree__node--child > .q-tree__node-header:before) {
		width: 12px;
	}
	&::v-deep(.q-tree--dense .q-tree__node-header:before) {
		width: 12px;
	}
	&::v-deep(.q-tree--dense .q-tree__node--child) {
		padding-left: $my_tree_width_base;
	}
	&::v-deep(.q-tree--dense .q-tree__node--child > .q-tree__node-header:before) {
		left: $my_tree_left;
	}

	&::v-deep(.q-tree__node-header-content) {
		font-weight: 500;
		color: #1f1814;
	}
	&::deep(.q-tree__node-body .q-tree__node-header .q-focus-helper) {
		// right: -4px;
		// left: -4px;
		// width: auto;
	}

	&::v-deep(.q-tree--dense .q-tree__arrow) {
		color: $ink-1;
	}

	&::v-deep(.q-tree__children .q-tree__node-header-content) {
		font-weight: 400;
		color: #5c5551;
	}
	&::v-deep(.q-tree__children .q-tree__node-header-content .header-top) {
		margin-right: map-get($space-sm, x);
	}

	&::v-deep(.q-tree__node-header) {
		padding-top: 4px;
		padding-bottom: 4px;
	}
}
</style>

<template>
	<bt-custom-dialog
		size="medium"
		v-bind="$attrs"
		persistent
		class=""
		:ok="true"
		:cancel="true"
		@hide="emits('hide')"
		@show="emits('show')"
		:title="title"
		:okLoading="loading"
	>
		<slot></slot>
		<q-inner-loading :showing="loading" style="z-index: 999999">
		</q-inner-loading>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
interface Props {
	title?: string;
	offsetRatio?: number;
	ok?: string | boolean;
	cancel?: string | boolean;
	loading?: boolean;
}
const props = withDefaults(defineProps<Props>(), {
	offsetRatio: 0
});
const base = 6;

const emits = defineEmits<{
	(event: 'hide'): void;
	(event: 'show'): void;
}>();
</script>

<style lang="scss" scoped>
.my-dialog-wrapper {
	border-radius: 12px;
}
.my-dialog-title {
}
</style>

<template>
	<bt-custom-dialog
		v-model="visible"
		class=""
		:title="title"
		persistent
		@update:model-value="change"
		v-bind="$attrs"
		:full-height="fullHeight"
		:full-width="fullWidth"
		:ok="ok"
		:cancel="cancel"
		:okLoading="loading"
	>
		<slot></slot>
		<q-inner-loading :showing="loading" style="z-index: 999999">
		</q-inner-loading>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { ref, withDefaults, watch } from 'vue';

interface Props {
	title?: string;
	modelValue?: boolean;
	fullWidth?: boolean;
	fullHeight?: boolean;
	ok?: string | boolean;
	cancel?: string | boolean;
	loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
	fullWidth: true,
	fullHeight: true,
	ok: false,
	cancel: false
});
const visible = ref();
const emits = defineEmits(['update:modelValue']);
const show = () => {
	visible.value = true;
};

watch(
	() => props.modelValue,
	(newValue, oldValue) => {
		if (newValue != visible.value) {
			visible.value = props.modelValue;
		}
	},
	{
		immediate: true
	}
);

const close = () => {
	visible.value = false;
};

const change = () => {
	emits('update:modelValue', visible.value);
};

defineExpose({
	show,
	close,
	hide: close
});
</script>
<style lang="scss" scoped>
.container {
	height: calc(100%);
	border-radius: 12px;
	position: relative;
}
</style>

<template>
	<div
		class="request-button"
		:style="{ borderColor: getBorderColor }"
		:class="loading ? '' : 'cursor-pointer'"
		@click="onClick"
	>
		<div
			ref="text"
			v-show="!loading"
			class="text-body3"
			:class="`text-${color}`"
		>
			{{ label }}
		</div>
		<bt-loading :loading="loading" />
	</div>
</template>

<script setup lang="ts">
import BtLoading from '../base/BtLoading.vue';
import { computed, ref } from 'vue';

const props = defineProps({
	label: {
		type: String,
		default: ''
	},
	loading: {
		type: Boolean,
		default: false
	},
	color: {
		type: String,
		default: 'orange-default'
	}
});
const text = ref();
const emit = defineEmits(['request']);
const onClick = () => {
	if (props.loading) {
		return;
	}
	emit('request');
};

const getBorderColor = computed(() => {
	return `var(--q-${props.color})`;
});
</script>

<style scoped lang="scss">
.request-button {
	display: inline-flex;
	height: 32px;
	padding: 8px 14px;
	justify-content: center;
	align-items: center;
	gap: 8px;
	border-radius: 8px;
	border: 1px solid;
}
</style>

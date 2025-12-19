<template>
	<div
		v-show="modelValue"
		class="my-bottom-sheet-container"
		@click.prevent="closeHanlder"
	>
		<div class="my-bottom-sheet-wrapper bg-white" @click.stop>
			<div class="my-bottom-sheet-header row items-center justify-between">
				<span class="text-h5 text-grey-10">{{ title }}</span>
				<QBtnToggleStyle size="xl">
					<q-btn
						icon="sym_r_close"
						flat
						dense
						size="md"
						class="q-pa-none"
						color="grey-8"
						@click="closeHanlder"
					/>
				</QBtnToggleStyle>
			</div>
			<div class="my-bottom-sheet-content q-pa-md">
				<slot></slot>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import QBtnToggleStyle from './QBtnToggleStyle.vue';
interface Props {
	modelValue: boolean;
	title?: string;
}

withDefaults(defineProps<Props>(), {});
const emits = defineEmits(['update:modelValue']);

const closeHanlder = () => {
	emits('update:modelValue', false);
};
</script>

<style lang="scss" scoped>
.my-bottom-sheet-container {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	bottom: 0;
	z-index: 99999999999;
	background: rgba(31, 24, 20, 0.4);
	.my-bottom-sheet-wrapper {
		position: absolute;
		bottom: 0;
		left: 0;
		right: 0;
		border-radius: 20px 20px 0 0;
		.my-bottom-sheet-header {
			padding: 14px 20px;
		}
	}
}
</style>

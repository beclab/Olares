<template>
	<div class="my-bottom-sheet-container" @click.prevent="closeHanlder">
		<div class="my-bottom-sheet-wrapper bg-white" @click.stop>
			<div class="my-bottom-sheet-header row items-center justify-between">
				<span class="text-h5 text-grey-10">{{ title }}</span>
				<QBtnToggleStyle size="xl">
					<q-btn
						icon="sym_r_close"
						flat
						dense
						size="md"
						class="q-pa-none bottom-sheet-btn-wapper"
						color="grey-8"
						@click="closeHanlder"
					/>
				</QBtnToggleStyle>
			</div>
			<div class="relative-position my-bottom-sheet-content">
				<slot></slot>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import QBtnToggleStyle from './QBtnToggleStyle.vue';
interface Props {
	title?: string;
}

withDefaults(defineProps<Props>(), {});
const emits = defineEmits(['close']);

const closeHanlder = () => {
	emits('close', false);
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
	border: 1px solid var(---5, #dbdbdb);

	.my-bottom-sheet-wrapper {
		position: absolute;
		top: 0px;
		bottom: 0;
		left: 0;
		right: 0;
		border-radius: 20px 20px 0 0;
		// border: 1px solid var(---, #ebebeb);

		.my-bottom-sheet-header {
			padding: 14px 20px;
			border-bottom: 1px solid var(---5, #dbdbdb);
		}
		.my-bottom-sheet-content {
			height: calc(100% - 50px);
		}
	}
}
.bottom-sheet-btn-wapper {
	&.q-btn.text-grey-8:before {
		border: none !important;
	}
}
</style>

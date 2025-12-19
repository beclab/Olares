<template>
	<div
		class="bg-grey-11-hover root row items-center justify-between"
		:style="{
			'pointer-events': disable ? 'none' : '',
			border: check ? '1px solid $blue' : ''
		}"
		@click="toggle"
	>
		<div class="text-ink-2 walletName">
			{{ label }}
		</div>
		<div class="walletCheck row items-center justify-center">
			<img
				class="check"
				:src="check ? 'walletIcon/checked.svg' : 'walletIcon/unChecked.svg'"
				alt=""
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
defineProps({
	label: {
		type: String,
		default: 'ETH',
		required: false
	},
	check: {
		type: Boolean,
		default: false,
		required: false
	},
	disable: {
		type: Boolean,
		default: false,
		required: false
	}
});
const emit = defineEmits(['toggle']);

const toggle = () => {
	emit('toggle');
};
</script>

<style lang="scss" scoped>
.root {
	width: 100%;
	height: 44px;
	cursor: pointer;
	border: 1px solid $separator;
	border-radius: 10px;
	margin-bottom: 10px;

	.walletCheck {
		width: 52px;
		height: 44px;
		.check {
			width: 12px;
			height: 12px;
		}
	}
	.walletName {
		width: calc(100% - 52px);
		font-size: map-get($map: $body2, $key: size);
	}
}
</style>

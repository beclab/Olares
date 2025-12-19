<template>
	<div
		class="bt-checkbox row justify-start items-center cursor-pointer"
		@click.stop="emit('update:modelValue', !modelValue)"
	>
		<q-img
			class="bt-checkbox__img"
			:src="
				circle
					? modelValue
						? circleCheckImg
						: circleUncheckImg
					: modelValue
					? checkImg
					: uncheckImg
			"
		/>
		<div
			v-if="label"
			class="bt-checkbox__label text-body3 text-ink-2"
			:style="{ maxWidth: maxWidth ? maxWidth : 'calc(100% - 32px)' }"
		>
			{{ label }}
		</div>
	</div>
</template>

<script setup lang="ts">
defineProps({
	modelValue: {
		type: Boolean,
		required: true
	},
	label: {
		type: String,
		required: false
	},
	circle: {
		type: Boolean,
		default: false
	},
	maxWidth: {
		type: String,
		default: ''
	},
	circleCheckImg: {
		type: String,
		default: 'wise/imgs/circle_check_box.svg'
	},
	circleUncheckImg: {
		type: String,
		default: 'wise/imgs/circle_uncheck_box.svg'
	},
	checkImg: {
		type: String,
		default: 'wise/imgs/check_box.svg'
	},
	uncheckImg: {
		type: String,
		default: 'wise/imgs/uncheck_box.svg'
	}
});
const emit = defineEmits(['update:modelValue']);
</script>

<style lang="scss" scoped>
.bt-checkbox {
	height: 32px;
	padding: 8px;

	&__img {
		width: 16px;
		height: 16px;
	}

	&__label {
		margin-left: 8px;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
}
</style>

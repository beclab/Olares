<template>
	<div
		class="bt-checkbox row justify-start items-center cursor-pointer q-mt-md"
		@click="emit('update:modelValue', !modelValue)"
	>
		<q-img
			class="bt-checkbox__img"
			:style="{
				'--icon-width': size + 'px',
				'--icon-height': size + 'px'
			}"
			:src="
				modelValue
					? '/imgs/checkbox/profile_check_box.svg'
					: $q.dark.isActive
					? 'imgs/checkbox/uncheck_box_dark.svg'
					: 'imgs/checkbox/uncheck_box_light.svg'
			"
		/>
		<div
			v-if="label"
			class="bt-checkbox__label text-subtitle2 text-ink-1 q-ml-sm"
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
	size: {
		type: Number,
		default: 20,
		required: false
	}
});
const emit = defineEmits(['update:modelValue']);
</script>

<style lang="scss" scoped>
.bt-checkbox {
	height: 20px;

	&__img {
		width: var(--icon-width);
		height: var(--icon-height);
	}

	&__label {
		max-width: 170px;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
}
</style>

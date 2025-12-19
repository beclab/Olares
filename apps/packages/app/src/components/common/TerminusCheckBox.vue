<template>
	<div
		class="bt-checkbox row justify-start items-center cursor-pointer"
		@click.stop="itemAction"
	>
		<q-img
			class="bt-checkbox__img"
			:style="{
				'--icon-width': size + 'px',
				'--icon-height': size + 'px'
			}"
			:src="
				modelValue
					? activeImage
					: $q.dark.isActive
					? normalDarkImage
					: normalImage
			"
		/>
		<div
			v-if="label"
			class="bt-checkbox__label text-body2 text-ink-2 q-ml-sm"
			:class="titleClasses"
		>
			{{ label }}
		</div>
	</div>
</template>

<script setup lang="ts">
const props = defineProps({
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
		default: 16,
		required: false
	},
	activeImage: {
		type: String,
		required: false,
		default: 'img/checkbox/check_box.svg'
	},
	normalImage: {
		type: String,
		required: false,
		default: 'img/checkbox/uncheck_box_light.svg'
	},
	normalDarkImage: {
		type: String,
		required: false,
		default: 'img/checkbox/uncheck_box_dark.svg'
	},
	titleClasses: {
		type: String,
		required: false,
		default: ''
	},
	hookSelect: {
		type: Boolean,
		required: false,
		default: false
	}
});
const emit = defineEmits(['update:modelValue', 'itemClick']);

const itemAction = () => {
	console.log('111');

	if (props.hookSelect) {
		emit('itemClick');
		return;
	}
	console.log('222');
	emit('update:modelValue', !props.modelValue);
};
</script>

<style lang="scss" scoped>
.bt-checkbox {
	height: 20px;

	&__img {
		width: var(--icon-width);
		height: var(--icon-height);
	}

	&__label {
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
}
</style>

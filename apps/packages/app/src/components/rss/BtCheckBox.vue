<template>
	<div
		class="bt-checkbox row justify-start items-center"
		:class="disabled ? 'bt-checkbox--disabled' : 'cursor-pointer'"
		@click.stop="handleCheckboxClick"
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
			v-if="label || linkLabel"
			class="bt-checkbox__label text-body3 text-ink-2"
			:style="{ maxWidth: maxWidth ? maxWidth : 'calc(100% - 32px)' }"
		>
			<span v-if="label">{{ label }}</span>
			<a
				v-if="linkLabel && link"
				:href="link"
				target="_blank"
				class="bt-checkbox__link"
			>
				{{ linkLabel }}
			</a>
		</div>
	</div>
</template>

<script setup lang="ts">
const props = defineProps({
	modelValue: {
		type: Boolean,
		required: true
	},
	disabled: {
		type: Boolean,
		default: false
	},
	label: {
		type: String,
		required: false
	},
	link: {
		type: String,
		required: false,
		default: ''
	},
	linkLabel: {
		type: String,
		required: false,
		default: ''
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

const handleCheckboxClick = (e: MouseEvent) => {
	const target = e.target as HTMLElement;
	if (target.tagName === 'A' || target.closest('a')) {
		return;
	}
	if (props.disabled) {
		return;
	}
	emit('update:modelValue', !props.modelValue);
};
</script>

<style lang="scss" scoped>
.bt-checkbox {
	height: 32px;
	padding: 8px;

	&--disabled {
		cursor: not-allowed;
		opacity: 0.5;
	}

	&__img {
		width: 16px;
		height: 16px;
	}

	&__link {
		color: $blue-default;
		text-decoration: underline;
		cursor: pointer;
		white-space: nowrap;
		margin-left: 4px;
	}

	&__label {
		margin-left: 8px;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
}
</style>

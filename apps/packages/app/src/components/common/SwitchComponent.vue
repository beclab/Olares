<template>
	<div
		class="toggle-switch text-caption row items-center justify-between q-px-sm"
		:class="{ checked: modelValue, disabled: disabled }"
		:data-title="title"
		@click.stop="toggle"
		:style="{ width: dynamicSwitchWidth + 'px' }"
	>
		<span ref="titleMeasureRef" class="title-measure">
			{{ title }}
		</span>

		<div class="text-caption on">
			{{ t('ON') }}
		</div>
		<div class="text-caption text-ink-3 off">
			{{ t('OFF') }}
		</div>
	</div>
</template>

<script setup>
import { ref, onMounted, watch, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	modelValue: {
		required: true,
		type: Boolean
	},
	title: {
		type: String,
		default: '',
		required: false
	},
	disabled: {
		type: Boolean,
		default: false,
		required: false
	}
});

const { t } = useI18n();
const emit = defineEmits(['update:modelValue']);

const titleMeasureRef = ref(null);
const dynamicSwitchWidth = ref(0);

const calculateSwitchWidth = () => {
	if (!titleMeasureRef.value) return;
	nextTick(() => {
		const titleActualWidth = titleMeasureRef.value.offsetWidth;
		const extraMargin = 40;
		dynamicSwitchWidth.value = titleActualWidth + extraMargin;
	});
};

onMounted(calculateSwitchWidth);
watch(() => props.title, calculateSwitchWidth);

const toggle = () => {
	if (props.disabled) return;
	emit('update:modelValue', !props.modelValue);
};
</script>

<style scoped lang="scss">
.toggle-switch {
	background: $background-3;
	border-radius: 14px;
	cursor: pointer;
	flex: none;
	height: 28px;
	position: relative;
	transition: background-color 150ms;

	&.disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.on {
		color: #ffffff66;
		margin-left: 2px;
	}

	.off {
		margin-right: 2px;
	}
}

.toggle-switch::before {
	background: #fff;
	border-radius: 10px;
	color: $ink-2;
	display: flex;
	align-items: center;
	justify-content: center;
	height: calc(100% - 8px);
	left: 4px;
	content: attr(data-title);
	position: absolute;
	top: 4px;
	transition: left 150ms;
	padding: 0 8px;
	will-change: left;
	white-space: nowrap;
}

.checked {
	background-color: $light-blue-default;
}

.checked::before {
	left: 36px;
}

.title-measure {
	position: absolute;
	left: -9999px;
	top: -9999px;
	visibility: hidden;
	white-space: nowrap;
	padding: 0 8px;
	font-size: inherit;
	font-family: inherit;
	font-weight: inherit;
}
</style>

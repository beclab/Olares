<template>
	<div
		class="column justify-center items-center"
		:class="icon || text || auto ? '' : 'full-width'"
		@click="handleParentClick"
	>
		<div
			:style="{
				'--root-padding': border ? '2px' : '0',
				'--root-border': `${border ? 2 : 0}px solid ${
					selected ? greenDefault.color.value : 'transparent'
				}`,
				width: border || auto ? 'auto' : '100%'
			}"
			class="picker-root column justify-center items-center"
		>
			<div
				:style="{
					width: icon || text ? '40px' : width,
					height: icon || text ? '40px' : height,
					'--content-border': `2px solid ${
						border
							? separator.color.value
							: selected
							? greenDefault.color.value
							: separator.color.value
					}`
				}"
				class="picker-border column justify-center items-center"
			>
				<slot
					name="default"
					:color="selected ? greenDefault.color.value : separator.color.value"
					:selected="selected"
					:fontSize="fontSize"
				/>

				<q-icon
					v-if="icon"
					:size="size"
					:color="selected ? 'light-green-default' : 'ink2'"
					:name="icon"
				/>

				<div
					v-if="text"
					class="picker-text text-body1"
					:class="selected ? 'text-light-green-default' : 'ink2'"
				>
					{{ text }}
				</div>
			</div>
		</div>

		<div v-if="label">
			<slot name="label" :label="label" :selected="selected" />
		</div>
	</div>
</template>

<script lang="ts" setup>
import { computed, getCurrentInstance, inject, onMounted, ref } from 'vue';
import { useColor } from '@bytetrade/ui';

const props = defineProps({
	value: {
		type: [String, Number],
		require: true
	},
	width: {
		type: String,
		default: '100%'
	},
	height: {
		type: String,
		default: 'auto'
	},
	isSelected: {
		type: Boolean,
		default: false
	},
	border: {
		type: Boolean,
		required: false
	},
	auto: {
		type: Boolean,
		required: false
	},
	insideBorder: {
		type: Boolean,
		default: true
	},
	icon: {
		type: String,
		required: false
	},
	text: {
		type: String,
		required: false
	},
	size: {
		type: String,
		default: '20px'
	},
	label: {
		type: String,
		required: false
	}
});

const addChildComponent: any = inject('addChildComponent');
const handleItemClick: any = inject('handleItemClick');
const selected = ref(true);
const selfRef = ref(getCurrentInstance());
const emit = defineEmits(['onSelected']);
const greenDefault = useColor('light-green-default');
const separator = useColor('separator');

const fontSize = computed(() => {
	if (props.width?.indexOf('px') > 0) {
		let number = parseInt(props.width?.substring(0, props.width?.length - 2));
		return (number * 16) / 375;
	}
	return 5.12;
});

const setSelected = (isSelected: boolean) => {
	selected.value = isSelected;
	if (isSelected) {
		emit('onSelected', props.value);
	}
};

defineExpose({ setSelected });

const handleParentClick = (event) => {
	if (handleItemClick && props.value !== '') {
		handleItemClick(selfRef, props.value);
	}
};

onMounted(() => {
	selected.value = props.isSelected;
	if (addChildComponent && props.value !== '') {
		addChildComponent(selfRef, props.value);
	}
});
</script>

<style scoped lang="scss">
.picker-root {
	height: auto;
	cursor: pointer;
	border-radius: 16px;
	padding: var(--root-padding);
	border: var(--root-border);

	.picker-border {
		border-radius: 12px;
		overflow: hidden;
		border: var(--content-border);

		.picker-text {
			width: 36px;
			height: 40px;
			display: flex;
			justify-content: center;
			align-items: center;
		}
	}
}
</style>

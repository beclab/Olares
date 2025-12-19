<template>
	<div
		ref="buttonRef"
		:class="selected ? 'background-picker-selected' : 'background-picker'"
		class="column justify-center items-center"
		@click="handleClick"
	>
		<div
			style="width: 116px; height: 232px; font-size: 6px; position: absolute"
			:style="selected ? 'top: 3px;left: 3px' : ''"
		>
			<slot />
		</div>

		<q-icon
			v-if="icon"
			size="32px"
			:color="selected ? 'ink-1' : 'ink-3'"
			:name="icon"
			@click="emit('onIconClick')"
		/>
		<div
			v-if="label"
			:class="
				selected
					? 'background-picker-selected__label'
					: 'background-picker__label'
			"
			@click="emit('onLabelClick')"
		>
			{{ label }}
		</div>
	</div>
</template>

<script lang="ts" setup>
import { getCurrentInstance, inject, onMounted, ref } from 'vue';

defineProps({
	icon: {
		type: String,
		required: false
	},
	label: {
		type: String,
		required: false
	}
});

const addChildComponent = inject('addChildComponent');
const handleItemClick = inject('handleItemClick');
const selected = ref(true);
const selfRef = ref(getCurrentInstance());
const emit = defineEmits(['onSelected', 'onIconClick', 'onLabelClick']);

const setSelected = (isSelected: boolean) => {
	selected.value = isSelected;
	if (isSelected) {
		emit('onSelected');
	}
};

defineExpose({ setSelected });

const handleClick = () => {
	if (handleItemClick) {
		handleItemClick(selfRef);
	}
};

onMounted(() => {
	if (addChildComponent) {
		addChildComponent(selfRef);
	}
});
</script>

<style scoped lang="scss">
.background-picker-selected {
	background: $background-3;
	border: 2px solid $green-default;
	height: 242px;
	width: 126px;
	justify-content: center;
	align-items: center;
	border-radius: 12px;
	cursor: pointer;
	text-decoration: none;
	position: relative;

	&__label {
		z-index: 10;
		margin-top: 4px;
		font-family: Roboto;
		font-size: 10px;
		font-weight: 400;
		line-height: 12px;
		letter-spacing: 0em;
		text-align: center;
		overflow: hidden;
		color: $ink-1;
		text-align: center;
		text-overflow: ellipsis;
	}
}

.background-picker {
	background: $background-3;
	border: 1px solid $separator;
	height: 234px;
	width: 118px;
	justify-content: center;
	align-items: center;
	border-radius: 12px;
	cursor: pointer;
	text-decoration: none;
	position: relative;

	&__label {
		z-index: 10;
		margin-top: 4px;
		font-family: Roboto;
		font-size: 10px;
		font-weight: 400;
		line-height: 12px;
		letter-spacing: 0em;
		text-align: center;
		overflow: hidden;
		color: $ink-3;
		text-align: center;
		text-overflow: ellipsis;
	}
}
</style>

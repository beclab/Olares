<template>
	<div class="row items-center justify-end">
		<div class="text-body3 text-ink-2 cpu-core q-mr-md">{{ label }}</div>
		<div class="column items-center justify-center" style="font-weight: 600">
			<div
				class="icon-parent column justify-center items-center"
				:class="Number(modelValue) < max ? '' : 'disable-hover'"
				@click="up"
				title="increase"
			>
				<q-icon
					size="8px"
					name="sym_r_expand_less"
					:class="Number(modelValue) < max ? 'text-ink-2' : 'text-background-5'"
				/>
			</div>
			<div
				class="icon-parent column justify-center items-center"
				:class="Number(modelValue) > min ? '' : 'disable-hover'"
				@click="down"
				title="decrease"
			>
				<q-icon
					size="8px"
					name="sym_r_expand_more"
					:class="Number(modelValue) > min ? 'text-ink-2' : 'text-background-5'"
				/>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
const props = defineProps({
	label: {
		type: String,
		default: '',
		required: false
	},
	modelValue: {
		type: String,
		required: true
	},
	min: {
		type: Number,
		required: false,
		default: 0
	},
	max: {
		type: Number,
		required: false,
		default: 100
	},
	unit: {
		type: Number,
		required: false,
		default: 1
	}
});

const up = () => {
	if (Number(props.modelValue) >= props.max) {
		return;
	}
	updateModelValue(Number(props.modelValue) + props.unit);
};

const down = () => {
	if (Number(props.modelValue) <= props.min) {
		return;
	}
	updateModelValue(Number(props.modelValue) - props.unit);
};

const updateModelValue = (value: number) => {
	emit('update:modelValue', value);
};

const emit = defineEmits(['update:modelValue']);
</script>

<style scoped lang="scss">
.icon-parent {
	width: 10px;
	height: 10px;

	&:hover {
		background: $background-3;
		filter: brightness(0.9);
	}
}

.disable-hover {
	pointer-events: none;
}
</style>

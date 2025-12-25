<template>
	<div class="color-picker row justify-between items-center q-mt-md">
		<div class="text-body1 text-ink-3">{{ label }}</div>
		<div class="row justify-end items-center">
			<q-input
				borderless
				input-class="text-body3 text-ink-1"
				class="color-edit q-ml-lg"
				:model-value="modelValue"
				@update:model-value="updateModelValue"
			/>
			<div
				class="color-picker-selected q-ml-sm"
				:style="{ background: modelValue }"
			>
				<q-popup-proxy cover transition-show="scale" transition-hide="scale">
					<q-color
						:model-value="modelValue"
						no-header
						no-footer
						@update:model-value="updateModelValue"
						default-view="palette"
					/>
				</q-popup-proxy>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
defineProps({
	label: {
		type: String,
		require: true
	},
	modelValue: {
		type: String,
		require: true
	}
});

const emit = defineEmits(['update:modelValue']);

const updateModelValue = (data: string) => {
	emit('update:modelValue', data);
};
</script>

<style scoped lang="scss">
.color-picker {
	width: 100%;
	height: auto;

	.color-edit {
		padding: 8px 12px;
		border-radius: 8px;
		border: 1px solid $input-stroke;
		width: 100px;
	}

	.color-picker-selected {
		width: 32px;
		height: 32px;
		border-radius: 50%;
		border: 1px solid $separator;
	}
}
</style>

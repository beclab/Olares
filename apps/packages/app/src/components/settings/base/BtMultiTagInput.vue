<template>
	<div class="multi-tag-input">
		<div v-if="modelValue.length > 0" class="multi-tag-chips row items-center">
			<div
				v-for="(tag, index) in modelValue"
				:key="`${tag}-${index}`"
				class="multi-tag-chip text-body3"
			>
				<span class="multi-tag-chip-label">{{ tag }}</span>
				<q-icon
					name="sym_r_close"
					size="14px"
					class="multi-tag-chip-remove"
					@click.stop="removeAt(index)"
				/>
			</div>
		</div>
		<terminus-edit
			:model-value="buffer"
			:hint-text="placeholder"
			:is-error="isError"
			:error-message="errorMessage"
			input-class="text-body3 text-ink-1"
			class="full-width"
			@update:modelValue="onBufferChange"
			@submit="addTag"
		>
			<template #right>
				<q-icon
					size="20px"
					name="sym_r_add"
					class="multi-tag-add"
					:class="canAdd ? 'multi-tag-add--active' : 'multi-tag-add--disabled'"
					@click="addTag"
				/>
			</template>
		</terminus-edit>
	</div>
</template>

<script lang="ts" setup>
import TerminusEdit from './TerminusEdit.vue';
import { computed, PropType, ref } from 'vue';

const props = defineProps({
	modelValue: {
		type: Array as PropType<string[]>,
		default: () => []
	},
	splitter: {
		type: String,
		default: ','
	},
	isError: {
		type: Boolean,
		default: false,
		required: false
	},
	errorMessage: {
		type: String,
		default: '',
		required: false
	},
	placeholder: {
		type: String,
		default: '',
		required: false
	}
});

const emit = defineEmits(['update:modelValue']);

const buffer = ref('');

const canAdd = computed(() => buffer.value.trim() !== '');

function onBufferChange(value: string) {
	buffer.value = value ?? '';
}

function addTag() {
	const sp = props.splitter || ',';
	const parts = buffer.value
		.split(sp)
		.map((part) => part.trim())
		.filter((part) => part !== '');
	if (parts.length === 0) {
		buffer.value = '';
		return;
	}
	const next = [...props.modelValue];
	for (const part of parts) {
		if (!next.includes(part)) {
			next.push(part);
		}
	}
	if (next.length !== props.modelValue.length) {
		emit('update:modelValue', next);
	}
	buffer.value = '';
}

function removeAt(index: number) {
	const next = props.modelValue.filter((_, i) => i !== index);
	emit('update:modelValue', next);
}
</script>

<style scoped lang="scss">
.multi-tag-input {
	width: 100%;
}

.multi-tag-chips {
	gap: 4px;
	flex-wrap: wrap;
	margin-bottom: 8px;
}

.multi-tag-chip {
	display: flex;
	align-items: center;
	gap: 4px;
	max-width: 100%;
	padding: 2px 6px;
	border-radius: 4px;
	background: $background-3;
	color: $ink-2;

	.multi-tag-chip-label {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.multi-tag-chip-remove {
		flex-shrink: 0;
		cursor: pointer;
		color: $ink-3;

		&:hover {
			color: $ink-1;
		}
	}
}

.full-width {
	width: 100%;
}

.multi-tag-add {
	&--active {
		cursor: pointer;
		color: $blue-default;
	}

	&--disabled {
		cursor: not-allowed;
		color: $ink-3;
	}
}
</style>

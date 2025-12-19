<template>
	<div class="form-item row q-mb-md">
		<div
			class="form-item-key text-subtitle2 text-ink-1"
			:style="{ width: `${keyWidth}px` }"
		>
			<span>{{ name }} </span>
			<span v-if="required" class="text-red-default q-ml-xs">*</span>
			<!-- <q-icon
				v-if="tip"
				class="text-ink-3 q-ml-xs"
				name="sym_r_help"
				size="16px"
			>
				<q-tooltip> {{ tip }} </q-tooltip>
			</q-icon> -->
		</div>
		<div class="form-item-value q-mb-lg">
			<slot></slot>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { defineProps } from 'vue';

interface Props {
	name: string;
	required: boolean;
	tip: string;
	keyWidth: string | number;
}
withDefaults(defineProps<Props>(), {
	required: true,
	tip: '',
	keyWidth: 140
});
</script>

<style lang="scss" scoped>
.form-item {
	.form-item-key {
		height: 40px;
		line-height: 40px;
		text-align: right;
		padding-right: 20px;
	}
	.form-item-value {
		flex: 1;
	}
}
</style>

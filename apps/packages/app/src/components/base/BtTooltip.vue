<template>
	<q-tooltip class="tooltip-normal row text-no-wrap" :offset="[0, 4]">
		<div class="tooltip-text" :style="{ maxWidth: maxWidth, textAlign: align }">
			{{ label }}
		</div>
		<bt-hot-key-icon
			v-if="hotkey.length > 0"
			class="q-ml-md"
			:hotkey="hotkey"
		/>
	</q-tooltip>
</template>

<script lang="ts" setup>
import BtHotKeyIcon from 'src/components/base/BtHotKeyIcon.vue';

defineProps({
	label: {
		type: String,
		require: false
	},
	hotkey: {
		type: String,
		default: ''
	},
	maxWidth: {
		type: String,
		default: 'auto'
	},
	align: {
		type: String as () => 'center' | 'start' | 'end',
		default: 'center'
	}
});
</script>

<style lang="scss" scoped>
.tooltip-text {
	display: inline-block;
	word-wrap: break-word;
	white-space: normal;
}
</style>

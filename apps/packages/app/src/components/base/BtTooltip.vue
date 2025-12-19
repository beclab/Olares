<template>
	<q-tooltip class="tooltip-normal row text-no-wrap" :offset="[0, 4]">
		<span>{{ label }}</span>
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
	}
});
</script>

<style lang="scss" scoped></style>

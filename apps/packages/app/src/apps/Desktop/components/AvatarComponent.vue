<template>
	<div
		class="desktop_avatar"
		:style="{
			width: `${size}px`,
			height: `${size}px`,
			borderRadius: `${size / 2}px`
		}"
	>
		<TerminusAvatar :info="tokenStore.terminus" :size="size" />
	</div>
</template>

<script lang="ts" setup>
import { useTokenStore } from '../../../stores/desktop/token';

const tokenStore = useTokenStore();

defineProps({
	size: {
		type: Number,
		required: false,
		default: 48
	}
});
</script>

<style lang="scss" scoped>
.desktop_avatar {
	border-radius: 50%;
	margin-bottom: 12px;
	cursor: pointer;
	overflow: hidden;
}
</style>

<template>
	<div
		class="relative-position terminus-account-avatar-bex"
		@mouseenter="enterHandler"
		@mouseleave="leaveHandler"
	>
		<slot></slot>
		<div
			class="user-list-popover q-px-md q-py-sm bg-tooltip-bg text-body3 text-ink-tooltip"
			v-if="$slots.tooltip"
		>
			<slot name="tooltip"></slot>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';

const visible = ref(false);

const enterHandler = () => {
	visible.value = true;
};

const leaveHandler = () => {
	visible.value = false;
};
</script>

<style lang="scss" scoped>
.terminus-account-avatar-bex {
	.user-list-popover {
		position: absolute;
		left: 0;
		// right: 0;
		min-width: 100%;
		top: calc(100% + 4px);
		border-radius: 8px;
		box-shadow: 0px 4px 10px 0px rgba(0, 0, 0, 0.2);
		display: none;
		white-space: nowrap;
	}
	&::after {
		content: '';
		position: absolute;
		top: 100%;
		left: 0;
		right: 0;
		height: 6px;
	}
	&:hover {
		.user-list-popover {
			display: block;
		}
	}
}
</style>

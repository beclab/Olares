<template>
	<div class="app-avatar-stack row items-center no-wrap">
		<div
			v-for="app in visibleApps"
			:key="app.appName"
			class="app-avatar-stack__item"
		>
			<q-img no-spinner :src="app.icon" class="app-avatar-stack__avatar">
				<q-tooltip>{{ app.title }}</q-tooltip>
			</q-img>
			<span v-if="showDot(app)" class="app-avatar-stack__dot" />
		</div>
		<span
			v-if="overflowCount > 0"
			class="app-avatar-stack__more text-body3 text-ink-2"
		>
			+{{ overflowCount }}
		</span>
	</div>
</template>

<script setup lang="ts">
import { APP_STATUS } from 'src/constant/constants';
import { AcceleratorAppView } from 'src/stores/settings/accelerator';
import { computed, PropType } from 'vue';

const props = defineProps({
	apps: {
		type: Array as PropType<AcceleratorAppView[]>,
		default: () => []
	},
	/** Max avatars before collapsing the rest into a "+N" chip. */
	max: {
		type: Number,
		default: 4
	},
	/** Show a running-status dot on each running app's avatar. */
	showStatusDot: {
		type: Boolean,
		default: true
	}
});

const visibleApps = computed(() => props.apps.slice(0, props.max));
const overflowCount = computed(() =>
	Math.max(props.apps.length - props.max, 0)
);

const showDot = (app: AcceleratorAppView): boolean =>
	props.showStatusDot && app.state === APP_STATUS.RUNNING;
</script>

<style scoped lang="scss">
.app-avatar-stack {
	gap: 4px;

	&__item {
		position: relative;
		width: 24px;
		height: 24px;
		flex-shrink: 0;
	}

	&__avatar {
		width: 24px;
		height: 24px;
		border-radius: 6px;
	}

	&__dot {
		position: absolute;
		right: -2px;
		bottom: -2px;
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: $green-default;
		border: 1.5px solid $background-1;
	}

	&__more {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		min-width: 24px;
		height: 24px;
		padding: 0 6px;
		border-radius: 4px;
		background: $background-3;
	}
}
</style>

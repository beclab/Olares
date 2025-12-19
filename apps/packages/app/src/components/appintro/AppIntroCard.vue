<template>
	<div
		v-if="contentSlot || description"
		class="app-intro-card-root bg-background-1"
		:style="{ '--cardPadding': deviceStore.isMobile ? '0' : '20px' }"
	>
		<div class="app-intro-card-title text-subtitle2 text-ink-1">
			{{ title }}
		</div>
		<slot v-if="contentSlot" name="content" />
		<expend-text-view :display-line="10" :text="description" />
	</div>
</template>

<script lang="ts" setup>
import { useSlots } from 'vue';
import ExpendTextView from '@apps/market/src/components/appintro/ExpendTextView.vue';
import { useDeviceStore } from '../../stores/settings/device';

defineProps({
	title: {
		type: String,
		default: '',
		require: false
	},
	description: {
		type: String,
		default: '',
		require: false
	}
});

const contentSlot = !!useSlots().content;
const deviceStore = useDeviceStore();
</script>

<style scoped lang="scss">
.app-intro-card-root {
	border-radius: 12px;
	width: 100%;
	height: auto;
	padding: var(--cardPadding);

	.app-intro-card-title {
		margin-bottom: 12px;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		max-width: 100%;
	}
}
</style>

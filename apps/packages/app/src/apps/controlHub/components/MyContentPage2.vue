<template>
	<div class="absolute-full bg-background-1" style="z-index: 999">
		<RouteBackBar
			class="my-content-page-header-extra"
			size="md"
			:title="title"
			:avatar="avatar"
		>
			<template #extra>
				<slot name="extra"></slot>
			</template>
		</RouteBackBar>
		<bt-scroll-area class="my-scroll-area-wrapper">
			<slot></slot>
		</bt-scroll-area>
	</div>
</template>

<script setup lang="ts">
interface Props {
	title: string;
	avatar?: string;
}
withDefaults(defineProps<Props>(), {});
import RouteBackBar from '@apps/control-panel-common/src/components/RouteBackBar.vue';
</script>

<style lang="scss" scoped>
.my-content-page-header-extra {
	position: absolute;
	top: 0;
	right: 8px;
	height: $content-header-height;
	z-index: 2;
	display: flex;
	align-items: center;
	left: 0;
}
.my-scroll-area-wrapper {
	height: calc(100vh - #{$content-header-height});
	margin-top: $content-header-height;
	background: $background-6;
	& > ::v-deep(.q-scrollarea__container) {
		min-width: 0px;
		overflow-x: hidden;
	}

	& > ::v-deep(.q-scrollarea__container > .q-scrollarea__content) {
		min-width: 0px;
		width: 100% !important;
	}
}
</style>

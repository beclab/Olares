<template>
	<div class="fixed-full bg-background-1" style="z-index: 999">
		<RouteBackBar :title="title" :avatar="avatar">
			<template #extra>
				<slot name="extra"></slot>
			</template>
		</RouteBackBar>
		<bt-scroll-area style="height: calc(100vh - 56px)">
			<MyLayoutContainer>
				<slot></slot>
			</MyLayoutContainer>
		</bt-scroll-area>
	</div>
</template>

<script setup lang="ts">
interface Props {
	title: string;
	avatar?: string;
}
withDefaults(defineProps<Props>(), {});
import RouteBackBar from './RouteBackBar.vue';
import MyLayoutContainer from './MyLayoutContainer.vue';
</script>

<style></style>

<template>
	<div
		class="windows-title-view-root row items-center justify-end"
		:style="`height:${height}px;`"
	>
		<div class="windows-title-view-root__title">
			<slot name="title" />
		</div>
		<windows-header-actions
			v-if="$q.platform.is.electron && $q.platform.is.win"
		/>
	</div>
</template>

<script setup lang="ts">
import WindowsHeaderActions from './WindowsHeaderActions.vue';

defineProps({
	height: {
		type: Number,
		default: 38,
		required: false
	}
});
</script>

<style scoped lang="scss">
.windows-title-view-root {
	width: 100%;
	height: 100%;
	-webkit-app-region: drag;
	-webkit-user-select: none;
	&__title {
		height: 100%;
	}
}
</style>

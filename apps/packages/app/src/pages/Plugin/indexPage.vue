<template>
	<div class="container">
		<comment-area />
	</div>
	<router-view></router-view>
</template>

<script lang="ts" setup>
import CommentArea from './CommentArea.vue';
</script>

<style lang="scss" scoped>
.container {
	width: 100%;
	height: 100%;
	position: absolute;
	left: 0;
	top: 0;
	// padding: 32px 20px 20px;
	background: $white;
	border-radius: 12px;
}
</style>

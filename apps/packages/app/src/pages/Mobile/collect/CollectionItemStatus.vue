<template>
	<div class="item-status-bg row items-center justify-center">
		<div class="collection-wrapper-operation">
			<slot name="status" />
		</div>
	</div>
</template>

<script setup lang="ts"></script>

<style scoped lang="scss">
.item-status-bg {
	width: 32px;
	height: 32px;
	border-radius: 8px;
}
.collection-wrapper-operation {
	padding: 6px;
	border-radius: 8px;
	background: $background-2;
	&:hover {
		background: $background-4;
	}
}
</style>

<template>
	<q-card class="simple-card-container" flat>
		<div class="row items-center q-mb-md">
			<q-img
				:src="avatar"
				:ratio="1"
				width="20px"
				spinner-size="xs"
				class="img"
			/>
			<span class="text-subtitle2 text-grey-10 q-ml-sm">{{ title }}</span>
		</div>
		<slot></slot>
	</q-card>
</template>

<script setup lang="ts">
interface Props {
	avatar: string;
	title: string;
}
withDefaults(defineProps<Props>(), {});
</script>

<style lang="scss" scoped>
.simple-card-container {
	border-radius: 0px;
	.img {
		border-radius: 50%;
	}
}
</style>

<template>
	<div class="func-btn-bg row items-center justify-end">
		<div class="func-btn" @click="funcOnClick">
			<div class="title setting-major-text">
				{{ title }}
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
withDefaults(
	defineProps<{
		title: string;
	}>(),
	{
		title: ''
	}
);

const funcOnClick = () => {
	emit('funcClick');
};

const emit = defineEmits(['funcClick']);
</script>

<style scoped lang="scss">
.func-btn-bg {
	.func-btn {
		height: 24px;
		background: #ffffff;
		border-radius: 8px;
		border: 1px solid #e0e2e9;
		width: auto;

		.title {
			font-size: 12px;
			font-weight: 400;
			line-height: 24px;
			margin-left: 16px;
			margin-right: 16px;
		}
	}

	.func-btn:hover {
		background-color: #f5f8fb;
	}
}
</style>

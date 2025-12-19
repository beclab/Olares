<template>
	<div class="dialog-common">
		<div class="row justify-between items-center">
			<span class="title setting-major-text"> {{ title }}</span>
			<q-btn flat @click="close">
				<q-icon name="close" size="16px" />
			</q-btn>
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

const close = () => {
	emit('closeAction');
};

const emit = defineEmits(['closeAction']);
</script>

<style scoped lang="scss">
.dialog-common {
	padding-top: 10px;
	width: 100%;
	padding-bottom: 10px;

	.title {
		height: 30px;
		font-size: 14px;
		font-weight: 500;
		line-height: 30px;
	}
}
</style>

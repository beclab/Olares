<template>
	<div
		class="row items-center justify-between full-width q-px-md"
		style="height: 64px"
	>
		<div class="row items-center full-height">
			<slot name="avatar" />
			<div class="text-body1 text-ink-2 q-ml-sm">
				{{ name }}
			</div>
		</div>
		<q-icon name="sym_r_close" color="ink-2" size="24px" @click="onClose" />
	</div>
</template>

<script setup lang="ts">
defineProps({
	name: {
		type: String,
		required: false,
		default: ''
	}
});

const onClose = () => {
	emits('close');
};

const emits = defineEmits(['close']);
</script>

<style scoped lang="scss">
.back-btn {
	width: 32px;
	height: 32px;
	border-radius: 8px;
	margin-right: 4px;

	&:hover {
		background-color: $background-3;
	}
}
</style>

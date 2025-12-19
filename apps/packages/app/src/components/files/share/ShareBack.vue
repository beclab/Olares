<template>
	<div class="row items-center">
		<div class="row items-center" style="flex: 0 0 32px">
			<div
				class="back-btn row items-center justify-center"
				@click="backToPrePage"
			>
				<q-icon
					name="sym_r_arrow_back_ios_new"
					size="20px"
					class="text-ink-2"
				/>
			</div>
		</div>
		<div
			class="text-h6 text-ink-1"
			style="flex: auto"
			:class="platform == 'mobile' ? ' row items-center justify-center' : ''"
		>
			{{ title }}
		</div>
		<slot name="end" />
	</div>
</template>

<script setup lang="ts">
defineProps({
	title: {
		type: String,
		required: false,
		default: ''
	},
	platform: {
		type: String,
		required: false,
		default: 'mobile'
	}
});

const backToPrePage = () => {
	emits('back');
};

const emits = defineEmits(['back']);
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

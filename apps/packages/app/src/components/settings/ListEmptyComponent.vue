<template>
	<div class="list-empty-root q-mt-md column items-center">
		<div class="bg-background-3 circle-bg column justify-center items-center">
			<q-icon size="20px" :name="icon" color="ink-2" />
		</div>
		<div class="text-ink-3 text-center q-mt-md">
			{{ message }}
		</div>
		<q-btn
			dense
			class="click-button q-px-md q-py-sm q-mt-lg text-body3 text-ink-2"
			:label="buttonLabel"
			@click="emit('onButtonClick')"
		/>
	</div>
</template>

<script setup lang="ts">
defineProps({
	icon: {
		type: String,
		default: ''
	},
	message: {
		type: String,
		default: ''
	},
	buttonLabel: {
		type: String,
		default: ''
	}
});

const emit = defineEmits(['onButtonClick']);
</script>

<style scoped lang="scss">
.list-empty-root {
	padding: 40px 20px;
	border: 1px solid $separator;
	border-radius: 12px;

	.circle-bg {
		width: 40px;
		height: 40px;
		border-radius: 50%;
	}

	.click-button {
		border: solid 1px $btn-stroke;
	}
}
</style>

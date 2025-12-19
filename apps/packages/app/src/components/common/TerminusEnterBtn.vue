<template>
	<div class="enter-btn-root row justify-center items-center" @click="action">
		<q-icon v-if="icon && icon.length > 0" :name="icon" size="20px" />
		<div
			class="enter-btn-root__title text-body3"
			:style="icon && icon.length > 0 ? 'margin-left: 8px;' : ''"
		>
			{{ title }}
		</div>
	</div>
</template>

<script setup lang="ts">
defineProps({
	icon: {
		type: String,
		required: false
	},
	title: {
		type: String,
		required: false,
		default: ''
	}
});

const emit = defineEmits(['sureAction']);

const action = () => {
	emit('sureAction');
};
</script>

<style scoped lang="scss">
.enter-btn-root {
	border: 1px solid $yellow;
	background: $yellow;
	border-radius: 8px;
	padding: 0 12px;
	height: 32px;

	&__title {
		text-align: left;
		color: $grey-10;
	}
}
</style>

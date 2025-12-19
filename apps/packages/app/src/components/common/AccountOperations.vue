<template>
	<div
		class="account-operation-root row justify-between items-center"
		:class="$q.dark.isActive ? 'dark-bg' : 'light-bg '"
	>
		<div class="row justify-start items-center">
			<div class="account-operation-root__img row justify-center items-center">
				<q-icon :name="`sym_r_${imageName}`" size="20px" />
			</div>

			<div class="account-operation-root__title text-subtitle1">
				{{ title }}
			</div>
		</div>

		<q-icon name="sym_r_keyboard_arrow_right" size="20px" />
	</div>
</template>

<script setup lang="ts">
defineProps({
	imageName: {
		type: String,
		required: true
	},
	title: {
		type: String,
		required: true
	}
});
</script>

<style scoped lang="scss">
.account-operation-root {
	width: 100%;
	padding: 20px;
	border-radius: 8px;

	&__img {
		width: 32px;
		height: 32px;
		border-radius: 50%;
		background-color: $yellow;
		color: $grey-10;
	}

	&__title {
		margin-left: 12px;
		color: $ink-1;
	}
}

.light-bg {
	background: linear-gradient(0deg, $yellow-1, $yellow-1);
}

.dark-bg {
	background: $background-3;
}
</style>

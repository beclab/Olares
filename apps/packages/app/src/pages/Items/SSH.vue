<template>
	<div class="row items-center justify-start totp-item">
		<q-icon name="sym_r_terminal" class="q-mr-sm" size="24px" color="grey-9" />
		<div class="totp-text text-h6 text-light-blue-default" v-if="dispaly">
			{{ password }}
		</div>
		<span class="totp-hidden" v-else>
			<i v-for="index in password.length" :key="index" class="bg-ink-1"></i>
		</span>
	</div>
</template>

<script lang="ts" setup>
defineProps({
	password: {
		type: String,
		required: true
	},
	dispaly: {
		type: Boolean,
		required: false,
		deault: true
	}
});
</script>

<style lang="scss" scoped>
.totp-item {
	// width: 140px;
	// background-color: red;
}
.totp-text {
	// font-family: Roboto;
	// font-size: 24px;
	line-height: 32px;
	// font-weight: 700;
}

.totp-hidden {
	// background-color: red;
	font-size: 24px;
	line-height: 32px;
	display: flex;
	align-items: center;
	justify-content: center;
	height: 32px;
	i {
		width: 10px;
		height: 10px;
		border-radius: 5px;
		display: inline-block;
		margin-left: 4px;
	}
	span {
		margin-left: 4px;
	}
}
</style>

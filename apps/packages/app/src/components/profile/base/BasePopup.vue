<template>
	<q-menu :self="self">
		<div class="column bt-menu">
			<slot />
		</div>
	</q-menu>
</template>

<script lang="ts" setup>
defineProps({
	self: {
		type: String,
		require: false
	}
});
</script>

<style lang="scss">
.q-menu {
	border-radius: 8px !important;
	background-color: #ffffff !important;
	box-shadow: 0 4px 10px 0 #00000033 !important;
}

.q-field__native,
.q-field__prefix,
.q-field__suffix,
.q-field__input {
	padding: 0 !important;
}

.q-field--dense .q-field__control,
.q-field--dense .q-field__marginal {
	height: auto !important;
}

.bt-menu {
	width: auto;
	padding: 8px;
}
</style>

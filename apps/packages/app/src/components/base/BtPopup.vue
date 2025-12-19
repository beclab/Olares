<template>
	<q-menu>
		<div class="column bt-menu" :style="{ padding: padding }">
			<slot />
		</div>
	</q-menu>
</template>

<script lang="ts" setup>
defineProps({
	padding: {
		type: String,
		default: '8px'
	}
});
</script>

<style lang="scss">
.q-menu {
	border-radius: 8px !important;
	background-color: $background-2 !important;
	box-shadow: 0 4px 10px 0 #00000033 !important;
}

::v-deep(.q-field__native .q-field__prefix .q-field__suffix .q-field__input) {
	padding: 0 !important;
}

::v-deep(.q-field--dense .q-field__control .q-field--dense .q-field__marginal) {
	height: auto !important;
}

::v-deep(.q-field__control) {
	height: auto !important;
}

.bt-menu {
	width: auto;
}
</style>

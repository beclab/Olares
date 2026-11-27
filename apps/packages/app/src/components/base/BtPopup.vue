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

<style lang="scss" scoped>
.bt-menu {
	width: auto;
}
</style>

<!--
	QMenu teleports its popup (carrying the `.q-menu` class) to <body>, so scoped
	styles can't reach it. These rules must stay global.
-->
<style lang="scss">
.q-menu {
	border-radius: 8px !important;
	background-color: $background-2 !important;
	box-shadow: 0 4px 10px 0 #00000033 !important;
}

.q-field__native .q-field__prefix .q-field__suffix .q-field__input {
	padding: 0 !important;
}

.q-field--dense .q-field__control .q-field--dense .q-field__marginal {
	height: auto !important;
}

.q-field__control {
	height: auto !important;
}
</style>

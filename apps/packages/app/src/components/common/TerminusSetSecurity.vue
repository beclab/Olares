<template>
	<div>
		<TerminusSetSecurityPassword></TerminusSetSecurityPassword>
		<TerminusSetSecurityAutoLock></TerminusSetSecurityAutoLock>
	</div>
</template>

<script setup lang="ts">
import TerminusSetSecurityPassword from './TerminusSetSecurityPassword.vue';
import TerminusSetSecurityAutoLock from './TerminusSetSecurityAutoLock.vue';
</script>

<style lang="scss" scoped>
.adminBtn {
	border: 1px solid $yellow;
	background-color: $yellow-1;
	display: inline-block;
	color: $ink-2;
	padding: 6px 12px;
	border-radius: 8px;
	cursor: pointer;

	&:hover {
		background-color: $yellow-3;
	}
}
</style>

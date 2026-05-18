<template>
	<bt-scroll-area>
		<slot />
	</bt-scroll-area>
</template>

<script setup lang="ts"></script>

<style scoped lang="scss">
::v-deep(.q-scrollarea__content) {
	max-width: 100%;
}
</style>

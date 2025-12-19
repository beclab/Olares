<template>
	<!-- <PadLoginLayout v-if="isPad" /> -->
	<PhoneLoginLayout />
</template>

<script setup lang="ts">
// import { useQuasar } from 'quasar';
// import PadLoginLayout from './mobiles/PadLoginLayout.vue';
import PhoneLoginLayout from './mobiles/PhoneLoginLayout.vue';
// import { computed } from 'vue';

// const $q = useQuasar();
</script>

<style lang="scss" scoped>
.layout-root {
	width: 100%;
	height: 100vh;

	.content-full {
		width: 100%;
		height: 100%;
	}

	.content-desktop {
		width: 100%;
		height: 100%;
	}

	.content-ios {
		@extend .content-full;
		padding-top: env(safe-area-inset-top);
		padding-bottom: env(safe-area-inset-bottom);
	}
}

.background-white {
	background-color: $background-1;
}
</style>

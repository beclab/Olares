<template>
	<div
		class="row justify-center content-center items-center layout-root"
		:class="hideBackground ? '' : 'background-white'"
	>
		<div
			:class="
				menuStore.useSafeArea && $q.platform.is.ios
					? 'content-ios'
					: $q.platform.is.desktop
					? 'content-desktop'
					: 'content-full'
			"
		>
			<router-view />
		</div>
	</div>
</template>

<script setup lang="ts">
import { useMenuStore } from 'src/stores/menu';
import { useQuasar } from 'quasar';
import { ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import '../../css/terminus.scss';

const $q = useQuasar();
const menuStore = useMenuStore();
const hideBackground = ref(false);

const Route = useRoute();

watch(
	() => Route.meta,
	() => {
		if (!Route.meta || Route.meta.backgroundHide === undefined) {
			hideBackground.value = false;
			return;
		}

		hideBackground.value = Route.meta.backgroundHide as boolean;
	}
);

if (Route.meta && Route.meta.backgroundHide != undefined) {
	hideBackground.value = Route.meta.backgroundHide as boolean;
}
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
	background-color: $white;
}
</style>

<template>
	<q-layout view="lhr lpr lfr" class="layout layout-app">
		<div
			class="mainlayout"
			:class="{
				'mainlayout-ios': $q.platform.is.ios
			}"
		>
			<q-page-container style="width: 100%; height: 100%">
				<router-view />
			</q-page-container>
		</div>
	</q-layout>
</template>

<script lang="ts" setup>
import 'src/css/mobile.scss';
const appHeight = () => {
	const doc = document.documentElement;
	doc.style.setProperty('--app-height', `${window.innerHeight}px`);
};
window.addEventListener('resize', appHeight);
appHeight();
</script>

<style lang="scss" scoped>
.mainlayout-ios {
	padding-top: env(safe-area-inset-top);
	padding-bottom: env(safe-area-inset-bottom);
}

:root {
	--app-height: 100%;
}

.mainlayout {
	height: 100vh;
	width: 100vw;
	height: var(--app-height);
}
</style>

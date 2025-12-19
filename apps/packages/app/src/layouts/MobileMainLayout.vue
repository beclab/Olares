<template>
	<q-layout view="lhr lpr lfr" class="layout layout-app">
		<div
			class="mainlayout"
			:style="{ width: isBex ? '100%' : '100vw' }"
			:class="{
				'mainlayout-ios': $q.platform.is.ios && menuStore.useSafeArea,
				'background-white': !menuStore.hideBackground
			}"
		>
			<q-page-container style="width: 100%; height: 100%">
				<div
					:class="
						tabbarShow ? 'container-show-tabbar' : 'container-hide-tabbar'
					"
				>
					<router-view />
				</div>
				<tabbar-component
					v-if="tabbarShow"
					class="tabbar"
					:class="$q.platform.is.ios ? 'tabbar-ios' : ''"
					:current="defaultIndex"
					@update-current="updateCurrent"
				/>
			</q-page-container>
		</div>
	</q-layout>
</template>

<script lang="ts" setup>
import TabbarComponent from '../components/common/TerminusTabbarComponent.vue';
import { useMobileMainLayout } from '../composables/mobile/useMobileMainLayout';
import '../css/terminus.scss';

const { menuStore, isBex, tabbarShow, defaultIndex, updateCurrent } =
	useMobileMainLayout();
</script>

<style lang="scss" scoped>
.mainlayout-ios {
	@extend .mainlayout;
	padding-top: env(safe-area-inset-top);
	padding-bottom: env(safe-area-inset-bottom);
}

.layout-app {
	perspective: 500;
	-webkit-perspective: 500;

	touch-callout: none;
	-webkit-touch-callout: none;
	user-select: none;
	-webkit-user-select: none;
}

.mainlayout {
	position: absolute;
}

.mainlayout {
	height: 100vh;
}

.background-white {
	background-color: $background-1;
}

.container-show-tabbar {
	height: calc(100% - 65px);
	width: 100%;
}

.container-hide-tabbar {
	height: 100%;
	width: 100%;
}

.rotate {
	animation: aniRotate 0.8s linear infinite;

	&:hover {
		background: transparent !important;
	}
}

@keyframes aniRotate {
	0% {
		transform: rotate(0deg);
	}

	50% {
		transform: rotate(180deg);
	}

	100% {
		transform: rotate(360deg);
	}
}

.tabbar {
	position: absolute;
	bottom: 1px;
	width: 100%;
}
</style>

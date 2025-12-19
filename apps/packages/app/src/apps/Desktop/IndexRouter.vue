<template>
	<component :is="currentComponent"></component>
</template>

<script setup>
import { computed } from 'vue';
import { DeviceType } from '@bytetrade/core';
import PcHome from './IndexPage.vue';
import LandscapeScreenIndex from './LandscapeScreenIndex.vue';
import MobileHome from './mobile/IndexPage.vue';

import { useTokenStore } from 'src/stores/token';
const tokenStore = useTokenStore();

const currentComponent = computed(() => {
	if (tokenStore.deviceInfo.device === DeviceType.MOBILE) {
		if (tokenStore.deviceInfo.isVerticalScreen) {
			return MobileHome;
		} else {
			return LandscapeScreenIndex;
		}
	} else {
		return PcHome;
	}
});
</script>

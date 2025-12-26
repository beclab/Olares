<template>
	<q-drawer
		v-model="store.leftDrawerOpen"
		:behavior="behavior"
		show-if-above
		:width="240"
		class="drawer"
		:class="{ borderRight: isWeb }"
	>
		<VaultsMenu />
	</q-drawer>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';
import { useMenuStore } from '../../stores/menu';
import VaultsMenu from './VaultsMenu.vue';
import { getAppPlatform } from '../../application/platform';
import { useDeviceStore } from '../../stores/device';

const store = useMenuStore();
const deviceStore = useDeviceStore();

const isWeb = ref(process.env.APPLICATION == 'VAULT');

const behavior = computed(function () {
	if (process.env.PLATFORM == 'MOBILE') {
		if (getAppPlatform().isPad && deviceStore.isLandscape) {
			return 'desktop';
		}
		return 'default';
	}
	if (process.env.IS_BEX) {
		return 'mobile';
	}
	return 'desktop';
});
</script>

<style lang="scss">
.drawer {
	padding-top: 6px;
	overflow: hidden !important;
	&.borderRight {
		border-right: 1px solid $separator;
	}
}
</style>

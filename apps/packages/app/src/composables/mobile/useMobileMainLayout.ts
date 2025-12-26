import { ref, onMounted, onUnmounted, watch, computed } from 'vue';
import { useRoute } from 'vue-router';
import { getAppPlatform } from '../../application/platform';
import { useMenuStore } from '../../stores/menu';
import { useLarepassWebsocketManagerStore } from '../../stores/larepassWebsocketManager';

export function useMobileMainLayout() {
	const Route = useRoute();
	const menuStore = useMenuStore();
	const socketStore = useLarepassWebsocketManagerStore();
	const isBex = process.env.IS_BEX;

	onMounted(async () => {
		getAppPlatform().homeMounted();
	});

	onUnmounted(() => {
		getAppPlatform().homeUnMounted();
	});

	const tabs = ref(getAppPlatform().tabbarItems);
	const defaultIndex = ref(-1);

	const updateCurrent = (index: number) => {
		defaultIndex.value = index;
	};

	watch(
		() => Route.meta,
		() => {
			if (!Route.meta || !(Route.meta as any).tabIdentify) {
				defaultIndex.value = -1;
				return;
			}
			defaultIndex.value = tabs.value.findIndex(
				(e) => e.identify === (Route.meta as any).tabIdentify
			);
		},
		{
			immediate: true
		}
	);

	watch(
		() => Route.path,
		() => {
			if (process.env.PLATFORM == 'MOBILE') {
				socketStore.restart();
			}
		},
		{
			immediate: true
		}
	);

	const tabbarShow = computed(() => {
		if (!getAppPlatform().isTabbarDisplay()) {
			return false;
		}
		return defaultIndex.value >= 0 && defaultIndex.value < tabs.value.length;
	});

	return {
		menuStore,
		isBex,
		tabbarShow,
		defaultIndex,
		updateCurrent
	};
}

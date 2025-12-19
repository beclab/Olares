import { ref, onMounted, onUnmounted } from 'vue';
import { withMountedGuard } from './withMountedGuard';

export function useAsyncMounted() {
	const isMounted = ref(false);
	onMounted(() => (isMounted.value = true));
	onUnmounted(() => (isMounted.value = false));

	return {
		isMounted,
		run: <T>(fn: () => Promise<T>) => withMountedGuard(isMounted, fn)
	};
}

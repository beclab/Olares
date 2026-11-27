import { computed } from 'vue';
import { useRoute } from 'vue-router';
export function useIsStudio() {
	const route = useRoute();

	return computed(() => !!route.meta.workloadActionHide);
}

import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { componentMeta } from '@apps/control-hub/src/router/const';
export function useIsStudio() {
	const route = useRoute();

	return computed(() => !!route.meta.workloadActionHide);
}

export function useIsStudio2() {
	const route = useRoute();

	return computed(() => route.meta.app === componentMeta.STUDIO);
}

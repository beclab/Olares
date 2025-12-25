import { Ref } from 'vue';

export function withMountedGuard<T, P extends any[]>(
	isMounted: Ref<boolean>,
	asyncFn: (...args: P) => Promise<T>
) {
	return async (...args: P): Promise<T | null> => {
		if (!isMounted.value) return null;
		try {
			const result = await asyncFn(...args);
			return isMounted.value ? result : null;
		} catch (e) {
			if (isMounted.value) throw e;
			return null;
		}
	};
}

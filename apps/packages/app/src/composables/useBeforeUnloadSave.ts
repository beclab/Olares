import { onMounted, onUnmounted } from 'vue';

export function useBeforeUnloadSave(hasUnsavedChanges: () => boolean) {
	const handleBeforeUnload = async (event: BeforeUnloadEvent) => {
		if (hasUnsavedChanges()) {
			event.preventDefault();
			event.returnValue = '';
		}
	};

	onMounted(() => {
		window.addEventListener('beforeunload', handleBeforeUnload);
	});

	onUnmounted(() => {
		window.removeEventListener('beforeunload', handleBeforeUnload);
	});
}

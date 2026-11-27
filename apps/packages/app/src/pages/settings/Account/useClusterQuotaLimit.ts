import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { get_cluster_resource } from 'src/types/resource';
import { notifyFailed } from 'src/utils/settings/btNotify';

/**
 * Caps a sub-user's CPU / memory quota at the cluster's remaining available
 * resources (total minus current usage). Both the create-user and
 * modify-quota dialogs share this so the limits stay consistent.
 *
 * When the cluster resource fetch fails the limits are unknown, so saving is
 * blocked (canSave=false) and an error toast is shown — we never silently let
 * an unbounded quota through.
 */
export function useClusterQuotaLimit() {
	const { t } = useI18n();

	const maxCpu = ref(0);
	const maxMemoryGi = ref(0);
	const ready = ref(false);
	const loadFailed = ref(false);

	const loadClusterLimit = async () => {
		try {
			const res = await get_cluster_resource();
			const cpuTotal = Number(res.user_cpu_total) || 0;
			const cpuUsage = Number(res.user_cpu_usage) || 0;
			const memTotal = Number(res.user_memory_total) || 0;
			const memUsage = Number(res.user_memory_usage) || 0;
			maxCpu.value = Math.max(0, cpuTotal - cpuUsage);
			// Match UserInfoPage display, which renders memory via getValueByUnit
			// (Gi = bytes / 1024^3). Using 1e9 here would drift ~7% from the
			// numbers the user sees on the account page.
			maxMemoryGi.value = Math.max(0, (memTotal - memUsage) / 1024 ** 3);
			ready.value = true;
			loadFailed.value = false;
		} catch (e) {
			console.error('useClusterQuotaLimit: failed to get cluster resource', e);
			ready.value = false;
			loadFailed.value = true;
			notifyFailed(t('errors.failed_to_get_cluster_resource'));
		}
	};

	onMounted(loadClusterLimit);

	const cpuMaxError = (val: string): string => {
		if (!ready.value) return '';
		if (Number(val) > maxCpu.value) {
			return t('errors.cpu_limit_exceeds_cluster', {
				max: maxCpu.value.toFixed(2)
			});
		}
		return '';
	};

	const memoryMaxError = (val: string): string => {
		if (!ready.value) return '';
		if (Number(val) > maxMemoryGi.value) {
			return t('errors.memory_limit_exceeds_cluster', {
				max: maxMemoryGi.value.toFixed(2)
			});
		}
		return '';
	};

	const canSave = computed(() => ready.value && !loadFailed.value);

	return {
		maxCpu,
		maxMemoryGi,
		ready,
		loadFailed,
		canSave,
		cpuMaxError,
		memoryMaxError,
		loadClusterLimit
	};
}

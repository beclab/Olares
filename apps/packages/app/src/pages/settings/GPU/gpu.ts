import { VRAMMode } from 'src/constant';
import { useApplicationStore } from 'src/stores/settings/application';
import { GPUInfo, useGPUStore } from 'src/stores/settings/gpu';
import { computed, ref } from 'vue';

export function useGPU() {
	const selectGpu = ref<GPUInfo | undefined>(undefined);

	const applicationStore = useApplicationStore();

	const gpuStore = useGPUStore();

	const selectApps = computed(() => {
		if (!selectGpu.value || !selectGpu.value.apps) {
			return [];
		}
		return selectGpu.value.apps.map((item) => {
			const app = applicationStore.applications.find(
				(e) => e.name == item.appName
			);
			return {
				app: app?.title || item.appName,
				icon: app?.icon || '',
				size: (item.memory || 0) * 1024 * 1024,
				value: item.appName,
				state: app?.state
			};
		});
	});

	const selectApplicationsOptions = computed(() => {
		const disabledAppsList = gpuStore.gpuList
			.filter((e) => e.nodeName != currentGpu.value?.nodeName)
			.map((e) => {
				return e.apps?.map((i) => i.appName) || ([] as string[]);
			});
		const disabledApps = mergeMultiple(disabledAppsList);

		return applicationStore.usegpuApplications
			.filter((e) => disabledApps.find((i) => i == e.name) == undefined)
			.filter((e) =>
				currentGpu.value?.sharemode == VRAMMode.Single
					? true
					: currentGpu.value?.apps?.find((i) => i.appName == e.name) ==
					  undefined
			)
			.map((app) => {
				return {
					label: app.title,
					value: app.name,
					icon: app.icon,
					state: app.state
				};
			});
	});

	const mergeMultiple = (arrays: string[][]) => {
		return [...new Set(arrays.flat())];
	};

	const currentGpu = computed(() => {
		if (selectGpu.value) {
			return selectGpu.value;
		}
		return undefined;
	});

	return {
		selectApps,
		selectGpu,
		applicationStore,
		selectApplicationsOptions,
		gpuStore,
		currentGpu
	};
}

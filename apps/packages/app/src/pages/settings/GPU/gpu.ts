import {
	getSuitableUnit,
	getSuitableValue,
	getValueByUnit
} from '@bytetrade/core';
import { VRAMMode } from 'src/constant';
import { useApplicationStore } from 'src/stores/settings/application';
import { GPUInfo, useGPUStore } from 'src/stores/settings/gpu';
import { computed, ref } from 'vue';
import { CPUResource, MemoryResource } from '@icebergtsn/k8s-resources';

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
				state: app?.state,
				memory: item.memory,
				minMemory: app?.requiredGpu ? parseMi(app.requiredGpu) : 0
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
			.filter((e) => {
				if (currentGpu.value?.sharemode == VRAMMode.MemorySlicing) {
					return (
						currentGpu.value?.apps?.find((i) => i.appName == e.name) ==
							undefined &&
						(!e.requiredGpu ||
							parseMi(e.requiredGpu) <= currentGpu.value!.memoryAvailable!)
					);
				} else {
					return currentGpu.value?.sharemode == VRAMMode.Single
						? true
						: currentGpu.value?.apps?.find((i) => i.appName == e.name) ==
								undefined;
				}
			})
			.map((app) => {
				return {
					label: app.title,
					value: app.name,
					icon: app.icon,
					state: app.state,
					minMemory: parseMi(app.requiredGpu)
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

	const parseMi = (raw: string | number | null | undefined) => {
		if (raw === undefined || raw === null || raw === '') {
			return 0;
		}
		try {
			let res: MemoryResource | null = null;
			if (typeof raw === 'number') {
				res = MemoryResource.fromBytes(raw);
			} else {
				const s = String(raw).trim();
				if (!s) {
					return 0;
				}
				if (/^\d (\.\d )?$/.test(s)) {
					res = MemoryResource.fromBytes(Number(s));
				} else {
					res = new MemoryResource(s);
				}
			}
			if (res) {
				return res.valueOf() / 1024 ** 2;
			}
			return 0;
		} catch (error) {
			return 0;
		}
	};

	return {
		selectApps,
		selectGpu,
		applicationStore,
		selectApplicationsOptions,
		gpuStore,
		currentGpu,
		parseMi
	};
}

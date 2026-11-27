import SelectComputeBindingDialog from 'src/pages/market/application/compute/SelectComputeBindingDialog.vue';
import SelectGpuModeDialog from 'src/pages/market/application/compute/SelectGpuModeDialog.vue';
import {
	ComputeAvailabilityRefetch,
	ComputeAvailabilityResponse,
	ComputeBindingItem,
	ComputeBindingValidation,
	ComputeModeOption,
	ComputeType
} from 'src/constant/compute';
import { QVueGlobals } from 'quasar';

export function pickGpuMode(
	q: QVueGlobals,
	appTitle: string,
	modes: ComputeModeOption[]
): Promise<ComputeType> {
	return new Promise((resolve, reject) => {
		q.dialog({
			component: SelectGpuModeDialog,
			componentProps: {
				appTitle,
				modes
			}
		})
			.onOk((result: { selectedGpuType: ComputeType }) => {
				if (result?.selectedGpuType) {
					resolve(result.selectedGpuType);
				} else {
					reject('Invalid selection'); // i18n-ignore: internal promise rejection, not user-visible
				}
			})
			.onCancel(
				() => reject('User canceled the operation') // i18n-ignore: internal promise rejection, not user-visible
			);
	});
}

export function pickComputeBinding(
	q: QVueGlobals,
	appTitle: string,
	availability: ComputeAvailabilityResponse,
	validation?: ComputeBindingValidation,
	appIcon = '',
	previousBinding?: ComputeBindingItem[],
	refetch?: ComputeAvailabilityRefetch
): Promise<ComputeBindingItem[]> {
	return new Promise((resolve, reject) => {
		q.dialog({
			component: SelectComputeBindingDialog,
			componentProps: {
				appTitle,
				appIcon,
				availability,
				validation,
				previousBinding,
				refetch
			}
		})
			.onOk((result: { computeBinding: ComputeBindingItem[] }) => {
				if (result?.computeBinding && result.computeBinding.length > 0) {
					resolve(result.computeBinding);
				} else {
					reject('Invalid binding'); // i18n-ignore: internal promise rejection, not user-visible
				}
			})
			.onCancel(
				() => reject('User canceled the operation') // i18n-ignore: internal promise rejection, not user-visible
			);
	});
}

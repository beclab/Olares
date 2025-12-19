<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('Switch GPU')"
		:skip="false"
		:ok="t('confirm')"
		size="medium"
		:cancel="t('cancel')"
		:platform="deviceStore.platform"
		@onSubmit="submitGPU"
		:ok-disabled="!enableCreate"
	>
		<div class="text-body3 q-mb-sm">
			{{ t('base.app') }}
		</div>
		<GPUSelect
			v-model="gpuId"
			:options="gpuList"
			:border="true"
			:height="40"
			:iconSize="24"
			classes="q-px-md"
			menuClasses="q-pa-xs"
			:menuItemHeight="40"
		/>
		<terminus-edit
			v-if="memeryInput"
			v-model="memoryLimit"
			:label="t('Memroy')"
			:show-password-img="false"
			class="q-mt-md"
			:is-error="
				memoryLimit.length > 0 && memoryLimitRule(memoryLimit).length > 0
			"
			:error-message="memoryLimitRule(memoryLimit)"
		>
			<template v-slot:right>
				<edit-number-right-slot v-model="memoryLimit" label="GB" :max="max" />
			</template>
		</terminus-edit>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import GPUSelect from './GPUSelect.vue';
import { computed, PropType, ref } from 'vue';
import { GPUInfo, useGPUStore } from 'src/stores/settings/gpu';
import { useDeviceStore } from 'src/stores/settings/device';
import { VRAMMode } from 'src/constant';
import EditNumberRightSlot from 'src/components/settings/EditNumberRightSlot.vue';
import TerminusEdit from 'src/components/settings/base/TerminusEdit.vue';
import { useQuasar } from 'quasar';
import ReminderDialogComponent from 'src/components/settings/ReminderDialogComponent.vue';

const props = defineProps({
	currentGPU: {
		type: Object as PropType<GPUInfo>,
		required: true
	},
	appName: {
		type: String,
		required: true
	},
	appTitle: {
		type: String,
		required: true
	}
});

const { t } = useI18n();

const gpuId = ref(props.currentGPU.id);

const gpuStore = useGPUStore();

const deviceStore = useDeviceStore();

const $q = useQuasar();

const CustomRef = ref();

const gpuList = gpuStore.gpuList.map((e) => {
	return {
		label: `${e.type}${e.index ? '-' + e.index : ''}(${e.nodeName})`,
		value: e.id,
		disable:
			props.appName != undefined &&
			props.currentGPU.id != e.id &&
			e.apps?.find((app) => app.appName == props.appName) != undefined
	};
});

const submitGPU = () => {
	const needUnbindNodes: GPUInfo[] = gpuStore.gpuList.filter(
		(e) =>
			e.nodeName != pCurrentGPU.value.nodeName &&
			e.apps?.find((app) => app.appName == props.appName) != undefined
	);

	$q.dialog({
		component: ReminderDialogComponent,
		componentProps: {
			message: t(
				'Are you sure to unbind “{app}” from the “{cgpus}” and bind to {tgpu}?',
				{
					app: props.appTitle,
					cgpus: needUnbindNodes
						.map(
							(e) => `${e.type}${e.index ? '-' + e.index : ''}(${e.nodeName})`
						)
						.join(','),
					tgpu: `${pCurrentGPU.value.type}${
						pCurrentGPU.value.index ? '-' + pCurrentGPU.value.index : ''
					}(${pCurrentGPU.value.nodeName})`
				}
			),
			title: t('Unbind App'),
			useCancel: true,
			confirmText: t('confirm'),
			cancelText: t('cancel')
		}
	}).onOk(async () => {
		try {
			await gpuStore.switchToVRAM(
				pCurrentGPU.value.id,
				props.appName,
				needUnbindNodes.map((e) => {
					return {
						id: e.id
					};
				}),
				pCurrentGPU.value.sharemode == VRAMMode.MemorySlicing
					? (Number(memoryLimit.value) * 1024).toFixed(0)
					: undefined
			);
			CustomRef.value.onDialogOK();
		} catch (error) {
			console.log(error.message);
		}
	});
};

const memoryLimit = ref(`${0}`);

const memoryLimitRule = (val: string) => {
	if (val.length === 0) {
		return t('errors.memory_limit_is_empty');
	}
	let rule = /^[+-]?(\d+\.?\d*|\.\d+)$/;
	if (!rule.test(val)) {
		return t('errors.only_valid_numbers_can_be_entered');
	}

	if (!pCurrentGPU.value.memoryAvailable) {
		return t('The maximum available space is {space}', {
			space: '0GB'
		});
	}

	if (pCurrentGPU.value.memoryAvailable - Number(val) * 1024 < 0) {
		return t('The maximum available space is {space}', {
			space:
				Number(
					Math.floor((pCurrentGPU.value.memoryAvailable * 100) / 1024).toFixed(
						2
					)
				) /
					100 +
				'GB'
		});
	}
	return '';
};

const pCurrentGPU = computed(() => {
	return gpuStore.gpuList.find((e) => e.id == gpuId.value) || props.currentGPU;
});

const max = computed(() => {
	return Math.floor((pCurrentGPU.value.memoryAvailable || 0) / 1024);
});

const memeryInput = computed(() => {
	return pCurrentGPU.value.sharemode == VRAMMode.MemorySlicing;
});

const enableCreate = computed(() => {
	if (!memeryInput.value) {
		return pCurrentGPU.value.id != props.currentGPU.id;
	}
	return (
		memoryLimitRule(memoryLimit.value).length == 0 &&
		Number(memoryLimit.value) > 0 &&
		pCurrentGPU.value.id != props.currentGPU.id
	);
});
</script>

<style scoped lang="scss"></style>

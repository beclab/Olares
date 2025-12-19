<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('modify_limits')"
		:skip="false"
		:ok="t('ok')"
		size="medium"
		:cancel="t('cancel')"
		:platform="deviceStore.platform"
		:okDisabled="!enableCreate"
		@onSubmit="onOKClick"
	>
		<terminus-edit
			v-model="cpuLimit"
			:label="t('cpu')"
			:show-password-img="false"
			:is-error="cpuLimit.length > 0 && cpuLimitRule(cpuLimit).length > 0"
			:error-message="cpuLimitRule(cpuLimit)"
		>
			<template v-slot:right>
				<edit-number-right-slot v-model="cpuLimit" label="core" />
			</template>
		</terminus-edit>

		<terminus-edit
			v-model="memoryLimit"
			:label="t('memory')"
			:show-password-img="false"
			class="q-mt-md"
			:is-error="
				memoryLimit.length > 0 && memoryLimitRule(memoryLimit).length > 0
			"
			:error-message="memoryLimitRule(memoryLimit)"
		>
			<template v-slot:right>
				<edit-number-right-slot v-model="memoryLimit" label="GB" />
			</template>
		</terminus-edit>
	</bt-custom-dialog>
</template>
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import TerminusEdit from 'src/components/settings/base/TerminusEdit.vue';
import EditNumberRightSlot from 'src/components/settings/EditNumberRightSlot.vue';
import { useDeviceStore } from 'src/stores/settings/device';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	cpu: {
		required: true,
		type: Number
	},
	memory: {
		required: true,
		type: Number
	}
});

const cpuLimit = ref('1');
const memoryLimit = ref('4');
const { t } = useI18n();
const deviceStore = useDeviceStore();
const CustomRef = ref();

onMounted(() => {
	cpuLimit.value = `${props.cpu}`;
	memoryLimit.value = `${props.memory}`;
});

function onOKClick() {
	CustomRef.value.onDialogOK({
		cpuLimit: cpuLimit.value,
		memoryLimit: memoryLimit.value
	});
}

const cpuLimitRule = (val: string) => {
	if (val.length === 0) {
		return t('errors.cpu_limit_is_empty');
	}
	let rule = /^[+-]?(\d+\.?\d*|\.\d+)$/;
	if (!rule.test(val)) {
		return t('errors.only_valid_numbers_can_be_entered');
	}
	return '';
};

const memoryLimitRule = (val: string) => {
	if (val.length === 0) {
		return t('errors.memory_limit_is_empty');
	}
	let rule = /^[+-]?(\d+\.?\d*|\.\d+)$/;
	if (!rule.test(val)) {
		return t('errors.only_valid_numbers_can_be_entered');
	}
	return '';
};

const enableCreate = computed(() => {
	return (
		cpuLimitRule(cpuLimit.value).length == 0 &&
		memoryLimitRule(memoryLimit.value).length == 0
	);
});
</script>
<style lang="scss" scoped>
.cpu-core {
	text-align: right;
}
</style>

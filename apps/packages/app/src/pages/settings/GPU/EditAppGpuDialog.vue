<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="title"
		:skip="false"
		:ok="t('confirm')"
		size="medium"
		:cancel="t('cancel')"
		:platform="deviceStore.platform"
		:ok-disabled="!enableCreate"
		@onSubmit="submitApp"
	>
		<div class="text-body3 q-mb-sm">
			{{ t('base.app') }}
		</div>
		<app-select
			v-model="appSelectMode"
			:options="selectApplicationsOptions"
			:border="true"
			:height="40"
			:iconSize="24"
			classes="q-px-md"
			menuClasses="q-pa-xs"
			:menuItemHeight="40"
		/>
		<terminus-edit
			v-if="memoryInput"
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
import EditNumberRightSlot from 'src/components/settings/EditNumberRightSlot.vue';
import AppSelect from 'src/pages/settings/Developer/pages/dialog/AppSelect.vue';
import TerminusEdit from 'src/components/settings/base/TerminusEdit.vue';
import { useDeviceStore } from 'src/stores/settings/device';
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { i18n } from 'src/boot/i18n';

interface Props {
	selectApplicationsOptions: {
		icon: string;
		state: string;
		label: string;
		value: string;
		isDefault?: boolean;
	}[];
	maxValue: number;
	memoryInput: boolean;
	title?: string;
	memeryInit: number;
}

const props = withDefaults(defineProps<Props>(), {
	selectApplicationsOptions: () => [],
	maxValue: 0,
	memoryInput: true,
	title: i18n.global.t('Bind App'),
	memeryInit: 0
});

const { t } = useI18n();

const CustomRef = ref();

const deviceStore = useDeviceStore();

const memoryLimit = ref(`${props.memeryInit}`);

const max = ref(Math.floor(props.maxValue / 1024));

const appSelectMode = ref('');

const memoryLimitRule = (val: string) => {
	if (val.length === 0) {
		return t('errors.memory_limit_is_empty');
	}
	let rule = /^[+-]?(\d+\.?\d*|\.\d+)$/;
	if (!rule.test(val)) {
		return t('errors.only_valid_numbers_can_be_entered');
	}

	if (props.maxValue - Number(val) * 1024 < 0) {
		return t('The maximum available space is {space}', {
			space:
				Number(Math.floor((props.maxValue * 100) / 1024).toFixed(2)) / 100 +
				'GB'
		});
	}
	return '';
};

const enableCreate = computed(() => {
	if (!props.memoryInput) {
		return appSelectMode.value != '';
	}
	return (
		appSelectMode.value &&
		appSelectMode.value.length > 0 &&
		memoryLimitRule(memoryLimit.value).length == 0 &&
		Number(memoryLimit.value) > 0
	);
});

if (props.selectApplicationsOptions.length > 0) {
	const defaultItem = props.selectApplicationsOptions.find(
		(e) => e.isDefault == true
	);
	appSelectMode.value = defaultItem
		? defaultItem.value
		: props.selectApplicationsOptions[0].value;
}

const submitApp = () => {
	CustomRef.value.onDialogOK({
		app: appSelectMode.value,
		memoryLimit: (Number(memoryLimit.value) * 1024).toFixed(0)
	});
};
</script>

<style scoped lang="scss">
.cpu-core {
	text-align: right;
}
</style>

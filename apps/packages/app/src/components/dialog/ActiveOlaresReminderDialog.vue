<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('step_select_proxy')"
		:ok="position ?? t('confirm')"
		:cancel="showCancel ? navigation ?? t('cancel') : false"
		size="small"
		platform="mobile"
		@onSubmit="onSubmit"
	>
		<div class="test-body2 text-ink-2 q-mb-sm">
			{{ t('step_proxy_text_1') }}
		</div>
		<bt-select
			v-model="host"
			:options="olaresTunnelsV2Options()"
			:border="true"
			color="text-blue-default"
		/>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import BtSelect from '../base/BtSelect.vue';
import { OlaresTunneV2Interface } from '../../utils/interface/frp';

import { PropType, ref } from 'vue';
import { i18n } from '../../boot/i18n';

const { t } = useI18n();

const props = defineProps({
	title: String,
	message: String,
	navigation: String,
	position: String,
	showCancel: Boolean,
	frpList: {
		type: Array as PropType<OlaresTunneV2Interface[]>,
		required: true,
		default: [] as OlaresTunneV2Interface[]
	}
});

const CustomRef = ref();

console.log('props.frpList ===>', props.frpList);

const host = ref('');
if (props.frpList.length > 0) {
	host.value = props.frpList[0].machine[0].host;
}

const olaresTunnelsV2Options = () => {
	return props.frpList.map((item: OlaresTunneV2Interface) => {
		const label = (item.name as any)[i18n.global.locale.value]
			? (item.name as any)[i18n.global.locale.value]
			: (item.name as any)['en-US']
			? (item.name as any)['en-US']
			: '';
		return {
			label: label,
			value: item.machine.length > 0 ? item.machine[0].host : '',
			enable: true
		};
	});
};

const onSubmit = () => {
	if (!host.value) {
		return;
	}
	CustomRef.value.onDialogOK(host.value);
};
</script>

<style lang="scss" scoped>
.biometric-unlock-dialog {
	padding: 20px;

	&__title {
		color: $ink-1;
	}

	&__desc {
		width: 100%;
		margin-top: 12px;
	}

	&__upgrade_bg {
		// margin-top: 40px;
		width: 100%;
		border: 1px solid $separator;

		border-radius: 8px;

		.checkbox-select {
			width: 16px;
			height: 16px;
			border-radius: 8px;
			position: relative;
			background: $light-blue-default;
		}

		.checkbox-select::before {
			content: '.';
			width: 6px;
			height: 6px;
			border-radius: 3px;
			background-color: white;
			color: transparent;
			position: absolute;
			top: 5px;
			left: 5px;
		}

		.checkbox-normal {
			width: 16px;
			height: 16px;
			border-radius: 8px;
			border: 1px solid $radio-stroke;
		}
	}
}
</style>

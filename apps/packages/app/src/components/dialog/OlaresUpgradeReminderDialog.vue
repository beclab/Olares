<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="title"
		:ok="position ?? t('confirm')"
		:cancel="showCancel ? navigation ?? t('cancel') : false"
		size="small"
		platform="mobile"
		@onSubmit="onSubmit"
	>
		<div class="test-body2 text-ink-2">
			{{ t('Please select your preferred upgrade method') }}
		</div>
		<div
			v-for="option in upgradeOptions"
			:key="option.value"
			class="biometric-unlock-dialog__upgrade_bg q-pa-md row items-center q-mt-md"
			@click="upgradeMode = option.value"
		>
			<div
				:class="{
					'checkbox-select': option.value == upgradeMode,
					'checkbox-normal': option.value != upgradeMode
				}"
			></div>
			<div class="q-ml-md">
				<div class="text-ink-1 text-subtitle2">
					{{ option.label }}
				</div>
				<div class="text-overline-m text-ink-3 q-mt-xs">
					{{ option.detail }}
				</div>
			</div>
		</div>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { upgradeModeOptions, UpgradeMode } from '../../constant/larepass';

import { ref } from 'vue';

const { t } = useI18n();

defineProps({
	title: String,
	message: String,
	navigation: String,
	position: String,
	showCancel: {
		default: true,
		type: Boolean
	},
	descAlign: {
		default: 'center',
		type: String as () => 'left' | 'center' | 'right'
	}
});

const CustomRef = ref();

const upgradeMode = ref(UpgradeMode.DOWNLOAD_AND_UPGRADE);

const upgradeOptions = ref(upgradeModeOptions());

const onSubmit = () => {
	CustomRef.value.onDialogOK(upgradeMode.value);
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

<template>
	<bt-custom-dialog
		ref="customRef"
		:title="t('Switch to {mode}', { mode: modeLabel })"
		:skip="false"
		:ok="t('Switch')"
		:cancel="t('cancel')"
		size="medium"
		:platform="deviceStore.platform"
		@onSubmit="onConfirm"
	>
		<div class="switch-support-type__body">
			<i18n-t
				keypath="switch_support_type_desc"
				tag="div"
				scope="global"
				class="text-ink-2"
				:class="deviceStore.isMobile ? 'text-body3-m' : 'text-body3'"
			>
				<template #stopped>
					<span class="switch-support-type__stopped">{{ t('Stopped') }}</span>
				</template>
			</i18n-t>

			<div v-if="appNames.length > 0" class="switch-support-type__apps q-mt-md">
				<div class="text-body3 switch-support-type__apps-label q-mb-xs">
					{{ t('Unbinding will be performed on:') }}
				</div>
				<div class="text-body2 text-ink-1">
					{{ appNameList }}
				</div>
			</div>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useDeviceStore } from 'src/stores/settings/device';

const props = defineProps<{
	modeLabel: string;
	appNames: string[];
}>();

const { locale, t } = useI18n();
const deviceStore = useDeviceStore();
const customRef = ref();
const appNameList = computed(() =>
	props.appNames.join(locale.value.startsWith('zh') ? '、' : ', ')
);

const onConfirm = () => {
	customRef.value?.onDialogOK();
};
</script>

<style scoped lang="scss">
.switch-support-type__body {
	padding: 0 4px;
	word-wrap: break-word;
	white-space: normal;

	.switch-support-type__stopped {
		color: $orange-default;
	}

	.switch-support-type__apps {
		border: 1px solid $orange-default;
		border-radius: 8px;
		padding: 12px;
		background-color: $orange-soft;

		&-label {
			color: $orange-default;
		}
	}
}
</style>

<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('update_secret')"
		:skip="false"
		size="medium"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		:okDisabled="value.length == 0"
		:platform="deviceStore.platform"
		@onSubmit="onOKClick"
	>
		<terminus-edit label="Secret" v-model="value" />
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import TerminusEdit from '../../../base/TerminusEdit.vue';
import { useDeviceStore } from 'src/stores/settings/device';

const value = ref('');
const deviceStore = useDeviceStore();

function onOKClick() {
	CustomRef.value.onDialogOK({
		value: value.value
	});
}

const { t } = useI18n();
const CustomRef = ref();
</script>
<style lang="scss"></style>

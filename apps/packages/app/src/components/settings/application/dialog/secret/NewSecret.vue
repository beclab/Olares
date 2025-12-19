<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('add_secret')"
		:skip="false"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		size="medium"
		:platform="deviceStore.platform"
		:okDisabled="!canSubmit"
		@onSubmit="onOKClick"
	>
		<terminus-edit :label="t('name')" v-model="key" />

		<terminus-edit :label="t('secret')" v-model="value" class="q-mt-md" />
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { computed, ref } from 'vue';
import TerminusEdit from '../../../base/TerminusEdit.vue';
import { useDeviceStore } from 'src/stores/settings/device';

const { t } = useI18n();

const key = ref('');
const value = ref('');

const canSubmit = computed(() => {
	return key.value !== '' && value.value !== '';
});

const CustomRef = ref();
const deviceStore = useDeviceStore();
function onOKClick() {
	CustomRef.value.onDialogOK({
		key: key.value,
		value: value.value
	});
}
</script>

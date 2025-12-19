<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('update_secret')"
		:skip="false"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		:platform="deviceStore.platform"
		@onSubmit="onOKClick"
	>
		<q-label class="q-pt-sm">{{
			t('delete_secret_id', {
				id: id
			})
		}}</q-label>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useDeviceStore } from 'src/stores/settings/device';

defineProps({
	id: {
		type: String,
		required: true
	}
});

const { t } = useI18n();

const CustomRef = ref();
const deviceStore = useDeviceStore();
async function onOKClick() {
	CustomRef.value.onDialogOK({});
}
</script>

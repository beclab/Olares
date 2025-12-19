<template>
	<q-dialog ref="dialogRef" @hide="onDialogCancel">
		<q-card class="q-dialog-plugin q-pa-lg" style="min-height: 200px">
			<q-input class="q-pt-sm" filled v-model="name" :label="t('name')" />

			<q-card-actions align="right" class="q-pt-md">
				<q-btn color="primary" :label="t('ok')" @click="onOKClick" />
				<q-btn color="primary" :label="t('cancel')" @click="onDialogCancel" />
			</q-card-actions>
		</q-card>
	</q-dialog>
</template>

<script setup lang="ts">
import { useNotificationStore } from 'src/stores/settings/notification';
import { useDialogPluginComponent, Loading } from 'quasar';
import { ref } from 'vue';

import { useI18n } from 'vue-i18n';
const { t } = useI18n();

defineEmits([...useDialogPluginComponent.emits]);

const applicationStore = useNotificationStore();

const name = ref('');

const { dialogRef, onDialogOK, onDialogCancel } = useDialogPluginComponent();

async function onOKClick() {
	Loading.show();

	try {
		await applicationStore.createNotifyPolicy({
			name: name.value,
			isDefault: false,
			status: 'Active',
			user: ''
		});
		Loading.hide();
		onDialogOK({});
	} catch (e) {
		console.log(e);
		Loading.hide();
	}
}
</script>

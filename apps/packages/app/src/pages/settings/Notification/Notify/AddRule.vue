<template>
	<q-dialog ref="dialogRef" @hide="onDialogCancel">
		<q-card class="q-dialog-plugin q-pa-lg" style="min-height: 200px">
			<q-select
				:label="t('sender')"
				class="q-py-sm"
				v-model="senderId"
				:options="applicationStore.senders"
				:option-value="
					(opt) => (Object(opt) === opt && 'id' in opt ? opt.id : null)
				"
				:option-label="
					(opt) =>
						Object(opt) === opt && 'name' in opt ? opt.name : '- Null -'
				"
				emit-value
				map-options
			/>
			<q-select
				:label="t('recipients')"
				class="q-py-sm"
				v-model="recipientsId"
				:options="recipientsList"
				:option-value="
					(opt) => (Object(opt) === opt && 'id' in opt ? opt.id : null)
				"
				:option-label="
					(opt) =>
						Object(opt) === opt && 'name' in opt ? opt.name : '- Null -'
				"
				emit-value
				map-options
			/>

			<q-card-actions align="right" class="q-pt-md">
				<q-btn color="primary" :label="t('ok')" @click="onOKClick" />
				<q-btn color="primary" :label="t('cancel')" @click="onDialogCancel" />
			</q-card-actions>
		</q-card>
	</q-dialog>
</template>

<script setup lang="ts">
import { useNotificationStore } from 'src/stores/settings/notification';
import { notifyFailed } from 'src/utils/settings/btNotify';
import { useDialogPluginComponent, Loading } from 'quasar';
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	id: {
		type: String,
		required: true
	}
});

const { t } = useI18n();

defineEmits([...useDialogPluginComponent.emits]);

const applicationStore = useNotificationStore();
const senderId = ref(applicationStore.senders[0].id);
const recipientsList = ref([]);
const recipientsId = ref();

function updateRecipients() {
	const sender = applicationStore.senders.find((f) => f.id == senderId.value);
	recipientsList.value = applicationStore.recipients.filter(
		(f) => f.type == sender.recipientType
	);
	if (recipientsList.value.length > 0) {
		recipientsId.value = recipientsList.value[0].id;
	} else {
		recipientsId.value = undefined;
	}
}
updateRecipients();

watch(
	() => senderId.value,
	() => {
		updateRecipients();
	}
);

const { dialogRef, onDialogOK, onDialogCancel } = useDialogPluginComponent();

async function onOKClick() {
	Loading.show();
	if (!recipientsId.value) {
		notifyFailed(t('need_recipients'));
		Loading.hide();
		return;
	}

	try {
		await applicationStore.createNotifyRule({
			notifyPolicyId: Number(props.id),
			sender: senderId.value,
			recipients: recipientsId.value,
			status: 'Active'
		});
		Loading.hide();
		onDialogOK({});
	} catch (e) {
		console.log(e);
		Loading.hide();
	}
}
</script>

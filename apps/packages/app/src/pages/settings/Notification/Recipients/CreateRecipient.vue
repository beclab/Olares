<template>
	<q-dialog ref="dialogRef" @hide="onDialogCancel">
		<q-card class="q-dialog-plugin q-pa-lg" style="min-height: 200px">
			<q-input
				class="q-pt-sm"
				filled
				v-model="recipients.name"
				:label="t('recipients')"
				readonly
			/>
			<q-input class="q-pt-sm" filled v-model="name" :label="t('name')" />

			<div class="q-pt-sm">{{ t('daata') }}</div>

			<q-input
				v-for="item in variablesInput"
				class="q-pt-sm"
				:key="item.key"
				filled
				v-model="item.value"
				:label="item.title"
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
import { recipientTemplates } from 'src/utils/settings/notification';
import { useDialogPluginComponent, Loading } from 'quasar';
import { useI18n } from 'vue-i18n';
import { ref } from 'vue';
const { t } = useI18n();

const props = defineProps({
	id: {
		type: String,
		required: true
	}
});

defineEmits([...useDialogPluginComponent.emits]);

const applicationStore = useNotificationStore();

const recipients = ref(
	applicationStore.recipients.find((re) => re.id == props.id)
);
const template = ref(
	recipientTemplates.find((se) => se.type == recipients.value.type)
);
const name = ref();
const variables = ref({});
const variablesInput = ref([]);

function convertFirstLetterToUpperCase(str) {
	return str
		.toLowerCase()
		.split(' ')
		.map((word) => word.charAt(0).toUpperCase() + word.slice(1))
		.join(' ');
}

function updateVariables() {
	let arr = [];
	console.log(template.value.validator.params);
	for (let key of template.value.validator.params) {
		arr.push({
			key: key,
			value: ref(variables.value[key]),
			title: convertFirstLetterToUpperCase(key.replace(/_/g, ''))
		});
	}

	variablesInput.value = arr;
}
updateVariables();

const { dialogRef, onDialogOK, onDialogCancel } = useDialogPluginComponent();

async function onOKClick() {
	Loading.show();

	try {
		const data = {};
		for (let item in variablesInput.value) {
			data[variablesInput.value[item].key] = variablesInput.value[item].value;
		}
		await applicationStore.createRecipientsAddress(props.id, name.value, data);
		Loading.hide();
		onDialogOK({});
	} catch (e) {
		console.log(e);
		Loading.hide();
	}
}
</script>

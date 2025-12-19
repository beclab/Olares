<template>
	<q-dialog ref="dialogRef" @hide="onDialogCancel">
		<q-card class="q-dialog-plugin q-pa-lg" style="min-height: 200px">
			<q-input
				class="q-pt-sm"
				filled
				v-model="sender"
				:label="t('sender')"
				readonly
			/>
			<q-input class="q-pt-sm" filled v-model="name" :label="t('name')" />

			<q-input
				class="q-pt-sm"
				filled
				v-model="app"
				:label="t('base.app')"
				:readonly="template.type != 'Application'"
			/>
			<div class="q-pt-sm">{{ t('credentials') }}</div>

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
import { senderTemplates } from 'src/utils/settings/notification';
import { useDialogPluginComponent, Loading } from 'quasar';
import { useI18n } from 'vue-i18n';
import { ref } from 'vue';

const { t } = useI18n();

const props = defineProps({
	sender: {
		type: String,
		required: true
	}
});

defineEmits([...useDialogPluginComponent.emits]);

const applicationStore = useNotificationStore();

const template = ref(senderTemplates.find((se) => se.name == props.sender));
const sender = ref(props.sender);
const name = ref();
const variables = ref({});
const variablesInput = ref([]);
const app = ref();
if (template.value.type != 'Application') {
	app.value = 'Notification';
}

function convertFirstLetterToUpperCase(str) {
	return str
		.toLowerCase()
		.split(' ')
		.map((word) => word.charAt(0).toUpperCase() + word.slice(1))
		.join(' ');
}

function updateVariables() {
	let arr = [];
	for (let key of template.value.credential.params) {
		arr.push({
			key: key.name,
			value: ref(variables.value[key.name]),
			title: convertFirstLetterToUpperCase(key.name.replace(/_/g, ''))
		});
	}

	variablesInput.value = arr;
}

updateVariables();

const { dialogRef, onDialogOK, onDialogCancel } = useDialogPluginComponent();

async function onOKClick() {
	Loading.show();

	try {
		const credential = {};
		for (let item in variablesInput.value) {
			credential[variablesInput.value[item].key] =
				variablesInput.value[item].value;
		}
		await applicationStore.createSender(
			{
				name: name.value,
				type: template.value.type,
				app: app.value,
				recipientType: template.value.recipientType,
				isEditable: true,
				user: '',
				status: 'Active'
			},
			credential
		);
		Loading.hide();
		onDialogOK({});
	} catch (e) {
		console.log(e);
		Loading.hide();
	}
}
</script>

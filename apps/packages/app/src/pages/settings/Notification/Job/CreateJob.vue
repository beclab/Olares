<template>
	<q-dialog ref="dialogRef" @hide="onDialogCancel">
		<q-card class="q-dialog-plugin q-pa-lg" style="min-height: 200px">
			<q-input class="q-pt-sm" filled v-model="name" :label="t('name')" />
			<q-select
				:label="t('application')"
				class="q-py-sm"
				v-model="app"
				:options="appList"
			/>
			<q-select
				:label="t('template')"
				class="q-py-sm"
				v-model="templateId"
				:options="templateList"
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
				:label="t('policy')"
				class="q-py-sm"
				v-model="policyId"
				:options="applicationStore.notifyPolicy"
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
				:label="t('language')"
				class="q-py-sm"
				v-model="language"
				:options="languageList"
			/>

			<div class="q-pt-sm">{{ t('variables') }}</div>

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
import { useDialogPluginComponent, Loading } from 'quasar';
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useNotificationStore } from '../../../../stores/settings/notification';

defineProps({
	list: {
		type: Array,
		required: true
	}
});

const { t } = useI18n();

defineEmits([...useDialogPluginComponent.emits]);

const applicationStore = useNotificationStore();

const apps = new Set();
applicationStore.templates.forEach((item) => {
	apps.add(item.appName);
});

const name = ref('');
const appList = [...apps];
const app = ref(appList[0]);
const templateList = ref(
	applicationStore.templates.filter((item) => item.appName === app.value)
);
const templateId = ref(
	templateList.value.length > 0 ? templateList.value[0].id : null
);
const language = ref(null);
const languageList = ref([]);
const variables = ref({});
const variablesInput = ref([]);
const policyId = ref(applicationStore.notifyPolicy[0].id);

function convertFirstLetterToUpperCase(str) {
	return str
		.toLowerCase()
		.split(' ')
		.map((word) => word.charAt(0).toUpperCase() + word.slice(1))
		.join(' ');
}

function updateVariables() {
	const t = applicationStore.templates.find(
		(item) => item.id === templateId.value
	);
	let arr = [];
	for (let key of t.variables) {
		arr.push({
			key: key,
			value: ref(variables.value[key]),
			title: convertFirstLetterToUpperCase(key.replace(/_/g, ''))
		});
	}

	variablesInput.value = arr;
}
updateVariables();

function updateLanguage() {
	applicationStore.templateContent
		.filter((tc) => tc.templateId == templateId.value)
		.forEach((tc) => {
			languageList.value.push(tc.language);
		});
	console.log(languageList.value);
	language.value = languageList.value[0];
}
updateLanguage();

watch(
	() => app.value,
	() => {
		templateList.value = applicationStore.templates.filter(
			(item) => item.appName === app.value
		);
		templateId.value =
			templateList.value.length > 0 ? templateList.value[0].id : null;
	}
);

watch(
	() => templateId.value,
	() => {
		updateVariables();
		updateLanguage();
	}
);

const { dialogRef, onDialogOK, onDialogCancel } = useDialogPluginComponent();

async function onOKClick() {
	Loading.show();

	try {
		const variables = {};
		for (let item in variablesInput.value) {
			console.log(
				variablesInput.value[item].key + ' ' + variablesInput.value[item].value
			);
			variables[variablesInput.value[item].key] =
				variablesInput.value[item].value;
		}
		await applicationStore.createJob({
			templateId: templateId.value,
			notifyPolicyId: policyId.value,
			name: name.value,
			language: language.value,
			rawMessage: { vars: variables }
		});
		Loading.hide();
		onDialogOK({});
	} catch (e) {
		Loading.hide();
	}
}
</script>

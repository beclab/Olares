<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="title"
		:skip="false"
		:ok="t('confirm')"
		size="medium"
		:cancel="t('cancel')"
		:platform="deviceStore.platform"
		:okDisabled="!enableCreate"
		@onSubmit="createUserName"
	>
		<terminus-edit
			v-model="name"
			:label="t('name')"
			:show-password-img="false"
			style="width: 100%"
			:isReadOnly="datasetID.length > 0"
			:is-error="nameIsError"
			:error-message="t('errors.knowledge_base_name_already_taken')"
		/>
		<div class="text-overline text-ink-2 q-mt-xs" v-if="!nameIsError">
			{{ t('knowldege_base_name') }}
		</div>

		<terminus-edit
			v-model="paths"
			:label="t('paths')"
			:show-password-img="false"
			:is-textarea="true"
			style="width: 100%"
			class="q-mt-lg"
			:is-error="pathsIsError"
			:error-message="t('errors.paths_must_start_with_data_home')"
		/>
		<div v-if="!pathsIsError" class="text-overline text-ink-2 q-mt-xs">
			{{ t('separate_different_paths_with_a_comma') }}
		</div>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import TerminusEdit from 'src/components/settings/base/TerminusEdit.vue';
import { useFilesStore, DatasetFolder } from 'src/stores/settings/files';
import { useDeviceStore } from 'src/stores/settings/device';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	datasetID: {
		required: false,
		type: String,
		default: ''
	}
});

const { t } = useI18n();

const title = ref(t('add_new_knowledge_base'));

const name = ref('');
const paths = ref('');
let dataset: DatasetFolder | undefined = undefined;
const fileStore = useFilesStore();

if (props.datasetID) {
	dataset = fileStore.datasets.find((e) => e.datasetID == props.datasetID);
	if (dataset) {
		name.value = dataset.datasetName;
		paths.value = dataset.paths ? dataset.paths.join(',') : '';
		title.value = t('save_knowledge_base', {
			base: `'${dataset.datasetName}'`
		});
	}
}

const enableCreate = computed(() => {
	return (
		name.value.length > 0 &&
		!nameIsError.value &&
		paths.value.length >= 11 &&
		(!dataset ||
			(dataset.paths == null && paths.value) ||
			dataset.paths.join(',') != paths.value) &&
		!pathsIsError.value
	);
});

const nameIsError = computed(() => {
	return (
		props.datasetID.length == 0 &&
		fileStore.datasets.find((e) => e.datasetName == name.value) != undefined
	);
});

const pathsIsError = computed(() => {
	return (
		paths.value != null &&
		paths.value.length > 0 &&
		paths.value
			.split(',')
			.find((e) => e.length >= 11 && !e.startsWith('/data/Home/')) != undefined
	);
	// return false;
});

const CustomRef = ref();

const deviceStore = useDeviceStore();

const createUserName = () => {
	CustomRef.value.onDialogOK({
		name: name.value,
		paths: paths.value.split(',')
	});
};
</script>

<style scoped lang="scss">
.cpu-core {
	text-align: right;
}
</style>

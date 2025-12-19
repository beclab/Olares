<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('Clone App')"
		:skip="false"
		:ok="t('confirm')"
		size="medium"
		:cancel="t('cancel')"
		:okDisabled="!isSetEnable"
		@onSubmit="submitCloneInfo"
	>
		<terminus-edit
			v-model="titleRef"
			:label="t('New Application Title')"
			style="width: 100%"
			:descriptions="[
				t(
					'Please provide a unique title for your cloned application to avoid confusion with existing ones.'
				)
			]"
			:error-message="data.titleValidation.message"
		/>

		<terminus-edit
			v-for="item in missingValues"
			class="q-mt-xs"
			:key="item.name"
			:label="
				t('Desktop shortcut name for', {
					name: item.name
				})
			"
			:model-value="item.title"
			:descriptions="[
				t('Please provide a unique title for the app entrance on Desktop')
			]"
			@update:modelValue="
				(data) => {
					updateMissingValues({ ...item, title: data });
				}
			"
		/>

		<terminus-edit
			v-for="item in invalidValues"
			class="q-mt-xs"
			:key="item.name"
			:label="
				t('Desktop shortcut name for', {
					name: item.name
				})
			"
			:model-value="item.title"
			:is-error="!!item.message"
			:error-message="item.message ?? ''"
			:descriptions="[
				t('Please provide a unique title for the app entrance on Desktop')
			]"
			@update:modelValue="
				(data) => {
					updateInvalidValues({ ...item, title: data });
				}
			"
		/>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import TerminusEdit from 'src/components/settings/base/TerminusEdit.vue';
import { AppCloneInfoResponse, CloneEntrance } from 'src/constant';
import { computed, onMounted, PropType, ref } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	data: {
		type: Object as PropType<AppCloneInfoResponse>,
		required: true
	}
});

const { t } = useI18n();
const CustomRef = ref();

const titleRef = ref('');
const missingValues = ref<CloneEntrance[]>([]);
const invalidValues = ref<CloneEntrance[]>([]);

onMounted(() => {
	if (props.data?.titleValidation) {
		titleRef.value = props.data?.titleValidation.title;
	}
	if (props.data.missingValues && props.data.missingValues.length > 0) {
		missingValues.value.push(...props.data.missingValues);
	}
	if (props.data.invalidValues && props.data.invalidValues.length > 0) {
		invalidValues.value.push(...props.data.invalidValues);
	}
});

const updateMissingValues = (newEntrance: CloneEntrance) => {
	const index = missingValues.value.findIndex(
		(item) => item.name === newEntrance.name
	);
	if (index !== -1) {
		missingValues.value[index] = { ...newEntrance };
	}
};

const updateInvalidValues = (newEntrance: CloneEntrance) => {
	const index = invalidValues.value.findIndex(
		(item) => item.name === newEntrance.name
	);
	if (index !== -1) {
		invalidValues.value[index] = { ...newEntrance, message: '' };
	}
};

const isSetEnable = computed(() => {
	const list = missingValues.value.concat(invalidValues.value);
	for (let i = 0; i < list.length; i++) {
		if (!list[i].title) {
			return false;
		}
	}
	if (!props.data?.titleValidation.isValid && !titleRef.value) {
		return false;
	}
	return true;
});

const submitCloneInfo = () => {
	const cloneInfoList = missingValues.value.concat(invalidValues.value);
	console.log(titleRef.value);
	console.log(cloneInfoList);
	CustomRef.value.onDialogOK({
		title: titleRef.value,
		entrances: cloneInfoList
	});
};
</script>

<style scoped lang="scss"></style>

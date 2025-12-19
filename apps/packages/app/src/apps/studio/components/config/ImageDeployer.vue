<template>
	<q-card class="image-container" flat>
		<q-card-section class="text-h6 text-ink-1">
			{{ t('docker.image_command_title') }}
		</q-card-section>
		<q-card-section class="q-py-none">
			<card-form-item
				v-if="!hiddenFields?.image"
				:name="t('docker.container_image')"
				:required="true"
				:tip="t('docker.container_image')"
			>
				<q-input
					ref="imageRef"
					dense
					borderless
					no-error-icon
					v-model.trim="containerConfig.image"
					class="form-item-input"
					input-class="text-ink-2"
					:placeholder="ruleConfig.imageName.placeholder"
					:rules="ruleConfig.imageName.rules"
					:disable="disabledFields?.image"
				>
				</q-input>
			</card-form-item>

			<card-form-item
				v-if="!hiddenFields?.startCmd"
				:name="t('docker.start_command')"
				:required="false"
				:tip="t('docker.start_command')"
			>
				<q-input
					dense
					borderless
					no-error-icon
					v-model.trim="containerConfig.startCmd"
					class="form-item-input"
					input-class="text-ink-2"
					:placeholder="ruleConfig.startCommand.placeholder"
					:disable="disabledFields?.startCmd"
				>
				</q-input>
			</card-form-item>

			<card-form-item
				v-if="!hiddenFields?.startCmdArgs"
				:name="t('docker.command_parameters')"
				:required="false"
				:tip="t('docker.command_parameters')"
			>
				<q-input
					dense
					borderless
					no-error-icon
					v-model.trim="containerConfig.startCmdArgs"
					class="form-item-input"
					input-class="text-ink-2"
					:placeholder="ruleConfig.startCmdArgs.placeholder"
				>
				</q-input>
			</card-form-item>

			<card-form-item
				v-if="!hiddenFields?.port"
				:name="t('docker.container_port')"
				:required="true"
				:tip="t('docker.container_port')"
			>
				<q-input
					ref="portRef"
					dense
					borderless
					no-error-icon
					v-model.trim="containerConfig.port"
					class="form-item-input"
					input-class="text-ink-2"
					:rules="ruleConfig.websitePort.rules"
					:placeholder="ruleConfig.websitePort.placeholder"
					:disable="disabledFields?.port"
				>
				</q-input>
			</card-form-item>
		</q-card-section>
	</q-card>
</template>

<script lang="ts" setup>
import { reactive, watch, ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { ruleConfig } from './../../types/config';
import CardFormItem from './../common/CardFormItem.vue';

interface Props {
	defaultValues?: {
		image?: string;
		startCmd?: string;
		startCmdArgs?: string;
		port?: string;
	};
	disabledFields?: {
		image?: boolean;
		startCmd?: boolean;
		startCmdArgs?: boolean;
		port?: boolean;
	};
	hiddenFields?: {
		image?: boolean;
		startCmd?: boolean;
		startCmdArgs?: boolean;
		port?: boolean;
	};
}

const props = withDefaults(defineProps<Props>(), {
	defaultValues: () => ({}),
	disabledFields: () => ({}),
	hiddenFields: () => ({})
});

const emits = defineEmits(['updateContainer']);

const { t } = useI18n();

const containerConfig = reactive({
	image: '',
	startCmd: '',
	startCmdArgs: '',
	port: '80'
});

onMounted(() => {
	if (props.defaultValues?.image) {
		containerConfig.image = props.defaultValues.image;
	}
	if (props.defaultValues?.startCmd) {
		containerConfig.startCmd = props.defaultValues.startCmd;
	}
	if (props.defaultValues?.startCmdArgs) {
		containerConfig.startCmdArgs = props.defaultValues.startCmdArgs;
	}
	if (props.defaultValues?.port) {
		containerConfig.port = props.defaultValues.port;
	}
});

const imageRef = ref();
const portRef = ref();

const updateContainer = () => {
	emits('updateContainer', containerConfig);
};

watch(
	() => containerConfig,
	() => {
		updateContainer();
	},
	{
		deep: true
	}
);

const validate = () => {
	imageRef.value.validate();
	portRef.value.validate();

	if (imageRef.value.hasError || portRef.value.hasError) {
		return false;
	}

	return true;
};

defineExpose({
	validate
});
</script>

<style lang="scss" scoped>
.image-container {
	margin: 20px 20px 0 20px;
	padding: 4px;
	border-radius: 12px;
	background-color: $background-1;
}
</style>
../../types/config../common/CardFormItem.vue

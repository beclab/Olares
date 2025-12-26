<template>
	<terminus-edit
		v-model="integrationInputs.accessKeyID"
		:label="t('integration.access_key_id')"
		:show-password-img="false"
		@update:model-value="setButtonStatus"
	/>
	<terminus-edit
		v-model="integrationInputs.accessKeySecret"
		class="q-mt-lg"
		:label="t('integration.access_key_secret')"
		:show-password-img="false"
		@update:model-value="setButtonStatus"
	/>
	<terminus-edit
		v-model="integrationInputs.endpoint"
		class="q-mt-lg"
		:label="t('integration.endpoint')"
		:show-password-img="false"
		@update:model-value="setButtonStatus"
	/>
	<terminus-edit
		v-model="integrationInputs.bucket"
		class="q-mt-lg"
		:label="t('bucket')"
		:show-password-img="false"
		:hint-text="t('integration.optional')"
		@update:model-value="setButtonStatus"
	/>
</template>

<script setup lang="ts">
import { AccountType } from '@bytetrade/core';
import TerminusEdit from '../../../../components/common/TerminusEdit.vue';
import { useI18n } from 'vue-i18n';
import { PropType, reactive } from 'vue';

const props = defineProps({
	accountType: {
		type: Object as PropType<AccountType>
	},
	buttonStatus: {
		type: Boolean,
		required: false,
		default: true
	}
});

const integrationInputs = reactive({
	accessKeyID: '',
	accessKeySecret: '',
	endpoint: '',
	bucket: ''
});

const { t } = useI18n();

const emit = defineEmits(['update:buttonStatus', 'update:buttonText']);

function setButtonStatus() {
	if (
		!integrationInputs.accessKeyID ||
		!integrationInputs.accessKeySecret ||
		!integrationInputs.endpoint
	) {
		emit('update:buttonStatus', false);
	} else {
		emit('update:buttonStatus', true);
	}
}

const allAccountValues = () => {
	return {
		name: integrationInputs.accessKeyID.trim(),
		type: props.accountType,
		raw_data: {
			expires_in: 0,
			expires_at: 0,
			refresh_token: integrationInputs.accessKeySecret.trim(),
			access_token: integrationInputs.accessKeySecret.trim(),
			endpoint: integrationInputs.endpoint.trim(),
			bucket: integrationInputs.bucket.trim()
		}
	};
};

defineExpose({ allAccountValues });
</script>

<style scoped lang="scss"></style>

<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="isParty ? t('third_party_domain') : t('third_level_domain')"
		:skip="false"
		size="medium"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		:okDisabled="!data || data.length == 0 || !isValidThirdLevelDomain(data)"
		:platform="deviceStore.platform"
		@onSubmit="onOKClick"
	>
		<terminus-edit
			:label="isParty ? t('third_party_domain') : t('third_level_domain')"
			v-model="data"
			:is-error="data.length > 0 && !isValidThirdLevelDomain(data)"
			:error-message="t('invalid_route_id')"
		/>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import TerminusEdit from '../../../base/TerminusEdit.vue';
import { useI18n } from 'vue-i18n';
import { useDeviceStore } from 'src/stores/settings/device';
import { isValidThirdLevelDomain } from 'src/components/settings/application/dialog/domain/isValidDomain';

defineProps({
	isParty: {
		type: Boolean,
		require: true
	}
});

const data = ref('');
const { t } = useI18n();

const CustomRef = ref();
const deviceStore = useDeviceStore();
async function onOKClick() {
	CustomRef.value.onDialogOK({
		data: data.value
	});
}
</script>

<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('add_account')"
		:skip="false"
		:ok="t('confirm')"
		size="medium"
		:platform="deviceStore.platform"
		:cancel="t('cancel')"
		@onSubmit="accountCreate"
	>
		<IntegrationAddList @itemClick="setItem" :backup="backup" />
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { IntegrationAccountInfo } from 'src/services/abstractions/integration/integrationService';
import ReminderDialogComponent from 'src/components/settings/ReminderDialogComponent.vue';
import IntegrationAddList from '../components/IntegrationAddList.vue';
import integraionService from 'src/services/integration/index';
import { useDeviceStore } from 'src/stores/device';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { ref } from 'vue';

const { t } = useI18n();

const $q = useQuasar();

const props = defineProps({
	backup: {
		type: Boolean,
		default: false
	}
});

const CustomRef = ref();
const deviceStore = useDeviceStore();
const cItem = ref<IntegrationAccountInfo | undefined>(
	props.backup
		? integraionService.supportBackupList[0]
		: integraionService.supportAuthList[0]
);
const accountCreate = async () => {
	if (!cItem.value) {
		return;
	}

	const webSupport = await integraionService.webSupport(cItem.value.type);
	if (!webSupport.status) {
		$q.dialog({
			component: ReminderDialogComponent,
			componentProps: {
				title: t('add_account'),
				message: webSupport.message,
				useCancel: false,
				confirmText: t('confirm')
			}
		});
		return;
	}

	CustomRef.value.onDialogCancel();
	console.log(props.backup);
	integraionService.getInstanceByType(cItem.value.type)?.signIn({
		quasar: $q,
		backup: props.backup
	});
};

const setItem = (item: IntegrationAccountInfo) => {
	cItem.value = item;
};
</script>

<style scoped lang="scss">
.cpu-core {
	text-align: right;
}
</style>

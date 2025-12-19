<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('Add ACL')"
		:skip="false"
		:ok="t('confirm')"
		size="medium"
		:cancel="t('cancel')"
		:platform="deviceStore.platform"
		:okDisabled="!isUpdateAble"
		@onSubmit="updateAcl"
	>
		<terminus-edit
			v-model="port"
			:label="t('port')"
			:show-password-img="false"
			style="width: 100%"
			:is-error="port.length > 0 && cpuLimitRule(port).length > 0"
			:error-message="cpuLimitRule(port)"
		/>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import TerminusEdit from 'src/components/settings/base/TerminusEdit.vue';
import { useDeviceStore } from 'src/stores/settings/device';
import { PortStatus, useAclStore } from 'src/stores/settings/acl';
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';

const deviceStore = useDeviceStore();

const { t } = useI18n();

const CustomRef = ref();

const aclStore = useAclStore();

const isUpdateAble = computed(() => {
	return cpuLimitRule(port.value).length == 0;
});

const cpuLimitRule = (val: string) => {
	if (val.length === 0) {
		return t('errors.cpu_limit_is_empty');
	}
	let rule = /^\d*$/;
	if (!rule.test(val)) {
		return t('errors.only_valid_numbers_can_be_entered');
	}
	const port = aclStore.allPortAppAclList.find(
		(e) =>
			(e.port.startsWith('*:') ? e.port.substring(2) : e.port) === val &&
			e.appName == 'olares-app'
	);

	if (port && port.status != PortStatus.Remove) {
		return t('Port number added');
	}
	return '';
};

const port = ref('');

const updateAcl = () => {
	console.log('add');
	CustomRef.value.onDialogOK(port.value);
};
</script>

<style scoped lang="scss">
.cpu-core {
	text-align: right;
}
</style>

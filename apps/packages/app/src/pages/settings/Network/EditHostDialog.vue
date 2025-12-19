<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="index > -1 ? t('hosts_edit') : t('add_hosts')"
		:skip="false"
		:ok="t('confirm')"
		size="medium"
		:cancel="t('cancel')"
		:platform="deviceStore.platform"
		:okDisabled="!isUpdateAble"
		@onSubmit="updateHost"
	>
		<terminus-edit
			v-model="hostName"
			:label="t('host_name')"
			:show-password-img="false"
			style="width: 100%"
			class=""
			:is-error="isHostError"
			:error-message="hostErrorMessage(hostName)"
		/>

		<terminus-edit
			v-model="ipAddress"
			:label="t('ip')"
			:show-password-img="false"
			class="q-mt-md"
			:is-error="ipAddress.length > 0 && !validator.isIP(ipAddress)"
			:error-message="ipErrorMessage(ipAddress)"
		/>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import TerminusEdit from 'src/components/settings/base/TerminusEdit.vue';
import { useTerminusDStore } from 'src/stores/settings/terminusd';
import { useDeviceStore } from 'src/stores/settings/device';
import { ref, PropType, onMounted, computed } from 'vue';
import { HostItem } from 'src/constant';
import { useI18n } from 'vue-i18n';
import validator from 'validator';
import { Loading } from 'quasar';

const { t } = useI18n();

const CustomRef = ref();

const deviceStore = useDeviceStore();

const props = defineProps({
	itemList: {
		type: Object as PropType<HostItem[]>,
		default: [] as HostItem[]
	},
	index: {
		type: Number,
		default: -1
	}
});

console.log(props.index);

const hostName = ref('');
const ipAddress = ref('');
const selected = ref<HostItem | undefined>(undefined);

const hostDuplicate = () => {
	if (selected.value) {
		if (
			props.itemList.find((v, index) => {
				if (index === props.index) {
					return false;
				} else {
					return v.host === hostName.value;
				}
			})
		) {
			return true;
		}
	} else {
		if (props.itemList.find((v) => v.host === hostName.value)) {
			return true;
		}
	}
};

const isHostError = computed(() => {
	return (
		(hostName.value && !validator.isFQDN(hostName.value)) || hostDuplicate()
	);
});

const hostErrorMessage = (val: string) => {
	if (val.length === 0) {
		return t('errors.host_is_empty');
	}
	//edit
	if (hostDuplicate()) {
		return t('errors.host_is_error_duplicate');
	}

	if (!validator.isFQDN(val)) {
		return t('errors.host_is_error');
	} else {
		return '';
	}
};

const ipErrorMessage = (val: string) => {
	if (val.length === 0) {
		return t('errors.ip_is_empty');
	}
	if (!validator.isIP(val)) {
		return t('errors.ip_is_error');
	} else {
		return '';
	}
};

const terminusDStore = useTerminusDStore();
const isUpdateAble = computed(() => {
	if (!validator.isFQDN(hostName.value)) {
		return false;
	}

	if (!validator.isIP(ipAddress.value)) {
		return false;
	}

	if (hostDuplicate()) {
		return false;
	}

	if (selected.value) {
		return (
			hostName.value !== selected.value.host ||
			ipAddress.value !== selected.value.ip
		);
	} else {
		return true;
	}
});

onMounted(() => {
	if (props.itemList.length > 0 && props.itemList[props.index]) {
		selected.value = props.itemList[props.index];
		hostName.value = selected.value.host;
		ipAddress.value = selected.value.ip;
	}
});

const updateHost = async () => {
	if (!isUpdateAble.value) {
		return;
	}

	Loading.show();

	let updateList: HostItem[] = [];
	console.log(selected.value);
	//edit
	if (selected.value) {
		updateList = props.itemList.map((item, index) => {
			if (props.index === index) {
				return {
					host: hostName.value,
					ip: ipAddress.value
				} as HostItem;
			} else {
				return item;
			}
		});
	} else {
		props.itemList.forEach((item) => {
			updateList.push(item);
		});

		updateList.push({
			host: hostName.value,
			ip: ipAddress.value
		});
	}

	console.log(updateList);

	try {
		await terminusDStore.updateHostsList(updateList);
		CustomRef.value.onDialogOK(updateList);
	} catch (error: any) {
		console.log(error);
	} finally {
		Loading.hide();
	}
};
</script>

<style scoped lang="scss">
.cpu-core {
	text-align: right;
}
</style>

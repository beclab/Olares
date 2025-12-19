<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('Add repo')"
		:skip="false"
		:ok="t('confirm')"
		size="medium"
		:cancel="t('cancel')"
		:platform="deviceStore.platform"
		:okDisabled="!isUpdateEnable"
		@onSubmit="updateEndpoints"
	>
		<terminus-edit
			v-model="repoName"
			:label="t('Repo name')"
			:show-password-img="false"
			style="width: 100%"
			class=""
		/>

		<terminus-edit
			v-model="endpoint"
			:label="t('Starting endpoint')"
			:show-password-img="false"
			class="q-mt-md"
		/>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import TerminusEdit from '../../../../../components/settings/base/TerminusEdit.vue';
import { useDeviceStore } from '../../../../../stores/settings/device';
import { ref, onMounted, computed } from 'vue';
import { useI18n } from 'vue-i18n';

// const props = defineProps({
// 	endpoints: {
// 		type: Array as PropType<string[]>,
// 		required: false,
// 		default: [] as string[]
// 	}
// });

const { t } = useI18n();

const CustomRef = ref();

const deviceStore = useDeviceStore();

const repoName = ref('');
const endpoint = ref('');

onMounted(() => {});

// const endpointDuplicate = () => {
// 	if (
// 		props.endpoints.find((v) => {
// 			return v == endpoint.value;
// 		})
// 	) {
// 		return true;
// 	}
// };

const isUpdateEnable = computed(() => {
	if (
		!repoName.value ||
		repoName.value.length == 0 ||
		!endpoint.value ||
		endpoint.value.length == 0
	) {
		return false;
	}
	// if (endpointDuplicate()) {
	// 	return false;
	// }
	return true;
});

const updateEndpoints = async () => {
	CustomRef.value.onDialogOK({
		repoName: repoName.value,
		endpoint: endpoint.value
	});
};
</script>

<style scoped lang="scss">
.cpu-core {
	text-align: right;
}
</style>

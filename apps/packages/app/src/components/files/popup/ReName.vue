<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('prompts.rename')"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		:ok-loading="submitLoading ? t('loading') : false"
		size="medium"
		@onSubmit="submit"
	>
		<div
			class="dialog-desc"
			:style="{ textAlign: isMobile ? 'center' : 'left' }"
		>
			<input
				class="input input--block text-ink-1"
				v-focus
				type="text"
				v-model.trim="name"
				@keyup.enter="submit"
			/>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { useQuasar } from 'quasar';
import { ref, onMounted } from 'vue';
import { dataAPIs } from '../../../api';
import { useFilesStore } from '../../../stores/files';
import { DriveType } from '../../../utils/interface/files';

import { useI18n } from 'vue-i18n';

const props = defineProps({
	item: {
		type: Object,
		required: false
	}
});

const $q = useQuasar();
const filesStore = useFilesStore();

const dataAPI = dataAPIs(DriveType.Sync);
const name = ref('');
const submitLoading = ref(false);
const isMobile = $q.platform.is.mobile;
const CustomRef = ref();
const { t } = useI18n();

onMounted(() => {
	name.value = props.item!.name;
});

const submit = async () => {
	if (name.value.length == 0) {
		return;
	}

	submitLoading.value = true;

	try {
		await dataAPI.renameRepo(props.item, name.value);
		submitLoading.value = false;
		CustomRef.value.onDialogOK();
		await filesStore.getMenu();
	} catch (error) {
		submitLoading.value = false;
	}
};
</script>

<style lang="scss" scoped>
.input {
	border-radius: 5px;
	border: 1px solid $input-stroke;
	background-color: transparent;
	&:focus {
		border: 1px solid $yellow-disabled;
	}
}
</style>

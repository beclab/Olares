<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('files.new_library')"
		:ok="t('create')"
		:cancel="t('cancel')"
		:okLoading="loading ? t('loading') : false"
		:size="$q.platform.is.mobile ? 'small' : 'medium'"
		:platform="$q.platform.is.mobile ? 'mobile' : 'web'"
		:okDisabled="!name || errorMessage.length > 0"
		@onSubmit="submit"
		@onHide="close"
		@onCancel="close"
	>
		<div class="card-content">
			<div class="text-body3 text-ink-3 q-mb-xs">
				{{ t('please_enter_a_library_name') }}
			</div>
			<input
				class="input input--block text-ink-1"
				v-focus
				type="text"
				v-model.trim="name"
				@keyup.enter="submit"
			/>
			<div class="text-negative" v-if="errorMessage">{{ errorMessage }}</div>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { syncUtil } from './../../../api';
import { useDataStore } from '../../../stores/data';
import { useFilesStore } from '../../../stores/files';

import { useI18n } from 'vue-i18n';
import { busEmit } from 'src/utils/bus';
import { notifyWarning } from 'src/utils/notifyRedefinedUtil';
import { useFileNameValidation } from 'src/composables/files/useFileNameValidation';

const CustomRef = ref();

const store = useDataStore();
const filesStore = useFilesStore();

const name = ref<string>('');
const loading = ref(false);

const { t } = useI18n();
const { errorMessage } = useFileNameValidation(name);

const submit = async () => {
	if (!name.value) {
		return false;
	}
	if (errorMessage.value) {
		notifyWarning(errorMessage.value);
		return false;
	}
	if (loading.value) {
		return;
	}
	loading.value = true;
	try {
		await syncUtil().createLibrary(name.value);
		loading.value = false;
	} catch (e) {
		loading.value = false;
	}

	busEmit('reposUpdate');

	store.closeHovers();
};

const close = () => {
	store.closeHovers();
};
</script>

<style lang="scss" scoped>
.card-content {
	padding: 0 0;
	.input {
		border-radius: 5px;
		border: 1px solid $input-stroke;
		background-color: transparent;
		&:focus {
			border: 1px solid $theme-input-focus-border;
		}
	}
}
</style>

<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('vault_t.upload_attachment')"
		:ok="t('vault_t.upload')"
		:cancel="t('vault_t.discard')"
		:size="$q.platform.is.mobile ? 'small' : 'medium'"
		:platform="$q.platform.is.mobile ? 'mobile' : 'web'"
		@onSubmit="onOKClick"
		:okLoading="loading ? t('transport_uploading') : ''"
	>
		<q-card-section class="q-pt-md q-px-none">
			<div class="text-color-title row flex flex-start">
				<div class="text-body3 text-ink-3 q-mb-sm">
					{{ t('vault_t.attachment_name') }}
				</div>
				<q-input
					class="uploadAtta"
					popup-content-class="options_selected_Account"
					borderless
					dense
					color="text-grey-4"
					v-model="name"
					input-style="padding: 0 8px;"
				/>
			</div>

			<q-card-section
				v-if="error"
				class="row flex flex-start q-pa-none q-pl-sm q-py-sm text-body3 text-red"
			>
				{{ error || t('error.index') }}
			</q-card-section>

			<q-card-section
				align="left"
				class="q-pl-sm q-py-sm text-body3 text-ink-2"
			>
				{{
					progress
						? t('vault_t.uploading_loaded_total', {
								loaded: format.formatFileSize(progress.loaded),
								total: format.formatFileSize(progress.total)
						  })
						: (getFileType(file.name) || t('vault_t.unkown_file_type')) +
						  ' - ' +
						  format.formatFileSize(file.size)
				}}
			</q-card-section>
		</q-card-section>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { app } from '../../globals';
import { ErrorCode } from '@didvault/sdk/src/core';
import { useI18n } from 'vue-i18n';
import { getFileType } from '@bytetrade/core';
import { format } from '../../utils/format';
import { useQuasar } from 'quasar';

const props = defineProps({
	itemID: {
		type: String,
		required: true
	},
	file: {
		type: File,
		required: true
	}
});

const loading = ref(false);
const error = ref('');
const name = ref(props.file.name);
let progress: any = ref(null);
const $q = useQuasar();

async function onOKClick() {
	if (loading.value) {
		return;
	}
	loading.value = true;
	progress.value = null;
	error.value = '';
	if (!name.value) {
		error.value = t('please_enter_a_name');
		return;
	}

	const att = await app.createAttachment(
		props.itemID!,
		props.file!,
		name.value
	);

	const upload = att.uploadProgress!;

	const handler = () => (progress.value = upload.uploadProgress);
	upload.addEventListener('progress', handler);
	try {
		await upload.completed;
	} catch (e) {
		console.error(e);
	}
	upload.removeEventListener('progress', handler);

	progress.value = null;
	loading.value = false;
	if (upload.error) {
		error.value =
			upload.error.code === ErrorCode.PROVISIONING_QUOTA_EXCEEDED
				? t('vault_t.storage_limit_exceeded')
				: t('vault_t.upload_failed_please_try_again');
	} else {
		loading.value = false;
		onDialogOK();
	}
}

const { t } = useI18n();
const CustomRef = ref();

const onDialogOK = () => {
	CustomRef.value.onDialogOK();
};
</script>

<style lang="scss" scoped>
.uploadAtta {
	width: 100%;
	border: 1px solid $input-stroke;
	border-radius: 8px;
	overflow: hidden;

	::v-deep(.q-field__control) {
		background: $background-1 !important;
		color: $ink-2;
	}

	::v-deep(.q-field__native) {
		color: $ink-2;
	}
}
.d-creatVault {
	.q-dialog-plugin {
		width: 400px;
		border-radius: 12px;

		.confirm {
			padding: 8px 12px;
		}
		.reset {
			padding: 8px 12px;
			border: 1px solid $separator;
		}

		.but-cancel-web {
			border-radius: 6px;
			border: 1px solid $separator;
		}
	}
}
</style>

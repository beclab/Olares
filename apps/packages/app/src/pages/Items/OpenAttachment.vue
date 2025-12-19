<template>
	<q-dialog
		ref="dialogRef"
		persistent
		:maximized="maximizedToggle"
		transition-show="slide-up"
		transition-hide="slide-down"
	>
		<q-card class="q-dialog-plugin column items-center justify-start">
			<q-card-section class="row items-center justify-between full-width">
				<div class="text-h6">{{ info.name }}</div>
				<div class="row items-center justify-center">
					<span class="operate q-mr-sm">
						<BtIcon src="trash3" style="padding: 6px" @click="onDelete" />
					</span>
					<span class="operate q-mr-sm" v-if="_attachment && info">
						<BtIcon src="download" style="padding: 6px" @click="onSaveDisk" />
					</span>
					<q-btn icon="close" flat dense @click="onDialogCancel" />
				</div>
			</q-card-section>
			<q-card-section
				class="q-pt-xs row items-center justify-center"
				style="height: calc(100% - 64px); width: 100vw; overflow: scroll"
			>
				<template v-if="_preview && mType == 'pdf'">
					<object
						class="content preview pdf stretch"
						type="application/pdf"
						:data="objectContent"
					></object>
				</template>
				<template v-else-if="_preview && mType == 'image'">
					<div
						class="content preview image stretch row items-center justify-center"
						style="width: 80vw; height: 80vh"
					>
						<img :src="objectContent" style="max-width: 100%" />
					</div>
				</template>
				<template v-else-if="_preview && (mType == 'text' || mType == 'code')">
					<pre
						class="content preview ${mType} stretch"
					><code>{{ objectContent }}</code></pre>
				</template>
				<template v-else>
					<div
						class="stretch centering vertical layout column items-center justify-center"
					>
						<div class="ellipis bold">
							{{ info.type || t('vault_t.unkown_file_type') }}
						</div>

						<div class="padded margined inverted red card" v-if="!_error">
							{{ _error }}
						</div>

						<div class="size" v-if="!!_error">
							{{
								_progress
									? t('vault_t.uploading_loaded_total', {
											loaded: format.formatFileSize(_progress.loaded),
											total: format.formatFileSize(_progress.total)
									  })
									: format.formatFileSize(info.size)
							}}
						</div>

						<div class="noPreview q-pa-sm">
							{{ t('vault_t.no_preview_available') }}
						</div>
					</div>
				</template>
			</q-card-section>
		</q-card>
	</q-dialog>
</template>

<script lang="ts" setup>
import { ref, computed } from 'vue';
import { useDialogPluginComponent, useQuasar } from 'quasar';
import { AttachmentInfo } from '@didvault/sdk/src/core';
import { app } from '../../globals';
import { mediaType } from '@didvault/sdk/src/util';
import { saveFile } from '@didvault/sdk/src/core';
import { Dialog } from 'quasar';
import { useI18n } from 'vue-i18n';
import { format } from '../../utils/format';
import { BtDialog } from '@bytetrade/ui';

const props = defineProps({
	itemID: {
		type: String,
		required: true
	},
	info: {
		type: AttachmentInfo,
		required: true
	}
});

const { dialogRef, onDialogOK, onDialogCancel } = useDialogPluginComponent();

const { t } = useI18n();

let _error = ref('');
let _progress: any = ref(null);
let _preview = ref(false);
let _attachment: any = ref(null);
let objectContent = ref();
let maximizedToggle = ref(true);
const $q = useQuasar();

const mType = computed(function () {
	if (!props.info) {
		return null;
	}
	return mediaType(props.info.type);
});

async function download() {
	if ($q.platform.is.nativeMobile && props.info.size > 5 * 1024 * 1024) {
		return;
	}
	_progress.value = null;
	_error.value = '';

	const att = await app.downloadAttachment(props.info!);
	const download = att.downloadProgress!;
	const handler = () => (_progress.value = download.progress);

	download.addEventListener('progress', handler);
	try {
		await download.completed;
	} catch (e) {
		console.error(e);
	}
	download.removeEventListener('progress', handler);

	_progress.value = null;

	if (download.error) {
		_error.value = t('vault_t.download_failed');
	} else {
		_attachment.value = att;

		const type = mediaType(props.info.type);
		if (type == 'pdf') {
			objectContent.value = await att.toObjectURL();
		} else if (type == 'image') {
			objectContent.value = await att.toObjectURL();
		} else if (type == 'text' || type == 'code') {
			objectContent.value = await att.toText();
		} else {
			objectContent.value = null;
		}

		_preview.value = true;
	}
}

download();

async function onSaveDisk() {
	if (!_attachment.value || !props.info) {
		throw new Error(t('vault_t.need_to_download_attachment_first'));
	}

	BtDialog.show({
		title: t('vault_t.save_to_disk'),
		message: t('vault_t.save_to_disk_message'),
		okStyle: {
			background: 'yellow-default',
			color: '#1F1F1F'
		},
		cancel: true,
		okText: t('base.confirm'),
		cancelText: t('base.cancel')
	})
		.then(async (res: any) => {
			if (res) {
				onDialogOK(false);
				saveFile(
					props.info.name,
					props.info.type,
					await _attachment.value.getData()
				);
			}
		})
		.catch((err: Error) => {
			console.log('click cancel', err);
		});
}

async function onDelete() {
	BtDialog.show({
		title: t('vault_t.delete_attachment'),
		message: t('vault_t.delete_attachment_message'),
		okStyle: {
			background: 'yellow-default',
			color: '#1F1F1F'
		},
		cancel: true,
		okText: t('base.confirm'),
		cancelText: t('base.cancel')
	})
		.then(async (res: any) => {
			if (res) {
				await app.deleteAttachment(props.itemID, props.info!);
				onDialogOK(true);
			}
		})
		.catch((err: Error) => {
			console.log('click cancel', err);
		});
}
</script>

<style lang="scss" scoped>
.operate {
	&:hover {
		border-radius: 4px;
	}
}

.noPreview {
	margin: 20px auto 0;
	display: inline-block;
	border: 1px solid $separator;
	border-radius: 5px;
}

.content {
	width: 60%;
	overflow: hidden;
	overflow-wrap: break-word;
	overflow: scroll;
}
</style>

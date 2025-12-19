<template>
	<q-dialog
		class="upload-dialog"
		position="bottom"
		ref="dialogRef"
		transition-show="jump-up"
		transition-hide="jump-down"
		transition-duration="300"
	>
		<div class="upload bg-background-2">
			<div class="header text-ink-1 text-subtitle2">
				{{ t('files.add_cloud_drive') }}
			</div>
			<div class="upload-body row items-center justify-between">
				<div
					class="body-item column items-center justify-center"
					@click="uploadFiles"
				>
					<q-icon name="sym_r_folder_copy" size="24px" />
					<div class="q-mt-sm">{{ t('files.select_file') }}</div>
				</div>

				<div
					v-if="$q.platform.is.nativeMobile && $q.platform.is.ios"
					class="body-item column items-center justify-center"
					@click="uploadImgOrVideo"
				>
					<q-icon name="sym_r_folder_copy" size="24px" />
					<div class="q-mt-sm">{{ t('files.select_image_video') }}</div>
				</div>

				<div
					class="body-item column items-center justify-center"
					@click="createFolder"
				>
					<q-icon name="sym_r_create_new_folder" size="24px" />
					<div class="q-mt-sm">{{ t('files.create_folder') }}</div>
				</div>
			</div>
		</div>
	</q-dialog>
</template>

<script setup lang="ts">
import { useDialogPluginComponent } from 'quasar';
import { useI18n } from 'vue-i18n';

const { dialogRef, onDialogOK } = useDialogPluginComponent();

const { t } = useI18n();

const uploadFiles = () => {
	onDialogOK('uploadFiles');
};

const createFolder = () => {
	onDialogOK('createFolder');
};

const uploadImgOrVideo = () => {
	onDialogOK('uploadImgOrVideo');
};
</script>

<style lang="scss" scoped>
.upload-dialog {
	.upload {
		width: 84% !important;
		height: 152px;
		border-radius: 12px;
		margin-bottom: 24px;
		overflow: hidden;
		padding: 0 20px;

		.header {
			width: 100%;
			height: 52px;
			line-height: 52px;
		}

		.upload-body {
			width: 100%;
			display: flex;
			flex-wrap: wrap;
			gap: 10px;
		}

		.body-item {
			flex: 1 1 80px;
			height: 76px;
			border-radius: 8px;
			border: 1px solid $separator;
		}
	}
}
</style>

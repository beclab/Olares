<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('files.upload_files')"
		:skip="false"
		:okLoading="false"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		size="medium"
		:okDisabled="!fileSavePathRef"
		@onSubmit="submit"
	>
		<div class="dialog-content dialog-padding q-py-md">
			<div v-if="files && files.length > 0">
				<TransferAnalysis
					:title="files[0].name"
					:detail="format.formatFileSize(files[0].size)"
				/>
				<div class="input-title text-body3 text-ink-3 q-mt-lg">
					{{ t('Upload to') }}
				</div>
				<TransfetSelectTo
					class="q-mt-xs"
					@setSelectPath="setSelectPath"
					:origins="origins"
				/>
			</div>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import TransferAnalysis from './TransferAnalysis.vue';
import TransfetSelectTo from './TransfetSelectTo.vue';
import { FilePath, useFilesStore } from 'src/stores/files';
import { DriveType } from 'src/utils/interface/files';
import { format } from 'src/utils/format';
import { filesIsV2 } from 'src/api';
import { PropType, ref } from 'vue';
import { i18n } from 'src/boot/i18n';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	confirmBtnTitle: {
		type: String,
		default: i18n.global.t('confirm'),
		required: false
	},
	cancelBtnTitle: {
		type: String,
		default: i18n.global.t('cancel'),
		required: false
	},
	files: {
		type: Object as PropType<
			{
				name: string;
				size: number;
			}[]
		>,
		required: true
	},
	origins: {
		type: Array as PropType<DriveType[]>,
		required: false,
		default: () => {
			if (filesIsV2()) {
				return [
					DriveType.Drive,
					// 0730 hide sync
					// DriveType.Sync,
					DriveType.External,
					DriveType.Cache,
					DriveType.Data,
					DriveType.GoogleDrive
				];
			}
			return [
				DriveType.Drive,
				DriveType.External,
				DriveType.Sync,
				DriveType.Cache,
				DriveType.Data,
				DriveType.GoogleDrive
			];
		}
	}
});

console.log('props ===>', props.origins);

const { t } = useI18n();

const CustomRef = ref();

const filesStore = useFilesStore();

const fileSavePathRef = ref<FilePath | undefined>(filesStore.currentPath[1]);

const setSelectPath = (fileSavePath: FilePath) => {
	fileSavePathRef.value = fileSavePath;
};

const submit = async () => {
	CustomRef.value.onDialogOK(fileSavePathRef.value);
};
</script>

<style lang="scss" scoped>
.card-dialog {
	.card-continer {
		width: 400px;
		border-radius: 12px;

		.dialog-padding {
			padding-left: 20px;
			padding-right: 20px;
		}

		.dialog-content {
			.progress {
				margin-bottom: 60px;
				margin-top: 20px;
				border: 1px solid $separator;
				border-radius: 10px;
				overflow: hidden;
			}
		}

		.common-border {
			border: 1px solid $separator;
			border-radius: 8px;
			height: 92px;
			overflow: hidden;
			// padding-top: 8px;
			// padding-bottom: 8px;
		}
	}
}
</style>

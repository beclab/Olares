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
				<bt-select-v3
					v-if="historySize > 0"
					:model-value="selector"
					:options="options"
					:fallback-label="fallbackLabel"
					:header-value="headerPath"
					show-selected-icon
					height="32px"
					class="q-mt-xs"
					@update:modelValue="setOptions"
					@header-click="onHeaderClick"
					@footer-click="onSelectFolder"
				>
					<template v-if="headerPath" #header>
						{{ decodeUrl(headerPath) }}
					</template>
					<template #footer>
						{{ t('Select...') }}
					</template>
				</bt-select-v3>
				<TransfetSelectTo
					ref="selectToDialog"
					v-show="historySize <= 0"
					class="q-mt-xs"
					@setSelectPath="setSelectPath"
					@setSelectCancel="setSelectCancel"
					:origins="origins"
				/>
			</div>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import TransferAnalysis from './TransferAnalysis.vue';
import TransfetSelectTo from './TransfetSelectTo.vue';
import BtSelectV3 from 'src/components/settings/base/BtSelectV3.vue';
import {
	getUploadHistory,
	saveUploadHistory
} from 'src/pages/Electron/Transfer/fileHistory';
import { FilePath, useFilesStore } from 'src/stores/files';
import { DriveType } from 'src/utils/interface/files';
import { computed, PropType, ref } from 'vue';
import { SelectorProps } from 'src/constant';
import { decodeUrl } from 'src/utils/encode';
import { format } from 'src/utils/format';
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
			return [
				DriveType.Drive,
				DriveType.Sync,
				DriveType.External,
				DriveType.Cache,
				DriveType.Data,
				DriveType.GoogleDrive
			];
		}
	},
	historySize: {
		type: Number,
		default: 0
	},
	defaultPath: {
		type: String,
		default: ''
	}
});

const { t } = useI18n();
const CustomRef = ref();
const selectToDialog = ref();
const filesStore = useFilesStore();
function filePathFromPath(path: string): FilePath {
	return new FilePath({
		path,
		isDir: true,
		driveType: DriveType.Drive,
		param: ''
	});
}

const selector = ref(props.defaultPath || '');
const fileSavePathRef = ref<FilePath | undefined>(
	props.defaultPath
		? filePathFromPath(props.defaultPath)
		: filesStore.currentPath[1]
);

const setSelectPath = (fileSavePath: FilePath) => {
	fileSavePathRef.value = fileSavePath;
	selector.value = fileSavePath.path;
};

const setSelectCancel = () => {
	selector.value = props.defaultPath || '';
	fileSavePathRef.value = props.defaultPath
		? filePathFromPath(props.defaultPath)
		: filesStore.currentPath[1];
};

const setOptions = (value: string) => {
	selector.value = value;
	const history = getUploadHistory(props.historySize);
	const pathInHistory = history.find((item) => item.path === value);
	if (pathInHistory) {
		fileSavePathRef.value = pathInHistory;
	} else if (value) {
		fileSavePathRef.value = filePathFromPath(value);
	}
};

const onSelectFolder = () => {
	selectToDialog.value.selectFolder();
};

const onHeaderClick = () => {
	const path = headerPath.value;
	if (path) {
		selector.value = path;
		fileSavePathRef.value = filePathFromPath(path);
	}
};

const options = computed(() => {
	if (props.historySize <= 0) return [];
	const history = getUploadHistory(props.historySize);
	return history.map(
		(item) =>
			({
				label: decodeUrl(item.path),
				value: item.path
			} as SelectorProps)
	);
});

const headerPath = computed(() => {
	const current =
		fileSavePathRef.value?.path ?? selector.value ?? props.defaultPath ?? '';
	if (!current) return '';
	const history = getUploadHistory(props.historySize);
	return history.some((item) => item.path === current) ? '' : current;
});

const fallbackLabel = computed(() =>
	selector.value ? decodeUrl(selector.value) : ''
);

const submit = async () => {
	if (props.historySize && fileSavePathRef.value) {
		saveUploadHistory(fileSavePathRef.value);
	}
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

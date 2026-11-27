<template>
	<bt-custom-dialog
		ref="dialogRef"
		:title="t('files.archive.compress_to_archive')"
		:ok="t('files.archive.create_archive')"
		:cancel="t('cancel')"
		size="medium"
		:platform="deviceStore.isMobile ? 'mobile' : 'web'"
		:position="deviceStore.isMobile ? 'bottom' : 'standard'"
		:okDisabled="onDisabled"
		:okLoading="loading ? t('loading') : false"
		@onSubmit="submit"
	>
		<div class="archive-dialog">
			<div class="archive-grid">
				<terminus-edit
					v-model="archiveName"
					title-class="archive-name-title"
					:label="t('files.archive.archive_name')"
					:hint-text="t('files.archive.enter_archive_name')"
					:inputHeight="40"
				/>
				<div>
					<div class="archive-name-title q-mb-xs">
						{{ t('files.archive.save_to') }}
					</div>
					<TransfetSelectTo
						ref="selectToDialog"
						class="q-mt-xs"
						:selectHeight="42"
						:defaultPathNumber="0"
						@setSelectPath="setSelectPath"
						@setSelectCancel="setSelectCancel"
						:origins="[driveType as DriveType]"
					/>
				</div>
			</div>
			<div>
				<div class="archive-name-title q-mb-sx">
					{{ $t('files.archive.format') }}
				</div>
				<bt-select
					v-model="format"
					:options="availableArchiveFormatOptions"
					:border="true"
				/>
				<div
					v-if="!isZipOrSeven"
					class="text-body3 text-orange-default q-mt-sm"
				>
					{{
						t('files.archive.format_feature_limited', { format: formatLabel })
					}}
				</div>
			</div>

			<div>
				<div class="text-body3 text-ink-3 q-mb-xs">
					{{ t('files.archive.compression_level') }}
				</div>
				<div class="row items-center no-wrap">
					<span>{{ t('files.archive.fast') }}</span>
					<q-slider
						v-model="compressionLevelNumber"
						trackSize="2px"
						trackColor="background-4"
						thumbColor="light-blue-default"
						thumbSize="12px"
						class="q-mx-sm archive-slider archive-slider--static"
						:min="1"
						:max="9"
						color="light-blue-default"
					/>
					<span>{{ t('files.archive.smallest') }}</span>
					<terminus-edit
						v-model="compressionLevel"
						class="archive-level-input"
						:inputMarginTop="0"
						:is-error="compressionLevelError"
						:inputHeight="40"
					/>
				</div>
				<div v-if="compressionLevelError" class="text-body3 text-negative">
					{{ compressionLevelRule(compressionLevel) }}
				</div>
			</div>
			<div class="archive-grid" v-if="isZipOrSeven">
				<terminus-edit
					v-model="password"
					:show-password-img="true"
					:is-error="
						password.length > 0 && password.length < archivePasswordLength
					"
					:error-message="passwordRule(password)"
					title-class="archive-name-title"
					:label="t('files.archive.password_optional')"
					:hint-text="t('files.archive.enter_password')"
					:inputHeight="40"
				/>

				<terminus-edit
					v-model="confirmPassword"
					:show-password-img="true"
					:is-error="confirmPassword.length > 0 && confirmPassword != password"
					:error-message="t('files.archive.password_mismatch')"
					title-class="archive-name-title"
					:label="t('files.archive.confirm_password')"
					:hint-text="t('files.archive.reenter_password')"
					:inputHeight="40"
				/>
			</div>
			<div :class="isZipOrSeven ? 'archive-grid row items-center' : ''">
				<div class="full-width" v-if="isZipOrSeven">
					<div class="text-body3 text-ink-2 q-mb-xs">
						{{ t('files.archive.split_volume_optional') }}
					</div>
					<div class="row no-wrap full-width">
						<terminus-edit
							v-model="splitVolume"
							:inputHeight="40"
							title-class="archive-name-title"
							:hint-text="t('files.archive.split_volume_placeholder')"
							style="flex: 1"
						/>
						<bt-select
							v-model="splitVolumeUnit"
							class="archive-split-unit q-ml-sm q-mt-xs"
							:options="splitVolumeUnitOptions"
							:border="true"
						/>
					</div>
				</div>
				<div>
					<div class="text-body3 text-ink-2 q-mb-xs">
						{{ t('files.archive.on_conflict') }}
					</div>
					<bt-select
						v-model="conflictStrategy"
						:label="t('files.archive.on_conflict')"
						:options="localizedConflictStrategyOptions"
						:border="true"
					/>
				</div>
			</div>
		</div>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { computed, PropType, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import {
	ArchiveConflict,
	ArchiveFormat,
	archiveAdvancedFormatOptions,
	archiveConflictOptions,
	canCompressItemsWithFormat
} from '../../../utils/interface/archive';
import {
	useFilesStore,
	FileItem,
	FilesIdType,
	FilePath
} from '../../../stores/files';
import { DriveType } from '../../../utils/interface/files';
import Taskmanager from '../../../services/olaresTask';
import TerminusEdit from '../../common/TerminusEdit.vue';
import BtSelect from '../../base/BtSelect.vue';
import TransfetSelectTo from '../../../pages/Electron/Transfer/TransfetSelectTo.vue';
import { useDeviceStore } from 'src/stores/settings/device';
const props = defineProps({
	origin_id: {
		type: Number,
		default: FilesIdType.PAGEID
	},
	format: {
		type: String,
		default: ArchiveFormat.Zip
	},
	driveType: {
		type: String,
		default: DriveType.Drive
	},
	defaultPath: {
		type: String,
		default: ''
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
				DriveType.Common,
				DriveType.GoogleDrive
			];
		}
	}
});

function filePathFromPath(path: string): FilePath {
	return new FilePath({
		path,
		isDir: true,
		driveType: props.driveType as DriveType,
		param: ''
	});
}

const { t } = useI18n();
const filesStore = useFilesStore();
const deviceStore = useDeviceStore();
const dialogRef = ref();
const format = ref<ArchiveFormat>(props.format as ArchiveFormat);
const archivePasswordLength = 6;
const compressionLevel = ref('5');
const compressionLevelNumber = computed({
	get() {
		const value = Number(compressionLevel.value);
		return Number.isFinite(value) ? value : 5;
	},
	set(value: number) {
		compressionLevel.value = String(value);
	}
});
const password = ref('');
const confirmPassword = ref('');
const splitVolume = ref('');
const splitVolumeUnit = ref('MB');
const fileSavePathRef = ref<FilePath | undefined>(
	props.defaultPath
		? filePathFromPath(props.defaultPath)
		: filesStore.currentPath[props.origin_id]
);
const conflictStrategy = ref(ArchiveConflict.Rename);
const selector = ref(props.defaultPath || '');
const loading = ref(false);

const splitVolumeUnitOptions = [
	// { label: 'KB', value: 'KB' },
	{ label: 'MB', value: 'MB' },
	{ label: 'GB', value: 'GB' }
];

const conflictStrategyOptions = archiveConflictOptions;
const localizedConflictStrategyOptions = computed(() =>
	conflictStrategyOptions.map((item) => ({
		...item,
		label: t(item.label)
	}))
);

const selectedItems = computed(
	() =>
		filesStore.selected[props.origin_id]
			.map((index: number) =>
				filesStore.getTargetFileItem(index, props.origin_id)
			)
			.filter(Boolean) as FileItem[]
);

const archiveName = ref('');

const isZipOrSeven = computed(() => {
	return (
		format.value === ArchiveFormat.Zip || format.value === ArchiveFormat.SevenZ
	);
});

const passwordMismatch = computed(
	() =>
		isZipOrSeven.value &&
		!!password.value &&
		password.value !== confirmPassword.value
);
const passwordError = computed(
	() =>
		isZipOrSeven.value &&
		password.value.length > 0 &&
		passwordRule(password.value).length > 0
);

const canCompressSelectedItems = computed(() =>
	canCompressItemsWithFormat(selectedItems.value, format.value)
);

const availableArchiveFormatOptions = computed(() =>
	archiveAdvancedFormatOptions.filter((item) =>
		canCompressItemsWithFormat(selectedItems.value, item.value)
	)
);

watch(
	availableArchiveFormatOptions,
	(options) => {
		if (!options.some((item) => item.value === format.value)) {
			format.value = ArchiveFormat.Zip;
		}
	},
	{ immediate: true }
);

const targetArchiveName = () => {
	if (archiveName.value.endsWith('.' + format.value)) {
		return archiveName.value;
	}
	return archiveName.value + '.' + format.value;
};

const splitVolumeMB = () => {
	if (!splitVolume.value) {
		return undefined;
	}

	const value = Number(splitVolume.value);
	if (!Number.isFinite(value) || value <= 0) {
		return undefined;
	}

	const unit = splitVolumeUnit.value.toUpperCase();
	const mbValue =
		unit === 'GB' ? value * 1024 : unit === 'KB' ? value / 1024 : value;
	return Math.max(1, Math.ceil(mbValue));
};

const submit = async () => {
	if (loading.value) {
		return;
	}
	if (
		passwordMismatch.value ||
		passwordError.value ||
		selectedItems.value.length === 0 ||
		!fileSavePathRef.value ||
		!canCompressSelectedItems.value
	) {
		return;
	}

	loading.value = true;
	try {
		const currentPath = filesStore.currentPath[props.origin_id];
		const task = await Taskmanager.archiveTask.createCompressTask({
			items: selectedItems.value,
			parentPath: fileSavePathRef.value.path || currentPath.path,
			driveType: props.driveType as DriveType,
			param: currentPath.param,
			format: format.value,
			targetFileName: targetArchiveName(),
			password: isZipOrSeven.value ? password.value : '',
			splitVolume: splitVolumeMB(),
			conflictStrategy: conflictStrategy.value,
			compressionLevel: Number(compressionLevel.value),
			preserveSymbolicLinks: true,
			node:
				filesStore.currentFileList[props.origin_id]?.node ||
				filesStore.currentNode[props.origin_id]?.name
		});
		if (task) {
			await Taskmanager.addTasks([task]);
		}
		dialogRef.value.onDialogOK();
	} finally {
		loading.value = false;
	}
};

const setSelectPath = (fileSavePath: FilePath) => {
	fileSavePathRef.value = fileSavePath;
	selector.value = fileSavePath.path;
};

const setSelectCancel = () => {
	selector.value = props.defaultPath || '';
	fileSavePathRef.value = props.defaultPath
		? filePathFromPath(props.defaultPath)
		: filesStore.currentPath[props.origin_id];
};

const compressionLevelRule = (val: string) => {
	if (val.length === 0) {
		return 'errors.compression_level_is_empty';
	}
	let rule = /^[+-]?(\d+\.?\d*|\.\d+)$/;
	if (!rule.test(val)) {
		return t('errors.only_valid_numbers_can_be_entered');
	}
	if (Number(val) < 1 || Number(val) > 9) {
		return t('files.archive.compression_level_range');
	}
	return '';
};

const passwordRule = (val: string) => {
	if (val.length < archivePasswordLength)
		return t('errors.at_least_number_digits_long', {
			number: archivePasswordLength
		});
	return '';
};

const compressionLevelError = computed(() => {
	return (
		compressionLevel.value.length > 0 &&
		compressionLevelRule(compressionLevel.value).length > 0
	);
});

const onDisabled = computed(() => {
	return (
		archiveName.value.length === 0 ||
		compressionLevel.value.length === 0 ||
		passwordMismatch.value ||
		passwordError.value ||
		compressionLevelError.value ||
		loading.value ||
		selectedItems.value.length === 0 ||
		!fileSavePathRef.value ||
		!canCompressSelectedItems.value
	);
});

const formatLabel = computed(
	() =>
		availableArchiveFormatOptions.value.find(
			(item) => item.value === format.value
		)?.label || format.value
);
</script>

<style scoped lang="scss">
.archive-dialog {
	width: 520px;
	display: flex;
	flex-direction: column;
	gap: 20px;
}

.archive-grid {
	display: grid;
	grid-template-columns: repeat(2, minmax(0, 1fr));
	gap: 12px;
}

@media (max-width: 599px) {
	.archive-dialog {
		width: calc(100vw - 40px);
		max-width: 100%;
	}

	.archive-grid {
		grid-template-columns: 1fr;
	}
}

.archive-full-row {
	width: 100%;
}

.archive-slider {
	flex: 1;
}

.archive-slider--static {
	:deep(.q-slider__thumb),
	:deep(.q-slider__thumb-shape),
	:deep(.q-slider__focus-ring) {
		transition: none !important;
	}

	:deep(.q-slider__focus-ring) {
		display: none !important;
		transform: none !important;
	}

	:deep(.q-slider__thumb-shape) {
		transform: scale(1) !important;
	}
}

.archive-level-input {
	width: 60px;
	margin-left: 12px;
	height: 40px;
	border-radius: 8px;
}

.archive-split-input {
	flex: 1;
}

.archive-split-unit {
	width: 72px;
}

::v-deep(.archive-grid > .scrollbar > .slider) {
	border-radius: 4px;
}
</style>

<style lang="scss">
.archive-name-title {
	color: $ink-3;
}
</style>

<template>
	<bt-custom-dialog
		ref="CustomRef"
		:skip="false"
		:ok="ok"
		:cancel="cancel"
		:persistent="true"
		size="medium"
		@onCancel="onCancel"
		:okDisabled="onDisabled"
		position="bottom"
		platform="mobile"
		@onSubmit="onSubmit"
	>
		<template v-slot:header>
			<share-back
				:title="
					!shareResult
						? t('files_popup_menu.Share to Public')
						: t('files.Link Details')
				"
				@back="onBack"
			>
				<template v-slot:end>
					<q-icon
						name="sym_r_close"
						color="ink-2"
						size="24px"
						@click="onClose"
					/>
				</template>
			</share-back>
		</template>
		<div
			class="dialog-desc"
			:class="{ 'dialog-content-height': shareResult != undefined }"
		>
			<template v-if="!shareResult">
				<div class="text-ink-2 text-subtitle1">
					{{ t('files.Add password') }}
				</div>
				<GeneratePassword v-model="publicPassword" :copy="true" />

				<div class="text-ink-2 text-subtitle1 q-mt-lg">
					{{ t('files.Set expiration') }}
				</div>
				<div class="row items-center q-mt-sm">
					<bt-check-box-component
						class="col"
						:model-value="setExpirationInDays == true"
						:label="t('files.In days')"
						@update:modelValue="setExpirationInDays = true"
					/>
					<bt-check-box-component
						:model-value="setExpirationInDays == false"
						:label="t('files.Exact date & time')"
						class="col"
						@update:modelValue="setExpirationInDays = false"
					/>
				</div>
				<terminus-edit
					v-if="setExpirationInDays"
					:inputHeight="38"
					v-model="days"
					class="q-mt-sm"
					:is-error="days.length > 0 && datesLimitRule(days).length > 0"
					:error-message="datesLimitRule(days)"
				></terminus-edit>

				<el-config-provider :locale="lang">
					<el-date-picker
						class="q-mt-sm"
						style="width: 100%; height: 40px"
						v-if="!setExpirationInDays"
						v-model="dateValue"
						type="datetime"
						format="YYYY-MM-DD HH:mm"
						value-format="YYYY-MM-DD HH:mm"
						placeholder="YYYY-MM-DD HH:mm"
						popper-class="share-link-date-picker-popper"
						:disabled-date="disabledDate"
						:editable="false"
					>
					</el-date-picker>
				</el-config-provider>

				<div
					class="row items-center justify-between full-width q-mt-lg"
					style="height: 48px"
				>
					<div class="text-ink-2">
						<q-icon name="sym_r_drive_folder_upload" size="24px" />
						<span class="q-ml-md text-subtitle2">
							{{ t('files.Allow upload only') }}
						</span>
					</div>

					<bt-switch
						size="sm"
						truthy-track-color="light-blue-default"
						v-model="uploadOnly"
					/>
				</div>

				<div
					class="row items-center justify-between full-width"
					style="height: 48px"
				>
					<div class="text-ink-2">
						<q-icon name="sym_r_upload" size="24px" />
						<span class="q-ml-md text-subtitle2">
							{{ t('files.File size limit') }}
						</span>
					</div>

					<bt-switch
						size="sm"
						truthy-track-color="light-blue-default"
						v-model="uploadLimiteOpen"
					/>
				</div>

				<template v-if="uploadLimiteOpen">
					<terminus-edit
						:inputHeight="38"
						v-model="uploadFileSizeLimit"
						:is-error="
							uploadFileSizeLimit.length > 0 &&
							filesSizeLimitRule(uploadFileSizeLimit).length > 0
						"
						:error-message="filesSizeLimitRule(uploadFileSizeLimit)"
					>
						<template v-slot:right>
							<div
								class="row items-center justify-end text-body2"
								style="margin-top: 7px"
								@click="editFileLimitUnit"
							>
								<div>
									{{
										diskUnitOptions().find((e) => e.value == uploadFileSizeUnit)
											?.label
									}}
								</div>
								<q-icon name="sym_r_expand_more" size="24px" color="ink-2" />
							</div>
						</template>
					</terminus-edit>
				</template>
			</template>
			<template v-else>
				<div class="share-link row items-center q-pa-md">
					<div class="share-info">
						<div class="text-ink-2 text-body2">{{ getShareLink }}</div>
					</div>
				</div>

				<div
					class="row items-center justify-between full-width q-mt-md"
					style="height: 48px"
					@click="copyLinkAndPassword"
				>
					<div class="text-ink-2">
						<q-icon name="sym_r_content_copy" size="24px" />
						<span class="q-ml-md text-subtitle2">
							{{ t('files.Copy link and password') }}
						</span>
					</div>
				</div>

				<div
					class="row items-center justify-between full-width"
					style="height: 48px"
					@click="removeShare"
				>
					<div class="text-negative">
						<q-icon name="sym_r_delete" size="24px" />
						<span class="q-ml-md text-subtitle2">
							{{ t('files.Delete link') }}
						</span>
					</div>
				</div>
			</template>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { ref, computed } from 'vue';
import { FilesIdType, useFilesStore } from '../../../../stores/files';

import { useDataStore } from '../../../../stores/data';

import TerminusEdit from 'src/components/common/TerminusEdit.vue';
import BtCheckBoxComponent from '../../../settings/base/BtCheckBoxComponent.vue';
import { ElDatePicker, ElConfigProvider } from 'element-plus';

import 'element-plus/dist/index.css';
import 'element-plus/theme-chalk/dark/css-vars.css';
import { formatFileModified } from '../../../../utils/file';

import GeneratePassword from '../GeneratePassword.vue';

import { usePublicShare, diskUnitOptions, DiskUnitMode } from './public';
import ShareMobileEditFileLimitDialog from './ShareMobileEditFileLimitDialog.vue';

import ShareBack from '../ShareBack.vue';
import { busEmit } from 'src/utils/bus';
import { useQuasar } from 'quasar';

const props = defineProps({
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	}
});

const { t } = useI18n();

const CustomRef = ref();

const $q = useQuasar();

const {
	copyLinkAndPassword,
	lang,
	shareResult,
	publicPassword,
	setExpirationInDays,
	dateValue,
	days,
	removeShare,
	disabledDate,
	getShareLink,
	uploadOnly,
	createPublicShare,
	onDisabled,
	datesLimitRule,
	onCancel,
	uploadLimiteOpen,
	uploadFileSizeLimit,
	filesSizeLimitRule,
	uploadFileSizeUnit
} = usePublicShare(props.origin_id);

const store = useDataStore();
const filesStore = useFilesStore();

const cancel = computed(() => {
	return false;
});

const ok = computed(() => {
	if (!!shareResult.value) {
		return false;
	}
	return t('confirm');
});

const onClose = () => {
	store.closeHovers();
	CustomRef.value.onDialogCancel();
};

const onBack = () => {
	const index = filesStore.selected[props.origin_id][0];
	store.closeHovers();
	busEmit('fileItemOpenOperation', index);
	CustomRef.value.onDialogCancel();
};

const onSubmit = async () => {
	if (!shareResult.value) {
		await createPublicShare();
	} else {
		store.closeHovers();
		CustomRef.value.onDialogOK();
	}
};

const editFileLimitUnit = () => {
	$q.dialog({
		component: ShareMobileEditFileLimitDialog,
		componentProps: {
			unit: uploadFileSizeUnit.value
		}
	}).onOk((value: DiskUnitMode) => {
		uploadFileSizeUnit.value = value;
	});
};
</script>

<style lang="scss" scoped>
.dialog-content-height {
	height: 260px;
}

.dialog-desc {
	width: 100%;
	padding: 0 0px;

	.share-link {
		width: 100%;
		min-height: 60px;
		border-radius: 8px;
		background: $background-6;

		.share-info {
			flex: 1;
		}

		.action-btn {
			height: 32px;
			width: 32px;
			border-radius: 4px;
		}
		.action-btn:hover {
			background-color: $background-3;
		}
	}
}
</style>

<style lang="scss">
.share-link-date-picker-popper {
	z-index: 99999 !important;
}
</style>

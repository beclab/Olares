<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('files_popup_menu.Share to Public')"
		:skip="false"
		:ok="ok"
		:cancel="cancel"
		:persistent="true"
		size="medium"
		@onCancel="onCancel"
		:okDisabled="onDisabled"
		@onSubmit="onSubmit"
	>
		<div class="dialog-desc">
			<template v-if="!shareResult">
				<div class="text-ink-2 text-subtitle1">
					{{ t('files.Add password') }}
				</div>
				<GeneratePassword v-model="publicPassword" />

				<div class="text-ink-2 text-subtitle1 q-mt-lg">
					{{ t('files.Set expiration') }}
				</div>
				<div class="row items-center q-mt-sm">
					<bt-check-box-component
						:model-value="setExpirationInDays == true"
						:label="t('files.In days')"
						@update:modelValue="setExpirationInDays = true"
					/>
					<bt-check-box-component
						:model-value="setExpirationInDays == false"
						:label="t('files.Exact date & time')"
						class="q-ml-lg"
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
				/>

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
					>
					</el-date-picker>
				</el-config-provider>
				<terminus-check-box
					class="q-mt-lg"
					v-model="uploadLimiteOpen"
					:label="t('files.File size limit')"
				/>
				<!-- <div class="row items-center justify-between full-width q-mt-lg">
					<div class="text-ink-2">
						{{ t('files.File size limit') }}
					</div>

					<bt-switch
						size="sm"
						truthy-track-color="light-blue-default"
						v-model="uploadLimiteOpen"
					/>
				</div> -->
				<template v-if="uploadLimiteOpen">
					<terminus-edit
						:inputHeight="38"
						v-model="uploadFileSizeLimit"
						class="q-mt-md"
						:is-error="
							uploadFileSizeLimit.length > 0 &&
							filesSizeLimitRule(uploadFileSizeLimit).length > 0
						"
						:error-message="filesSizeLimitRule(uploadFileSizeLimit)"
					>
						<template v-slot:right>
							<bt-select
								:offset="[50, 5]"
								style="margin-top: 7px"
								v-model="uploadFileSizeUnit"
								:options="diskUnitOptions()"
							/>
						</template>
					</terminus-edit>
				</template>
				<terminus-check-box
					class="q-mt-lg"
					v-model="uploadOnly"
					:label="t('files.Allow upload only')"
				/>
			</template>
			<template v-else>
				<div class="share-link row items-center q-pa-md">
					<div class="share-info">
						<div class="text-ink-2 text-body2">{{ getShareLink }}</div>
						<div class="text-ink-3 text-body3 q-mt-sm">
							{{
								t('expire_time') +
								': ' +
								formatFileModified(
									shareResult.expire_time || '',
									'YYYY-MM-DD HH:mm'
								)
							}}
						</div>
					</div>
					<div class="row items-center">
						<div
							class="action-btn row items-center justify-center text-ink-3"
							@click="removeShare"
						>
							<q-icon name="sym_r_delete" size="20px" />
						</div>
						<div
							class="action-btn row items-center justify-center text-ink-3"
							@click="copyLinkAndPassword"
						>
							<q-icon name="sym_r_content_copy" size="20px"></q-icon>
						</div>
					</div>
				</div>
			</template>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { ref, computed, onMounted } from 'vue';
import { FilesIdType, useFilesStore } from '../../../../stores/files';

import { useDataStore } from '../../../../stores/data';

import TerminusEdit from 'src/components/common/TerminusEdit.vue';
import BtCheckBoxComponent from '../../../settings/base/BtCheckBoxComponent.vue';
import { ElDatePicker, ElConfigProvider } from 'element-plus';
import TerminusCheckBox from '../../../common/TerminusCheckBox.vue';

import 'element-plus/dist/index.css';
import 'element-plus/theme-chalk/dark/css-vars.css';
import { formatFileModified } from '../../../../utils/file';

import GeneratePassword from '../GeneratePassword.vue';
import BtSelect from '../../../base/BtSelect.vue';

import { usePublicShare, diskUnitOptions } from './public';

const props = defineProps({
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	}
});

const { t } = useI18n();

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

const oldName = () => {
	const index = filesStore.selected[props.origin_id][0];
	return filesStore.getTargetFileItem(index, props.origin_id)?.name;
};

const cancel = computed(() => {
	if (!shareResult.value) {
		return t('cancel');
	}
	return false;
});

const ok = computed(() => {
	return t('confirm');
});

const CustomRef = ref();

const onSubmit = async () => {
	if (!shareResult.value) {
		await createPublicShare();
	} else {
		store.closeHovers();
		CustomRef.value.onDialogOK();
	}
};

onMounted(async () => {});
</script>

<style lang="scss" scoped>
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

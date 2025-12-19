<template>
	<page-title-component :show-back="true" :title="restoreTitle" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<bt-form>
			<bt-list first :label="t('basic_backup_info')">
				<bt-form-item
					v-if="restoreType === BackupLocationType.fileSystem"
					:title="t('backup_path')"
				>
					<transfet-select-to
						class="q-mt-xs"
						@setSelectPath="setBackupPath"
						:origins="backupOriginsRef"
						:master-node="true"
					>
						<template v-slot:default>
							<div class="row justify-end items-center">
								<div
									class="text-body1 text-ink-1"
									style="width: calc(100% - 40px)"
									v-if="backupUrl"
								>
									{{ backupUrl }}
								</div>
								<q-btn
									class="text-ink-2 btn-size-sm btn-no-text btn-no-border"
									icon="sym_r_edit_square"
									outline
									no-caps
								/>
							</div>
						</template>
					</transfet-select-to>
				</bt-form-item>

				<error-message-tip
					v-else
					:is-error="!urlRule"
					:error-message="t('errors.naming_is_not_compliant')"
					:with-popup="true"
				>
					<bt-form-item :title="t('backup_url')" :width-separator="false">
						<bt-edit-view
							style="width: 200px"
							v-model="backupUrl"
							:right="true"
							:placeholder="t('backup_url')"
						/>
					</bt-form-item>
				</error-message-tip>

				<error-message-tip
					:width-separator="
						restoreType !== BackupLocationType.space ||
						resourcesType === BackupResourcesType.files
					"
					:is-error="password.length < 4"
				>
					<bt-form-item :width-separator="false">
						<template v-slot:title>
							<div class="column">
								<div>{{ t('restore_password') }}</div>
								<div class="row" v-if="!deviceStore.isMobile">
									<div
										style="width: 16px; height: 16px"
										class="row items-center justify-center"
									>
										<q-icon
											:name="
												password.length >= 4 ? 'sym_r_check' : 'sym_r_clear'
											"
											class="text-ink-3"
											size="16px"
										/>
									</div>
									<div class="text-body3 text-ink-3">
										{{ t('must_have_at_least_4_characters') }}
									</div>
								</div>
							</div>
						</template>

						<bt-edit-view
							v-model="password"
							:is-password="true"
							style="width: 200px"
							:right="true"
							:placeholder="t('please_enter_a_password')"
						/>
					</bt-form-item>

					<template v-slot:reminder v-if="deviceStore.isMobile">
						<div
							class="row bg-background-3 items-center q-mb-sm q-px-xs"
							style="width: 100%; height: 24px; border-radius: 4px"
						>
							<div
								style="width: 16px; height: 16px"
								class="row items-center justify-center"
							>
								<q-icon
									:name="password.length >= 4 ? 'sym_r_check' : 'sym_r_clear'"
									class="text-ink-3"
									size="16px"
								/>
							</div>
							<div class="q-ml-sm text-overline-m text-ink-3">
								{{ t('must_have_at_least_4_characters') }}
							</div>
						</div>
					</template>
				</error-message-tip>

				<bt-form-item
					v-if="restoreType !== BackupLocationType.space"
					:title="t('select_a_snapshot')"
					:width-separator="resourcesType === BackupResourcesType.files"
				>
					<div>
						<div
							v-if="snapshotStatus === SnapshotStatus.INITIALIZED"
							class="text-body1 text-negative"
						>
							{{ t('no_available_snapshots') }}
						</div>
						<div
							v-if="snapshotStatus === SnapshotStatus.NODATA"
							class="text-orange-default text-body1 cursor-pointer"
							@click="showSnapshotsDialog"
						>
							query snapshots
						</div>
						<div
							v-if="snapshotStatus === SnapshotStatus.COMPLETED"
							class="text-info cursor-pointer single-line"
							style="max-width: 200px"
							@click="showSnapshotsDialog"
						>
							{{ selectSnapshot?.id }}
						</div>
					</div>
				</bt-form-item>

				<bt-form-item
					v-if="resourcesType === BackupResourcesType.files"
					:title="t('Restore location')"
				>
					<transfet-select-to
						class="q-mt-xs"
						@setSelectPath="setRestorePath"
						:origins="backupOriginsRef"
						:master-node="true"
					>
						<template v-slot:default>
							<div class="row justify-end items-center">
								<div
									class="text-body1 text-ink-1"
									style="width: calc(100% - 40px)"
									v-if="restorePathRef"
								>
									{{ restorePathRef }}
								</div>
								<q-btn
									class="text-ink-2 btn-size-sm btn-no-text btn-no-border"
									icon="sym_r_edit_square"
									outline
									no-caps
								/>
							</div>
						</template>
					</transfet-select-to>
				</bt-form-item>

				<error-message-tip
					v-if="resourcesType === BackupResourcesType.files"
					:is-error="!dirNameRule"
					:error-message="t('errors.naming_is_not_compliant')"
					:with-popup="true"
					:width-separator="false"
				>
					<bt-form-item :title="t('New folder name')" :width-separator="false">
						<bt-edit-view
							style="width: 200px"
							v-model="dirName"
							:right="true"
							:placeholder="t('New folder name')"
						/>
					</bt-form-item>
				</error-message-tip>
			</bt-list>
		</bt-form>
		<div class="row justify-end">
			<q-btn
				dense
				flat
				:disable="!canSubmit"
				class="confirm-btn q-mt-lg q-mt-lg"
				:label="t('start_restore')"
				@click="onSubmit"
				:loading="isLoading"
			/>
		</div>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import ErrorMessageTip from '../../../../components/settings/base/ErrorMessageTip.vue';
import PageTitleComponent from '../../../../components/settings/PageTitleComponent.vue';
import BtFormItem from '../../../../components/settings/base/BtFormItem.vue';
import TransfetSelectTo from '../../../Electron/Transfer/TransfetSelectTo.vue';
import BtEditView from '../../../../components/settings/base/BtEditView.vue';
import BtForm from '../../../../components/settings/base/BtForm.vue';
import ParseUrlSnapshotDialog from './ParseUrlSnapshotDialog.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import { useDeviceStore } from 'src/stores/settings/device';
import { useBackupStore } from 'src/stores/settings/backup';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { base64ToString } from '@didvault/sdk/src/core';
import { useRoute, useRouter } from 'vue-router';
import { FilePath } from 'src/stores/files';
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import {
	BackupLocationType,
	backupOriginsRef,
	BackupResourcesType,
	MENU_TYPE,
	RestoreSnapshotInfo
} from 'src/constant';

const $q = useQuasar();
const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const restorePathRef = ref();
const dirName = ref('');
const password = ref('');
const backupUrl = ref('');
const isLoading = ref(false);
const backupStore = useBackupStore();
const deviceStore = useDeviceStore();
const resourcesType = ref(BackupResourcesType.files);
const enum SnapshotStatus {
	INITIALIZED,
	COMPLETED,
	NODATA
}
const snapshotStatus = ref(SnapshotStatus.INITIALIZED);
const selectSnapshot = ref<RestoreSnapshotInfo | null>(null);
const restoreType = route.params.type as string;

const restoreTitle = computed(() => {
	switch (restoreType) {
		case BackupLocationType.fileSystem:
			return t('from_local_path');
		case BackupLocationType.space:
			return t('from_space_url');
		case BackupLocationType.awsS3:
			return t('from_aws_s3_url');
		case BackupLocationType.tencentCloud:
			return t('from_tencent_cos_url');
		default:
			return '';
	}
});

const setBackupPath = (fileSavePath: FilePath) => {
	backupUrl.value = fileSavePath.decodePath;
};

const setRestorePath = (fileSavePath: FilePath) => {
	restorePathRef.value = fileSavePath.decodePath;
};

const urlRule = computed(() => {
	return backupUrl.value.length > 0;
});

const dirNameRule = computed(() => {
	return dirName.value.length > 0;
});

const showSnapshotsDialog = () => {
	$q.dialog({
		component: ParseUrlSnapshotDialog,
		componentProps: {
			url: backupUrl.value,
			pwd: password.value
		}
	}).onOk((data) => {
		if (data) {
			snapshotStatus.value = SnapshotStatus.COMPLETED;
			selectSnapshot.value = data.data;
			resourcesType.value = data.type;
		}
	});
};

function getBackupTypeFromUrl(url: string): string | null {
	try {
		const decodeUrl = base64ToString(url);
		const parsedUrl = new URL(decodeUrl);
		return parsedUrl.searchParams.get('backupType');
	} catch (error) {
		console.error('URL parse error:', error);
		return null;
	}
}

watch(
	() => [backupUrl.value, password.value],
	() => {
		if (
			backupUrl.value.length > 0 &&
			restoreType === BackupLocationType.space
		) {
			const type = getBackupTypeFromUrl(backupUrl.value);
			if (type === '2') {
				resourcesType.value = BackupResourcesType.app;
			} else {
				resourcesType.value = BackupResourcesType.files;
			}
		}

		if (
			backupUrl.value.length > 0 &&
			password.value.length > 0 &&
			restoreType !== BackupLocationType.space
		) {
			resourcesType.value = BackupResourcesType.files;
			snapshotStatus.value = SnapshotStatus.NODATA;
			selectSnapshot.value = null;
		}
	}
);

const onSubmit = () => {
	isLoading.value = true;
	backupStore
		.restoreCustomUrl(
			backupUrl.value,
			password.value,
			restorePathRef.value,
			dirName.value,
			restoreType !== BackupLocationType.space,
			selectSnapshot.value
		)
		.then(() => {
			BtNotify.show({
				type: NotifyDefinedType.SUCCESS,
				message: t('success')
			});
			router.push({ name: MENU_TYPE.Restore });
		})
		.catch((e) => {
			console.error(e);
		})
		.finally(() => {
			isLoading.value = false;
		});
};

const canSubmit = computed(() => {
	if (resourcesType.value === BackupResourcesType.files) {
		switch (restoreType) {
			case BackupLocationType.fileSystem:
			case BackupLocationType.tencentCloud:
			case BackupLocationType.awsS3:
				return (
					!!backupUrl.value &&
					!!password.value &&
					!!restorePathRef.value &&
					!!selectSnapshot.value &&
					urlRule.value &&
					!!dirName.value &&
					dirNameRule.value
				);
			case BackupLocationType.space:
				return (
					!!backupUrl.value &&
					!!password.value &&
					!!restorePathRef.value &&
					urlRule.value &&
					!!dirName.value &&
					dirNameRule.value
				);
			default:
				return false;
		}
	} else if (resourcesType.value === BackupResourcesType.app) {
		switch (restoreType) {
			case BackupLocationType.fileSystem:
			case BackupLocationType.tencentCloud:
			case BackupLocationType.awsS3:
				return (
					!!backupUrl.value &&
					!!password.value &&
					!!selectSnapshot.value &&
					urlRule.value
				);
			case BackupLocationType.space:
				return !!backupUrl.value && !!password.value && urlRule.value;
			default:
				return false;
		}
	}
	return false;
});
</script>

<style scoped lang="scss"></style>

<template>
	<page-title-component
		:custom-back="true"
		:show-back="true"
		:title="t('add_backup')"
		@on-back-click="onBackup"
	/>
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<bt-form v-model:can-submit="formAble">
			<bt-list first :label="t('basic_backup_info')">
				<bt-form-item
					v-if="backupType === BackupResourcesType.app"
					:title="t('application')"
					:margin-top="false"
				>
					<bt-app-select v-model="backupApp" :options="backupAppOptions" />
				</bt-form-item>
				<bt-form-item :title="t('backup_location')" :margin-top="false">
					<div class="row justify-end items-center full-width">
						<div
							v-if="
								locationData &&
								locationData.type === BackupLocationType.fileSystem
							"
							style="width: calc(100% - 40px)"
							class="text-body1 text-ink-1"
						>
							{{ locationData.data.decodePath }}
						</div>
						<div
							class="row justify-end items-center"
							style="max-width: calc(100% - 40px)"
							v-if="
								locationData && locationData.type === BackupLocationType.space
							"
						>
							<q-img
								class="backup-location-img"
								:src="getBackupIconByLocation(locationData.type)"
							/>
							<span
								class="text-body1 text-ink-1 q-ml-xs single-line"
								style="max-width: calc(100% - 30px)"
								>{{ locationData.data.name }}</span
							>
						</div>
						<div
							class="row justify-end items-center"
							v-if="
								locationData &&
								locationData.type === BackupLocationType.tencentCloud
							"
							style="max-width: calc(100% - 40px)"
						>
							<q-img
								class="backup-location-img"
								:src="getBackupIconByLocation(locationData.type)"
							/>
							<span
								class="text-body1 text-ink-1 q-ml-xs single-line"
								style="max-width: calc(100% - 30px)"
								>{{ locationData.data.name }}</span
							>
						</div>

						<div
							class="row justify-end items-center"
							style="max-width: calc(100% - 40px)"
							v-if="
								locationData && locationData.type === BackupLocationType.awsS3
							"
						>
							<q-img
								class="backup-location-img"
								:src="getBackupIconByLocation(locationData.type)"
							/>
							<span
								class="text-body1 text-ink-1 q-ml-xs single-line"
								style="max-width: calc(100% - 30px)"
								>{{ locationData.data.name }}</span
							>
						</div>
						<q-btn
							class="text-ink-2 btn-size-sm btn-no-text btn-no-border"
							icon="sym_r_edit_square"
							outline
							no-caps
							@click="onLocationClick"
						/>
					</div>

					<template v-slot:bottom>
						<div
							v-if="
								locationData &&
								(locationData.type === BackupLocationType.awsS3 ||
									locationData.type === BackupLocationType.tencentCloud)
							"
							class="location-raw-data column justify-start items-center q-mx-lg q-mb-lg"
						>
							<div
								v-if="locationData.data.raw_data"
								class="row justify-between items-center full-width"
							>
								<span class="text-ink-3 text-body3" style="max-width: 30%">{{
									t('bucket_name')
								}}</span>
								<span
									v-if="locationData.data.raw_data.bucket"
									class="text-ink-1 text-body3 text-right"
									style="max-width: 70%"
									>{{ locationData.data.raw_data.bucket }}</span
								>
							</div>

							<div
								v-if="locationData.data.raw_data"
								class="row justify-between items-center full-width"
							>
								<span class="text-ink-3 text-body3" style="max-width: 30%">{{
									t('sever_endpoint')
								}}</span>
								<span
									v-if="locationData.data.raw_data.endpoint"
									class="text-ink-1 text-body3 text-right"
									style="max-width: 70%"
									>{{ locationData.data.raw_data.endpoint }}</span
								>
							</div>
						</div>
					</template>
				</bt-form-item>

				<bt-form-item
					v-if="locationData && locationData.type === BackupLocationType.space"
					:title="t('backup_region')"
				>
					<bt-select v-model="spaceRegion" :options="spaceRegionOptions" />
				</bt-form-item>

				<bt-form-item
					v-if="backupType === BackupResourcesType.files"
					:title="t('backup_path')"
				>
					<transfet-select-to
						class="q-mt-xs"
						@setSelectPath="setSelectPath"
						:origins="BackupPathOrigins"
						:master-node="true"
					>
						<template v-slot:default>
							<div class="row justify-end items-center">
								<div
									class="text-body1 text-ink-1"
									style="width: calc(100% - 40px)"
									v-if="backupPathRef"
								>
									{{ backupPathRef }}
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
					:is-error="!nameRule"
					:error-message="t('errors.naming_is_not_compliant')"
					:with-popup="true"
					:width-separator="false"
				>
					<bt-form-item :title="t('backup_name')" :width-separator="false">
						<bt-edit-view
							style="width: 200px"
							v-model="name"
							:right="true"
							:placeholder="t('please_input_the_backup_name')"
						/>
					</bt-form-item>
				</error-message-tip>
			</bt-list>

			<bt-list :label="t('schedule_and_security')">
				<bt-form-item :title="t('snapshot_frequency')">
					<bt-select v-model="frequency" :options="frequencyOptions" />
				</bt-form-item>

				<bt-form-item :title="t('run_backup_at')">
					<div class="row items-center justify-end">
						<bt-select
							v-model="monthDay"
							:options="monthOption"
							v-if="frequency == BackupFrequency.Monthly"
						/>

						<bt-select
							v-model="weekDay"
							:options="weekOption"
							v-if="frequency == BackupFrequency.Weekly"
						/>

						<div class="text-body2 text-ink-1 q-ml-md">
							{{ time }}
						</div>

						<q-icon
							size="20px"
							name="sym_r_access_time"
							color="ink-1"
							class="time-clock"
							@click="onTimeDialog"
						/>
					</div>
				</bt-form-item>
				<error-message-tip :is-error="password.length < 4">
					<bt-form-item :width-separator="false">
						<template v-slot:title>
							<div class="column">
								<div>{{ t('backup_password') }}</div>
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
				<error-message-tip
					:is-error="password2.length === 0 || password !== password2"
					:error-message="t('errors.passwords_are_inconsistent')"
					:width-separator="false"
					:with-popup="true"
				>
					<bt-form-item :title="t('confirm_password')" :width-separator="false">
						<bt-edit-view
							v-model="password2"
							:is-password="true"
							style="width: 200px"
							:right="true"
							:placeholder="t('please_confirm_a_password')"
						/>
					</bt-form-item>
				</error-message-tip>
			</bt-list>
		</bt-form>
		<div class="row justify-end q-mb-md">
			<q-btn
				dense
				flat
				:disable="!canSubmit"
				class="confirm-btn q-px-md q-mt-lg"
				:label="t('submit')"
				@click="checkPath"
				:loading="isLoading"
			/>
		</div>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import TransfetSelectTo from '../../../Electron/Transfer/TransfetSelectTo.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import ErrorMessageTip from 'src/components/settings/base/ErrorMessageTip.vue';
import BtAppSelect from 'src/components/settings/base/BtAppSelect.vue';
import BtEditView from 'src/components/settings/base/BtEditView.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import BaseTimeDialog from 'src/components/base/BaseTimeDialog.vue';
import BtSelect from 'src/components/settings/base/BtSelect.vue';
import BackupLocationDialog from './BackupLocationDialog.vue';
import BtForm from 'src/components/settings/base/BtForm.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import { getOlaresSpaceRegion } from 'src/api/settings/backup';
import { notifyFailed } from 'src/utils/settings/btNotify';
import { useBackupStore } from 'src/stores/settings/backup';
import { useDeviceStore } from 'src/stores/settings/device';
import { computed, onMounted, ref } from 'vue';
import { BtDialog, useColor } from '@bytetrade/ui';
import { BackupFrequency } from '@bytetrade/core';
import { useRoute, useRouter } from 'vue-router';
import { FilePath } from 'src/stores/files';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import {
	ApplicationSelectorState,
	getBackupIconByLocation,
	BackupResourcesType,
	BackupLocationType,
	OlaresSpaceRegion,
	BackupPathOrigins,
	frequencyOptions,
	SelectorProps,
	monthOption,
	weekOption,
	MENU_TYPE
} from 'src/constant';

const $q = useQuasar();
const { t } = useI18n();
const backupPathRef = ref();
const backupType = ref<BackupResourcesType>(BackupResourcesType.files);
const backupStore = useBackupStore();
const deviceStore = useDeviceStore();
const isLoading = ref(false);
const router = useRouter();
const route = useRoute();
const backupApp = ref();
const backupAppOptions = ref<ApplicationSelectorState[]>();
const nameRule = computed(() => {
	return name.value.length > 0;
});
const frequency = ref(BackupFrequency.Daily);
const weekDay = ref(1);
const monthDay = ref(1);

const time = ref('03:00');
const password = ref('');
const password2 = ref('');

const formAble = ref(false);
const locationData = ref(null);
const spaceRegion = ref<string | null>(null);
const spaceRegionOptions = ref<SelectorProps[]>([]);

const setSelectPath = (fileSavePath: FilePath) => {
	backupPathRef.value = fileSavePath.decodePath;
};

onMounted(() => {
	const path = route.params.backup_path as string;
	backupType.value = route.params.backup_type as string;
	if (path) {
		backupPathRef.value = decodeURIComponent(path);
	}
	if (backupType.value === BackupResourcesType.app) {
		backupAppOptions.value = backupStore.getSupportApplicationOptions();
		if (backupAppOptions.value.length > 0) {
			backupApp.value = backupAppOptions.value[0];
		}
	}
});

const onBackup = () => {
	router.replace({ name: MENU_TYPE.Backup });
};

const name = ref('');

const canSubmit = computed(() => {
	if (!formAble.value) {
		return false;
	}

	if (password.value != password2.value) {
		return false;
	}

	if (!locationData.value || !locationData.value.data) {
		return false;
	}

	if (locationData.value.type === BackupLocationType.space) {
		if (!spaceRegion.value) {
			return false;
		}

		if (
			!spaceRegionOptions.value.find((item) => item.value === spaceRegion.value)
		)
			return false;
	}

	if (backupType.value === BackupResourcesType.files && !backupPathRef.value) {
		return false;
	}

	if (backupType.value === BackupResourcesType.app && !backupApp.value) {
		return false;
	}

	return true;
});

const onTimeDialog = () => {
	$q.dialog({
		component: BaseTimeDialog,
		componentProps: {
			time: time.value
		}
	}).onOk((data) => {
		time.value = data;
	});
};

const onLocationClick = () => {
	$q.dialog({
		component: BackupLocationDialog
	}).onOk((data) => {
		console.log(data);
		locationData.value = data;
		if (locationData.value.type === BackupLocationType.space) {
			spaceRegion.value = null;
			spaceRegionOptions.value = [];
			getOlaresSpaceRegion()
				.then((res) => {
					res.forEach((region) => {
						spaceRegionOptions.value.push({
							label: region.regionId,
							value: region.cloudName + '_' + region.regionId
						});
					});
					if (spaceRegionOptions.value.length > 0) {
						spaceRegion.value = spaceRegionOptions.value[0].value;
					}
					console.log(spaceRegionOptions.value);
				})
				.catch((e) => {
					console.error(e);
				});
		}
	});
};

async function checkPath() {
	if (backupType.value === BackupResourcesType.files) {
		if (
			backupPathRef.value &&
			!!backupStore.backupList.find((item) => item.path === backupPathRef.value)
		) {
			const { color: blue } = useColor('blue-default');
			const { color: textInk } = useColor('ink-on-brand');

			BtDialog.show({
				title: t('backup_creation_reminder'),
				message: t('already_backup_path'),
				okStyle: {
					background: blue.value,
					color: textInk.value
				},
				okText: t('base.confirm'),
				cancelText: t('base.cancel'),
				cancel: true
			})
				.then((res) => {
					if (res) {
						onSubmit();
					} else {
						console.log('click cancel');
					}
				})
				.catch((err) => {
					console.log('click error', err);
				});
			return;
		}
	} else if (backupType.value === BackupResourcesType.app) {
		const findSameBackupLocation = backupStore.backupList
			.filter((item) => item.backupType === BackupResourcesType.app)
			.find(
				(item) =>
					item.location === locationData.value.type &&
					backupApp.value.value === item.backupAppTypeName
			);

		if (
			backupType.value === BackupResourcesType.app &&
			!!findSameBackupLocation
		) {
			const { color: blue } = useColor('blue-default');
			const { color: textInk } = useColor('ink-on-brand');

			BtDialog.show({
				title: t('backup_creation_reminder'),
				message: t('already_backup_origin'),
				okStyle: {
					background: blue.value,
					color: textInk.value
				},
				okText: t('base.confirm'),
				cancel: true
			}).catch((err) => {
				console.log('click error', err);
			});
			return;
		}
	}

	onSubmit();
}

async function onSubmit() {
	let region: OlaresSpaceRegion;
	if (locationData.value.type === BackupLocationType.space) {
		if (!spaceRegion.value) {
			notifyFailed('Please select a space region');
			return;
		}
		const regionArray = spaceRegion.value.split('_');
		region = {
			cloudName: regionArray[0],
			regionId: regionArray[1]
		};
	}

	isLoading.value = true;
	backupStore
		.createBackupPlan(
			backupType.value,
			name.value,
			password.value,
			locationData.value,
			{
				snapshotFrequency: frequency.value,
				timesOfDay: time.value,
				dayOfWeek: weekDay.value,
				dateOfMonth: monthDay.value
			},
			backupApp.value,
			backupPathRef.value,
			region
		)
		.then(() => {
			onBackup();
		})
		.catch((e) => {
			console.error(e);
		})
		.finally(() => {
			isLoading.value = false;
		});
}
</script>

<style lang="scss" scoped>
.backup-location-img {
	width: 24px;
	height: 24px;
}

.location-raw-data {
	display: flex;
	padding: 12px;
	flex-direction: column;
	align-items: flex-start;
	gap: 8px;
	align-self: stretch;
	border-radius: 8px;
	background: $background-6;
}
</style>

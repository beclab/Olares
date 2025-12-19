<template>
	<page-title-component :show-back="true" :title="t('Manage backup task')" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<bt-form v-model:can-submit="canSubmit">
			<bt-list first>
				<bt-form-item
					v-if="backup && backup.backupType === BackupResourcesType.app"
					:title="t('Backup App')"
				>
					<div class="text-body2 text-ink-2">
						{{ backup.backupAppTypeName }}
					</div>
				</bt-form-item>

				<bt-form-item
					v-if="backup && backup.backupType === BackupResourcesType.files"
					:title="t('backup_path')"
				>
					<div class="text-body2 text-ink-2">{{ backup.path }}</div>
				</bt-form-item>

				<bt-form-item v-if="backup" :title="t('backup_name')">
					<div class="text-body2 text-ink-2">{{ backup.name }}</div>
				</bt-form-item>

				<bt-form-item :title="t('snapshot_frequency')">
					<div v-if="frequency" class="text-body2 text-ink-2 q-ml-md">
						{{ frequency?.label }}
					</div>
				</bt-form-item>

				<bt-form-item :title="t('run_backup_at')" :width-separator="false">
					<div class="row justify-end items-center">
						<div
							v-if="frequency && frequency?.value === BackupFrequency.Weekly"
							class="text-body2 text-ink-2 q-ml-md"
						>
							{{ weekDay }}
						</div>
						<div
							v-if="frequency && frequency?.value === BackupFrequency.Monthly"
							class="text-body2 text-ink-2 q-ml-md"
						>
							{{ monthDay }}
						</div>
						<div class="text-body2 text-ink-2 q-ml-md">
							{{ time }}
						</div>
					</div>
				</bt-form-item>
			</bt-list>
		</bt-form>
		<div class="row justify-end">
			<q-btn
				dense
				flat
				class="cancel-btn q-px-md q-mt-lg q-mr-md"
				:label="t('manager')"
			>
				<bt-popup style="width: 176px">
					<bt-popup-item
						v-close-popup
						:title="t('edit')"
						active-soft="background-hover"
						active-text="text-ink-2"
						icon="sym_r_edit_square"
						@on-item-click="onUpdate"
					/>
					<bt-popup-item
						v-if="backup.backupPolicies.enabled"
						v-close-popup
						active-soft="background-hover"
						active-text="text-ink-2"
						:title="t('pause')"
						icon="sym_r_pause_circle"
						@on-item-click="onPause"
					/>
					<bt-popup-item
						v-else
						active-soft="background-hover"
						active-text="text-ink-2"
						v-close-popup
						:title="t('resume')"
						icon="sym_r_play_circle"
						@on-item-click="onResume"
					/>
					<bt-popup-item
						active-soft="background-hover"
						active-text="text-ink-2"
						v-close-popup
						:title="t('delete')"
						icon="sym_r_delete"
						@on-item-click="onDelete"
					/>
				</bt-popup>
			</q-btn>

			<!--			<q-btn-->
			<!--				dense-->
			<!--				flat-->
			<!--				class="cancel-btn q-px-md q-mt-lg q-mr-md"-->
			<!--				:label="t('restore')"-->
			<!--				@click="onRestore"-->
			<!--			/>-->

			<q-btn
				dense
				flat
				class="confirm-btn q-px-md q-mt-lg"
				:label="t('snapshot_now')"
				@click="onSubmit"
			/>
		</div>

		<bt-list class="bg-background-1" :label="t('size')">
			<bt-form-item
				:title="t('Source Size')"
				:data="restoreSize"
				:width-separator="true"
				:description="
					t('The total space your files currently occupy on your device.')
				"
			/>

			<bt-form-item
				:title="t('backup_size')"
				:data="backupSize"
				:width-separator="false"
				:description="
					t(
						'The actual storage space used by your backup file, which can differ due to compression or version history.'
					)
				"
			/>
		</bt-list>

		<bt-list v-if="snapshots.length > 0" :label="t('snapshots')">
			<q-table
				tableHeaderStyle="height: 32px;"
				table-header-class="text-body3 text-ink-2"
				flat
				wrap-cells
				class="q-px-lg q-pt-md"
				:bordered="false"
				:rows="snapshots"
				:columns="columns"
				row-key="id"
				v-model:pagination="pagination"
				@request="onRequest"
			>
				<template v-slot:body-cell-action="props">
					<q-td :props="props">
						<div class="row items-center justify-end">
							<q-btn
								class="q-mr-xs btn-size-sm btn-no-text text-grey-8"
								icon="sym_r_chevron_right"
								color="ink-2"
								outline
								@click.stop="gotoSnapshot(props.row.id)"
								no-caps
							/>
						</div>
					</q-td>
				</template>

				<template v-slot:body-cell-status="props">
					<q-td :props="props">
						<div class="row items-center">
							<q-img
								class="backup-status-img q-mr-sm"
								:src="getBackupStatusImg(props.row.status)"
							/>
							<div>
								{{
									props.row.status === BackupStatus.running &&
									props.row.progress
										? props.row.status +
										  `(${(props.row.progress / 100).toFixed(0)} %)`
										: props.row.status
								}}
							</div>
						</div>
					</q-td>
				</template>
			</q-table>
		</bt-list>
		<div v-else class="empty-parent column items-center">
			<q-img src="settings/default_empty.svg" class="empty-image" />
			<div class="empty-text">{{ t('no_snapshots') }}</div>
		</div>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import { BtDialog, BtNotify, NotifyDefinedType, useColor } from '@bytetrade/ui';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import { binaryInsert, CompareBackup } from 'src/utils/rss-utils';
import { getBackupSnapshot } from 'src/api/settings/snapshot';
import BtList from 'src/components/settings/base/BtList.vue';
import BtPopupItem from '../../../../components/base/BtPopupItem.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import SnapshotFrequencyDialog from './SnapshotFrequencyDialog.vue';
import { ref, onMounted, computed, onBeforeUnmount } from 'vue';
import BtPopup from '../../../../components/base/BtPopup.vue';
import { useBackupStore } from 'src/stores/settings/backup';
import BtForm from 'src/components/settings/base/BtForm.vue';
import { timestampToTime } from './FormatBackupTime';
import { BackupFrequency } from '@bytetrade/core';
import { date, format, useQuasar } from 'quasar';
import { bus } from 'src/utils/bus';
import { useRouter } from 'vue-router';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import {
	BackupPlanDetail,
	frequencyOptions,
	getBackupStatusImg,
	BackupMessage,
	weekOption,
	BackupStatus,
	BackupSnapshot,
	BackupResourcesType
} from 'src/constant';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const backupStore = useBackupStore();
const { humanStorageSize } = format;

const time = ref();
const $q = useQuasar();
const frequency = ref();
const weekDay = ref('');
const monthDay = ref('');
const canSubmit = ref(false);
const backupId: string = route.params.backupId as string;
const backup = ref<BackupPlanDetail | null>(null);
const snapshots = ref<BackupSnapshot[]>([]);
const backupSize = computed(() => {
	if (backup.value) {
		try {
			return humanStorageSize(Number(backup.value.size));
		} catch (e) {
			return '0';
		}
	}
	return '0';
});

const restoreSize = computed(() => {
	if (backup.value) {
		try {
			return humanStorageSize(Number(backup.value.restoreSize));
		} catch (e) {
			return '0';
		}
	}
	return '0';
});

function updateBackupDetail(data: BackupMessage) {
	console.log(data);
	if (data && data.backupId === backupId) {
		const snapShot: BackupSnapshot = snapshots.value.find(
			(item: BackupSnapshot) => item.id === data.id
		);
		if (snapShot) {
			snapShot.status = data.status;
			if (data.progress) {
				snapShot.progress = data.progress;
			}
			if (data.size) {
				snapShot.size = data.size;
			}
			if (data.totalSize && backup.value) {
				backup.value.size = data.totalSize;
			}
			if (data.restoreSize && backup.value) {
				backup.value.restoreSize = data.restoreSize;
			}
		} else {
			getBackupSnapshot(backupId, data.id)
				.then((newSnapshot: BackupSnapshot) => {
					if (newSnapshot && newSnapshot.id) {
						const list = snapshots.value;
						snapshots.value = binaryInsert(list, newSnapshot, CompareBackup);
					}
				})
				.catch((e) => {
					console.log(e);
				});
		}
	}
}

onMounted(async () => {
	bus.on('backup_state_event', updateBackupDetail);
	await configDefaultValue();
	getSnapshots();
});

onBeforeUnmount(() => {
	bus.off('backup_state_event', updateBackupDetail);
});

async function onSubmit() {
	if (backup.value) {
		const { color: blue } = useColor('blue-default');
		const { color: textInk } = useColor('ink-on-brand');

		BtDialog.show({
			title: t('add_backup_snapshot'),
			message: t('add_backup_snapshot_prompt'),
			okStyle: {
				background: blue.value,
				color: textInk.value
			},
			okText: t('add_now'),
			cancelText: t('base.cancel'),
			cancel: true
		})
			.then((res) => {
				if (res) {
					backupStore.createBackupSnapShot(backupId).catch((e) => {
						console.log(e);
					});
				} else {
					console.log('click cancel');
				}
			})
			.catch((err) => {
				console.log('click error', err);
			});
	}
}

async function configDefaultValue() {
	console.log(backupId);
	backup.value = await backupStore.getBackupDetails(backupId);
	console.log(backup.value);
	if (backup.value) {
		if (backup.value.backupPolicies) {
			const frequencyOption = frequencyOptions.value.find(
				(item) => item.value === backup.value.backupPolicies.snapshotFrequency
			);
			if (frequencyOption) {
				frequency.value = frequencyOption;
			}
			const realTime = timestampToTime(
				Number(backup.value.backupPolicies.timespanOfDay)
			);
			console.log(realTime);
			time.value = realTime;
		}

		if (
			backup.value.backupPolicies &&
			backup.value.backupPolicies.snapshotFrequency === BackupFrequency.Weekly
		) {
			const day1 = backup.value.backupPolicies.dayOfWeek;
			if (day1 > 0) {
				const options = weekOption.value.find((item) => item.value === day1);
				if (options) {
					weekDay.value = options.label;
				}
			}
		}
		if (
			backup.value.backupPolicies &&
			backup.value.backupPolicies.snapshotFrequency === BackupFrequency.Monthly
		) {
			const day2 = backup.value.backupPolicies.dateOfMonth;
			if (day2 > 0) {
				monthDay.value = t('monthly_day', { day: day2 });
			}
		}
	}
}

// async function onRestore() {
// 	if (backup.value) {
// 		router.push('/backup/restore_existing_backup/' + backupId);
// 	}
// }

async function onDelete() {
	if (backup.value) {
		const { color: blue } = useColor('blue-default');
		const { color: textInk } = useColor('ink-on-brand');
		BtDialog.show({
			title: t('delete_backup'),
			message: t('confirm_delete_backup'),
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
					backupStore
						.deleteBackupPlan(backupId)
						.then(() => {
							BtNotify.show({
								type: NotifyDefinedType.SUCCESS,
								message: t('success')
							});
							router.back();
						})
						.catch((e) => {
							console.error(e);
						});
				} else {
					console.log('click cancel');
				}
			})
			.catch((err) => {
				console.log('click error', err);
			});
	}
}

async function onUpdate() {
	if (backup.value) {
		$q.dialog({
			component: SnapshotFrequencyDialog,
			componentProps: {
				backupId,
				policy: backup.value.backupPolicies
			}
		}).onOk(async () => {
			await configDefaultValue();
		});
	}
}

async function onPause() {
	if (backup.value) {
		await backupStore
			.pauseBackup(backup.value.id)
			.then((res) => {
				backup.value = res;
				BtNotify.show({
					type: NotifyDefinedType.SUCCESS,
					message: t('successful')
				});
			})
			.catch((e) => {
				console.error(e);
			});
	}
}

async function onResume() {
	if (backup.value) {
		await backupStore
			.resumeBackup(backup.value.id)
			.then((res) => {
				backup.value = res;
				BtNotify.show({
					type: NotifyDefinedType.SUCCESS,
					message: t('successful')
				});
			})
			.catch((e) => {
				console.error(e);
			});
	}
}

async function gotoSnapshot(snapshotId: string) {
	if (backup.value) {
		router.push('/backup/' + backupId + '/' + snapshotId);
	}
}

const columns = [
	{
		name: 'createTime',
		align: 'left',
		label: t('create_time'),
		field: 'createAt',
		format: (val: any) => {
			return date.formatDate(val * 1000, 'YYYY-MM-DD HH:mm');
		},
		sortable: false
	},
	{
		name: 'size',
		align: 'left',
		label: t('size'),
		field: 'size',
		format: (val: any) => {
			try {
				return humanStorageSize(Number(val));
			} catch (e) {
				return 0;
			}
		},
		sortable: false
	},
	{
		name: 'status',
		align: 'left',
		label: t('status'),
		field: 'status',
		sortable: false
	},
	{
		name: 'action',
		align: 'right',
		label: t('action'),
		field: 'Action',
		sortable: false
	}
];

const pagination = ref({
	page: 1,
	rowsPerPage: 6,
	rowsNumber: 0
});

const onRequest = (props: {
	pagination: {
		sortBy: string;
		page: number;
		rowsPerPage: number;
	};
	filter?: any;
	getCellValue: (col: any, row: any) => any;
}) => {
	if (!backupId) return;
	const { page, rowsPerPage } = props.pagination;
	const params = {
		offset: rowsPerPage * (page - 1),
		limit: rowsPerPage,
		backupId
	};
	pagination.value.page = page;
	pagination.value.rowsPerPage = rowsPerPage;
	getSnapshots(params);
};

const getSnapshots = (
	params: any = {
		backupId,
		offset: pagination.value.rowsPerPage * (pagination.value.page - 1),
		limit: pagination.value.rowsPerPage
	}
) => {
	return backupStore
		.getSnapshots(params.backupId, params.offset, params.limit)
		.then((response: any) => {
			console.log(response);
			snapshots.value = response.snapshots;
			pagination.value.rowsNumber = response.totalCount;
		});
};
</script>

<style lang="scss" scoped>
.resource-title {
	color: $ink-1;
	margin-top: 12px;
	margin-bottom: 8px;
}

.backup-arrow-bg {
	width: 24px;
	height: 24px;
	border-radius: 8px;
	background: $background-3;
	text-align: right;
}

.empty-parent {
	width: 100%;
	height: 400px;

	.empty-image {
		margin-top: 40px;
		width: 160px;
		height: 160px;
	}
}
</style>

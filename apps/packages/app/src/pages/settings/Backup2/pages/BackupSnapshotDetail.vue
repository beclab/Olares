<template>
	<page-title-component :show-back="true" :title="t('snapshot_details')" />

	<bt-scroll-area class="nav-height-scroll-area-conf">
		<bt-list first class="full-width">
			<bt-form-item :title="t('size')" :data="snapshotSize" />
			<bt-form-item :title="t('backup_type')" :data="backupType" />
			<bt-form-item
				v-if="snap?.status !== BackupStatus.running"
				:title="t('status')"
				:width-separator="
					(snap?.status === BackupStatus.failed ||
						snap?.status === BackupStatus.rejected) &&
					!!snap?.message
				"
			>
				<div class="row justify-end items-center">
					<q-img
						class="backup-status-img q-mr-sm"
						:src="getBackupStatusImg(snap?.status)"
					/>
					<div>
						{{ snap?.status }}
					</div>
				</div>
			</bt-form-item>
			<bt-form-item
				v-if="snap?.status === BackupStatus.running"
				:width-separator="false"
			>
				<template v-slot:all>
					<div class="column justify-start full-width q-px-lg">
						<q-linear-progress
							class="full-width"
							:value="Number(snap?.progress / 10000)"
							size="4px"
							color="info"
						/>
						<div class="text-info text-body2 q-mt-sm">
							{{ t('backup_running_message') }}
						</div>
					</div>
				</template>
			</bt-form-item>
			<bt-form-item
				v-if="
					(snap?.status === BackupStatus.failed ||
						snap?.status === BackupStatus.rejected) &&
					!!snap?.message
				"
				:title="t('message')"
				:width-separator="false"
			>
				<div
					class="text-body1 text-negative failed-message-width cursor-pointer"
					@click="onCopy"
				>
					{{ t(snap?.message) }}
				</div>
			</bt-form-item>
		</bt-list>

		<div class="row full-width justify-end">
			<q-btn
				v-if="
					snap?.status === BackupStatus.pending ||
					snap?.status === BackupStatus.running
				"
				dense
				flat
				class="cancel-btn q-px-md q-mt-lg"
				:label="t('cancel')"
				:loading="isCanceling"
				@click="onCancel"
			/>

			<!--			<q-btn-->
			<!--				v-if="snap?.status === BackupStatus.completed"-->
			<!--				dense-->
			<!--				flat-->
			<!--				class="cancel-btn q-px-md q-mt-lg"-->
			<!--				:label="t('restore')"-->
			<!--				@click="onRestore"-->
			<!--			/>-->
		</div>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { bus } from 'src/utils/bus';
import { useRoute } from 'vue-router';
import { format } from 'quasar';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { useBackupStore } from 'src/stores/settings/backup';
import BtList from 'src/components/settings/base/BtList.vue';
import { ref, onMounted, computed, onBeforeUnmount } from 'vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import {
	BackupSnapshotDetail,
	getBackupStatusImg,
	BackupStatus,
	SnapshotType,
	BackupMessage
} from 'src/constant';
import { getApplication } from 'src/application/base';

const { t } = useI18n();
const route = useRoute();
// const router = useRouter();
const { humanStorageSize } = format;
const isCanceling = ref(false);
const backupStore = useBackupStore();
const snap = ref<BackupSnapshotDetail | null>(null);
const backupId = route.params.backupId as string;
const snapshotId = route.params.snapshotId as string;

const snapshotSize = computed(() => {
	if (snap.value) {
		try {
			return humanStorageSize(Number(snap.value.size));
		} catch (e) {
			console.log(e.message);
		}
	}
	return '0';
});

const backupType = computed(() => {
	if (snap.value) {
		switch (snap.value.snapshotType) {
			case SnapshotType.Incremental:
				return t('incremental');
			case SnapshotType.Fully:
				return t('fully');
			case SnapshotType.Unknown:
				return t('unknown');
			default:
				return t('unknown');
		}
	}
	return t('unknown');
});

async function getDetails() {
	return backupStore.getSnapShotDetail(backupId, snapshotId).then((res) => {
		snap.value = res;
	});
}

function updateSnapShotDetail(data: BackupMessage) {
	if (
		data &&
		data.backupId === backupId &&
		data.id === snapshotId &&
		snap.value
	) {
		snap.value.progress = data.progress;
		snap.value.status = data.status;
		snap.value.message = data.message;
	}
}

onMounted(() => {
	bus.on('backup_state_event', updateSnapShotDetail);
	getDetails().catch((e) => {
		console.error(e);
	});
});

onBeforeUnmount(() => {
	bus.off('backup_state_event', updateSnapShotDetail);
});

const onCancel = () => {
	isCanceling.value = true;
	backupStore
		.cancelBackupSnapShot(backupId, snapshotId)
		.then(() => {
			BtNotify.show({
				type: NotifyDefinedType.SUCCESS,
				message: t('success')
			});
		})
		.catch((e) => {
			console.error(e);
		})
		.finally(() => {
			isCanceling.value = false;
			getDetails().catch((e) => {
				console.error(e);
			});
		});
};

const onCopy = () => {
	if (snap.value?.message) {
		getApplication()
			.copyToClipboard(snap.value?.message)
			.then(() => {
				BtNotify.show({
					type: NotifyDefinedType.SUCCESS,
					message: t('copy_success')
				});
			})
			.catch((e) => {
				BtNotify.show({
					type: NotifyDefinedType.FAILED,
					message: t('copy_failure_message', e.message)
				});
			});
	}
};

// const onRestore = () => {
// 	router.push('/backup/restore_existing_backup/' + backupId + '/' + snapshotId);
// };
</script>
<style lang="scss" scoped>
.failed-message-width {
	max-width: 100%;
	padding-top: 12px;
	padding-bottom: 12px;
	text-align: right;
	//overflow-wrap: break-word;
	//word-wrap: break-word;
	word-break: break-all;
	white-space: normal;
}
</style>

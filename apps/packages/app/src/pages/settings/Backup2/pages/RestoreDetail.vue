<template>
	<page-title-component :show-back="true" :title="t('restore_details')" />

	<bt-scroll-area class="nav-height-scroll-area-conf">
		<bt-list first>
			<bt-form-item :title="t('backup_name')" :data="restore?.name" />

			<bt-form-item
				v-if="restore && restore?.backupType === BackupResourcesType.app"
				:title="t('Backup App')"
				:data="restore?.backupAppTypeName"
			/>

			<bt-form-item
				v-if="restore && restore?.backupType === BackupResourcesType.files"
				:title="t('backup_path')"
				:data="restore?.backupPath"
			/>
			<bt-form-item
				:title="t('snapshot')"
				:data="
					date.formatDate(restore?.snapshotTime * 1000, 'YYYY-MM-DD HH:mm')
				"
			/>
			<bt-form-item
				v-if="restore && restore?.backupType === BackupResourcesType.files"
				:title="t('Restore location')"
				:data="restore?.restorePath"
			/>
			<bt-form-item
				v-if="restore?.status !== BackupStatus.running"
				:title="t('status')"
				:width-separator="
					(restore?.status === BackupStatus.failed ||
						restore?.status === BackupStatus.rejected) &&
					!!restore?.message
				"
			>
				<div
					class="row justify-end items-center"
					:class="getRestoreColorClass(restore?.status)"
				>
					<div class="status-bg q-mr-xs row items-center justify-center">
						<div
							class="status-node"
							:class="getRestoreColorClass(restore?.status, 'bg')"
						/>
					</div>
					{{ restore?.status }}
				</div>
			</bt-form-item>
			<bt-form-item
				v-if="restore?.status === BackupStatus.running"
				:width-separator="false"
			>
				<template v-slot:all>
					<div class="column justify-start full-width q-px-lg">
						<q-linear-progress
							class="full-width"
							:value="Number(restore?.progress / 10000)"
							size="4px"
							color="info"
						/>
						<div class="text-info text-body2 q-mt-sm">
							{{ t('restoring_message') }}
						</div>
					</div>
				</template>
			</bt-form-item>
			<bt-form-item
				v-if="
					(restore?.status === BackupStatus.failed ||
						restore?.status === BackupStatus.rejected) &&
					!!restore?.message
				"
				:title="t('message')"
				:width-separator="false"
			>
				<div
					class="text-body1 text-negative failed-message-width cursor-pointer"
					@click="onCopy"
				>
					{{ t(restore?.message) }}
				</div>
			</bt-form-item>
		</bt-list>

		<div class="row justify-end items-center">
			<q-btn
				v-if="
					restore?.status === BackupStatus.pending ||
					restore?.status === BackupStatus.running
				"
				dense
				flat
				class="cancel-btn q-px-md q-mt-lg"
				:label="t('cancel')"
				@click="onCancel"
			/>

			<q-btn
				dense
				flat
				v-if="
					restore?.status === BackupStatus.completed &&
					restore?.backupType === BackupResourcesType.files
				"
				class="confirm-btn q-px-md q-mt-lg"
				:label="t('open_in_files')"
				@click="onOpen"
			/>

			<q-btn
				dense
				flat
				v-if="
					restore?.status === BackupStatus.completed &&
					restore?.backupType === BackupResourcesType.app
				"
				class="confirm-btn q-px-md q-mt-lg"
				:label="t('login.open_the_app')"
				@click="onOpenApp"
			/>
		</div>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { bus } from 'src/utils/bus';
import { useRoute } from 'vue-router';
import { date } from 'quasar';
import { onBeforeUnmount, onMounted, ref } from 'vue';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { useBackupStore } from 'src/stores/settings/backup';
import BtList from 'src/components/settings/base/BtList.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import {
	BackupResourcesType,
	BackupStatus,
	getRestoreColorClass,
	RestoreMessage,
	RestorePlanDetail
} from 'src/constant';
import { useApplicationStore } from 'src/stores/settings/application';
import { getApplication } from 'src/application/base';

const { t } = useI18n();
const route = useRoute();
const isCanceling = ref(false);
const backupStore = useBackupStore();
const restore = ref<RestorePlanDetail | null>(null);
const restoreId = route.params.restoreId as string;

async function getDetails() {
	return backupStore.getRestoreDetails(restoreId).then((res) => {
		restore.value = res;
	});
}

function updateRestoreDetail(data: RestoreMessage) {
	console.log(data);
	console.log(restoreId);
	console.log(data.id);
	console.log(restore.value);
	if (data && data.id === restoreId && restore.value) {
		restore.value.progress = data.progress;
		console.log(restore.value.progress);
		restore.value.status = data.status;
		restore.value.message = data.message;
	}
}

onMounted(() => {
	bus.on('restore_state_event', updateRestoreDetail);
	getDetails().catch((e) => {
		console.error(e);
	});
});

onBeforeUnmount(() => {
	bus.off('restore_state_event', updateRestoreDetail);
});

const onOpen = async () => {
	if (restore.value && restore.value.restorePath) {
		let url = backupStore.getModuleSever('files', 'https:');
		url = url + restore.value.restorePath;
		window.open(url);
	}
};

const onOpenApp = async () => {
	if (restore.value && restore.value.backupAppTypeName) {
		const applicationStore = useApplicationStore();
		const app = applicationStore.getApplicationById(
			restore.value.backupAppTypeName
		);
		if (app) {
			const entrance = app.entrances?.find((item) => !item.invisible);
			if (entrance) {
				let url = backupStore.getModuleSever(entrance.id, 'https:');
				window.open(url);
			}
		}
	}
};

const onCopy = () => {
	if (restore.value?.message) {
		getApplication()
			.copyToClipboard(restore.value?.message)
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

const onCancel = async () => {
	isCanceling.value = true;
	backupStore
		.cancelRestore(restoreId)
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
</script>
<style lang="scss" scoped>
.status-bg {
	width: 20px;
	height: 20px;

	.status-node {
		width: 8px;
		height: 8px;
		border-radius: 4px;
	}
}

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

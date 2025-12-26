<template>
	<bt-grid
		class="bg-background-1"
		:repeat-count="1"
		:show-progress="plan?.status === BackupStatus.running"
		:progress="Number(plan?.progress / 10000)"
	>
		<template v-slot:title>
			<div
				class="text-subtitle2 plan-item-title row justify-between items-center clickable-view full-width"
				@click="gotoRestore"
			>
				<div
					class="restore-plan-name single-line"
					style="width: calc(100% - 20px)"
				>
					{{ plan?.name }}
				</div>
				<q-icon name="sym_r_chevron_right" size="20px" />
			</div>
		</template>
		<template v-slot:grid>
			<div
				class="row justify-between items-center full-width"
				style="height: 36px"
			>
				<div
					class="row justify-start items-center"
					style="width: calc(100% - 120px)"
				>
					<q-img
						v-if="plan.backupType === BackupResourcesType.app"
						class="application-logo"
						:src="AppIcon"
					/>
					<q-img v-else class="folder-img" src="/img/folder-default.svg" />
					<div
						class="column justify-start q-ml-sm"
						style="max-width: calc(100% - 40px)"
					>
						<div class="text-body3 text-ink-1 single-line">
							{{
								plan.backupType === BackupResourcesType.files
									? plan.path
									: plan.backupAppTypeName
							}}
						</div>
						<div
							class="single-line text-body3 q-mt-xs"
							:class="getRestoreColorClass(plan?.status)"
						>
							{{ statusText }}
						</div>
					</div>
				</div>

				<div class="row justify-end items-center text-body3 text-ink-3">
					<q-icon size="16px" name="sym_r_browse_gallery" />
					<span class="q-ml-xs">{{
						date.formatDate(props.plan?.snapshotTime * 1000, 'YYYY-MM-DD HH:mm')
					}}</span>
				</div>
			</div>
		</template>
	</bt-grid>
</template>

<script lang="ts" setup>
import { date } from 'quasar';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import BtGrid from '../base/BtGrid.vue';
import { computed, PropType } from 'vue';
import {
	RestorePlan,
	BackupStatus,
	getRestoreColorClass,
	BackupResourcesType
} from 'src/constant';
import { useBackupStore } from 'src/stores/settings/backup';

const { t } = useI18n();
const router = useRouter();
const backupStore = useBackupStore();

const props = defineProps({
	plan: {
		type: Object as PropType<RestorePlan>,
		require: true
	}
});

async function gotoRestore() {
	if (props.plan && props.plan.id) {
		router.push('/backup/restore/' + props.plan.id);
	}
}

const AppIcon = computed(() => {
	if (props.plan?.backupType === BackupResourcesType.app) {
		const list = backupStore.getSupportApplicationOptions();
		const options = list.find(
			(item) => item.value === props.plan.backupAppTypeName
		);
		if (options) {
			return options.app.icon;
		}
	}

	return '/img/folder-default.svg';
});

const statusText = computed(() => {
	const startTime =
		props.plan?.createAt == 0
			? '-'
			: date.formatDate(props.plan?.createAt * 1000, 'YYYY-MM-DD HH:mm');
	const endTime =
		props.plan?.endAt == 0
			? '-'
			: date.formatDate(props.plan?.endAt * 1000, 'YYYY-MM-DD HH:mm');
	switch (props.plan.status) {
		case BackupStatus.pending:
			return t('restore_pending', { time: startTime });
		case BackupStatus.running:
			return t('restoring_message');
		case BackupStatus.completed:
			return t('restore_completed', { time: endTime });
		case BackupStatus.failed:
			return t('restore_failed', { time: endTime });
		case BackupStatus.canceled:
			return t('restore_canceled', { time: endTime });
		case BackupStatus.rejected:
			return t('restore_rejected');
		default:
			return t('unknown');
	}
});
</script>

<style scoped lang="scss">
.plan-item-title {
	margin-bottom: 14px;
	color: $ink-1;

	.restore-plan-name {
		max-width: calc(100% - 32px);
	}
}

.folder-img {
	width: 31px;
	height: 25px;
}

.application-logo {
	width: 32px;
	height: 30px;
	border-radius: 8px;
}

.backup-indicator {
	width: 2px;
	height: 2px;
	border-radius: 50%;
	background: $background-5;
}

.location-img {
	width: 32px;
	height: 32px;
}
</style>

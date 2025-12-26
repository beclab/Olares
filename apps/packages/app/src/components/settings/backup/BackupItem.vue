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
				@click="gotoBackup"
			>
				<div
					class="row justify-start items-center"
					style="width: calc(100% - 20px)"
				>
					<span
						class="text-ink-1 text-subtitle2 single-line backup-plan-name"
						style="max-width: calc(100% - 180px)"
					>
						{{ plan?.name }}
					</span>
					<div
						class="row justify-start items-center text-info bg-background-3 q-ml-sm q-px-sm single-line"
						style="border-radius: 4px; max-width: 180px; height: 20px"
					>
						<q-icon size="12px" name="sym_r_device_reset" />
						<span class="text-overline q-ml-xs">{{ nextSnapShot }}</span>
					</div>
				</div>
				<q-icon name="sym_r_chevron_right" size="20px" />
			</div>
		</template>
		<template v-slot:grid>
			<div
				class="row justify-start items-center full-width"
				style="height: 32px"
			>
				<div class="row justify-start items-center col-5">
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
						<div
							class="text-body3 text-ink-1 single-line"
							style="max-width: 100%"
						>
							{{
								plan.backupType === BackupResourcesType.files
									? plan.path
									: plan.backupAppTypeName
							}}
						</div>
						<div
							v-if="size"
							class="text-overline text-ink-3 q-mt-xs single-line"
							style="max-width: 100%"
						>
							{{ size }}
						</div>
					</div>
				</div>

				<div class="row justify-center items-center col-2">
					<span class="backup-indicator q-mr-xs" />
					<span class="backup-indicator q-mr-xs" />
					<span class="backup-indicator q-mr-xs" />

					<q-img
						class="backup-status-img"
						:src="getBackupStatusImg(plan?.status)"
					/>
					<span class="backup-indicator q-ml-xs" />
					<span class="backup-indicator q-ml-xs" />
					<span class="backup-indicator q-ml-xs" />
				</div>

				<div class="row justify-start items-center col-5">
					<q-img
						class="location-img"
						:src="getBackupIconByLocation(plan.location)"
					/>
					<div
						class="column justify-start q-ml-sm"
						style="max-width: calc(100% - 40px)"
					>
						<div
							class="text-body3 text-ink-1 single-line"
							style="max-width: 100%"
						>
							{{ locationTitle }}
						</div>
						<div
							v-if="locationName"
							class="text-overline text-ink-3 q-mt-xs single-line"
							style="max-width: 100%"
						>
							{{ locationName }}
						</div>
					</div>
				</div>
			</div>
		</template>
	</bt-grid>
</template>

<script lang="ts" setup>
import { computed, PropType } from 'vue';
import { date, format } from 'quasar';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import BtGrid from '../base/BtGrid.vue';
import {
	BackupLocationType,
	BackupPlan,
	BackupResourcesType,
	BackupStatus,
	getBackupIconByLocation,
	getBackupStatusImg
} from 'src/constant';
import humanStorageSize = format.humanStorageSize;
import { useBackupStore } from 'src/stores/settings/backup';

const router = useRouter();
const backupStore = useBackupStore();

const props = defineProps({
	plan: {
		type: Object as PropType<BackupPlan>,
		require: true
	}
});

const { t } = useI18n();
const AppIcon = computed(() => {
	if (props.plan && props.plan?.backupType === BackupResourcesType.app) {
		const list = backupStore.getSupportApplicationOptions();
		const options = list.find(
			(item) => item.value === props.plan.backupAppTypeName
		);
		if (options) {
			console.log(options.app.url);
			return options.app.icon;
		}
	}

	return '/img/folder-default.svg';
});

const size = computed(() => {
	if (props.plan && props.plan.size) {
		try {
			return humanStorageSize(Number(props.plan.size));
		} catch (e) {
			return '';
		}
	} else {
		return '';
	}
});

const nextSnapShot = computed(() => {
	if (props.plan && props.plan.nextBackupTimestamp) {
		return (
			t('next_backup') +
			date.formatDate(props.plan.nextBackupTimestamp * 1000, 'MMM DD, h:mm A')
		);
	}
	return '-';
});

const locationName = computed(() => {
	if (props.plan && props.plan.location) {
		switch (props.plan.location) {
			case BackupLocationType.fileSystem:
				return t('local_directory');
			case BackupLocationType.space:
			case BackupLocationType.awsS3:
			case BackupLocationType.tencentCloud:
				return props.plan.locationConfigName;
			default:
				return props.plan.locationConfigName;
		}
	}
	return '';
});

const locationTitle = computed(() => {
	if (props.plan && props.plan.location) {
		switch (props.plan.location) {
			case BackupLocationType.fileSystem:
				return props.plan.locationConfigName;
			case BackupLocationType.space:
				return 'Olares Space';
			case BackupLocationType.awsS3:
				return 'AWS S3';
			case BackupLocationType.tencentCloud:
				return 'Tencent COS';
			default:
				return props.plan.locationConfigName;
		}
	}
	return '';
});

async function gotoBackup() {
	if (props.plan && props.plan.id) {
		router.push('/backup/' + props.plan.id);
	}
}
</script>

<style scoped lang="scss">
.plan-item-title {
	margin-bottom: 14px;
	color: $ink-1;

	.backup-plan-name {
		max-width: 60%;
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

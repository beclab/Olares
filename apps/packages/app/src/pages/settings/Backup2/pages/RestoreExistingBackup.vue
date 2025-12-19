<template>
	<page-title-component
		:show-back="true"
		:title="t('add_restore_from_existing_backup')"
	/>
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<bt-form>
			<bt-list first :label="t('basic_backup_info')">
				<bt-form-item :title="t('select_a_backup')" :margin-top="false">
					<bt-select
						v-model="selectBackupId"
						:options="backupStore.restoreOptions"
					/>
				</bt-form-item>

				<bt-form-item :title="t('select_a_snapshot')">
					<bt-select
						v-if="snapshotOptions.length > 0"
						v-model="selectSnapshotId"
						:options="snapshotOptions"
					/>
					<div v-else class="text-negative text-body1">
						{{ t('no_available_snapshots') }}
					</div>
				</bt-form-item>

				<bt-form-item :title="t('Restore location')" :width-separator="false">
					<transfet-select-to
						class="q-mt-xs"
						@setSelectPath="setSelectPath"
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
			</bt-list>
		</bt-form>
		<div class="row justify-end">
			<q-btn
				dense
				flat
				:disable="!canSubmit"
				class="confirm-btn q-mt-lg q-px-md"
				:label="t('start_restore')"
				@click="onSubmit"
				:loading="isLoading"
			/>
		</div>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import PageTitleComponent from '../../../../components/settings/PageTitleComponent.vue';
import BtFormItem from '../../../../components/settings/base/BtFormItem.vue';
import TransfetSelectTo from '../../../Electron/Transfer/TransfetSelectTo.vue';
import BtSelect from '../../../../components/settings/base/BtSelect.vue';
import BtForm from '../../../../components/settings/base/BtForm.vue';
import { useDeviceStore } from '../../../../stores/settings/device';
import { BackupStatus, backupOriginsRef } from '../../../../constant';
import { useBackupStore } from '../../../../stores/settings/backup';
import { FilePath } from '../../../../stores/files';
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { date } from 'quasar';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import BtList from '../../../../components/settings/base/BtList.vue';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const restorePathRef = ref();
const selectBackupId = ref();
const selectSnapshotId = ref();
const isLoading = ref(false);
const deviceStore = useDeviceStore();
const backupStore = useBackupStore();
const snapshotOptions = ref([]);

onMounted(async () => {
	const backupId = route.params.backupId as string;
	if (backupId) {
		const backup = await backupStore.getBackupDetails(backupId);
		console.log(backup);
		if (backup) {
			selectBackupId.value = backup.id;
		}
	}
});

const setSelectPath = (fileSavePath: FilePath) => {
	restorePathRef.value = fileSavePath.decodePath;
};

const onSubmit = () => {
	isLoading.value = true;
	backupStore
		.restoreBackup(selectSnapshotId.value, restorePathRef.value)
		.then(() => {
			BtNotify.show({
				type: NotifyDefinedType.SUCCESS,
				message: t('success')
			});
			router.push({ path: '/backup' });
		})
		.catch((e) => {
			console.error(e);
		})
		.finally(() => {
			isLoading.value = false;
		});
};

watch(
	() => selectBackupId.value,
	() => {
		if (selectBackupId.value) {
			snapshotOptions.value = [];
			selectSnapshotId.value = '';
			backupStore
				.getSnapshots(selectBackupId.value, 0, 100)
				.then((response: any) => {
					snapshotOptions.value = response.snapshots
						.filter((item) => item.status === BackupStatus.completed)
						.map((snapshot) => {
							return {
								label: date.formatDate(
									snapshot.createAt * 1000,
									'YYYY-MM-DD HH:mm'
								),
								value: snapshot.id,
								enable: true
							};
						});

					const completeList = response.snapshots.filter(
						(snapshot) => snapshot.status === BackupStatus.completed
					);
					if (completeList.length > 0) {
						const snapshotId = route.params.snapshotId as string;
						if (
							snapshotId &&
							!!snapshotOptions.value.find((item) => item.value === snapshotId)
						) {
							selectSnapshotId.value = snapshotId;
						} else {
							selectSnapshotId.value = completeList[0].id;
						}
					}
				});
		}
	}
);

const canSubmit = computed(() => {
	return (
		!!selectSnapshotId.value && !!selectBackupId.value && !!restorePathRef.value
	);
});
</script>

<style scoped lang="scss"></style>

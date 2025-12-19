<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('select_a_backup_location')"
		:skip="false"
		:ok="t('confirm')"
		size="medium"
		:platform="deviceStore.platform"
		:cancel="t('cancel')"
		@onSubmit="onConfirm"
		:okDisabled="!OKAble"
	>
		<div class="text-body1 text-ink-3 q-mb-sm">
			{{ t('backup_to_local_directory') }}
		</div>

		<transfet-select-to
			class="q-mt-xs"
			@setSelectPath="onPathClick"
			:origins="backupOriginsRef"
			:master-node="true"
		>
			<template v-slot:default>
				<div
					v-if="fileSavePathRef"
					class="row justify-between items-center item-selected cursor-pointer integration-account-item"
				>
					<div
						class="row justify-start items-center"
						style="max-width: calc(100% - 30px)"
					>
						<q-img class="folder-img" src="/img/folder-default.svg" />
						<span
							class="text-subtitle2 text-ink-1 q-ml-md single-line"
							style="max-width: calc(100% - 80px)"
							>{{ fileSavePathRef.decodePath }}</span
						>
						<q-icon
							class="text-ink-1 q-ml-sm"
							size="20px"
							name="sym_r_edit_square"
						/>
					</div>

					<bt-check-box-component
						:model-value="selectLocation.key === BackupLocationType.fileSystem"
					/>
				</div>
				<div
					v-else
					class="row justify-start items-center item-empty cursor-pointer"
				>
					<q-icon class="text-info" size="30px" name="sym_r_add" />
					<span class="text-subtitle2 text-info q-ml-md">{{
						t('add_local_path')
					}}</span>
				</div>
			</template>
		</transfet-select-to>

		<div class="text-body1 text-ink-3 q-mt-lg q-mb-sm">
			{{ t('backup_to_online_storage') }}
		</div>

		<div class="column grid-item">
			<account-item
				v-for="(item, index) in integrationStore.backupAccounts"
				:key="`${item.type}_${item.name}`"
				:title="item.type"
				:available="item.available"
				:detail="item.name"
				class="integration-account-item"
				:side="true"
				:selectable="true"
				:style="index == 0 ? '' : 'margin-top:12px'"
				@account-click="onlineClick(item, `${item.type}_${item.name}`)"
				:selected="selectLocation.key === `${item.type}_${item.name}`"
			>
				<template v-slot:avatar>
					<!-- <setting-avatar :size="40" style="margin-left: 8px" /> -->
					<q-img
						width="40px"
						height="40px"
						:noSpinner="true"
						:src="integrationStore.getAccountIcon(item)"
					/>
				</template>
			</account-item>

			<div
				class="row justify-start items-center item-empty cursor-pointer integration-account-item"
				@click="addAccount"
			>
				<q-icon class="text-info" size="30px" name="sym_r_add" />
				<span class="text-subtitle2 text-info q-ml-md">{{
					t('add_account')
				}}</span>
			</div>
		</div>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { FilePath } from 'src/stores/files';
import { useDeviceStore } from 'src/stores/settings/device';
import { IntegrationAccountMiniData } from '@bytetrade/core';
import { useIntegrationStore } from 'src/stores/settings/integration';
import TransfetSelectTo from '../../../Electron/Transfer/TransfetSelectTo.vue';
import AccountItem from '../../../../components/settings/account/AccountItem.vue';
import AddIntegrationDialog from '../../Integration/dialog/AddIntegrationDialog.vue';
import BtCheckBoxComponent from '../../../../components/settings/base/BtCheckBoxComponent.vue';
import {
	backupOriginsRef,
	BackupLocationType,
	getBackupLocationTypeByIntegrationAccount
} from 'src/constant';

const $q = useQuasar();
const { t } = useI18n();
const CustomRef = ref();
const router = useRouter();
const fileSavePathRef = ref();
const onlineAccount = ref(null);
const deviceStore = useDeviceStore();
const integrationStore = useIntegrationStore();
const selectLocation = ref<{
	type: BackupLocationType | null;
	key: string;
	data: any;
}>({ type: null, key: '', data: null });

const onConfirm = () => {
	if (!selectLocation.value.type || !selectLocation.value.data) {
		console.error('error 2');
		return;
	}

	CustomRef.value.onDialogOK({
		type: selectLocation.value.type,
		data: selectLocation.value.data
	});
};

const onPathClick = (fileSavePath: FilePath) => {
	fileSavePathRef.value = fileSavePath;
	selectLocation.value.type = BackupLocationType.fileSystem;
	selectLocation.value.data = fileSavePathRef.value;
	selectLocation.value.key = BackupLocationType.fileSystem;
};

const onlineClick = async (item: IntegrationAccountMiniData, key: string) => {
	selectLocation.value.type = getBackupLocationTypeByIntegrationAccount(item);
	selectLocation.value.key = key;
	if (
		selectLocation.value.type === BackupLocationType.tencentCloud ||
		selectLocation.value.type === BackupLocationType.awsS3
	) {
		const integrationStore = useIntegrationStore();
		onlineAccount.value = await integrationStore.getAccountFullData(item);
		console.log(onlineAccount.value);
	} else {
		onlineAccount.value = item;
	}

	selectLocation.value.data = onlineAccount.value;
};

const OKAble = computed(() => {
	return selectLocation.value.type && selectLocation.value.data;
});

const addAccount = () => {
	if (deviceStore.isMobile) {
		router.push({
			path: '/integration/add',
			query: {
				backup: 1
			}
		});
	} else {
		$q.dialog({
			component: AddIntegrationDialog,
			componentProps: {
				backup: true
			}
		}).onOk(() => {});
	}
};
</script>

<style scoped lang="scss">
.item-empty {
	border-radius: 12px;
	border: 1px dashed $separator;
	background: $background-1;
	padding: 12px;

	.empty-img {
		width: 40px;
		height: 40px;
	}
}

.item-selected {
	border-radius: 12px;
	border: 1px solid $separator;
	background: $background-1;
	padding: 12px;

	.folder-img {
		width: 39px;
		height: 31px;
	}
}

.grid-item {
	display: grid;
	grid-template-columns: 1fr;
}
</style>

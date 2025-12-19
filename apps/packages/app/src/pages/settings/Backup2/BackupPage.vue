<template>
	<page-title-component
		:show-back="false"
		:title="t(`home_menus.${MENU_TYPE.Backup.toLowerCase()}`)"
	>
		<template v-slot:end>
			<q-btn
				flat
				dense
				v-if="backupStore.backupList.length !== 0"
				@click="goBackupFile"
			>
				<q-icon size="20px" name="sym_r_add" color="info" />
				<bt-popup
					v-if="backupStore.getSupportApplicationOptions().length > 0"
					style="width: 176px"
					@click.stop
				>
					<bt-popup-item
						v-close-popup
						active-soft="background-hover"
						active-text="text-ink-2"
						:title="t('Backup Files')"
						@on-item-click="newBackup(BackupResourcesType.files)"
					/>
					<bt-popup-item
						v-close-popup
						active-soft="background-hover"
						active-text="text-ink-2"
						:title="t('Backup App')"
						@on-item-click="newBackup(BackupResourcesType.app)"
					/>
				</bt-popup>
			</q-btn>
		</template>
	</page-title-component>
	<app-menu-empty
		v-if="backupStore.backupList.length === 0"
		:title="t('Add Backup Task')"
		:button-label="t('add_backup')"
		:message="
			t(
				'Create a backup task to regularly save your important data and keep it secure.'
			)
		"
		:menu-type="MENU_TYPE.Backup"
		@click="goBackupFile"
	>
		<bt-popup
			v-if="backupStore.getSupportApplicationOptions().length > 0"
			style="width: 176px"
			@click.stop
		>
			<bt-popup-item
				v-close-popup
				active-soft="background-hover"
				active-text="text-ink-2"
				:title="t('Backup Files')"
				@on-item-click="newBackup(BackupResourcesType.files)"
			/>
			<bt-popup-item
				v-close-popup
				active-soft="background-hover"
				active-text="text-ink-2"
				:title="t('Backup App')"
				@on-item-click="newBackup(BackupResourcesType.app)"
			/>
		</bt-popup>
	</app-menu-empty>
	<bt-scroll-area v-else class="nav-height-scroll-area-conf">
		<backup-item
			v-for="plan of backupStore.backupList"
			:key="plan.id"
			:plan="plan"
		/>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { MENU_TYPE } from 'src/constant';
import { BackupResourcesType } from 'src/constant';
import BtPopup from '../../../components/base/BtPopup.vue';
import { useBackupStore } from 'src/stores/settings/backup';
import BtPopupItem from '../../../components/base/BtPopupItem.vue';
import AppMenuEmpty from 'src/components/settings/AppMenuEmpty.vue';
import BackupItem from '../../../components/settings/backup/BackupItem.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';

const { t } = useI18n();
const router = useRouter();
const backupStore = useBackupStore();

function newBackup(type: BackupResourcesType) {
	router.push('/backup/create_backup/' + type);
}

function goBackupFile() {
	if (backupStore.getSupportApplicationOptions().length === 0) {
		router.push('/backup/create_backup/' + BackupResourcesType.files);
	}
}
</script>

<style scoped lang="scss">
.backup-border {
	padding: 20px;
	border-radius: 12px;
	border: 1px solid $separator;

	.space-logo {
		width: 24px;
		height: 24px;
	}
}

.add-btn {
	border-radius: 8px;
	border: 1px solid $separator;
	cursor: pointer;
	padding: 5px 8px;
	text-decoration: none;

	.add-title {
		color: $ink-2;
	}
}

.add-btn-padding-mobile {
	padding: 8px 10px;
}

.add-btn:hover {
	background-color: $background-3;
}

.terminus-space-backup-title {
	text-align: center;
	width: 336px;
}
</style>

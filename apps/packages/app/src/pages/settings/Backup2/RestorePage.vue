<template>
	<page-title-component
		:show-back="false"
		:title="t(`home_menus.${MENU_TYPE.Restore.toLowerCase()}`)"
	>
		<template v-slot:end>
			<q-btn flat dense v-if="backupStore.restoreList.length !== 0">
				<q-icon size="20px" name="sym_r_add" color="info" />
				<bt-popup style="width: 176px">
					<bt-popup-item
						v-close-popup
						active-soft="background-hover"
						active-text="text-ink-2"
						:title="t('from_local_path')"
						@on-item-click="newRestore(BackupLocationType.fileSystem)"
					/>
					<bt-popup-item
						v-close-popup
						active-soft="background-hover"
						active-text="text-ink-2"
						:title="t('from_space_url')"
						@on-item-click="newRestore(BackupLocationType.space)"
					/>
					<bt-popup-item
						v-close-popup
						active-soft="background-hover"
						active-text="text-ink-2"
						:title="t('from_tencent_cos_url')"
						@on-item-click="newRestore(BackupLocationType.tencentCloud)"
					/>
					<bt-popup-item
						v-close-popup
						active-soft="background-hover"
						active-text="text-ink-2"
						:title="t('from_aws_s3_url')"
						@on-item-click="newRestore(BackupLocationType.awsS3)"
					/>
				</bt-popup>
			</q-btn>
		</template>
	</page-title-component>
	<app-menu-empty
		v-if="backupStore.restoreList.length === 0"
		:title="t('Add restore task')"
		:button-label="t('Add restore task')"
		:message="
			t(
				'Create a restore task to recover your data from backups and quickly retrieve what you need.'
			)
		"
		:menu-type="MENU_TYPE.Restore"
	>
		<bt-popup style="width: 176px">
			<bt-popup-item
				v-close-popup
				active-soft="background-hover"
				active-text="text-ink-2"
				:title="t('from_local_path')"
				@on-item-click="newRestore(BackupLocationType.fileSystem)"
			/>
			<bt-popup-item
				v-close-popup
				active-soft="background-hover"
				active-text="text-ink-2"
				:title="t('from_space_url')"
				@on-item-click="newRestore(BackupLocationType.space)"
			/>
			<bt-popup-item
				v-close-popup
				active-soft="background-hover"
				active-text="text-ink-2"
				:title="t('from_tencent_cos_url')"
				@on-item-click="newRestore(BackupLocationType.tencentCloud)"
			/>
			<bt-popup-item
				v-close-popup
				active-soft="background-hover"
				active-text="text-ink-2"
				:title="t('from_aws_s3_url')"
				@on-item-click="newRestore(BackupLocationType.awsS3)"
			/>
		</bt-popup>
	</app-menu-empty>
	<bt-scroll-area v-else class="nav-height-scroll-area-conf">
		<restore-item
			v-for="plan of backupStore.restoreList"
			:key="plan.id"
			:plan="plan"
		/>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { MENU_TYPE } from 'src/constant';
import BtPopup from '../../../components/base/BtPopup.vue';
import { useBackupStore } from 'src/stores/settings/backup';
import BtPopupItem from '../../../components/base/BtPopupItem.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import RestoreItem from '../../../components/settings/backup/RestoreItem.vue';
import { BackupLocationType } from 'src/constant';
import AppMenuEmpty from 'src/components/settings/AppMenuEmpty.vue';

const { t } = useI18n();
const router = useRouter();
const backupStore = useBackupStore();

function newRestore(type: BackupLocationType) {
	router.push('/backup/restoreOptions/' + type);
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

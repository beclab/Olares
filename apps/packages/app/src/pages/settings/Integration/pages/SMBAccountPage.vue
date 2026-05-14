<template>
	<page-title-component
		:title="t('SMB account management')"
		:show-back="true"
	/>

	<bt-scroll-area
		v-if="accountList.length > 0"
		class="nav-height-scroll-area-conf"
	>
		<bt-list first>
			<bt-form-item
				v-for="(item, index) in accountList"
				:key="item.id"
				:title="item.name"
				:width-separator="index !== accountList.length - 1"
			>
				<q-icon
					name="sym_r_delete"
					class="cursor-pointer"
					color="ink-2"
					size="24px"
					@click="deleteAccount(item.id)"
				/>
			</bt-form-item>
		</bt-list>
		<div v-if="accountList.length > 0" class="row justify-end q-mt-lg">
			<q-btn
				dense
				flat
				no-caps
				class="confirm-btn q-px-md q-mb-lg"
				:label="t('Add account')"
				@click="addAccount"
			/>
		</div>
		<div style="height: 20px" />
	</bt-scroll-area>
	<app-menu-empty
		v-else
		:title="t('No SMB accounts')"
		:message="
			t(
				'Add an account to easily back up, restore, and sync your important data.'
			)
		"
		:button-label="t('Add account')"
		image="settings/imgs/root/smb.svg"
		@on-button-click="addAccount"
	/>
</template>

<script lang="ts" setup>
import AddSMBAccountDialog from 'src/pages/settings/Integration/dialog/AddSMBAccountDialog.vue';
import ReminderDialogComponent from 'src/components/settings/ReminderDialogComponent.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import AppMenuEmpty from 'src/components/settings/AppMenuEmpty.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import { deleteSMBAccount, getSMBAccountList } from 'src/api/settings/smb';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { useI18n } from 'vue-i18n';
import { onMounted, ref } from 'vue';
import { useQuasar } from 'quasar';

const { t } = useI18n();
const accountList = ref([]);
const $q = useQuasar();

onMounted(async () => {
	accountList.value = await getSMBAccountList();
});

const addAccount = async () => {
	$q.dialog({
		component: AddSMBAccountDialog,
		componentProps: {}
	}).onOk(async () => {
		accountList.value = await getSMBAccountList();
	});
};

const deleteAccount = async (id: string) => {
	$q.dialog({
		component: ReminderDialogComponent,
		componentProps: {
			title: t('delete_account'),
			message: t('Are you sure to delete the account?'),
			confirmText: t('confirm'),
			cancelText: t('cancel')
		}
	}).onOk(async () => {
		deleteSMBAccount([id])
			.then(async () => {
				BtNotify.show({
					type: NotifyDefinedType.SUCCESS,
					message: t('success')
				});
				const index = accountList.value.findIndex((item) => item.id === id);
				if (index > -1) {
					accountList.value.splice(index, 1);
				}
			})
			.catch((e) => {
				BtNotify.show({
					type: NotifyDefinedType.FAILED,
					message: e.response.data.message || e.message
				});
			});
	});
};
</script>

<style lang="scss" scoped>
.terminus-cloud-page {
	width: 100%;
	height: calc(100% - 56px);

	.terminus-cloud-icon {
		width: 170px;
	}
}
</style>

<template>
	<page-title-component :show-back="true" :title="t('account_settings')">
		<template v-slot:end v-if="deviceStore.isMobile">
			<div class="row justify-center items-center" @click="deleteAction()">
				<q-icon name="sym_r_delete" color="ink-1" size="20px" />
			</div>
		</template>
	</page-title-component>
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<bt-list v-if="accountData" first>
			<account-item
				:border="false"
				:title="accountData.name"
				:available="accountData.available"
				:clickable="false"
				:side="false"
				:detail="`${t('Authorized time')}: ${formattedDate(
					accountData.create_at
				)}`"
			>
				<template v-slot:avatar>
					<q-img
						width="40px"
						height="40px"
						:noSpinner="true"
						:src="getAccountIcon(accountData)"
					/>
				</template>
			</account-item>
			<bt-separator
				v-if="detailInfo && detailInfo.scopes.length > 0"
				:offset="20"
			/>
			<div
				class="text-subtitle1 text-ink-1 q-mt-md q-px-md"
				v-if="detailInfo?.title"
			>
				{{ detailInfo.title }}
			</div>
			<div
				class="text-body2 text-ink-2 q-mt-sm q-px-md q-pb-md"
				v-if="detailInfo && detailInfo.scopes.length > 0"
			>
				<div
					v-for="(item, index) in detailInfo?.scopes"
					:key="index"
					class="row"
					:class="index > 0 ? 'q-mt-sm' : ''"
				>
					<q-icon :name="item.icon" size="20px" />
					<div class="q-ml-sm" style="max-width: calc(100% - 30px)">
						{{ item.introduce }}
					</div>
				</div>
			</div>
		</bt-list>
		<div
			class="row justify-end q-mt-md"
			v-if="accountData && !deviceStore.isMobile"
		>
			<q-btn
				dense
				flat
				class="cancel-btn q-px-md"
				:label="t('delete')"
				@click="deleteAction"
			/>
		</div>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import { useI18n } from 'vue-i18n';
import { onMounted, ref } from 'vue';
import { useIntegrationStore } from 'src/stores/settings/integration';
import { IntegrationPermissions } from 'src/services/abstractions/integration/integrationService';
import AccountItem from 'src/components/settings/account/AccountItem.vue';
import { date, useQuasar } from 'quasar';
import { getRequireImage } from 'src/utils/settings/helper';
import integraionService from 'src/services/integration/index';
import ReminderDialogComponent from 'src/components/settings/ReminderDialogComponent.vue';
import { notifyFailed } from 'src/utils/settings/btNotify';
import { useDeviceStore } from 'src/stores/settings/device';
import BtSeparator from 'src/components/settings/base/BtSeparator.vue';
import { AccountType, IntegrationAccountMiniData } from '@bytetrade/core';
import BtList from 'src/components/settings/base/BtList.vue';

const { t } = useI18n();

const route = useRoute();

const router = useRouter();

const $q = useQuasar();

const integrationStore = useIntegrationStore();
const deviceStore = useDeviceStore();

const type = route.params.type as AccountType;
const name = route.params.name as string;
const accountData = ref<IntegrationAccountMiniData | undefined>();

const detailInfo = ref<IntegrationPermissions | undefined>();

onMounted(async () => {
	try {
		const data = await integrationStore.getAccountByTypeAndName(type, name);
		if (data && data.length > 0) {
			accountData.value = data[0];
			detailInfo.value = await integraionService
				.getInstanceByType(accountData.value!.type)
				?.permissions();
		}
	} catch (error) {
		console.log(error);
	}
});

const formattedDate = (datetime: number) => {
	if (datetime == 0) {
		return '--';
	}
	return date.formatDate(datetime, 'YYYY-MM-DD HH:mm:ss');
};

const getAccountIcon = (data: IntegrationAccountMiniData) => {
	const account = integrationStore.getAccountByType(data);
	if (!account) {
		return '';
	}
	return getRequireImage(`integration/${account.detail.icon}`);
};

const deleteAction = async () => {
	$q.dialog({
		component: ReminderDialogComponent,
		componentProps: {
			title: t('delete_item', {
				item: t('account')
			}),
			message: t('are_you_sure_you_want_to_delete_item', {
				item: name
			}),
			useCancel: true
		}
	}).onOk(async () => {
		try {
			$q.loading.show();
			await integrationStore.deleteAccount(accountData.value!);
			integrationStore.accounts = integrationStore.accounts.filter(
				(e) =>
					e.type != accountData.value?.type ||
					(e.type == accountData.value?.type &&
						e.name != accountData.value?.name)
			);
			$q.loading.hide();
			router.back();
		} catch (error) {
			$q.loading.hide();
			notifyFailed(error.message);
		}
	});
};
</script>

<style scoped lang="scss"></style>

<template>
	<page-title-component
		:show-back="true"
		:title="t('Link Your Accounts & Data')"
	/>
	<app-menu-empty
		v-if="integrationStore.accounts.length === 0"
		:title="t('No accounts connected')"
		:button-label="t('integration.add_account')"
		:message="
			t(
				'Add an account to easily back up, restore, and sync your important data across platforms.'
			)
		"
		:menu-type="MENU_TYPE.Integration"
		@on-button-click="addAccount"
	/>
	<bt-scroll-area v-else class="nav-height-scroll-area-conf">
		<account-item
			v-for="(item, index) in integrationStore.accounts"
			:key="`${item.type}_${item.name}`"
			:title="item.name"
			:available="item.available"
			:detail="`${t('Authorized time')}:${formattedDate(item.create_at)}`"
			@account-click="clickCloud(item)"
			:style="[
				deviceStore.isMobile ? 'height: 64px' : '',
				index === 0 ? 'margin-top:12px' : 'margin-top:20px'
			]"
			:side="item.type == AccountType.Space ? false : true"
		>
			<template v-slot:avatar>
				<q-img
					width="40px"
					height="40px"
					:noSpinner="true"
					:src="integrationStore.getAccountIcon(item)"
				/>
			</template>
		</account-item>

		<list-bottom-func-btn
			@funcClick="addAccount"
			class="q-mt-md"
			:title="t('add_account')"
		/>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import ListBottomFuncBtn from 'src/components/settings/ListBottomFuncBtn.vue';
import AccountItem from 'src/components/settings/account/AccountItem.vue';
import AddIntegrationDialog from '../dialog/AddIntegrationDialog.vue';
import AppMenuEmpty from 'src/components/settings/AppMenuEmpty.vue';
import { IntegrationAccountMiniData, AccountType } from '@bytetrade/core';
import { useIntegrationStore } from 'src/stores/settings/integration';
import integraionService from 'src/services/integration/index';
import { useDeviceStore } from 'src/stores/settings/device';
import { useRouter } from 'vue-router';
import { date, useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { MENU_TYPE } from 'src/constant';

const { t } = useI18n();

const router = useRouter();
const deviceStore = useDeviceStore();

const $q = useQuasar();

const integrationStore = useIntegrationStore();

function clickCloud(account: IntegrationAccountMiniData) {
	if (account.type === AccountType.Space) {
		return;
	}
	const path = integraionService
		.getInstanceByType(account.type)
		?.detailPath(account);
	if (path) {
		router.push({ path });
	}
}

const formattedDate = (datetime: number) => {
	if (datetime <= 0) {
		return '--';
	}
	return date.formatDate(datetime, 'YYYY-MM-DD HH:mm:ss');
};

const addAccount = () => {
	if (deviceStore.isMobile) {
		router.push('/integration/add');
	} else {
		$q.dialog({
			component: AddIntegrationDialog
		}).onOk(() => {});
	}
};
</script>

<style scoped lang="scss">
.cookie-manage-icon {
	width: 40px;
	height: 40px;
	background: linear-gradient(180deg, #fcfcfc 0%, #f2f2f2 100%);
	border-radius: 50%;
}
.application-logo {
	width: 40px;
	height: 40px;
}

.application-name {
	color: $ink-1;
}

.application-label {
	color: $ink-2;
}

.add-btn {
	border-radius: 8px;
	padding: 6px 12px;
	border: 1px solid $separator;
	cursor: pointer;
	text-decoration: none;

	.add-title {
		color: $ink-2;
	}
}

.add-btn:hover {
	background-color: $background-3;
}
</style>

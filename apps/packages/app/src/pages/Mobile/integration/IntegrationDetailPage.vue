<template>
	<div class="integration-detail-root">
		<terminus-title-bar
			right-icon="sym_r_more_horiz"
			@on-right-click="moreClick"
			:title="t('integration.account_settings')"
		/>
		<terminus-scroll-area class="integration-detail-scroll">
			<template v-slot:content>
				<div
					v-if="accountData"
					class="detail-border-item"
					:style="
						detailInfo && detailInfo.scopes.length > 0
							? 'padding-bottom: 16px'
							: ''
					"
				>
					<integration-item
						:border="false"
						:title="accountData.name"
						:available="accountData.available"
						:clickable="false"
						:side="false"
						:detail="`Authorized time: ${formattedDate(accountData.create_at)}`"
					>
						<template v-slot:avatar>
							<q-img
								width="40px"
								height="40px"
								:noSpinner="true"
								:src="getAccountIcon(accountData)"
							/>
						</template>
					</integration-item>
					<q-separator
						class="separator-background q-mt-sm"
						v-if="detailInfo && detailInfo.scopes.length > 0"
					/>
					<div
						class="text-subtitle1 text-ink-1 q-mt-lg"
						v-if="detailInfo?.title"
					>
						{{ detailInfo.title }}
					</div>
					<div
						class="text-body2 text-ink-2 q-mt-sm"
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
				</div>
			</template>
		</terminus-scroll-area>
	</div>
</template>

<script setup lang="ts">
import TerminusTitleBar from '../../../components/common/TerminusTitleBar.vue';
import TerminusScrollArea from '../../../components/common/TerminusScrollArea.vue';
import { date, useQuasar } from 'quasar';
import IntegrationActionsDialog from './IntegrationActionsDialog.vue';
import TerminusTipDialog from '../../../components/dialog/TerminusTipDialog.vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import {
	AccountType,
	IntegrationAccountMiniData,
	IntegrationPermissions
} from '../../../services/abstractions/integration/integrationService';
import integrationService from '../../../services/integration';
import { ref } from 'vue';
import { useIntegrationStore } from '../../../stores/integration';
import IntegrationItem from './IntegrationItem.vue';
import { onMounted } from 'vue';
import { getRequireImage } from '../../../utils/imageUtils';
import { notifyFailed } from '../../../utils/notifyRedefinedUtil';

const $q = useQuasar();

const { t } = useI18n();

const route = useRoute();
const router = useRouter();

const integrationStore = useIntegrationStore();

const type = route.params.type as AccountType;
const name = (route.params.name || '') as string;
const accountData = ref<IntegrationAccountMiniData | undefined>();
const detailInfo = ref<IntegrationPermissions | undefined>();
onMounted(async () => {
	accountData.value = integrationStore.accounts.find(
		(e) => e.type == type && e.name == name
	);
	detailInfo.value = await integrationService
		.getInstanceByType(accountData.value!.type)
		?.permissions();
	try {
		if (!accountData.value) {
			const accounts = await integrationStore.getAccountByTypeAndName(
				type,
				name
			);
			if (accounts && accounts.length > 0) {
				accountData.value = accounts[0];
			}
		}
	} catch (error) {
		console.log(error);
	}
	if (!accountData.value) {
		router.back();
	}
});

const moreClick = () => {
	$q.dialog({
		component: IntegrationActionsDialog,
		componentProps: {}
	}).onOk(async (operate: string) => {
		if (operate == 'delete') {
			if (!accountData.value) {
				return;
			}
			$q.dialog({
				component: TerminusTipDialog,
				componentProps: {
					title: t('delete_account'),
					message: accountData.value!.name,
					navigation: t('cancel'),
					position: t('delete')
				}
			}).onOk(async () => {
				$q.loading.show();
				try {
					await integrationStore.deleteAccount(accountData.value);
					$q.loading.hide();
				} catch (error) {
					$q.loading.hide();
					notifyFailed(error.message);
					return;
				}
				router.back();
			});
		}
	});
};

const formattedDate = (datetime: number) => {
	if (datetime <= 0) {
		return '--';
	}
	return date.formatDate(datetime, 'YYYY-MM-DD HH:mm:ss');
};
const getAccountIcon = (data: IntegrationAccountMiniData) => {
	const account = integrationService.getAccountByType(data.type);
	if (!account) {
		return '';
	}
	return getRequireImage(`setting/integration/${account.detail.icon}`);
};
</script>

<style scoped lang="scss">
.integration-detail-root {
	width: 100%;
	height: 100%;

	.integration-detail-scroll {
		width: 100%;
		height: calc(100% - 56px);

		.detail-border-item {
			border: 1px solid $separator;
			border-radius: 12px;
			margin-top: 12px;
			margin-left: 20px;
			margin-right: 20px;
			padding: 0px 16px;
		}
	}
}
</style>

<template>
	<page-title-component :show-back="true" :title="t('olares_space')" />

	<bt-scroll-area class="nav-height-scroll-area-conf">
		<div
			class="terminus-cloud-page column items-center"
			v-if="!bindSpaceAccount"
		>
			<QRCodeLogin @success="integrationStore.getAccount('all')">
				<template v-slot:mode>
					<div class="row items-center justify-center">
						<q-img class="terminus-cloud-icon" src="settings/scan.svg" />
						<div class="q-ml-md text-h4 text-ink-1">
							{{ t('login.scan_to_log_in') }}
						</div>
					</div>
				</template>
			</QRCodeLogin>
		</div>
		<div v-else>
			<bt-list>
				<q-item style="padding: 0" class="item-margin-left q-my-sm">
					<div class="terminus-info row items-center justify-between">
						<div class="row items-center">
							<setting-avatar :size="56" style="margin-left: 8px" />
							<div class="column justify-between" style="margin-left: 8px">
								<div class="row items-center">
									<div class="text-h4 person-name q-mr-sm">
										{{ adminStore.user.name }}
									</div>
									<div
										v-if="integrationStore.plan"
										class="q-px-md text-subtitle3 text-white plan row items-center"
										:class="`bg-${integrationStore.planLevel.color}`"
									>
										{{ integrationStore.planLevel.text }}
									</div>
								</div>
								<div class="text-body3 person-id q-mt-xs">
									{{ adminStore.terminus.terminusName }}
								</div>
							</div>
						</div>
						<div class="row items-center item-margin-right">
							<list-bottom-func-btn
								:title="t('Go to Olares Space')"
								:height="32"
								@funcClick="goSpace"
							/>
						</div>
					</div>
				</q-item>
			</bt-list>
			<bt-list v-if="reverseProxyMode !== ReverseProxyMode.NoNeed">
				<bt-form-item :width-separator="false">
					<template #title>
						<span class="text-body1 text-ink-1">
							{{ t('Reverse Proxy') }}
							<q-btn dense>
								<q-icon name="sym_r_help" size="16px" color="ink-3"></q-icon>
								<q-tooltip>{{
									t('Forwards requests to backend servers.')
								}}</q-tooltip>
							</q-btn>
						</span>
					</template>
					<div class="row items-center justify-between">
						{{
							reverseProxyOptions().find((e) => e.value == reverseProxyMode)
								?.label
						}}
					</div>
				</bt-form-item>
			</bt-list>

			<bt-list v-if="integrationStore.backupPlan">
				<storage-useage
					:title="
						t('backup_usage') +
						': ' +
						t('Storage used', {
							Storage: calculateSize(
								integrationStore.backupPlan?.usageSize || 0
							)
						})
					"
					:tip="t('Forwards requests to backend servers.')"
					:total="integrationStore.backupPlan?.totalSize"
					:used="integrationStore.backupPlan?.usageSize"
					:overage="0"
				/>
			</bt-list>

			<bt-list v-if="integrationStore.trafficPlan">
				<storage-useage
					:title="
						t('Traffic usage') +
						': ' +
						t('Storage used', {
							Storage: calculateSize(
								integrationStore.trafficPlan?.usedSize || 0
							)
						})
					"
					:tip="t('Forwards requests to backend servers.')"
					:total="integrationStore.trafficPlan?.totalSize"
					:used="integrationStore.trafficPlan?.usedSize"
					:overage="0"
				/>
			</bt-list>
			<list-bottom-func-btn
				:title="t('Log out')"
				style="margin-top: 12px"
				@click="deleteAction"
			/>
		</div>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import ReminderDialogComponent from 'src/components/settings/ReminderDialogComponent.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import ListBottomFuncBtn from 'src/components/settings/ListBottomFuncBtn.vue';
import SettingAvatar from 'src/components/settings/base/SettingAvatar.vue';
import StorageUseage from 'src/components/settings/StorageUseage.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import QRCodeLogin from 'src/components/settings/QRCodeLogin.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import { useIntegrationStore } from 'src/stores/settings/integration';
import { ReverseProxyMode, reverseProxyOptions } from 'src/constant';
import { useNetworkStore } from 'src/stores/settings/network';
import { getSuitableValue } from 'src/utils/monitoring';
import { useAdminStore } from 'src/stores/settings/admin';
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';

const { t } = useI18n();
const integrationStore = useIntegrationStore();
const adminStore = useAdminStore();
const networkStore = useNetworkStore();
const $q = useQuasar();

const bindSpaceAccount = computed(() => {
	if (integrationStore.accountLoading) {
		return true;
	}
	return integrationStore.spaceAccount.length > 0;
});

const goSpace = () => {
	window.open('https://space.olares.com');
};

watch(
	() => integrationStore.spaceAccount,
	async () => {
		if (integrationStore.spaceAccount.length > 0) {
			await integrationStore.getUsage();
		}
	},
	{
		immediate: true
	}
);

const reverseProxyMode = ref(ReverseProxyMode.NoNeed);

watch(
	() => networkStore.reverseProxy,
	() => {
		if (networkStore.reverseProxy) {
			reverseProxyMode.value = networkStore.reverseProxy
				.enable_cloudflare_tunnel
				? ReverseProxyMode.CloudFlare
				: networkStore.reverseProxy.enable_frp
				? networkStore.reverseProxy.frp_auth_method == 'jws'
					? ReverseProxyMode.OlaresTunnel
					: ReverseProxyMode.SelfBuiltFrp
				: ReverseProxyMode.NoNeed;
		}
	}
);

const calculateSize = (size: number) => {
	return getSuitableValue(size.toString(), 'disk');
};

const deleteAction = async () => {
	const accountData = integrationStore.spaceAccount[0];

	$q.dialog({
		component: ReminderDialogComponent,
		componentProps: {
			title: t('delete_item', {
				item: t('account')
			}),
			message: t('are_you_sure_you_want_to_delete_item', {
				item: accountData.name
			}),
			useCancel: true
		}
	}).onOk(async () => {
		try {
			$q.loading.show();
			await integrationStore.deleteAccount(accountData);
			integrationStore.accounts = integrationStore.accounts.filter(
				(e) =>
					e.type != accountData.type ||
					(e.type == accountData.type && e.name != accountData.name)
			);
			$q.loading.hide();
		} catch (error) {
			$q.loading.hide();
		}
	});
};

onMounted(() => {
	networkStore.configReverseProxy();
});
</script>

<style scoped lang="scss">
.terminus-cloud-page {
	width: 100%;
	height: calc(100% - 56px);
	.terminus-cloud-icon {
		width: 32px;
	}
}
.terminus-info {
	width: 100%;
	height: 78px;

	.person-name {
		color: $ink-1;
	}

	.person-id {
		color: $ink-3;
	}

	.plan {
		height: 24px;
		border-radius: 4px;
	}
}

.legend-item {
	width: 8px;
	height: 8px;
	border-radius: 4px;
}
</style>

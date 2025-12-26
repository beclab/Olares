<template>
	<page-title-component :show-back="true" :title="t('domain_setup')" />
	<bt-scroll-area v-if="isShared" class="nav-height-scroll-area-conf">
		<bt-list>
			<bt-form-item
				:title="t('endpoint')"
				:width-separator="!isShared"
				:dataWidthP="60"
			>
				<div
					class="row justify-end items-center full-width"
					@click="setCopyInfo(entranceUrl)"
				>
					<div
						style="
							max-width: calc(100% - 40px);
							overflow: hidden;
							white-space: nowrap;
							text-overflow: ellipsis;
							margin-right: 10px;
						"
						class="text-body-3 text-ink-3"
					>
						{{ entranceUrl }}
					</div>
					<q-icon name="sym_r_content_copy" size="24px" />
				</div>
			</bt-form-item>
		</bt-list>
	</bt-scroll-area>

	<bt-scroll-area
		v-else-if="application && domainInfo"
		class="nav-height-scroll-area-conf"
	>
		<bt-list>
			<bt-form-item :title="t('endpoint')" :dataWidthP="60">
				<div
					class="row justify-end items-center full-width"
					@click="setCopyInfo(entranceUrl)"
				>
					<div
						style="
							max-width: calc(100% - 40px);
							overflow: hidden;
							white-space: nowrap;
							text-overflow: ellipsis;
							margin-right: 10px;
						"
						class="text-body-3 text-ink-3"
					>
						{{ entranceUrl }}
					</div>
					<q-icon name="sym_r_content_copy" size="24px" />
				</div>
			</bt-form-item>
			<bt-form-item
				:title="t('default_route_id')"
				:margin-top="false"
				:data="entranceRef?.id"
				:width-separator="!application.isSysApp"
			/>

			<bt-form-item
				v-if="!application.isSysApp"
				:title="t('setup_custom_route_id')"
			>
				<div class="row justify-between items-center">
					<bt-edit-view
						style="width: calc(100% - 32px)"
						:right="true"
						:is-read-only="true"
						v-model="third_level_domain"
					/>
					<bt-icon
						v-if="tryAddThirdLevelDomain"
						name="delete"
						@click="deleteDomain(false)"
					/>
					<bt-icon v-else name="add" @click="addDomain(false)" />
				</div>
			</bt-form-item>

			<error-message-tip
				v-if="!application.isSysApp"
				:is-error="isCertError"
				:error-message="
					domainInfo?.cname_status === 'cert-invalid'
						? t('Invalid HTTPS certificate. Please check and upload again.')
						: t('No https certificate, please upload')
				"
				:width-separator="tryAddThirdPartyDomain"
				:with-popup="true"
			>
				<bt-form-item
					:title="t('setup_custom_domain')"
					:width-separator="false"
					:dataWidthP="third_party_domain ? 50 : 0"
				>
					<div class="row justify-end items-center full-width">
						<div
							v-if="third_party_domain"
							style="
								max-width: calc(100% - 40px);
								overflow: hidden;
								white-space: nowrap;
								text-overflow: ellipsis;
								margin-right: 10px;
							"
							class="text-body-3 text-ink-3"
						>
							{{ third_party_domain }}
						</div>
						<q-btn
							dense
							v-if="tryAddThirdPartyDomain && isCertError"
							class="row items-center justify-center more-btn"
						>
							<bt-icon name="more_horiz" />
							<q-menu class="popup-menu bg-background-2" :offset="[100, 5]">
								<q-list dense padding style="width: 100px">
									<q-item
										clickable
										v-close-popup
										class="domain-operation-item q-pb-md"
										@click="editDomain(true)"
									>
										<div
											class="domain-operation-div row justify-start items-center text-ink-2"
										>
											<q-icon name="sym_r_edit_square" size="16px" />
											<div class="domain-operation-text text-body3">
												{{ t('edit') }}
											</div>
										</div>
									</q-item>
									<q-item
										clickable
										v-close-popup
										class="domain-operation-item q-pb-md"
										@click="deleteDomain(true)"
									>
										<div
											class="domain-operation-div row justify-start items-center text-ink-2"
										>
											<q-icon name="sym_r_delete" size="16px" />
											<div class="domain-operation-text text-body3">
												{{ t('remove') }}
											</div>
										</div>
									</q-item>
								</q-list>
							</q-menu>
						</q-btn>

						<bt-icon
							v-else-if="tryAddThirdPartyDomain"
							name="delete"
							@click="deleteDomain(true)"
						/>
						<bt-icon v-else name="add" @click="addDomain(true)" />
					</div>
				</bt-form-item>
			</error-message-tip>

			<bt-form-item
				v-if="
					show_third_domain_cname_setButton &&
					tryAddThirdPartyDomain &&
					!application.isSysApp
				"
				:title="t('activation_third_party_domain')"
				:width-separator="false"
			>
				<q-btn
					dense
					class="active-btn q-px-md submit-btn-margin"
					:label="t('activation')"
					@click="setCName"
					color="primary"
				/>
			</bt-form-item>

			<bt-form-item
				v-if="
					!show_third_domain_cname_setButton &&
					tryAddThirdPartyDomain &&
					!application.isSysApp
				"
				:title="t('status')"
				:width-separator="false"
			>
				<div class="row items-center">
					<div
						:class="
							domainInfo?.cname_status === 'active'
								? 'status-activated'
								: 'status-pending'
						"
					/>
					<div class="form-item-data">
						{{ third_domain_status }}
					</div>
				</div>
			</bt-form-item>
		</bt-list>
	</bt-scroll-area>
</template>

<script lang="ts" setup>
import { ref, onMounted, computed, watch, onUnmounted } from 'vue';
import { useRoute } from 'vue-router';
import { useQuasar } from 'quasar';
import { useApplicationStore } from 'src/stores/settings/application';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import AddDomain from 'src/components/settings/application/dialog/domain/AddDomain.vue';
import AddThirdPartDomain from 'src/components/settings/application/dialog/domain/AddThirdPartDomain.vue';
import ActivationDomain from 'src/components/settings/application/dialog/domain/ActivationDomain.vue';
import { SetupDomain } from 'src/constant/global';
import BtIcon from 'src/components/settings/base/BtIcon.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import BtEditView from 'src/components/settings/base/BtEditView.vue';
import ErrorMessageTip from 'src/components/settings/base/ErrorMessageTip.vue';
import { ReverseProxyMode } from 'src/constant/index';
import { useI18n } from 'vue-i18n';
import { useAdminStore } from 'src/stores/settings/admin';
import { useNetworkStore } from 'src/stores/settings/network';
import {
	notifySuccess,
	notifyWarning
} from '../../../../utils/settings/btNotify';
import { getApplication } from 'src/application/base';

const { t } = useI18n();

const applicationStore = useApplicationStore();
const Route = useRoute();
const $q = useQuasar();
const adminStore = useAdminStore();
const networkStore = useNetworkStore();

const application = ref(
	applicationStore.getApplicationById(Route.params.name as string)
);

const application_name = ref(Route.params.name as string);
const entrance_name = Route.params.entrance as string;

const isShared = ref((Route.params.shared as string) != '');

const tryAddThirdLevelDomain = ref<boolean>(false);
const tryAddThirdPartyDomain = ref<boolean>(false);

const domainInfo = ref<SetupDomain | undefined>();
const third_level_domain = ref<string>('');
const third_party_domain = ref<string>('');
const third_party_cert = ref<string>('');
const third_party_key = ref<string>('');

function clearValue() {
	third_party_domain.value = '';
	third_level_domain.value = '';
	tryAddThirdLevelDomain.value = false;
	tryAddThirdPartyDomain.value = false;
}

const addDomain = (isParty: boolean) => {
	if (!entranceRef.value) {
		return;
	}
	console.log('entranceRef.v', entranceRef.value);

	// if (entranceRef.value.authLevel == )

	$q.dialog({
		component: isParty ? AddThirdPartDomain : AddDomain,
		componentProps: {
			isParty,
			reverseProxyMode: reverseProxyMode.value
		}
	}).onOk((data: { data: string; cert?: string; key?: string }) => {
		if (isParty) {
			tryAddThirdPartyDomain.value = third_party_domain.value !== data.data;
			if (tryAddThirdPartyDomain.value) {
				third_party_domain.value = data.data;
				third_party_cert.value = data.cert || '';
				third_party_key.value = data.key || '';
				onSubmit();
			}
		} else {
			tryAddThirdLevelDomain.value = third_level_domain.value !== data.data;
			if (tryAddThirdLevelDomain.value) {
				third_level_domain.value = data.data;
				onSubmit();
			}
		}
	});
};

const deleteDomain = (isParty: boolean) => {
	if (isParty) {
		third_party_domain.value = '';
		tryAddThirdPartyDomain.value = false;
	} else {
		third_level_domain.value = '';
		tryAddThirdLevelDomain.value = false;
	}
	onSubmit();
};

async function updateDomain() {
	domainInfo.value = await applicationStore.getDomainSetup(
		application_name.value,
		entrance_name
	);

	clearValue();
	if (domainInfo.value?.third_party_domain) {
		third_party_domain.value = domainInfo.value!.third_party_domain;
		tryAddThirdPartyDomain.value = true;
	}

	if (domainInfo.value?.third_level_domain) {
		third_level_domain.value = domainInfo.value.third_level_domain;
		tryAddThirdLevelDomain.value = true;
	}
	startRefreshCNameStatus();
}

onMounted(async () => {
	application.value = applicationStore.getApplicationById(
		Route.params.name as string
	);
	if (!isShared.value) {
		configNetworkData();
		networkStore.configReverseProxy();
		await updateDomain();
	}
});

onUnmounted(() => {
	clearRefreshCNameTimer();
});

const show_third_domain_cname_setButton = computed(() => {
	return domainInfo.value?.cname_target_status == 'unset';
});

const third_domain_status = computed(() => {
	if (!domainInfo.value?.cname_target) {
		return '';
	} else {
		if (domainInfo.value?.cname_target_status == 'unset') {
			return t('please_add_a_dns_record_record_cname_to_cname', {
				record: third_party_domain.value,
				cname: domainInfo.value?.cname_target
			});
		} else if (domainInfo.value?.cname_target_status == 'set') {
			if (domainInfo.value?.cname_status == 'unset') {
				return t('checking_cname_records_usually_takes_1_hour');
			} else if (domainInfo.value?.cname_status == 'pending') {
				return t('wait_for_cname_to_be_activated');
			} else if (domainInfo.value?.cname_status == 'active') {
				return t('activated');
			}
		}
		return (
			domainInfo.value?.cname_target_status +
			' / ' +
			domainInfo.value?.cname_status
		);
	}
});

async function setCName() {
	$q.dialog({
		component: ActivationDomain,
		componentProps: {
			thirdPartyDomain: third_party_domain.value,
			cnameTarget: domainInfo.value?.cname_target
		}
	}).onOk(async () => {
		try {
			await applicationStore.setupCName(application_name.value, entrance_name);
		} catch (e: any) {
			notifyWarning(e.message || t('errors.network_error'));
		}

		await updateDomain();
		await applicationStore.getEntrances(application_name.value);
	});
}

async function onSubmit() {
	let obj = domainInfo.value || {
		third_level_domain: '',
		third_party_domain: '',
		key: '',
		cert: ''
	};

	if (tryAddThirdLevelDomain.value) {
		if (!third_level_domain.value) {
			notifyWarning(t('errors.third_level_domain_is_empty'));
			return;
		}

		obj.third_level_domain = third_level_domain.value;
	} else {
		obj.third_level_domain = '';
	}

	if (tryAddThirdPartyDomain.value) {
		if (!third_party_domain.value) {
			notifyWarning(t('errors.third_party_domain_is_empty'));
			return;
		}

		obj.third_party_domain = third_party_domain.value;
		obj.cert = third_party_cert.value;
		obj.key = third_party_key.value;
	} else {
		obj.third_party_domain = '';
		obj.cert = '';
		obj.key = '';
	}

	try {
		await applicationStore.setupDomain(
			application_name.value,
			entrance_name,
			obj
		);
	} catch (e: any) {
		// notifyWarning(e.message || t('errors.network_error'));
	}

	await updateDomain();
	await applicationStore.getEntrances(application_name.value);
}

const entranceRef = ref(
	application.value?.entrances?.find((e) => e.name == entrance_name)
);

const sharedEntranceRef = ref(
	application.value?.sharedEntrances?.find((e) => e.name == entrance_name)
);

const entranceUrl = computed(() => {
	if (isShared.value) {
		if (!sharedEntranceRef.value) {
			return '';
		}
		return 'http://' + sharedEntranceRef.value.url;
	}
	if (!entranceRef.value) {
		return '';
	}
	return (
		'https://' +
		entranceRef.value.id +
		'.' +
		adminStore.olaresId.replace('@', '.')
	);
});

const setCopyInfo = (info: string) => {
	getApplication()
		.copyToClipboard(info)
		.then(() => {
			notifySuccess(t('copy_successfully'));
		});
};

watch(
	() => networkStore.reverseProxy,
	() => {
		configNetworkData();
	}
);

const reverseProxyMode = ref(ReverseProxyMode.NoNeed);

const configNetworkData = () => {
	if (networkStore.reverseProxy) {
		reverseProxyMode.value = networkStore.reverseProxy.enable_cloudflare_tunnel
			? ReverseProxyMode.CloudFlare
			: networkStore.reverseProxy.enable_frp
			? networkStore.reverseProxy.frp_auth_method == 'jws'
				? ReverseProxyMode.OlaresTunnel
				: ReverseProxyMode.SelfBuiltFrp
			: ReverseProxyMode.NoNeed;
	}
};

const refreshCNameTimer = ref();

const startRefreshCNameStatus = () => {
	if (refreshCNameTimer.value) {
		return;
	}
	if (
		!domainInfo.value ||
		(domainInfo.value.cname_status !== 'pending' &&
			(!isCertError.value || !certEditCompleted.value))
	) {
		return;
	}

	refreshCNameTimer.value = setInterval(async () => {
		if (
			domainInfo.value?.cname_status == 'pending' ||
			(isCertError.value && certEditCompleted.value)
		) {
			domainInfo.value = await applicationStore.getDomainSetup(
				application_name.value,
				entrance_name
			);
			if (!isCertError.value) {
				certEditCompleted.value = false;
			}
			return;
		}
		clearRefreshCNameTimer();
	}, 5000);
};

const clearRefreshCNameTimer = () => {
	if (refreshCNameTimer.value) {
		clearInterval(refreshCNameTimer.value);
		refreshCNameTimer.value = undefined;
	}
};

const isCertError = computed(() => {
	return (
		domainInfo.value?.cname_status === 'cert-invalid' ||
		domainInfo.value?.cname_status === 'cert-not-found'
	);
});

const certEditCompleted = ref(false);
const editDomain = (isParty: boolean) => {
	$q.dialog({
		component: isParty ? AddThirdPartDomain : AddDomain,
		componentProps: {
			isParty,
			reverseProxyMode: reverseProxyMode.value,
			domain: isParty ? third_party_domain.value : ''
		}
	}).onOk(async (data: { data: string; cert?: string; key?: string }) => {
		if (isParty) {
			tryAddThirdPartyDomain.value = true;
			third_party_domain.value = data.data;
			third_party_cert.value = data.cert || '';
			third_party_key.value = data.key || '';
			certEditCompleted.value = true;
			await onSubmit();
		}
	});
};
</script>

<style scoped lang="scss">
.status-pending {
	width: 8px;
	height: 8px;
	background-color: $warning;
	border-radius: 50%;
	margin-right: 6px;
}

.status-activated {
	width: 8px;
	height: 8px;
	background-color: $green;
	border-radius: 50%;
	margin-right: 6px;
}

.active-btn {
	font-size: map-get($map: $body3, $key: size) !important;
	font-weight: map-get($map: $body3, $key: weight) !important;
	line-height: map-get($map: $body3, $key: line-height) !important;
	letter-spacing: map-get($map: $body3, $key: letter-spacing) !important;
	text-align: center !important;
	border-radius: 8px !important;
	text-transform: capitalize !important;
	background-color: $blue-default !important;
	color: #fff !important;
	height: 32px;
}

.more-btn {
	width: 20px;
	height: 20px;
	min-height: 20px !important;
	border-radius: 4px !important;
}

.domain-operation-item {
	// width: 135px;
	height: 32px;
	margin: 0;
	padding: 0;
	border-radius: 4px;

	.domain-operation-div {
		width: 100%;
		height: 100%;
		padding: 8px;

		.domain-operation-text {
			margin-left: 8px;
		}
	}
}
</style>

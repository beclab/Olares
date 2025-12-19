<template>
	<page-title-component :show-back="true" :title="t('Reverse Proxy')" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<bt-list first v-if="networkStore.reverseProxy">
			<bt-form-item
				:title="t('Set up a reverse proxy for external access')"
				:margin-top="false"
				:chevron-right="false"
				:widthSeparator="
					reverseProxyMode != ReverseProxyMode.CloudFlare &&
					reverseProxyMode != ReverseProxyMode.NoNeed
				"
			>
				<bt-select
					v-if="reverseProxyMode != ReverseProxyMode.NoNeed"
					v-model="reverseProxyMode"
					:options="reverseProxyModelOptions"
				/>
				<div v-else>No need (IP Direct)</div>
			</bt-form-item>

			<bt-form-item
				v-if="reverseProxyMode == ReverseProxyMode.OlaresTunnel"
				:title="t('Select an Olares tunnel')"
				:margin-top="false"
				:chevron-right="false"
				:widthSeparator="false"
			>
				<bt-select v-model="olaresTunnelMode" :options="olaresTunnelsOptions" />
			</bt-form-item>

			<error-message-tip
				v-if="reverseProxyMode == ReverseProxyMode.SelfBuiltFrp"
				:with-popup="true"
				:width-separator="true"
			>
				<bt-form-item :title="t('Server Address')" :width-separator="false">
					<bt-edit-view
						style="width: 200px"
						v-model="serverAddress"
						:right="true"
						:placeholder="t('Please input server address')"
					/>
				</bt-form-item>
			</error-message-tip>

			<error-message-tip
				v-if="reverseProxyMode == ReverseProxyMode.SelfBuiltFrp"
				:with-popup="true"
				:width-separator="true"
			>
				<bt-form-item :title="t('Port')" :width-separator="false">
					<bt-edit-view
						style="width: 200px"
						v-model="port"
						:right="true"
						:placeholder="t('Please input port')"
					/>
				</bt-form-item>
			</error-message-tip>

			<error-message-tip
				v-if="reverseProxyMode == ReverseProxyMode.SelfBuiltFrp"
				:with-popup="true"
				:width-separator="authMethod == 'token'"
			>
				<bt-form-item :title="t('Auth Method')" :width-separator="false">
					<bt-select v-model="authMethod" :options="frpAuthMethod()" />
				</bt-form-item>
			</error-message-tip>

			<error-message-tip
				v-if="
					reverseProxyMode == ReverseProxyMode.SelfBuiltFrp &&
					authMethod == 'token'
				"
				:with-popup="true"
				:width-separator="false"
			>
				<bt-form-item :title="t('Token')" :width-separator="false">
					<bt-edit-view
						style="width: 200px"
						v-model="token"
						:right="true"
						:placeholder="t('Please input token')"
					/>
				</bt-form-item>
			</error-message-tip>
		</bt-list>
		<div
			class="row justify-end"
			v-if="reverseProxyMode != ReverseProxyMode.NoNeed"
		>
			<q-btn
				dense
				flat
				class="confirm-btn q-px-md"
				style="margin-top: 20px"
				:label="t('save')"
				@click="onSubmit"
			/>
		</div>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import AffectedDomainDialog from '../../../components/settings/network/dialog/AffectedDomainDialog.vue';
import PageTitleComponent from '../../../components/settings/PageTitleComponent.vue';
import ErrorMessageTip from '../../../components/settings/base/ErrorMessageTip.vue';
import BtEditView from '../../../components/settings/base/BtEditView.vue';
import BtFormItem from '../../../components/settings/base/BtFormItem.vue';
import BtSelect from '../../../components/settings/base/BtSelect.vue';
import { useApplicationStore } from '../../../stores/settings/application';
import { useNetworkStore } from '../../../stores/settings/network';
import BtList from 'src/components/settings/base/BtList.vue';
import { ApplicationCustonDomain } from 'src/constant/global';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import {
	frpAuthMethod,
	olaresTunnelDefaultValue,
	ReverseProxyMode,
	reverseProxyOptions
} from '../../../constant';

const { t } = useI18n();
const $q = useQuasar();
const networkStore = useNetworkStore();
const applicationStore = useApplicationStore();

const reverseProxyMode = ref(ReverseProxyMode.NoNeed);
const olaresTunnelMode = ref('');
const serverAddress = ref('');
const port = ref('');
const authMethod = ref('');
const token = ref('');

let reverseProxyModelOptions = ref(reverseProxyOptions());

watch(
	() => networkStore.reverseProxy,
	() => {
		configData();
	}
);

const configData = () => {
	if (networkStore.reverseProxy) {
		reverseProxyMode.value = networkStore.reverseProxy.enable_cloudflare_tunnel
			? ReverseProxyMode.CloudFlare
			: networkStore.reverseProxy.enable_frp
			? networkStore.reverseProxy.frp_auth_method == 'jws'
				? ReverseProxyMode.OlaresTunnel
				: ReverseProxyMode.SelfBuiltFrp
			: ReverseProxyMode.NoNeed;

		if (
			reverseProxyMode.value == ReverseProxyMode.OlaresTunnel ||
			reverseProxyMode.value == ReverseProxyMode.SelfBuiltFrp
		) {
			reverseProxyModelOptions.value = reverseProxyModelOptions.value.filter(
				(e) => e.value != ReverseProxyMode.CloudFlare
			);
		}

		const olaresOptions = networkStore.olaresTunnelsV2Options();
		olaresTunnelsOptions.value = olaresOptions;

		const option = olaresOptions.find(
			(e) => e.value == networkStore.reverseProxy?.frp_server
		);

		if (option) {
			olaresTunnelMode.value = option.value;
		} else if (olaresOptions && olaresOptions.length > 0) {
			olaresTunnelMode.value = olaresOptions[0].value;
		}

		if (reverseProxyMode.value != ReverseProxyMode.OlaresTunnel) {
			serverAddress.value = networkStore.reverseProxy.frp_server;
			port.value = `${
				networkStore.reverseProxy.frp_port == 0
					? ''
					: networkStore.reverseProxy.frp_port
			}`;
			authMethod.value = networkStore.reverseProxy.frp_auth_method;
			token.value = networkStore.reverseProxy.frp_auth_token;
		}
	}
};

const olaresTunnelsOptions = ref<any>();

onMounted(async () => {
	configData();
	await networkStore.configReverseProxy();
	await applicationStore.getEntranceSetupDomain();
});

const onSubmit = async () => {
	if (!networkStore.reverseProxy) {
		return;
	}

	if (reverseProxyMode.value == ReverseProxyMode.SelfBuiltFrp) {
		if (serverAddress.value.length == 0 || port.value.length == 0) {
			return;
		}
		if (authMethod.value == 'token' && token.value.length == 0) {
			return;
		}
	}
	let reminderMessage = t(
		'During the reverse proxy switch, Olares may be inaccessible for 10 minutes.'
	);
	let affectedDomains = [] as ApplicationCustonDomain[];
	if (
		networkStore.reverseProxy.enable_cloudflare_tunnel &&
		(reverseProxyMode.value == ReverseProxyMode.OlaresTunnel ||
			reverseProxyMode.value == ReverseProxyMode.SelfBuiltFrp)
	) {
		reminderMessage = t(
			'After switching to FRP, the custom domain will no longer be valid. To restore functionality, you need to upload an HTTPS certificate on the Applications > Entrances page. The switch may take up to 10 minutes to complete, during which Olares may be inaccessible.'
		);
		affectedDomains = applicationStore.customDomainApplications;
	} else if (reverseProxyMode.value == ReverseProxyMode.CloudFlare) {
		reminderMessage = t(
			"After switching to Cloudflare, the custom domain will be updated and resolved to work with Cloudflare's network. The switch may take up to 10 minutes to complete, during which Olares may be inaccessible."
		);
	}

	$q.dialog({
		component: AffectedDomainDialog,
		componentProps: {
			title: t('Switch reverse proxy'),
			message: reminderMessage,
			useCancel: true,
			confirmText: t('confirm'),
			cancelText: t('cancel'),
			affectedDomains
		}
	}).onOk(async () => {
		confirmSwitch();
	});
	// console.log(
	// 	'networkStore.reverseProxy.enable_frp ===>',
	// 	networkStore.reverseProxy.enable_frp
	// );
	// console.log(reverseProxyMode.value);
};

const confirmSwitch = async () => {
	if (!networkStore.reverseProxy) {
		return;
	}
	if (
		reverseProxyMode.value == ReverseProxyMode.OlaresTunnel ||
		reverseProxyMode.value == ReverseProxyMode.SelfBuiltFrp
	) {
		networkStore.reverseProxy.enable_frp = true;
		networkStore.reverseProxy.enable_cloudflare_tunnel = false;
	} else if (reverseProxyMode.value == ReverseProxyMode.CloudFlare) {
		networkStore.reverseProxy.enable_frp = false;
		networkStore.reverseProxy.enable_cloudflare_tunnel = true;
	}

	if (reverseProxyMode.value == ReverseProxyMode.OlaresTunnel) {
		Object.assign(networkStore.reverseProxy, olaresTunnelDefaultValue);
		networkStore.reverseProxy.frp_server = olaresTunnelMode.value;
	} else if (reverseProxyMode.value == ReverseProxyMode.SelfBuiltFrp) {
		let server = serverAddress.value;
		if (
			serverAddress.value.startsWith('http://') ||
			serverAddress.value.startsWith('https://')
		) {
			server = server.split('//')[1];
		}
		networkStore.reverseProxy.frp_server = server;
		networkStore.reverseProxy.frp_port = Number(port.value);
		networkStore.reverseProxy.frp_auth_method = authMethod.value;
		networkStore.reverseProxy.frp_auth_token = token.value;
	}
	$q.loading.show();
	await networkStore
		.updateReverseProxy(networkStore.reverseProxy)
		.then(() => {
			BtNotify.show({
				type: NotifyDefinedType.SUCCESS,
				message: t('success')
			});
		})
		.catch((err) => {
			console.error(err);
			BtNotify.show({
				type: NotifyDefinedType.FAILED,
				message: err?.response?.data || t('failed')
			});
		});
	$q.loading.hide();
};

onMounted(() => {});
</script>

<style scoped lang="scss"></style>

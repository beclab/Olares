<template>
	<page-title-component :show-back="true" :title="t('olares_space')">
	</page-title-component>

	<bt-scroll-area class="nav-height-scroll-area-conf">
		<div class="mode-title-class q-mt-lg">{{ t('binding_info') }}</div>

		<bt-list v-if="load">
			<q-item style="padding: 0" class="item-margin-left q-my-sm">
				<div class="terminus-info row items-center justify-between">
					<div class="row items-center">
						<setting-avatar :size="40" style="margin-left: 8px" />
						<div class="column justify-between" style="margin-left: 8px">
							<div class="row items-center">
								<div class="text-subtitle2 person-name q-mr-sm">
									{{ adminStore.user.name }}
								</div>
								<account-connect-status v-if="!bindTerminusName" />
							</div>
							<div class="text-body3 person-id">
								{{ '@' + adminStore.olaresId.split('@')[1] }}
							</div>
						</div>
					</div>
					<div
						class="row items-center item-margin-right"
						v-if="!bindTerminusName"
					>
						<q-btn
							dense
							class="confirm-btn bound-btn"
							:label="t('bound')"
							flat
							@click="bindName"
						/>
					</div>
					<div
						v-else
						class="row items-center justify-end item-margin-right"
						@click="unbindName(bindTerminusName.id)"
					>
						<div class="unbound-bg row items-center justify-center">
							<q-icon name="sym_r_link_off" size="16px" color="ink-1" />
						</div>
					</div>
				</div>
			</q-item>
			<bt-separator />
			<q-item
				class="text-body3 reminder-info row items-center item-margin-left item-margin-right"
				v-if="!bindTerminusName"
			>
				{{
					t(
						'only_after_binding_the_olares_id_can_you_continue_to_bind_avatar_message'
					)
				}}
			</q-item>
			<q-item
				v-else
				class="metamask-info q-my-md item-margin-left item-margin-right"
			>
				<div class="metamask-info__content row items-center justify-between">
					<div
						class="row items-center"
						:style="`width:calc(100% - ${
							blockChainStore.account && blockChainStoreAddressAuthed
								? 120
								: 120
						}px);`"
					>
						<q-img
							src="account/metamask.png"
							:noSpinner="true"
							width="40px"
							height="40px"
						/>
						<div
							class="column justify-between"
							style="margin-left: 8px; width: calc(100% - 52px)"
						>
							<div class="row items-center">
								<div class="text-subtitle2 name q-mr-sm">
									{{
										blockChainStore.account
											? t('username_is_metaMask_wallet', {
													username: adminStore.user.name
											  })
											: 'MetaMask'
									}}
								</div>
								<account-connect-status
									v-if="
										blockChainStore.account && !blockChainStoreAddressAuthed
									"
								/>
							</div>
							<div class="text-body3 text">
								{{
									blockChainStore.account
										? blockChainStore.account
										: t(
												'you_need_to_bind_a_blockchain_wallet_to_obtain_the_corresponding_nft_image_message'
										  )
								}}
							</div>
						</div>
					</div>
					<div class="row items-center" v-if="!blockChainStore.account">
						<q-btn
							dense
							class="confirm-btn bound-btn"
							:label="t('import')"
							flat
							@click="connectWallet"
						/>
					</div>
					<div
						class="row items-center"
						v-else-if="!blockChainStoreAddressAuthed"
					>
						<q-btn
							dense
							class="confirm-btn bound-btn"
							:label="t('bound')"
							flat
							@click="onAddSign"
						/>
					</div>
					<div v-else class="row items-center justify-end">
						<q-btn
							dense
							class="confirm-btn bound-btn q-mr-sm"
							:label="t('nft_images')"
							flat
							@click="gotoNFTPage(blockChainStore.account)"
						/>
						<div
							class="unbound-bg row items-center justify-center"
							@click="onRemoveSign"
						>
							<q-icon name="sym_r_link_off" size="16px" color="ink-1" />
						</div>
					</div>
				</div>
			</q-item>
			<bt-separator
				class="q-mt-sm"
				v-if="
					bindTerminusName &&
					blockChainStore.account &&
					!blockChainStoreAddressAuthed
				"
			/>
			<q-item
				class="text-body3 reminder-info q-my-md row items-center item-margin-left item-margin-right"
				v-if="
					bindTerminusName &&
					blockChainStore.account &&
					!blockChainStoreAddressAuthed
				"
			>
				<div
					v-html="
						t(
							'you_need_to_bind_a_blockchain_wallet_to_obtain_the_corresponding_nft_image_message_more'
						)
					"
				></div>
			</q-item>
		</bt-list>
		<div
			v-else
			class="row items-center justify-center"
			style="width: 100%; height: 200px"
		>
			<q-spinner-dots color="primary" size="2em" />
		</div>
	</bt-scroll-area>
</template>

<script lang="ts" setup>
import { useQuasar } from 'quasar';
import { ref, watch, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useAdminStore } from 'src/stores/settings/admin';
import { useSpaceStore, BindTerminusName } from 'src/stores/settings/space';
import { useBlockchainStore } from 'src/stores/settings/blockchain';
import { toUtf8Bytes, keccak256, ethers } from 'ethers';
import ResolverABI from 'src/core/RootResolverABI';
import baseinfo from 'src/core/baseinfo';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import SettingAvatar from 'src/components/settings/base/SettingAvatar.vue';
import AccountConnectStatus from 'src/components/settings/account/AccountConnectStatus.vue';
import AccountRequestSignDialog from 'src/components/settings/account/AccountRequestSignDialog.vue';
import { notifyFailed, notifyWarning } from 'src/utils/settings/btNotify';
import BtSeparator from 'src/components/settings/base/BtSeparator.vue';
import BtList from 'src/components/settings/base/BtList.vue';

import { useI18n } from 'vue-i18n';
const { t } = useI18n();

const $q = useQuasar();
const blockChainStore = useBlockchainStore();
const adminStore = useAdminStore();
const router = useRouter();
const spaceStore = useSpaceStore();
const bindTerminusName = ref<BindTerminusName | undefined>(undefined);

interface AuthenticatedAddress {
	index: number;
	address: string;
}

const load = ref(false);
type JsonRpcProvider = ethers.JsonRpcProvider;
let authenticatedAddress = ref<AuthenticatedAddress[]>([]);
const ethLoading = ref(false);

const blockChainStoreAddressAuthed = ref(false);

// const account_type = 'space';
onMounted(async () => {
	// name.value = 'integration-account:space';
	// const res: any = accountStore.secrets.find((item) =>
	// 	item.name.startsWith(name.value)
	// );
	// if (res) {
	// 	secert.value = JSON.parse(res.value);
	// } else {
	// 	router.push({ path: '/integration/login/' + account_type });
	// 	return;
	// }
	await setBindTerminusName(await spaceStore.getTerminsNames());

	load.value = true;
});

async function setBindTerminusName(name: BindTerminusName | undefined) {
	bindTerminusName.value = name;
	if (name) {
		if (typeof window.ethereum != 'undefined') {
			if (window.ethereum.chainId && window.ethereum.selectedAddress) {
				await connectWallet();
			}
		}
	}
}

async function bindName() {
	try {
		await spaceStore.requestBindTerminsNames();
		$q.dialog({
			component: AccountRequestSignDialog,
			componentProps: {
				title: t('bound_olares_id'),
				info: t(
					'Please_authorize_and_verify_on_the_larePass_mobile_phone_to_bind_your_olares_id'
				),
				imagePath: 'account/bind-terminus-space.png'
			}
		}).onDismiss(async () => {
			setBindTerminusName(await spaceStore.getTerminsNames());
		});
	} catch (error: any) {
		console.log(error.message);
	}
}

async function unbindName(id: number) {
	$q.loading.show();
	try {
		await spaceStore.unBindTerminsNames(id);
		setBindTerminusName(undefined);
		$q.loading.hide();
	} catch (error: any) {
		console.log(error.message);
		$q.loading.hide();
	}
}

async function gotoNFTPage(address: string) {
	router.push({ path: '/integration/detail/space/' + address });
}

function getProvider(): JsonRpcProvider {
	const provider = new ethers.JsonRpcProvider(process.env.NODE_RPC);
	return provider;
}

const connectWallet = async () => {
	if (typeof window.ethereum === 'undefined') {
		$q.dialog({
			component: AccountRequestSignDialog,
			componentProps: {
				title: t('import_wallet'),
				info: t('import_wallet_dialog_message')
			}
		});
		return;
	}
	await blockChainStore.unlockWallet(true);
	await updateAddress();
};
watch(
	() => blockChainStore.account,
	() => {
		if (blockChainStore.account) {
			updateAddress();
		}
	}
);

function getContractRootResolver(): ethers.Contract {
	const contract = new ethers.Contract(
		process.env.CONTRACT_ROOT_RESOLVER
			? process.env.CONTRACT_ROOT_RESOLVER
			: '',
		ResolverABI.abi,
		getProvider()
	);
	return contract;
}

async function updateAddress() {
	ethLoading.value = false;
	authenticatedAddress.value = [];

	try {
		const contract = getContractRootResolver();
		const ad = [
			...(await contract.getAuthenticationAddresses(
				adminStore.olaresId.replace('@', '.')
			))
		];

		for (var i = 0; i < ad.length; i++) {
			authenticatedAddress.value.push({
				index: i,
				address: ad[i][1].toString()
			});
		}
	} catch (error) {
		/* empty */
	} finally {
		ethLoading.value = true;
	}

	if (authenticatedAddress.value.length > 0 && blockChainStore.account) {
		const items = authenticatedAddress.value.filter(
			(item: AuthenticatedAddress) =>
				item.address.toLowerCase() == blockChainStore.account!.toLowerCase()
		);
		blockChainStoreAddressAuthed.value = items.length > 0;
	} else {
		blockChainStoreAddressAuthed.value = false;
	}
}

const onAddSign = async () => {
	if (!blockChainStore.account || !blockChainStore.network) {
		notifyWarning(t('need_connect_first'));
		return;
	}

	const items = authenticatedAddress.value.filter(
		(item: AuthenticatedAddress) =>
			item.address.toLowerCase() == blockChainStore.account!.toLowerCase()
	);
	if (items.length > 0) {
		notifyFailed(t('this_address_has_been_authenticated'));
		return;
	}

	try {
		$q.loading.show();
		let signData = await baseinfo.onAddAuthenticate(
			blockChainStore.network,
			blockChainStore.account,
			adminStore.olaresId
		);
		const uuid = keccak256(
			toUtf8Bytes(
				JSON.stringify(Object.assign({ t: new Date().getTime() }, signData))
			)
		).substring(0, 34);

		await spaceStore.requestTermiPassSignEthAddress(
			signData.signed,
			signData.data,
			uuid
		);
		$q.dialog({
			component: AccountRequestSignDialog,
			componentProps: {
				title: t('bind_blockchain_wallet'),
				info: t('bind_blockchain_wallet_dialog_message'),
				imagePath: 'account/bind-terminus-space.png'
			}
		}).onDismiss(async () => {
			await updateAddress();
		});
	} catch (e: any) {
		console.log(e.message);
		notifyFailed(t('bound_faild'));
	} finally {
		$q.loading.hide();
	}
};

const onRemoveSign = async () => {
	if (!blockChainStore.account || !blockChainStore.network) {
		notifyWarning(t('need_connect_first'));
		return;
	}

	try {
		$q.loading.show();
		let signData = await baseinfo.onRemoveAuthenticate(
			blockChainStore.network,
			blockChainStore.account,
			adminStore.olaresId
		);

		const uuid = keccak256(
			toUtf8Bytes(
				JSON.stringify(Object.assign({ t: new Date().getTime() }, signData))
			)
		).substring(0, 34);

		const index = authenticatedAddress.value.filter(
			(item: AuthenticatedAddress) =>
				item.address.toLowerCase() == blockChainStore.account!.toLowerCase()
		)[0].index;
		await spaceStore.requestTermiPassSignRemoveEthAddress(
			signData.signed,
			signData.data,
			uuid,
			index
		);

		$q.dialog({
			component: AccountRequestSignDialog,
			componentProps: {
				title: t('unbind_blockchain_wallet'),
				info: t('unbind_blockchain_wallet_dialog_message'),
				imagePath: 'account/bind-terminus-space.png'
			}
		}).onDismiss(async () => {
			await updateAddress();
		});
	} catch (e: any) {
		notifyFailed(t('unbound_faild'));
	} finally {
		$q.loading.hide();
	}
};
</script>
<style lang="scss" scoped>
.terminus-info {
	width: 100%;

	.person-name {
		color: $ink-1;
	}

	.person-id {
		color: $ink-2;
	}
}

.bound-btn {
	height: 24px;
	padding: 4px 8px;
}

.reminder-info {
	color: $ink-2;
	padding: 0;
}

.metamask-info {
	padding: 0;
	width: calc(100% - 40px);

	&__content {
		width: 100%;

		.name {
			color: $ink-1;
		}

		.text {
			color: $ink-2;
		}
	}
}

.unbound-bg {
	width: 24px;
	height: 24px;
	border-radius: 8px;
	background: $background-3;
}
</style>

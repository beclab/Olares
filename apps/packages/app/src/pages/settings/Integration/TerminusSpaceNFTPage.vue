<template>
	<page-title-component :show-back="true" :title="t('nft_images')" />
	<bt-scroll-area
		class="nav-height-scroll-area-conf nft-header-height"
		v-if="nft_loaded"
	>
		<div class="row q-gutter-md" v-if="nftImageSet.length > 0">
			<template v-for="item in nftImageSet" :key="item">
				<bio-avatar-selector
					:src="item.image"
					:selected="
						current_nft?.token_id === item.token_id &&
						current_nft.image == item.image &&
						current_nft.token_address == item.token_address
					"
					@click="onItemClick(item)"
					:bound="
						BoundNFTAddressList.find(
							(e) =>
								e.tokenID == item.token_id &&
								e.chainType == item.chain &&
								e.tokenAddress == item.contract_id
						)
					"
				/>
				<!-- <div>sasdsd</div>
				<div>{{ item.contract_id }}</div> -->
			</template>
		</div>
		<div v-else class="column items-center" style="width: 100%; height: 100%">
			<q-img src="settings/backup_empty.svg" class="empty-image" />
			<div class="empty-text">{{ t('no_nft_available') }}</div>
		</div>
	</bt-scroll-area>
	<div v-else class="row items-center justify-center nft-header-height">
		<q-spinner-hourglass color="primary" size="2em" />
	</div>
	<div class="bottom-part">
		<bt-separator />
		<div class="bottom-part__content row items-center justify-between">
			<div
				class="text-body3 text-ink-1 row items-center justify-center reminder"
			>
				{{ t('select_the_image_to_complete_binding_nft_message') }}
			</div>
			<q-btn
				dense
				class="confirm-btn bound-btn q-mr-md"
				:disable="current_nft == undefined"
				:label="t('bound')"
				flat
				@click="bindNFT"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { NFTItem, ProfileNFT } from 'src/utils/settings/nft';
import BioAvatarSelector from 'src/components/settings/BioAvatarSelector.vue';
import { useSpaceStore, BoundNFTAddressModel } from 'src/stores/settings/space';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import AccountRequestSignDialog from 'src/components/settings/account/AccountRequestSignDialog.vue';
import BtSeparator from 'src/components/settings/base/BtSeparator.vue';
import { notifyFailed } from 'src/utils/settings/btNotify';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';
import { ref, onMounted } from 'vue';
import { useQuasar } from 'quasar';
import axios from 'axios';

const { t } = useI18n();

const spaceStore = useSpaceStore();

const route = useRoute();

const $q = useQuasar();
const nft_loaded = ref(false);
const nftImageSet = ref<ProfileNFT[]>([]);

const address = ref<string>('');
let current_nft = ref<ProfileNFT | undefined>(undefined);

const BoundNFTAddressList = ref<BoundNFTAddressModel[]>([]);

async function updateNFT(address: string) {
	try {
		nft_loaded.value = false;
		const data: any = await axios.post(
			'https://ndbq.ursa-services.bttcdn.com/batch_query_fast_only_img_by_address',
			{ address, limit_in_collection: 10, cache: true }
		);
		if (data.code == 0) {
			const nfts: NFTItem[] = (nftImageSet.value = data.nfts);
			let arr: ProfileNFT[] = [];

			for (const nft of nfts) {
				if (
					nft.token_meta.address == '0x2f0146ca3c62a33177959565ae9df2f86cf01551'
				) {
					continue;
				}
				for (const token of nft.tokenDatas) {
					let image = token.image;
					if (!token.image) {
						continue;
					}
					if (token.image) {
						if (token.image.startsWith('ipfs://')) {
							image = token.image.replace(
								'ipfs://',
								'https://gateway.ipfs.io/ipfs/'
							);
						}
					}
					arr.push({
						chain: nft.chain,
						token_name: nft.token_meta.name,
						token_address: nft.token_meta.address,
						token_id: token.token_id,
						nft_name: token.name,
						image: image,
						contract_id: token.contract_id
					});
				}
			}
			nftImageSet.value = arr;
		}
	} catch (e: any) {
		notifyFailed(e.message);
	} finally {
		nft_loaded.value = true;
	}
}

async function requestNFTAddress() {
	try {
		const addressData = await spaceStore.getNFTAddress();
		if (addressData && addressData.address) {
			BoundNFTAddressList.value = addressData.address;
		}
	} catch (error) {
		/* empty */
	}
}

async function bindNFT() {
	if (!current_nft.value) {
		return;
	}

	const hasBound = BoundNFTAddressList.value.find(
		(e) =>
			e.tokenID == current_nft.value!.token_id &&
			e.chainType == current_nft.value?.chain &&
			e.tokenAddress == current_nft.value.contract_id
	);

	if (hasBound != undefined) {
		notifyFailed(t('had_bound'));
		return;
	}

	$q.loading.show();
	try {
		await spaceStore.requestTermiPassSignNFTBinding(
			current_nft.value.token_id,
			current_nft.value.token_address,
			current_nft.value.token_name,
			current_nft.value.nft_name,
			current_nft.value.image,
			current_nft.value.chain,
			address.value
		);

		$q.dialog({
			component: AccountRequestSignDialog,
			componentProps: {
				title: t('bound_nft_avatar'),
				info: t(
					'please_authorize_and_verify_on_the_larePass_mobile_phone_to_bind_your_nft_avatar'
				),
				imagePath: 'account/bind-nft-avatar.png'
			}
		}).onDismiss(async () => {
			requestNFTAddress();
		});
	} finally {
		$q.loading.hide();
	}
}

onMounted(async () => {
	address.value = route.params.address as string;

	updateNFT(address.value).then(() => {
		if (nftImageSet.value.length == 0) {
			updateNFT('0xb81a0e6c38c3fec8a171cfe9631f60127a0c5bfd');
		}
	});
	requestNFTAddress();
});

const onItemClick = (nft: ProfileNFT) => {
	if (current_nft.value == nft) {
		current_nft.value = undefined;
		return;
	}
	current_nft.value = nft;
};
</script>
<style lang="scss" scoped>
.nft-header-height {
	height: calc(100% - 56px - 81px);
	width: 100%;
	padding-bottom: 0px;

	.empty-image {
		margin-top: 144px;
		width: 160px;
		height: 160px;
	}
}

.bottom-part {
	height: 80px;
	width: 100%;

	&__content {
		width: 100%;
		height: 100%;

		.reminder {
			padding-left: 20px;
			padding-right: 20px;
			width: calc(100% - 80px);
			height: 100%;
		}

		.bound-btn {
			height: 24px;
			padding: 4px 8px;
		}
	}
}
</style>

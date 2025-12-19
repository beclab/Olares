import axios from 'axios';
import { defineStore } from 'pinia';
import { useTokenStore } from './token';

export type BindTerminusName = {
	id: number;
	terminusName: string;
	selfHosted: boolean;
	updateAt: string;
	clusterID: string;
	userid: string;
	createAt: string;
	status: string;
};

export type SpaceState = {
	bindTerminusName: BindTerminusName | undefined;
};

export interface BoundNFTAddressModel {
	terminusName: string;
	nftName: string;
	tokenID: string;
	tokenName: string;
	chainType: string;
	ethAddressID: string;
	updateAt: string;
	userid: string;
	createAt: string;
	tokenAddress: string;
	imageUrl: string;
	terminusNameID: string;
	ethAddress: string;
	did: string;
	status: string;
}

export const useSpaceStore = defineStore('space', {
	state: () => {
		return {
			bindTerminusName: undefined
		} as SpaceState;
	},

	getters: {
		//
	},

	actions: {
		async getTerminsNames(): Promise<BindTerminusName | undefined> {
			const tokenStore = useTokenStore();
			const data: any = await axios.get(
				`${tokenStore.url}/api/cloud/getBindingTerminusNames`
			);
			if (data) {
				this.bindTerminusName = data;
			} else {
				this.bindTerminusName = undefined;
			}

			return this.bindTerminusName;
		},

		async getNFTAddress(): Promise<any> {
			const tokenStore = useTokenStore();
			const data: any = await axios.get(
				`${tokenStore.url}/api/cloud/getNFTAddress`
			);
			return data;
		},

		async requestBindTerminsNames() {
			const tokenStore = useTokenStore();
			const data: any = await axios.post(
				`${tokenStore.url}/api/cloud/sign/requestTerminusName`,
				{}
			);
			return data;
		},

		async unBindTerminsNames(id: number) {
			const tokenStore = useTokenStore();
			const data: any = await axios.post(
				`${tokenStore.url}/api/cloud/sign/unBindTerminusNames`,
				{ id }
			);
			return data;
		},

		async requestTermiPassSignEthAddress(
			sigFromAddressPrivKey: string,
			signData: any,
			uuid: string
		) {
			const tokenStore = useTokenStore();
			const data = axios.post(
				`${tokenStore.url}/api/cloud/sign/requestTermiPassSignEthAddress`,
				{ sigFromAddressPrivKey, signData, uuid }
			);
			return data;
		},

		async requestTermiPassSignRemoveEthAddress(
			sigFromAddressPrivKey: string,
			signData: any,
			uuid: string,
			index: number
		) {
			const tokenStore = useTokenStore();
			const data = axios.post(
				`${tokenStore.url}/api/cloud/sign/requestTermiPassSignRemoveEthAddress`,
				{ sigFromAddressPrivKey, signData, uuid, index }
			);
			return data;
		},

		async requestTermiPassSignNFTBinding(
			token_id: string,
			token_address: string,
			token_name: string,
			nft_name: string,
			image: string,
			chain_type: string,
			owner: string
		) {
			const tokenStore = useTokenStore();
			const data = axios.post(
				`${tokenStore.url}/api/cloud/sign/requestTermiPassSignNFT`,
				{
					token_id,
					token_address,
					token_name,
					nft_name,
					image,
					chain_type,
					owner
				}
			);
			return data;
		}
	}
});

import { defineStore } from 'pinia';
import { User } from 'src/types/User';
import axios from 'axios';
import { Encoder, OlaresInfo, DefaultOlaresInfo } from '@bytetrade/core';

export enum CloudLoginState {
	NOT_CHECK = 0,
	YES = 1,
	NO = 2
}
export type UserStoreState = {
	info: OlaresInfo;
	user: User | null;
	nfts: NFTAddress[];
	cloud_login: CloudLoginState;
	isMobile: boolean;
};

export interface VerifiablePresentation {
	'@context'?: any;
	id?: string;
	holder?: string;
	type: any;
	presentation_submission: any;
	verifiableCredential?: any[];
	proof?: any;
}

export interface NFTAddress {
	terminusName: string;
	nftName: string;
	tokenID: string;
	tokenName: string;
	ethAddressID: number;
	updateAt: number;
	chainType: string;
	userid: string;
	createAt: number;
	tokenAddress: string;
	imageUrl: string;
	terminusNameID: number;
	vp: string;
	id: number;
	ethAddress: string;
	did: string;
	status: string;
}

export const useUserStore = defineStore('userStore', {
	state: () => {
		return {
			info: DefaultOlaresInfo,
			user: null,
			nfts: [],
			cloud_login: CloudLoginState.NOT_CHECK,
			isMobile: window.innerWidth < 768
		} as UserStoreState;
	},

	getters: {
		getUser(): User | null {
			return this.user;
		},
		olaresId(): string {
			return this.info?.olaresId || this.info?.terminusName || '';
		}
	},

	actions: {
		setUser(user: User) {
			this.user = user;
		},

		setInfo(nn: OlaresInfo) {
			this.info = nn;
		},

		async getNftAddress() {
			try {
				this.cloud_login = CloudLoginState.NOT_CHECK;

				const reponse: any = await axios.get(
					window.location.origin + '/api/cloud/getNFTAddress'
				);
				const data = reponse.data;
				console.log(data);
				console.log(data.data);
				if (data.code != 0) {
					this.cloud_login = CloudLoginState.NO;
					this.nfts = [];
					return;
				}

				this.nfts = [];
				this.cloud_login = CloudLoginState.YES;

				for (let i = 0; i < data.data.address.length; ++i) {
					console.log(data.data.address[i]);
					const address: NFTAddress = data.data.address[i];
					if (address.imageUrl.startsWith('ipfs://')) {
						address.imageUrl = address.imageUrl.replace(
							'ipfs://',
							'https://gateway.ipfs.io/ipfs/'
						);
					}
					this.nfts.push(address);
					console.log(address);
					const vp: VerifiablePresentation = JSON.parse(address.vp);
					console.log(vp);
					const vcstr = Encoder.bytesToString(
						Encoder.base64UrlToBytes(vp.verifiableCredential![0].split('.')[1])
					);
					console.log(vcstr);
					const vc = JSON.parse(vcstr);
					console.log(vc);
					console.log(vc.vc.credentialSubject.image);
				}
			} catch (e) {
				console.log(e);
				this.cloud_login = CloudLoginState.NO;
				this.nfts = [];
			}
		},

		loadAvatar() {
			if (!this.user || !this.user.header || !this.user.header.avatarUrl) {
				return {
					imageUrl: '',
					avatar: ''
				};
			}
			if (this.user?.header.avatarUrl?.startsWith('http')) {
				return {
					imageUrl: this.user?.header.avatarUrl,
					avatar: this.user?.header.avatarUrl
				};
			} else {
				// const re = new RegExp('^[1-3][0-9]\\.png');
				// const re2 = new RegExp('^[0-9]\\.png');
				const re = new RegExp('^[1-3]?[0-9]\\.png');
				if (re.test(this.user?.header.avatarUrl)) {
					console.log('re true');
					return {
						imageUrl: this.user?.header.avatarUrl,
						avatar: this.user?.header.avatarUrl
					};
				} else {
					console.log('re false');
					try {
						const vp = JSON.parse(this.user?.header.avatarUrl);
						if (vp) {
							const vcstr = Encoder.bytesToString(
								Encoder.base64UrlToBytes(
									vp.verifiableCredential![0].split('.')[1]
								)
							);
							console.log(vcstr);
							const vc = JSON.parse(vcstr);
							console.log(vc);
							console.log(vc.vc.credentialSubject.image);
							let imageUrl = vc.vc.credentialSubject.image;
							if (imageUrl.startsWith('ipfs://')) {
								imageUrl = imageUrl.replace(
									'ipfs://',
									'https://gateway.ipfs.io/ipfs/'
								);
							}

							console.log(imageUrl);
							return {
								imageUrl: imageUrl,
								avatar: this.user?.header.avatarUrl
							};
						}
					} catch (e) {
						console.log(e);
						return {
							imageUrl: this.user?.header.avatarUrl,
							avatar: this.user?.header.avatarUrl
						};
					}
				}
			}
		}
	}
});

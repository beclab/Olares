// Copyright 2023 bytetrade
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import { defineStore } from 'pinia';
import { i18n } from 'src/boot/i18n';
import { notifyFailed } from 'src/utils/settings/btNotify';

export type DevelopingApps = {
	// walletHolderEvm: WalletHolder
	// cloudHolder: CloudHolder
	inited: boolean;
	account: string | undefined;
	isConnected: boolean;
	estimated: string;
	network: number | undefined;
	address: string;
};
let ethereum: any | null = null;

export const useBlockchainStore = defineStore('blockchain', {
	state() {
		return {
			// walletHolderEvm: new WalletHolder(),
			// cloudHolder: new CloudHolder()
			inited: false,
			account: undefined,
			isConnected: false,
			estimated: '',
			network: undefined,
			address: '0x1016A0886b4AeD69043367d501a99cfBAaB052B5'
		} as DevelopingApps;
	},
	actions: {
		saveAccount(account: string | undefined) {
			this.account = account;
		},
		saveNetwork(network: any) {
			this.network = network;
		},
		async unlockWallet(isManual: boolean) {
			if (process.env.SERVER) {
				return;
			}

			if (typeof window.ethereum === 'undefined') {
				if (isManual == true) {
					notifyFailed(
						i18n.global.t('errors.please_install_an_defi_wallet_like_metaMask')
					);
				}
			} else {
				ethereum = window.ethereum;
				ethereum
					.enable()
					.catch((reason: any) => {
						if (reason.code === 4001) {
							console.log('User rejected the request');
						} else if (reason.code == -32002) {
							console.log(
								'Request of type wallet_requestPermissions alread…ng for origin http://localhost:8080. Please wait'
							);
						}
					})
					.then(async (accountsMM: any) => {
						// You also should verify the user is on the correct network:

						const chainId = (window as any).ethereum.chainId;

						if (!accountsMM || accountsMM.length == 0) {
							return;
						}

						this.saveAccount(accountsMM[0]);
						this.saveNetwork(chainId);

						this.inited = true;
					})
					.catch((e: any) => {
						console.log(e);
						notifyFailed(i18n.global.t('errors.connect_wallet_failed'));
					});
			}

			if (ethereum) {
				const handleAccountsChanged = (accounts: any) => {
					if (accounts.length > 0) {
						this.saveAccount(accounts[0]);
					} else {
						this.saveAccount(undefined);
					}
				};

				ethereum.on('accountsChanged', handleAccountsChanged);

				ethereum.on('chainChanged', (chainId: any) => {
					let cid = parseInt(chainId);
					if (!cid) {
						cid = parseInt(('' + chainId).substring(2), 10);
					}

					this.saveNetwork(cid);

					if (!this.account) {
						ethereum
							.request({ method: 'eth_accounts' })
							.then(handleAccountsChanged)
							.catch((err: any) => {
								// Some unexpected error.
								// For backwards compatibility reasons, if no accounts are available,
								// eth_accounts will return an empty array.
								console.error(err);
							});
					}
				});

				ethereum.on('connect', (connectInfo: any) => {
					console.log('connect ' + JSON.stringify(connectInfo));
				});

				ethereum.on('disconnect', () => {
					this.saveAccount(undefined);
					this.saveNetwork(undefined);
				});
			}
		}
	}
});

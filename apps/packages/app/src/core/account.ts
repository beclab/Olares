// import { ethers } from "ethers";
import { Ref, ref } from 'vue';

export type WALLET_ADDRESS = string;

export class WalletHolder {
	accounts: Ref<WALLET_ADDRESS[]>;
	chain: Ref<number>;

	constructor() {
		this.accounts = ref(['']);
		this.chain = ref(NaN);
	}
}

export class CloudHolder {
	cloudId: Ref<string>;

	constructor() {
		this.cloudId = ref('');
	}
}

export enum Mode {
	METAMASK,
	CLOUD
}

export class Account {
	mode: Mode = Mode.CLOUD;
	walletHolderEvm: WalletHolder;
	cloudHolder: CloudHolder;

	constructor() {
		this.walletHolderEvm = new WalletHolder();
		this.cloudHolder = new CloudHolder();
		this.watchAccount();
		this.initCloud();
	}

	initCloud = () => {
		setTimeout(() => {
			this.cloudHolder.cloudId.value = 'aaa';
		}, 3 * 1000);
	};

	watchAccount() {
		if (!window.ethereum) {
			return;
		}
		try {
			(window as any).ethereum.on('accountsChanged', (accounts: string[]) => {
				this.walletHolderEvm.accounts.value = accounts;
				this.walletHolderEvm.chain.value = (window as any).ethereum.chainId;
			});
			(window as any).ethereum.on('chainChanged', (chainId: any) => {
				this.walletHolderEvm.chain.value = chainId;
			});
			(window as any).ethereum.on('connect', (connectInfo: any) => {
				console.log('connect ' + JSON.stringify(connectInfo));
			});
		} catch (error) {
			console.error('watchAccount error');
			console.error(error);
		}
	}

	getUserAddress(): Promise<WalletHolder> {
		return new Promise((resolve, reject) => {
			(window as any).ethereum
				.request({ method: 'eth_requestAccounts' })
				.then((accounts: any) => {
					this.walletHolderEvm.accounts.value = accounts;
					this.walletHolderEvm.chain.value = (window as any).ethereum.chainId;
				})
				.catch((err: any) => {
					console.error('wallet accounts:');
					console.error(err);
					reject(err);
				});
			resolve(this.walletHolderEvm);
		});
	}
}

const account = new Account();
export default account;

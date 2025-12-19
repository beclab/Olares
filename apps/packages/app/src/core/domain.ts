import { watch, computed } from 'vue';
import account from './account';
import { Ref, ref } from 'vue';
import { ethers } from 'ethers';
import ABI from './TerminusDIDABI';
//import { apiCloud, apiDidGate } from 'boot/axios';

type JsonRpcProvider = ethers.JsonRpcProvider;

export interface DomainNode {
	id: string;
	name: string;
	did: string;
	note: string;
	allowSubdomain: boolean;
	owner: string;
	cloudId: string;

	subdomains: DomainNode[];
}

export class Domain {
	allDomains: Ref<DomainNode[]>;
	domains: Ref<DomainNode[]>;
	choosed: Ref<DomainNode | undefined>;
	subChoosed: Ref<DomainNode | undefined>;
	domainFetching: Ref<string>;

	domainsForCloud: Ref<DomainNode[]>;
	domainsForWallet: Ref<DomainNode[]>;

	pageState: Ref<string>;

	viewDomain = computed(() => {
		if (domain.pageState.value == 'settings') {
			return undefined;
		}

		if (domain.pageState.value == 'basedomain') {
			return this.choosed.value;
		}

		if (domain.pageState.value == 'subdomain') {
			return this.subChoosed.value;
		}

		return undefined;
	});

	constructor() {
		this.pageState = ref('');
		this.domainFetching = ref('false');
		this.allDomains = ref([]);
		this.domains = ref([]);
		this.domainsForCloud = ref([]);
		this.domainsForWallet = ref([]);
		this.choosed = ref(undefined);
		this.subChoosed = ref(undefined);

		watch(
			() => account.walletHolderEvm.accounts.value,
			() => {
				this.domainFetching.value = 'true';
				this.fetchDomainsFromBlock();
			}
		);
		watch(
			() => account.cloudHolder.cloudId.value,
			() => {
				this.domainFetching.value = 'true';
				this.fetchDomainsFromCloud();
			}
		);

		this.fetchDomainsFromBlock();

		const findSubdomain = (domain: DomainNode) => {
			const arr: DomainNode[] = [];
			this.allDomains.value.forEach((item) => {
				if (
					item.name.endsWith(domain.name) &&
					domain.name.split('.').length + 1 == item.name.split('.').length
				) {
					const subdomains = findSubdomain(item);
					item.subdomains = subdomains;
					arr.push(item);
				}
			});
			return arr;
		};

		watch(
			() => this.allDomains.value,
			() => {
				const newArr: DomainNode[] = Object.assign([], this.domains.value);
				newArr.forEach((item) => {
					const domains = findSubdomain(item);
					item.subdomains = domains;
				});

				this.domains.value = newArr;
			}
		);

		watch(
			() => this.domainsForCloud.value,
			() =>
				(this.domains.value = Object.assign(
					[],
					this.domainsForCloud.value,
					this.domainsForWallet.value
				))
		);
		watch(
			() => this.domainsForWallet.value,
			() =>
				(this.domains.value = Object.assign(
					[],
					this.domainsForCloud.value,
					this.domainsForWallet.value
				))
		);
	}

	getProvider(): JsonRpcProvider {
		const provider = new ethers.JsonRpcProvider(
			'https://optimism-goerli.publicnode.com'
		);
		return provider;
	}

	getContract(): ethers.Contract {
		const contract = new ethers.Contract(
			process.env.CONTRACT_DID ? process.env.CONTRACT_DID : '',
			ABI.abi,
			this.getProvider()
		);
		return contract;
	}

	fetchDomainsFromCloud = async () => {
		// apiCloud.post('getAllDomain',
		//     {
		//         cloud_id: account.cloudHolder.cloudId.value
		//     }, {
		//     headers: {
		//         'Content-Type': 'application/json',
		//     },
		// }).then(async resp => {
		//     console.log('resp:')
		//     console.log(resp)
		//     if (resp.data.code == 0) {
		//         const contract = this.getContract()
		//         const arr: DomainNode[] = []
		//         for (const domain of resp.data.domains) {
		//             if (domain.name == 'ee.com') continue;
		//             // const tokenId = await contract.tokenIdOf(domain.name)
		//             const resp = await apiDidGate.get(`domain/faster_search/${domain.name}`)
		//             console.log('resp:')
		//             console.log(resp)
		//             if (resp.status == 200 && resp.data.code == 0) {
		//                 arr.push(resp.data.data)
		//             }
		//         }
		//         this.domainsForCloud.value = arr
		//         // console.log(this.domains)
		//         this.domainFetching.value = 'false'
		//     }
		// }).catch(error => {
		//     console.error(error)
		// })
	};

	fetchDomainsFromBlock = async () => {
		const address = account.walletHolderEvm.accounts.value[0]; //'0x945e9704d2735b420363071bb935acf2b9c4b814'//account.walletHolderEvm.accounts.value[0]
		const contract = this.getContract();
		const balance = await contract.balanceOf(address);

		const arr: DomainNode[] = [];
		for (let index = 0; index < balance; index++) {
			const tokenId = await contract.tokenOfOwnerByIndex(address, index);
			const meta = await contract.getMetadata(tokenId);
			const owner = await contract.ownerOf(tokenId);
			arr.push({
				id: tokenId,
				name: meta['0'],
				did: meta['1'],
				note: meta['2'],
				allowSubdomain: meta['3'],
				subdomains: [],
				owner: owner,
				cloudId: ''
			});
		}
		//TODO FIXME fetch cloudId

		this.domainsForWallet.value = arr;
		this.domainFetching.value = 'false';
	};

	fetchSubdomain = async (callback: any) => {
		const contract = this.getContract();
		const totalSupply = await contract.totalSupply();
		const arr: DomainNode[] = [];
		for (let index = 0; index < totalSupply; index++) {
			const tokenId = await contract.tokenByIndex(index);
			const meta = await contract.getMetadata(tokenId);
			const owner = await contract.ownerOf(tokenId);
			arr.push({
				id: tokenId,
				name: meta['0'],
				did: meta['1'],
				note: meta['2'],
				allowSubdomain: meta['3'],
				subdomains: [],
				owner: owner,
				cloudId: ''
			});
		}
		//TODO FIXME fetch cloudId

		this.allDomains.value = arr;

		callback();
	};
}

const domain = new Domain();
export default domain;

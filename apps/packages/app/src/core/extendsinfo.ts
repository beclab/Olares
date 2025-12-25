import { watch } from 'vue';
import domain from './domain';
import { Ref, ref } from 'vue';
import { ethers } from 'ethers';
import ABI from './TerminusDIDABI';
import Registry from './RegistryABI';
import axios from 'axios';

type JsonRpcProvider = ethers.JsonRpcProvider;

export interface Tag {
	key: string;
	value: string;
}

export interface ExtendsInfoData {
	tags: Tag[];
}

export class ExtendsInfo {
	extendsInfo: Ref<ExtendsInfoData>;
	resolverABIS: Map<string, any>;

	constructor() {
		this.resolverABIS = new Map();
		this.extendsInfo = ref({
			tags: []
		});

		watch(
			() => domain.choosed.value,
			() => {
				this.fetchExtendsInfo();
			}
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

	getContractRegistrar(): ethers.Contract {
		const contract = new ethers.Contract(
			process.env.CONTRACT_REGISTRY ? process.env.CONTRACT_REGISTRY : '',
			Registry.abi,
			this.getProvider()
		);
		return contract;
	}

	loadResolver = async (resolver: string) => {
		if (!this.resolverABIS.has(resolver)) {
			await axios
				.get(
					`http://api-goerli-optimism.etherscan.io/api?module=contract&action=getabi&address=${resolver}&format=raw`
				)
				.then((resp) => {
					this.resolverABIS.set(resolver, resp.data);
				})
				.catch((error) => {
					console.error(error);
				});
		}

		return new ethers.Contract(
			resolver,
			this.resolverABIS.get(resolver),
			this.getProvider()
		);
	};

	fetchExtendsInfo = async () => {
		const contract = this.getContract();
		const contractRegistrar = this.getContractRegistrar();
		const tagKeys = await contract.getTagKeys(domain.choosed.value?.id);

		const tags: Tag[] = [];
		for (const tag of tagKeys) {
			if (tag < 0xffff) {
				continue;
			}

			const ret = await contractRegistrar.getterOf(
				domain.choosed.value?.name,
				tag
			);
			const resolver = ret.substring(0, 42);
			const selector = '0x' + ret.substring(42, 50);

			const contractResolver = await this.loadResolver(resolver);

			const fragments = contractResolver.interface.fragments;
			let value = '';
			for (const fragment of fragments) {
				const functionFragment = <ethers.FunctionFragment>fragment;
				if (functionFragment.selector == selector) {
					value = await contractResolver
						.getFunction(functionFragment)
						.staticCall(domain.choosed.value?.id);
				}
			}

			tags.push({
				key: tag,
				value: value
			});
		}

		this.extendsInfo.value = { tags };
	};
}

const extendsInfo = new ExtendsInfo();
export default extendsInfo;

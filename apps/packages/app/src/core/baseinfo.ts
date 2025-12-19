import { watch } from 'vue';
import domain, { DomainNode } from './domain';
import { Ref, ref } from 'vue';
import { ethers } from 'ethers';
import Registry from './RegistryABI';
import ABI from './TerminusDIDABI';
import ResolverABI from './RootResolverABI';
// import JSEncrypt from 'jsencrypt'
// import { Buffer } from 'buffer'
import CryptoJS from 'crypto-js';
import TerminusDIDABI from './TerminusDIDABI';
import { notifyFailed, notifySuccess } from 'src/utils/settings/btNotify';
import { i18n } from 'src/boot/i18n';

const SignatureAlogorithm = {
	ECDSA: 0
};
const Action = {
	Add: 0,
	Remove: 1
};

type JsonRpcProvider = ethers.JsonRpcProvider;

export interface AuthenticatedAddress {
	'0': number;
	'1': string;
}

export interface BaseInfoData extends DomainNode {
	ip: string;
	rsaKey: string;
	resolver: string;
	authenticateds: AuthenticatedAddress[];
}

const getHex = (data: any) => {
	let hex = '';
	for (let i = 0; i < data.length; i++) {
		const code = data.charCodeAt(i);
		let hexCode = code.toString(16);
		if (hexCode.length < 2) {
			hexCode = '0' + hexCode;
		}
		hex += hexCode;
	}
	return hex;
};

const getBase64 = (data: any) => {
	const hex = data;
	const arr: number[] = [];
	for (let i = 0; i < hex.length; i += 2) {
		const code = parseInt(hex.slice(i, i + 2), 16);
		arr.push(code);
	}

	let str = '';
	for (let i = 0; i < arr.length; i++) {
		const char = String.fromCharCode(arr[i]);
		str += char;
	}
	const base64 = btoa(str);
	return base64;
};

export class BaseInfo {
	baseInfo: Ref<BaseInfoData | undefined>;

	constructor() {
		this.baseInfo = ref(undefined);

		watch(
			() => domain.viewDomain.value,
			() => {
				if (domain.viewDomain.value != undefined) {
					this.fetchBaseInfo();
				}
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

	getContractRootResolver(): ethers.Contract {
		const contract = new ethers.Contract(
			process.env.CONTRACT_ROOT_RESOLVER
				? process.env.CONTRACT_ROOT_RESOLVER
				: '',
			ResolverABI.abi,
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

	fetchBaseInfo = async () => {
		this.baseInfo.value = undefined;

		const contract = this.getContractRootResolver();
		const ip: string = await contract.dnsARecord(domain.viewDomain.value?.id);
		const rsaKey = await contract.rsaPubKey(domain.viewDomain.value?.id);

		const ipStr = `${parseInt(ip.substring(2, 4), 16)}.${parseInt(
			ip.substring(4, 6),
			16
		)}.${parseInt(ip.substring(6, 8), 16)}.${parseInt(
			ip.substring(8, 10),
			16
		)}`;

		let pubKey = '';
		try {
			if (rsaKey != '0x') {
				const str = String.fromCharCode(
					...rsaKey.match(/.{1,2}/g).map((byte: any) => parseInt(byte, 16))
				);
				const encoded = btoa(str);
				const base64 = getBase64(rsaKey.replaceAll('0x', ''));

				console.log(
					CryptoJS.enc.Base64.stringify(CryptoJS.enc.Hex.parse(rsaKey))
				);
				console.log(encoded);
				console.log(base64);
				pubKey = `-----BEGIN RSA PUBLIC KEY-----
                        ${base64}
                        -----END RSA PUBLIC KEY-----`;
			}
		} catch (error) {
			console.error(error);
		}

		const resolver = await (
			await this.getContractRegistrar()
		).customResolver(domain.viewDomain.value?.id);
		const authenticatedAddress = await contract.authenticationAddress(
			domain.viewDomain.value?.id
		);

		const baseInfo: BaseInfoData = Object.assign(
			{
				resolver: resolver,
				ip: ipStr,
				rsaKey: pubKey, //key.exportPublic('pkcs8-public-pem')
				authenticateds: authenticatedAddress
			},
			domain.viewDomain.value
		);

		this.baseInfo.value = baseInfo;

		console.log('fetchBaseInfo this.baseInfo.value:');
		console.log(this.baseInfo.value);
	};

	updateIp = async (ipStr: string) => {
		console.log(ipStr);

		const v2 =
			'((\\d{1,2})|(1\\d{2})|(2[0-4]\\d)|(25[0-5]))\\.((\\d{1,2})|(1\\d{2})|(2[0-4]\\d)|(25[0-5]))\\.((\\d{1,2})|(1\\d{2})|(2[0-4]\\d)|(25[0-5]))\\.((\\d{1,2})|(1\\d{2})|(2[0-4]\\d)|(25[0-5]))';
		const v3 =
			'((\\d{1,2})|(1\\d{2})|(2[0-4]\\d)|(25[0-5]))\\.((\\d{1,2})|(1\\d{2})|(2[0-4]\\d)|(25[0-5]))\\.((\\d{1,2})|(1\\d{2})|(2[0-4]\\d)|(25[0-5])).((d{1,2})|(1d{2})|(2[0-4]d)|(25[0-5]))';
		const reg = new RegExp(v3, 'g');
		const reg2 = new RegExp(v2, 'g');

		const isIp = ipStr.match(reg);
		const isIp2 = ipStr.match(reg2);
		console.log('isIp:');
		console.log(isIp);
		console.log('isIp2:');
		console.log(isIp2);

		if (isIp != null || isIp2 != null) {
			const provider = new ethers.BrowserProvider((window as any).ethereum);
			const signer = await provider.getSigner();

			const contract = new ethers.Contract(
				process.env.CONTRACT_ROOT_RESOLVER
					? process.env.CONTRACT_ROOT_RESOLVER
					: '',
				ResolverABI.abi,
				signer
			);
			const ipArr = ipStr.split('.');

			const ipBytes = `0x${parseInt(ipArr[0]).toString(16)}${parseInt(
				ipArr[1]
			).toString(16)}${parseInt(ipArr[2]).toString(16)}${parseInt(
				ipArr[3]
			).toString(16)}`;

			console.log(ethers.getBytes(ipBytes));
			console.log(ipBytes);
			console.log(this.baseInfo.value?.name);

			await contract
				.setDnsARecord(this.baseInfo.value?.name, ipBytes)
				.then(() => {
					notifySuccess(i18n.global.t('submit_succeed'));
				})
				.catch((error) => {
					console.log('error1:');
					console.log(error);

					notifyFailed(error.message);
				});
		} else {
			//TODO notify error
		}
	};

	updateRSAPublicKey = async (rsaPublicKey: string) => {
		const keyBase64 = rsaPublicKey
			.replace('-----BEGIN PUBLIC KEY-----', '')
			.replace('-----END PUBLIC KEY-----', '')
			.replace('\n', '');

		// const keyHex = CryptoJS.enc.Hex.stringify(CryptoJS.enc.Base64.parse(rsaPublicKey))
		const keyHex = getHex(atob(rsaPublicKey));

		const provider = new ethers.BrowserProvider((window as any).ethereum);
		const signer = await provider.getSigner();
		const contract = new ethers.Contract(
			process.env.CONTRACT_ROOT_RESOLVER
				? process.env.CONTRACT_ROOT_RESOLVER
				: '',
			ResolverABI.abi,
			signer
		);

		console.log('keyHex:');
		console.log(keyHex);
		console.log(keyBase64);
		console.log(rsaPublicKey);

		await contract
			.setRsaPubKey(this.baseInfo.value?.name, `0x${keyHex}`)
			.then(() => {
				notifySuccess(i18n.global.t('submit_succeed'));
			})
			.catch((error) => {
				console.log('error1:');
				console.log(error);

				notifyFailed(error.message);
			});
	};

	createSubdomain = async (
		owner: string,
		doamin: string,
		did: string,
		notes: string,
		allowSubdomain: boolean
	) => {
		const provider = new ethers.BrowserProvider((window as any).ethereum);
		const signer = await provider.getSigner();
		const contract = new ethers.Contract(
			process.env.CONTRACT_REGISTRY ? process.env.CONTRACT_REGISTRY : '',
			Registry.abi,
			signer
		);

		console.log('owner:');
		console.log(owner);

		console.log('doamin:');
		console.log(doamin);

		console.log('did:');
		console.log(did);

		console.log('notes:');
		console.log(notes);

		console.log('allowSubdomain:');
		console.log(allowSubdomain);

		await contract
			.register(owner, {
				domain: doamin,
				did: did,
				notes: notes,
				allowSubdomain: allowSubdomain
			})
			.then(() => {
				notifySuccess(i18n.global.t('submit_succeed'));
			})
			.catch((error) => {
				console.log('error1:');
				console.log(error);

				notifyFailed(error.message);
			});
	};

	changeOwner = async (newOwner: string) => {
		const provider = new ethers.BrowserProvider((window as any).ethereum);
		const signer = await provider.getSigner();
		const contract = new ethers.Contract(
			process.env.CONTRACT_DID ? process.env.CONTRACT_DID : '',
			TerminusDIDABI.abi,
			signer
		);

		console.log('changeOwner:');
		console.log(this.baseInfo.value?.owner);
		console.log(newOwner);
		console.log(this.baseInfo.value?.id);

		await contract
			.transferFrom(
				this.baseInfo.value?.owner,
				newOwner,
				this.baseInfo.value?.id
			)
			.then(() => {
				notifySuccess(i18n.global.t('submit_succeed'));
			})
			.catch((error) => {
				console.log('error1:');
				console.log(error);
				notifyFailed(error.message);
			});
	};

	changeResolver = async (newResolver: string) => {
		const provider = new ethers.BrowserProvider((window as any).ethereum);
		const signer = await provider.getSigner();
		const contract = new ethers.Contract(
			process.env.CONTRACT_REGISTRY ? process.env.CONTRACT_REGISTRY : '',
			Registry.abi,
			signer
		);

		await contract
			.setCustomResolver(this.baseInfo.value?.name, newResolver)
			.then(() => {
				notifySuccess(i18n.global.t('submit_succeed'));
			})
			.catch((error) => {
				console.log('error1:');
				console.log(error);
				notifyFailed(error.message);
			});
	};

	onRemoveAuthenticate = async (
		network: number,
		address: string,
		terminusName: string
	) => {
		const signDomain = {
			name: 'Terminus DID Root Tagger',
			version: '1',
			chainId: network,
			verifyingContract: process.env.CONTRACT_ROOT_RESOLVER //"0x2a50Ef86370Fc2226352b4DbAFE8fDE020BF9219"
		};

		const types = {
			AuthAddressReq: [
				{ name: 'addr', type: 'address' },
				{ name: 'algorithm', type: 'uint8' },
				{ name: 'domain', type: 'string' },
				{ name: 'signAt', type: 'uint256' },
				{ name: 'action', type: 'uint8' }
			]
		};

		const addr = address;
		const value = {
			addr: addr,
			algorithm: SignatureAlogorithm.ECDSA,
			domain: terminusName.replace('@', '.'),
			signAt: parseInt(new Date().getTime() / 1000 + ''),
			action: Action.Remove
		};

		console.log(value);
		// ethers.TypedDataEncoder
		// ethers.signTypedData()
		const provider = new ethers.BrowserProvider((window as any).ethereum);
		const signer = await provider.getSigner();
		const signed = await signer.signTypedData(signDomain, types, value);
		return {
			data: value,
			signed
			//index: index
		};
	};

	onAddAuthenticate = async (
		network: number,
		address: string,
		terminusName: string
	) => {
		console.log('network', network);
		const signDomain = {
			name: 'Terminus DID Root Tagger',
			version: '1',
			chainId: network, //11155420,
			verifyingContract: process.env.CONTRACT_ROOT_RESOLVER
		};

		const types = {
			AuthAddressReq: [
				{ name: 'addr', type: 'address' },
				{ name: 'algorithm', type: 'uint8' },
				{ name: 'domain', type: 'string' },
				{ name: 'signAt', type: 'uint256' },
				{ name: 'action', type: 'uint8' }
			]
		};

		const value = {
			addr: address,
			algorithm: SignatureAlogorithm.ECDSA,
			domain: terminusName.replace('@', '.'),
			signAt: parseInt(new Date().getTime() / 1000 + ''),
			action: Action.Add
		};

		console.log(value);
		// ethers.TypedDataEncoder
		// ethers.signTypedData()
		const provider = new ethers.BrowserProvider((window as any).ethereum);
		const signer = await provider.getSigner();

		console.log(signDomain);
		console.log(types);
		console.log(value);

		const signed = await signer.signTypedData(signDomain, types, value);
		console.log('signed:');
		console.log(signed);

		return {
			data: value,
			signed
		};
	};
}

const baseInfo = new BaseInfo();
export default baseInfo;

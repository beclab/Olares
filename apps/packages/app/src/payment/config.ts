import type { ERC20ContractAbiItem } from './types';

export const WEB3_CONFIG = {
	CHAIN: [
		{
			chainId: '0x1',
			chainName: 'Ethereum Mainnet',
			nativeCurrency: {
				name: 'Ethereum',
				symbol: 'ETH',
				decimals: 18
			},
			rpcUrls: [
				'https://mainnet.infura.io/v3/YOUR_API_KEY',
				'https://api.mycryptoapi.com/eth',
				'https://cloudflare-eth.com'
			],
			blockExplorerUrls: ['https://etherscan.io/'],
			maxPriorityFeePerGasGwei: '2',
			maxFeePerGasGwei: '50'
		},

		{
			chainId: '0x38',
			chainName: 'Binance Smart Chain Mainnet',
			nativeCurrency: {
				name: 'Binance Coin',
				symbol: 'BNB',
				decimals: 18
			},
			rpcUrls: [
				'https://bsc-dataseed1.binance.org',
				'https://bsc-dataseed2.binance.org',
				'https://bsc-dataseed3.binance.org'
			],
			blockExplorerUrls: ['https://bscscan.com/'],
			maxPriorityFeePerGasGwei: '5',
			maxFeePerGasGwei: '5'
		},

		{
			chainId: '0xa',
			chainName: 'Optimism',
			nativeCurrency: {
				name: 'Ethereum',
				symbol: 'ETH',
				decimals: 18
			},
			rpcUrls: ['https://mainnet.optimism.io'],
			blockExplorerUrls: ['https://optimistic.etherscan.io/'],
			maxPriorityFeePerGasGwei: '0.001',
			maxFeePerGasGwei: '0.01'
		},
		{
			chainId: '18',
			chainName: 'CyberMiles Mainnet',
			nativeCurrency: {
				name: 'CyberMiles Token',
				symbol: 'CMT',
				decimals: 18
			},
			rpcUrls: ['https://mainnet.cybermiles.io'],
			blockExplorerUrls: ['https://explorer.cybermiles.io/'],
			maxPriorityFeePerGasGwei: '1',
			maxFeePerGasGwei: '20'
		}
	],

	CONTRACT: {
		abi: [
			{
				inputs: [
					{ internalType: 'address', name: 'to', type: 'address' },
					{ internalType: 'uint256', name: 'amount', type: 'uint256' }
				],
				name: 'transfer',
				outputs: [{ internalType: 'bool', name: '', type: 'bool' }],
				stateMutability: 'nonpayable',
				type: 'function'
			},
			{
				inputs: [{ internalType: 'address', name: 'account', type: 'address' }],
				name: 'balanceOf',
				outputs: [{ internalType: 'uint256', name: '', type: 'uint256' }],
				stateMutability: 'view',
				type: 'function'
			},
			{
				inputs: [],
				name: 'decimals',
				outputs: [{ internalType: 'uint8', name: '', type: 'uint8' }],
				stateMutability: 'view',
				type: 'function'
			}
		] as ERC20ContractAbiItem[]
	}
};

export const ENCRYPT_CONFIG = {
	RSA_HASH_ALG: 'SHA-256' as const,
	RSA_MAX_PLAINTEXT_LENGTH: 190, // Maximum length of plaintext for RSA-OAEP with SHA-256
	OLARES_PAY_HEADER: {
		identifier: [0x4f, 0x4c, 0x52, 0x50], // "OLRP"
		version: 0x01,
		reserved: [0x00, 0x00]
	}
} as const;

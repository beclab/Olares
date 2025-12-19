export default {
	abi: [
		{
			inputs: [
				{
					internalType: 'address',
					name: 'terminusDID_',
					type: 'address'
				},
				{
					internalType: 'address',
					name: 'operator_',
					type: 'address'
				}
			],
			stateMutability: 'nonpayable',
			type: 'constructor'
		},
		{
			inputs: [],
			name: 'InvalidAction',
			type: 'error'
		},
		{
			inputs: [
				{
					internalType: 'address',
					name: 'addr',
					type: 'address'
				},
				{
					internalType: 'bytes',
					name: 'signature',
					type: 'bytes'
				}
			],
			name: 'InvalidAddressSignature',
			type: 'error'
		},
		{
			inputs: [],
			name: 'InvalidIndex',
			type: 'error'
		},
		{
			inputs: [
				{
					internalType: 'string',
					name: 'domain',
					type: 'string'
				},
				{
					internalType: 'string',
					name: 'tagName',
					type: 'string'
				}
			],
			name: 'RootTagNoExists',
			type: 'error'
		},
		{
			inputs: [
				{
					internalType: 'uint256',
					name: 'signAt',
					type: 'uint256'
				},
				{
					internalType: 'uint256',
					name: 'blockchainCurTimeStamp',
					type: 'uint256'
				}
			],
			name: 'SignatureIsValidOnlyInOneHour',
			type: 'error'
		},
		{
			inputs: [],
			name: 'Unauthorized',
			type: 'error'
		},
		{
			inputs: [],
			name: 'UnsupportedSigAlgorithm',
			type: 'error'
		},
		{
			inputs: [],
			name: 'DOMAIN_SEPARATOR',
			outputs: [
				{
					internalType: 'bytes32',
					name: '',
					type: 'bytes32'
				}
			],
			stateMutability: 'view',
			type: 'function'
		},
		{
			inputs: [
				{
					components: [
						{
							internalType: 'address',
							name: 'addr',
							type: 'address'
						},
						{
							internalType: 'enum SignatureHelper.SigAlg',
							name: 'algorithm',
							type: 'uint8'
						},
						{
							internalType: 'string',
							name: 'domain',
							type: 'string'
						},
						{
							internalType: 'uint256',
							name: 'signAt',
							type: 'uint256'
						},
						{
							internalType: 'enum SignatureHelper.Action',
							name: 'action',
							type: 'uint8'
						}
					],
					internalType: 'struct SignatureHelper.AuthAddressReq',
					name: 'authAddressReq',
					type: 'tuple'
				},
				{
					internalType: 'bytes',
					name: 'sigFromAddressPrivKey',
					type: 'bytes'
				},
				{
					internalType: 'bytes',
					name: 'sigFromDomainOwnerPrivKey',
					type: 'bytes'
				}
			],
			name: 'addAuthenticationAddress',
			outputs: [],
			stateMutability: 'nonpayable',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'string',
					name: 'domain',
					type: 'string'
				}
			],
			name: 'getAuthenticationAddresses',
			outputs: [
				{
					components: [
						{
							internalType: 'enum SignatureHelper.SigAlg',
							name: 'algorithm',
							type: 'uint8'
						},
						{
							internalType: 'address',
							name: 'addr',
							type: 'address'
						}
					],
					internalType: 'struct RootTagger.AuthAddress[]',
					name: '',
					type: 'tuple[]'
				}
			],
			stateMutability: 'view',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'string',
					name: 'domain',
					type: 'string'
				}
			],
			name: 'getDnsARecord',
			outputs: [
				{
					internalType: 'bytes4',
					name: '',
					type: 'bytes4'
				}
			],
			stateMutability: 'view',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'string',
					name: 'domain',
					type: 'string'
				}
			],
			name: 'getLatestDID',
			outputs: [
				{
					internalType: 'string',
					name: '',
					type: 'string'
				}
			],
			stateMutability: 'view',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'string',
					name: 'domain',
					type: 'string'
				}
			],
			name: 'getRsaPubKey',
			outputs: [
				{
					internalType: 'bytes',
					name: '',
					type: 'bytes'
				}
			],
			stateMutability: 'view',
			type: 'function'
		},
		{
			inputs: [],
			name: 'operator',
			outputs: [
				{
					internalType: 'address',
					name: '',
					type: 'address'
				}
			],
			stateMutability: 'view',
			type: 'function'
		},
		{
			inputs: [
				{
					components: [
						{
							internalType: 'address',
							name: 'addr',
							type: 'address'
						},
						{
							internalType: 'enum SignatureHelper.SigAlg',
							name: 'algorithm',
							type: 'uint8'
						},
						{
							internalType: 'string',
							name: 'domain',
							type: 'string'
						},
						{
							internalType: 'uint256',
							name: 'signAt',
							type: 'uint256'
						},
						{
							internalType: 'enum SignatureHelper.Action',
							name: 'action',
							type: 'uint8'
						}
					],
					internalType: 'struct SignatureHelper.AuthAddressReq',
					name: 'authAddressReq',
					type: 'tuple'
				},
				{
					internalType: 'bytes',
					name: 'sigFromDomainOwnerPrivKey',
					type: 'bytes'
				},
				{
					internalType: 'uint256',
					name: 'index',
					type: 'uint256'
				}
			],
			name: 'removeAuthenticationAddress',
			outputs: [],
			stateMutability: 'nonpayable',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'string',
					name: 'domain',
					type: 'string'
				},
				{
					internalType: 'bytes4',
					name: 'ipv4',
					type: 'bytes4'
				}
			],
			name: 'setDnsARecord',
			outputs: [],
			stateMutability: 'nonpayable',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'string',
					name: 'domain',
					type: 'string'
				},
				{
					internalType: 'string',
					name: 'latestDID',
					type: 'string'
				}
			],
			name: 'setLatestDID',
			outputs: [],
			stateMutability: 'nonpayable',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'string',
					name: 'domain',
					type: 'string'
				},
				{
					internalType: 'bytes',
					name: 'pubKey',
					type: 'bytes'
				}
			],
			name: 'setRsaPubKey',
			outputs: [],
			stateMutability: 'nonpayable',
			type: 'function'
		},
		{
			inputs: [],
			name: 'terminusDID',
			outputs: [
				{
					internalType: 'address',
					name: '',
					type: 'address'
				}
			],
			stateMutability: 'view',
			type: 'function'
		}
	]
};

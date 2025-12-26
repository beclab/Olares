export default {
	abi: [
		{
			inputs: [
				{
					internalType: 'address',
					name: 'registry_',
					type: 'address'
				},
				{
					internalType: 'address',
					name: 'rootResolver_',
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
			inputs: [
				{
					internalType: 'address',
					name: 'resolver',
					type: 'address'
				}
			],
			name: 'BadResolver',
			type: 'error'
		},
		{
			inputs: [
				{
					internalType: 'address',
					name: 'owner',
					type: 'address'
				}
			],
			name: 'OwnableInvalidOwner',
			type: 'error'
		},
		{
			inputs: [
				{
					internalType: 'address',
					name: 'account',
					type: 'address'
				}
			],
			name: 'OwnableUnauthorizedAccount',
			type: 'error'
		},
		{
			inputs: [],
			name: 'Unauthorized',
			type: 'error'
		},
		{
			inputs: [
				{
					internalType: 'uint256',
					name: 'tokenId',
					type: 'uint256'
				},
				{
					internalType: 'uint256',
					name: 'key',
					type: 'uint256'
				}
			],
			name: 'UnsupportedTag',
			type: 'error'
		},
		{
			anonymous: false,
			inputs: [
				{
					indexed: true,
					internalType: 'address',
					name: 'previousOwner',
					type: 'address'
				},
				{
					indexed: true,
					internalType: 'address',
					name: 'newOwner',
					type: 'address'
				}
			],
			name: 'OwnershipTransferStarted',
			type: 'event'
		},
		{
			anonymous: false,
			inputs: [
				{
					indexed: true,
					internalType: 'address',
					name: 'previousOwner',
					type: 'address'
				},
				{
					indexed: true,
					internalType: 'address',
					name: 'newOwner',
					type: 'address'
				}
			],
			name: 'OwnershipTransferred',
			type: 'event'
		},
		{
			inputs: [],
			name: 'acceptOwnership',
			outputs: [],
			stateMutability: 'nonpayable',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'uint256',
					name: 'tokenId',
					type: 'uint256'
				}
			],
			name: 'customResolver',
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
					internalType: 'string',
					name: 'domain',
					type: 'string'
				},
				{
					internalType: 'uint256',
					name: 'key',
					type: 'uint256'
				}
			],
			name: 'getterOf',
			outputs: [
				{
					type: 'bytes32'
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
			inputs: [],
			name: 'owner',
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
			inputs: [],
			name: 'pendingOwner',
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
					internalType: 'address',
					name: 'tokenOwner',
					type: 'address'
				},
				{
					components: [
						{
							internalType: 'string',
							name: 'domain',
							type: 'string'
						},
						{
							internalType: 'string',
							name: 'did',
							type: 'string'
						},
						{
							internalType: 'string',
							name: 'notes',
							type: 'string'
						},
						{
							internalType: 'bool',
							name: 'allowSubdomain',
							type: 'bool'
						}
					],
					internalType: 'struct Metadata',
					name: 'metadata',
					type: 'tuple'
				}
			],
			name: 'register',
			outputs: [
				{
					internalType: 'uint256',
					name: 'tokenId',
					type: 'uint256'
				}
			],
			stateMutability: 'nonpayable',
			type: 'function'
		},
		{
			inputs: [],
			name: 'registry',
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
			inputs: [],
			name: 'renounceOwnership',
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
					internalType: 'uint256',
					name: 'key',
					type: 'uint256'
				}
			],
			name: 'resolverOf',
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
			inputs: [],
			name: 'rootResolver',
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
					internalType: 'string',
					name: 'domain',
					type: 'string'
				},
				{
					internalType: 'address',
					name: 'resolver',
					type: 'address'
				}
			],
			name: 'setCustomResolver',
			outputs: [],
			stateMutability: 'nonpayable',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'address',
					name: 'operator_',
					type: 'address'
				}
			],
			name: 'setOperator',
			outputs: [],
			stateMutability: 'nonpayable',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'address',
					name: 'registry_',
					type: 'address'
				}
			],
			name: 'setRegistry',
			outputs: [],
			stateMutability: 'nonpayable',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'address',
					name: 'rootResolver_',
					type: 'address'
				}
			],
			name: 'setRootResolver',
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
					internalType: 'uint256',
					name: 'key',
					type: 'uint256'
				},
				{
					internalType: 'bytes',
					name: 'value',
					type: 'bytes'
				}
			],
			name: 'setTag',
			outputs: [
				{
					internalType: 'bool',
					name: 'addedOrRemoved',
					type: 'bool'
				}
			],
			stateMutability: 'nonpayable',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'uint256',
					name: 'key',
					type: 'uint256'
				}
			],
			name: 'tagGetter',
			outputs: [
				{
					internalType: 'bytes4',
					name: '',
					type: 'bytes4'
				}
			],
			stateMutability: 'pure',
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
					internalType: 'address',
					name: 'owner',
					type: 'address'
				}
			],
			name: 'traceOwner',
			outputs: [
				{
					internalType: 'uint256',
					name: 'domainLevel',
					type: 'uint256'
				},
				{
					internalType: 'uint256',
					name: 'ownedLevel',
					type: 'uint256'
				},
				{
					internalType: 'string',
					name: 'ownedDomain',
					type: 'string'
				}
			],
			stateMutability: 'view',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'address',
					name: 'newOwner',
					type: 'address'
				}
			],
			name: 'transferOwnership',
			outputs: [],
			stateMutability: 'nonpayable',
			type: 'function'
		}
	]
};

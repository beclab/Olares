export default {
	abi: [
		{
			inputs: [],
			stateMutability: 'nonpayable',
			type: 'constructor'
		},
		{
			inputs: [
				{
					internalType: 'address',
					name: 'target',
					type: 'address'
				}
			],
			name: 'AddressEmptyCode',
			type: 'error'
		},
		{
			inputs: [
				{
					internalType: 'address',
					name: 'implementation',
					type: 'address'
				}
			],
			name: 'ERC1967InvalidImplementation',
			type: 'error'
		},
		{
			inputs: [],
			name: 'ERC1967NonPayable',
			type: 'error'
		},
		{
			inputs: [
				{
					internalType: 'uint256',
					name: 'tokenId',
					type: 'uint256'
				}
			],
			name: 'ERC721DuplicateToken',
			type: 'error'
		},
		{
			inputs: [
				{
					internalType: 'address',
					name: 'sender',
					type: 'address'
				},
				{
					internalType: 'uint256',
					name: 'tokenId',
					type: 'uint256'
				},
				{
					internalType: 'address',
					name: 'owner',
					type: 'address'
				}
			],
			name: 'ERC721IncorrectOwner',
			type: 'error'
		},
		{
			inputs: [
				{
					internalType: 'address',
					name: 'operator',
					type: 'address'
				},
				{
					internalType: 'uint256',
					name: 'tokenId',
					type: 'uint256'
				}
			],
			name: 'ERC721InsufficientApproval',
			type: 'error'
		},
		{
			inputs: [
				{
					internalType: 'address',
					name: 'approver',
					type: 'address'
				}
			],
			name: 'ERC721InvalidApprover',
			type: 'error'
		},
		{
			inputs: [
				{
					internalType: 'address',
					name: 'operator',
					type: 'address'
				}
			],
			name: 'ERC721InvalidOperator',
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
			name: 'ERC721InvalidOwner',
			type: 'error'
		},
		{
			inputs: [
				{
					internalType: 'address',
					name: 'receiver',
					type: 'address'
				}
			],
			name: 'ERC721InvalidReceiver',
			type: 'error'
		},
		{
			inputs: [
				{
					internalType: 'address',
					name: 'sender',
					type: 'address'
				}
			],
			name: 'ERC721InvalidSender',
			type: 'error'
		},
		{
			inputs: [
				{
					internalType: 'uint256',
					name: 'tokenId',
					type: 'uint256'
				}
			],
			name: 'ERC721NonexistentToken',
			type: 'error'
		},
		{
			inputs: [
				{
					internalType: 'address',
					name: 'owner',
					type: 'address'
				},
				{
					internalType: 'uint256',
					name: 'index',
					type: 'uint256'
				}
			],
			name: 'ERC721OutOfBoundsIndex',
			type: 'error'
		},
		{
			inputs: [],
			name: 'FailedInnerCall',
			type: 'error'
		},
		{
			inputs: [
				{
					internalType: 'string',
					name: 'label',
					type: 'string'
				}
			],
			name: 'InvalidDomainLabel',
			type: 'error'
		},
		{
			inputs: [],
			name: 'InvalidInitialization',
			type: 'error'
		},
		{
			inputs: [
				{
					internalType: 'string',
					name: 'domain',
					type: 'string'
				}
			],
			name: 'InvalidRegistration',
			type: 'error'
		},
		{
			inputs: [],
			name: 'InvalidTagDefinition',
			type: 'error'
		},
		{
			inputs: [],
			name: 'NotInitializing',
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
			inputs: [
				{
					internalType: 'string',
					name: 'domain',
					type: 'string'
				},
				{
					internalType: 'string',
					name: 'name',
					type: 'string'
				}
			],
			name: 'RedefinedTag',
			type: 'error'
		},
		{
			inputs: [],
			name: 'TagInvalidOp',
			type: 'error'
		},
		{
			inputs: [],
			name: 'TagTooManyEntries',
			type: 'error'
		},
		{
			inputs: [],
			name: 'UUPSUnauthorizedCallContext',
			type: 'error'
		},
		{
			inputs: [
				{
					internalType: 'bytes32',
					name: 'slot',
					type: 'bytes32'
				}
			],
			name: 'UUPSUnsupportedProxiableUUID',
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
					internalType: 'string',
					name: 'domain',
					type: 'string'
				},
				{
					internalType: 'string',
					name: 'name',
					type: 'string'
				}
			],
			name: 'UndefinedTag',
			type: 'error'
		},
		{
			inputs: [
				{
					internalType: 'string',
					name: 'domain',
					type: 'string'
				}
			],
			name: 'UnregisteredDomain',
			type: 'error'
		},
		{
			anonymous: false,
			inputs: [
				{
					indexed: true,
					internalType: 'address',
					name: 'owner',
					type: 'address'
				},
				{
					indexed: true,
					internalType: 'address',
					name: 'approved',
					type: 'address'
				},
				{
					indexed: true,
					internalType: 'uint256',
					name: 'tokenId',
					type: 'uint256'
				}
			],
			name: 'Approval',
			type: 'event'
		},
		{
			anonymous: false,
			inputs: [
				{
					indexed: true,
					internalType: 'address',
					name: 'owner',
					type: 'address'
				},
				{
					indexed: true,
					internalType: 'address',
					name: 'operator',
					type: 'address'
				},
				{
					indexed: false,
					internalType: 'bool',
					name: 'approved',
					type: 'bool'
				}
			],
			name: 'ApprovalForAll',
			type: 'event'
		},
		{
			anonymous: false,
			inputs: [
				{
					indexed: false,
					internalType: 'uint64',
					name: 'version',
					type: 'uint64'
				}
			],
			name: 'Initialized',
			type: 'event'
		},
		{
			anonymous: false,
			inputs: [
				{
					indexed: true,
					internalType: 'bytes32',
					name: 'hash',
					type: 'bytes32'
				},
				{
					indexed: false,
					internalType: 'string[]',
					name: 'value',
					type: 'string[]'
				}
			],
			name: 'OffchainStringArray',
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
			anonymous: false,
			inputs: [
				{
					indexed: true,
					internalType: 'address',
					name: 'from',
					type: 'address'
				},
				{
					indexed: true,
					internalType: 'address',
					name: 'to',
					type: 'address'
				},
				{
					indexed: true,
					internalType: 'uint256',
					name: 'tokenId',
					type: 'uint256'
				}
			],
			name: 'Transfer',
			type: 'event'
		},
		{
			anonymous: false,
			inputs: [
				{
					indexed: true,
					internalType: 'uint256',
					name: 'tokenId',
					type: 'uint256'
				}
			],
			name: 'TransferByParentOwner',
			type: 'event'
		},
		{
			anonymous: false,
			inputs: [
				{
					indexed: true,
					internalType: 'uint256',
					name: 'tokenId',
					type: 'uint256'
				}
			],
			name: 'TransferBySuperAdmin',
			type: 'event'
		},
		{
			anonymous: false,
			inputs: [
				{
					indexed: true,
					internalType: 'address',
					name: 'implementation',
					type: 'address'
				}
			],
			name: 'Upgraded',
			type: 'event'
		},
		{
			inputs: [],
			name: 'UPGRADE_INTERFACE_VERSION',
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
			inputs: [],
			name: 'acceptOwnership',
			outputs: [],
			stateMutability: 'nonpayable',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'string',
					name: 'from',
					type: 'string'
				},
				{
					internalType: 'string',
					name: 'to',
					type: 'string'
				},
				{
					internalType: 'string',
					name: 'name',
					type: 'string'
				},
				{
					internalType: 'bytes',
					name: 'value',
					type: 'bytes'
				}
			],
			name: 'addTag',
			outputs: [],
			stateMutability: 'nonpayable',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'address',
					name: 'to',
					type: 'address'
				},
				{
					internalType: 'uint256',
					name: 'tokenId',
					type: 'uint256'
				}
			],
			name: 'approve',
			outputs: [],
			stateMutability: 'nonpayable',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'address',
					name: 'owner',
					type: 'address'
				}
			],
			name: 'balanceOf',
			outputs: [
				{
					internalType: 'uint256',
					name: '',
					type: 'uint256'
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
					internalType: 'string',
					name: 'name',
					type: 'string'
				},
				{
					internalType: 'bytes',
					name: 'abiType',
					type: 'bytes'
				},
				{
					internalType: 'string[][]',
					name: 'fieldNames',
					type: 'string[][]'
				}
			],
			name: 'defineTag',
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
			name: 'getApproved',
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
					internalType: 'bytes32',
					name: 'hash',
					type: 'bytes32'
				}
			],
			name: 'getFieldNamesEventBlock',
			outputs: [
				{
					internalType: 'uint256',
					name: '',
					type: 'uint256'
				}
			],
			stateMutability: 'view',
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
			name: 'getMetadata',
			outputs: [
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
					internalType: 'struct TerminusDID.Metadata',
					name: '',
					type: 'tuple'
				}
			],
			stateMutability: 'view',
			type: 'function'
		},
		//  {
		//   "inputs": [{
		//     "internalType": "string",
		//     "name": "domain",
		//     "type": "string"
		//   }],
		//   "name": "getMetadata",
		//   "outputs": [{
		//     "components": [{
		//       "internalType": "string",
		//       "name": "domain",
		//       "type": "string"
		//     }, {
		//       "internalType": "string",
		//       "name": "did",
		//       "type": "string"
		//     }, {
		//       "internalType": "string",
		//       "name": "notes",
		//       "type": "string"
		//     }, {
		//       "internalType": "bool",
		//       "name": "allowSubdomain",
		//       "type": "bool"
		//     }],
		//     "internalType": "struct TerminusDID.Metadata",
		//     "name": "",
		//     "type": "tuple"
		//   }],
		//   "stateMutability": "view",
		//   "type": "function"
		// },
		{
			inputs: [
				{
					internalType: 'string',
					name: 'domain',
					type: 'string'
				},
				{
					internalType: 'string',
					name: 'name',
					type: 'string'
				}
			],
			name: 'getTagABIType',
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
			inputs: [
				{
					internalType: 'string',
					name: 'from',
					type: 'string'
				},
				{
					internalType: 'string',
					name: 'to',
					type: 'string'
				}
			],
			name: 'getTagCount',
			outputs: [
				{
					internalType: 'uint256',
					name: '',
					type: 'uint256'
				}
			],
			stateMutability: 'view',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'string',
					name: 'from',
					type: 'string'
				},
				{
					internalType: 'string',
					name: 'to',
					type: 'string'
				},
				{
					internalType: 'string',
					name: 'name',
					type: 'string'
				},
				{
					internalType: 'uint256[]',
					name: 'elemPath',
					type: 'uint256[]'
				}
			],
			name: 'getTagElem',
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
			inputs: [
				{
					internalType: 'string',
					name: 'from',
					type: 'string'
				},
				{
					internalType: 'string',
					name: 'to',
					type: 'string'
				},
				{
					internalType: 'string',
					name: 'name',
					type: 'string'
				},
				{
					internalType: 'uint256[]',
					name: 'elemPath',
					type: 'uint256[]'
				}
			],
			name: 'getTagElemLength',
			outputs: [
				{
					internalType: 'uint256',
					name: '',
					type: 'uint256'
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
					internalType: 'string',
					name: 'name',
					type: 'string'
				},
				{
					internalType: 'uint256',
					name: 'index',
					type: 'uint256'
				}
			],
			name: 'getTagFieldNamesHashByIndex',
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
					internalType: 'string',
					name: 'from',
					type: 'string'
				},
				{
					internalType: 'string',
					name: 'to',
					type: 'string'
				},
				{
					internalType: 'uint256',
					name: 'index',
					type: 'uint256'
				}
			],
			name: 'getTagNameByIndex',
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
				},
				{
					internalType: 'string',
					name: 'name',
					type: 'string'
				}
			],
			name: 'getTagType',
			outputs: [
				{
					internalType: 'bytes',
					name: 'abiType',
					type: 'bytes'
				},
				{
					internalType: 'bytes32[]',
					name: 'fieldNamesHash',
					type: 'bytes32[]'
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
					internalType: 'string',
					name: 'name',
					type: 'string'
				}
			],
			name: 'getTagger',
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
					name: 'from',
					type: 'string'
				},
				{
					internalType: 'string',
					name: 'to',
					type: 'string'
				},
				{
					internalType: 'string',
					name: 'name',
					type: 'string'
				}
			],
			name: 'hasTag',
			outputs: [
				{
					internalType: 'bool',
					name: '',
					type: 'bool'
				}
			],
			stateMutability: 'view',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'string',
					name: 'name_',
					type: 'string'
				},
				{
					internalType: 'string',
					name: 'symbol_',
					type: 'string'
				}
			],
			name: 'initialize',
			outputs: [],
			stateMutability: 'nonpayable',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'address',
					name: 'owner',
					type: 'address'
				},
				{
					internalType: 'address',
					name: 'operator',
					type: 'address'
				}
			],
			name: 'isApprovedForAll',
			outputs: [
				{
					internalType: 'bool',
					name: '',
					type: 'bool'
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
			name: 'isRegistered',
			outputs: [
				{
					internalType: 'bool',
					name: '',
					type: 'bool'
				}
			],
			stateMutability: 'view',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'bytes[]',
					name: 'data',
					type: 'bytes[]'
				}
			],
			name: 'multicall',
			outputs: [
				{
					internalType: 'bytes[]',
					name: 'results',
					type: 'bytes[]'
				}
			],
			stateMutability: 'nonpayable',
			type: 'function'
		},
		{
			inputs: [],
			name: 'name',
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
			inputs: [
				{
					internalType: 'uint256',
					name: 'tokenId',
					type: 'uint256'
				}
			],
			name: 'ownerOf',
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
					internalType: 'string',
					name: 'from',
					type: 'string'
				},
				{
					internalType: 'string',
					name: 'to',
					type: 'string'
				},
				{
					internalType: 'string',
					name: 'name',
					type: 'string'
				},
				{
					internalType: 'uint256[]',
					name: 'elemPath',
					type: 'uint256[]'
				}
			],
			name: 'popTagElem',
			outputs: [],
			stateMutability: 'nonpayable',
			type: 'function'
		},
		{
			inputs: [],
			name: 'proxiableUUID',
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
					internalType: 'string',
					name: 'from',
					type: 'string'
				},
				{
					internalType: 'string',
					name: 'to',
					type: 'string'
				},
				{
					internalType: 'string',
					name: 'name',
					type: 'string'
				},
				{
					internalType: 'uint256[]',
					name: 'elemPath',
					type: 'uint256[]'
				},
				{
					internalType: 'bytes',
					name: 'value',
					type: 'bytes'
				}
			],
			name: 'pushTagElem',
			outputs: [],
			stateMutability: 'nonpayable',
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
					internalType: 'struct TerminusDID.Metadata',
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
			inputs: [
				{
					internalType: 'string',
					name: 'from',
					type: 'string'
				},
				{
					internalType: 'string',
					name: 'to',
					type: 'string'
				},
				{
					internalType: 'string',
					name: 'name',
					type: 'string'
				}
			],
			name: 'removeTag',
			outputs: [],
			stateMutability: 'nonpayable',
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
					internalType: 'address',
					name: 'from',
					type: 'address'
				},
				{
					internalType: 'address',
					name: 'to',
					type: 'address'
				},
				{
					internalType: 'uint256',
					name: 'tokenId',
					type: 'uint256'
				}
			],
			name: 'safeTransferFrom',
			outputs: [],
			stateMutability: 'nonpayable',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'address',
					name: 'from',
					type: 'address'
				},
				{
					internalType: 'address',
					name: 'to',
					type: 'address'
				},
				{
					internalType: 'uint256',
					name: 'tokenId',
					type: 'uint256'
				},
				{
					internalType: 'bytes',
					name: 'data',
					type: 'bytes'
				}
			],
			name: 'safeTransferFrom',
			outputs: [],
			stateMutability: 'nonpayable',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'address',
					name: 'operator',
					type: 'address'
				},
				{
					internalType: 'bool',
					name: 'approved',
					type: 'bool'
				}
			],
			name: 'setApprovalForAll',
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
					internalType: 'string',
					name: 'domain',
					type: 'string'
				},
				{
					internalType: 'string',
					name: 'name',
					type: 'string'
				},
				{
					internalType: 'address',
					name: 'tagger',
					type: 'address'
				}
			],
			name: 'setTagger',
			outputs: [],
			stateMutability: 'nonpayable',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'bytes4',
					name: 'interfaceId',
					type: 'bytes4'
				}
			],
			name: 'supportsInterface',
			outputs: [
				{
					internalType: 'bool',
					name: '',
					type: 'bool'
				}
			],
			stateMutability: 'view',
			type: 'function'
		},
		{
			inputs: [],
			name: 'symbol',
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
					internalType: 'uint256',
					name: 'index',
					type: 'uint256'
				}
			],
			name: 'tokenByIndex',
			outputs: [
				{
					internalType: 'uint256',
					name: '',
					type: 'uint256'
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
			name: 'tokenIdOf',
			outputs: [
				{
					internalType: 'uint256',
					name: '',
					type: 'uint256'
				}
			],
			stateMutability: 'pure',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'address',
					name: 'owner',
					type: 'address'
				},
				{
					internalType: 'uint256',
					name: 'index',
					type: 'uint256'
				}
			],
			name: 'tokenOfOwnerByIndex',
			outputs: [
				{
					internalType: 'uint256',
					name: '',
					type: 'uint256'
				}
			],
			stateMutability: 'view',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'uint256',
					name: '',
					type: 'uint256'
				}
			],
			name: 'tokenURI',
			outputs: [
				{
					internalType: 'string',
					name: '',
					type: 'string'
				}
			],
			stateMutability: 'pure',
			type: 'function'
		},
		{
			inputs: [],
			name: 'totalSupply',
			outputs: [
				{
					internalType: 'uint256',
					name: '',
					type: 'uint256'
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
					name: 'from',
					type: 'address'
				},
				{
					internalType: 'address',
					name: 'to',
					type: 'address'
				},
				{
					internalType: 'uint256',
					name: 'tokenId',
					type: 'uint256'
				}
			],
			name: 'transferFrom',
			outputs: [],
			stateMutability: 'nonpayable',
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
		},
		{
			inputs: [
				{
					internalType: 'string',
					name: 'from',
					type: 'string'
				},
				{
					internalType: 'string',
					name: 'to',
					type: 'string'
				},
				{
					internalType: 'string',
					name: 'name',
					type: 'string'
				},
				{
					internalType: 'uint256[]',
					name: 'elemPath',
					type: 'uint256[]'
				},
				{
					internalType: 'bytes',
					name: 'value',
					type: 'bytes'
				}
			],
			name: 'updateTagElem',
			outputs: [],
			stateMutability: 'nonpayable',
			type: 'function'
		},
		{
			inputs: [
				{
					internalType: 'address',
					name: 'newImplementation',
					type: 'address'
				},
				{
					internalType: 'bytes',
					name: 'data',
					type: 'bytes'
				}
			],
			name: 'upgradeToAndCall',
			outputs: [],
			stateMutability: 'payable',
			type: 'function'
		}
	]
};

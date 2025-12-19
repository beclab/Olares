import { AddressValidator } from '../addressValidator';
import type { ChainConfig, OrderDataBase } from '../types';
import { Contract } from 'web3-eth-contract';
import { WEB3_CONFIG } from '../config';
import { logger } from '../logger';
import { ref } from 'vue';
import Web3 from 'web3';
import BigNumber from 'bignumber.js';

class ChainStrategy {
	static async switchToTargetChain(
		ethereum: Window['ethereum'],
		chainConfig: ChainConfig
	): Promise<void> {
		try {
			await ethereum.request({
				method: 'wallet_switchEthereumChain',
				params: [{ chainId: chainConfig.chainId }]
			});
			logger.log(
				`✅ Switched to chain: ${chainConfig.chainName} (${chainConfig.chainId})`
			);
		} catch (switchError: unknown) {
			const error = switchError as { code?: number; message?: string };
			if (error.code === 4902) {
				await ethereum.request({
					method: 'wallet_addEthereumChain',
					params: [chainConfig]
				});
				logger.log(
					`✅ Added chain: ${chainConfig.chainName} (${chainConfig.chainId})`
				);
			} else {
				throw new Error(
					`Chain switch failed: ${error.message || 'Unknown error'}`
				);
			}
		}
	}
}

export class Web3Service {
	private web3: Web3 | null = null;
	private isConnected = ref(false);
	private userAccount = ref('');
	private contract: Contract<any> | null = null;
	private chainConfig: ChainConfig;

	private initContract(contractAddress: string): void {
		if (!this.web3) {
			throw new Error('Web3 instance not initialized (connect wallet first)');
		}
		this.contract = new this.web3.eth.Contract(
			WEB3_CONFIG.CONTRACT.abi,
			contractAddress
		) as Contract<any>;
	}

	async connectWallet(
		contractAddress: string,
		chainConfig: ChainConfig
	): Promise<void> {
		logger.log('Connecting to wallet...');
		try {
			if (!window.ethereum) {
				throw new Error('MetaMask not detected (please install MetaMask)');
			}

			this.chainConfig = chainConfig;
			await ChainStrategy.switchToTargetChain(window.ethereum, chainConfig);

			const accounts = (await window.ethereum.request({
				method: 'eth_requestAccounts'
			})) as string[];

			if (!accounts || accounts.length === 0) {
				throw new Error('No accounts authorized');
			}

			this.userAccount.value = accounts[0];
			this.web3 = new Web3(window.ethereum);
			this.initContract(contractAddress);

			this.isConnected.value = true;
			logger.log(`✅ Wallet connected: ${this.userAccount.value}`);
			await this.queryBalance();
		} catch (error) {
			const errorMsg =
				error instanceof Error ? error.message : 'connectWallet error';
			logger.log(`❌ Wallet connection failed: ${errorMsg}`);
			throw error;
		}
	}

	/**
	 * Query user's token balance in the smallest unit (wei-like units)
	 * @returns Balance in the smallest unit as a string to maintain precision
	 */
	async queryBalance(): Promise<string> {
		if (!this.contract || !this.userAccount.value) {
			throw new Error('contract not initialized or no account connected');
		}

		try {
			const balanceWei = await this.contract.methods
				.balanceOf(this.userAccount.value)
				.call();
			const decimals = await this.contract.methods.decimals().call();

			// Log formatted balance for readability
			const balanceFormatted = new BigNumber(Number(balanceWei).toString())
				.dividedBy(new BigNumber(10).exponentiatedBy(Number(decimals)))
				.toString();
			logger.log(`Balance: ${balanceFormatted} (${balanceWei} smallest units)`);

			// Return raw balance string to maintain precision
			return Number(balanceWei).toString();
		} catch (error) {
			const errorMsg =
				error instanceof Error ? error.message : 'queryBalance error';
			logger.log(`❌ balance query failed: ${errorMsg}`);
			throw error;
		}
	}

	async signOrder(orderData: OrderDataBase): Promise<OrderDataBase> {
		if (!this.web3 || !window.ethereum || !this.userAccount.value) {
			throw new Error('Web3 not connected (connect wallet first)');
		}

		logger.log('Starting order signature...');
		console.log(orderData);
		try {
			const productId = orderData.product[0]?.product_id || 'N/A';
			const message = `Order Signature Confirmation
Sender: ${orderData.from}
Receiver: ${orderData.to}
Product ID: ${productId}
Nonce: ${orderData.nonce}`;

			const hexMessage = this.web3.utils.utf8ToHex(message);

			const signature = (await window.ethereum.request({
				method: 'personal_sign',
				params: [hexMessage, this.userAccount.value]
			})) as string;

			const signedOrder = {
				...orderData,
				signature,
				signer: this.userAccount.value
			};

			logger.log('✅ Order signed successfully');
			return signedOrder;
		} catch (error: unknown) {
			const err = error as { code?: number; message?: string };
			if (err.code === 4001) {
				const cancelMsg = 'User cancelled order signature';
				logger.log(cancelMsg);
				throw new Error(cancelMsg);
			}
			const errorMsg =
				error instanceof Error ? error.message : 'signOrder error';
			logger.log(`❌ Order signature failed: ${errorMsg}`);
			throw error;
		}
	}

	/**
	 * Detect if the current chain supports EIP-1559 (London Upgrade)
	 * @returns Returns true if supported, otherwise false
	 */
	async isEIP1559Supported(): Promise<boolean> {
		// Check if Web3 instance exists
		if (!this.web3) {
			logger.log('Web3 not initialized, cannot detect EIP-1559 support');
			return false;
		}

		try {
			// Get latest block (false is faster, doesn't need full transaction details)
			const latestBlock = await this.web3.eth.getBlock('latest', false);

			// EIP-1559 enabled chains must have baseFeePerGas field in blocks
			// This field is the core indicator of EIP-1559, representing the block's base fee
			const isSupported =
				latestBlock &&
				'baseFeePerGas' in latestBlock &&
				latestBlock.baseFeePerGas !== null &&
				latestBlock.baseFeePerGas !== undefined;

			if (isSupported) {
				logger.log(
					`✅ Current chain supports EIP-1559 (baseFee: ${
						latestBlock && latestBlock.baseFeePerGas
							? this.web3.utils.fromWei(latestBlock.baseFeePerGas, 'gwei')
							: ' latestBlock null '
					} Gwei)`
				);
			} else {
				logger.log(
					'Current chain does not support EIP-1559, will use legacy gasPrice'
				);
			}

			return isSupported;
		} catch (error) {
			const errorMsg =
				error instanceof Error ? error.message : 'isEIP1559Supported error';
			logger.log(`Failed to detect EIP-1559 support: ${errorMsg}`);
			// Default to not supported on error, using legacy gasPrice is safer
			return false;
		}
	}

	/**
	 * Get EIP-1559 maxPriorityFeePerGas and maxFeePerGas from Web3
	 * @param web3 Web3 instance
	 * @param gas Fallback gas configuration (used when dynamic fetching fails)
	 * @returns String parameters in WEI units
	 */
	async getEIP1559Fees(
		web3: Web3,
		gas: {
			maxPriorityFeePerGasGwei: string;
			maxFeePerGasGwei: string;
		}
	): Promise<{
		maxPriorityFeePerGas: string;
		maxFeePerGas: string;
	}> {
		try {
			// 1. Get fee history from the last 4 blocks
			// Parameters: [number of blocks, latest block identifier, priority fee percentile array]
			const feeHistory = await web3.eth.getFeeHistory(4, 'latest', [50]);

			// 2. Calculate average priority fee (average from median values of multiple blocks)
			let totalPriorityFee = BigInt(0);
			let count = 0;

			for (const rewards of feeHistory.reward) {
				if (rewards && rewards.length > 0 && rewards[0]) {
					totalPriorityFee += BigInt(rewards[0]);
					count++;
				}
			}

			// Calculate average priority fee, use default value of 1 Gwei if no data available
			let maxPriorityFeePerGas: bigint;
			if (count > 0) {
				maxPriorityFeePerGas = totalPriorityFee / BigInt(count);
				// // Set minimum priority fee to 0.01 Gwei (some chains may return extremely small values)
				// const minPriorityFee = BigInt(web3.utils.toWei('0.01', 'gwei'));
				// if (maxPriorityFeePerGas < minPriorityFee) {
				// 	maxPriorityFeePerGas = minPriorityFee;
				// }
			} else {
				// If no historical data, use default value of 1 Gwei
				maxPriorityFeePerGas = BigInt(web3.utils.toWei('1', 'gwei'));
			}

			// 3. Get the base fee for the next block (baseFeePerGas)
			// The last value in baseFeePerGas array is a prediction for the next block
			const nextBaseFee =
				feeHistory.baseFeePerGas[feeHistory.baseFeePerGas.length - 1];

			if (!nextBaseFee) {
				throw new Error('Unable to retrieve base fee');
			}

			// 4. Calculate maxFeePerGas
			// Formula: maxFeePerGas = (baseFee * 2) + maxPriorityFeePerGas
			// Multiply by 2 to handle potential base fee increases in the next block (EIP-1559 allows up to 12.5% increase per block)
			const baseFeeWithBuffer = BigInt(nextBaseFee) * BigInt(2);
			const maxFeePerGas = baseFeeWithBuffer + maxPriorityFeePerGas;

			logger.log(`EIP-1559 Gas Fee Calculation:`);
			logger.log(
				`  - Base Fee (next block): ${web3.utils.fromWei(
					nextBaseFee,
					'gwei'
				)} Gwei`
			);
			logger.log(
				`  - Priority Fee: ${web3.utils.fromWei(
					maxPriorityFeePerGas,
					'gwei'
				)} Gwei`
			);
			logger.log(
				`  - Max Total Fee: ${web3.utils.fromWei(maxFeePerGas, 'gwei')} Gwei`
			);

			return {
				maxPriorityFeePerGas: maxPriorityFeePerGas.toString(),
				maxFeePerGas: maxFeePerGas.toString()
			};
		} catch (error) {
			console.error(
				'Failed to get EIP-1559 fees, using fallback config:',
				error
			);
			logger.log('⚠Using fallback gas fees from config file');
			return {
				maxPriorityFeePerGas: web3.utils.toWei(
					gas.maxPriorityFeePerGasGwei,
					'gwei'
				),
				maxFeePerGas: web3.utils.toWei(gas.maxFeePerGasGwei, 'gwei')
			};
		}
	}

	async transfer(
		contractAddress: string,
		toAddress: string,
		amount: string,
		encryptedOrderData: string
	): Promise<string> {
		if (!this.web3 || !this.contract || !this.userAccount.value) {
			throw new Error('Web3 not connected (connect wallet first)');
		}

		logger.log('Starting transfer...');
		try {
			if (!AddressValidator.isValidAddress(toAddress)) {
				throw new Error(`Invalid receiver address: ${toAddress}`);
			}
			if (!amount) {
				throw new Error(`Invalid transfer amount: ${amount}`);
			}
			if (!encryptedOrderData) {
				throw new Error('Encrypted order data is required');
			}

			const transferData = this.contract.methods
				.transfer(toAddress, amount)
				.encodeABI();

			const combinedData = transferData + encryptedOrderData;

			const txParams: any = {
				from: this.userAccount.value,
				to: contractAddress, // USDC contract address
				data: combinedData
			};

			// 1. Estimate gas limit
			const estimatedGas = await this.web3.eth.estimateGas(txParams);
			const GAS_LIMIT_MULTIPLIER = 1.2;
			const gasLimit = Math.floor(Number(estimatedGas) * GAS_LIMIT_MULTIPLIER);

			// 2. Detect if EIP-1559 is supported
			const is1559 = await this.isEIP1559Supported();

			// 3. Set gas fee parameters based on chain type
			let estimatedGasCostWei: bigint;

			if (
				is1559 &&
				this.chainConfig &&
				this.chainConfig.maxFeePerGasGwei &&
				this.chainConfig.maxPriorityFeePerGasGwei
			) {
				// EIP-1559 mode
				const { maxPriorityFeePerGas, maxFeePerGas } =
					await this.getEIP1559Fees(this.web3, {
						maxFeePerGasGwei: this.chainConfig.maxFeePerGasGwei,
						maxPriorityFeePerGasGwei: this.chainConfig.maxPriorityFeePerGasGwei
					});

				txParams.maxFeePerGas = maxFeePerGas;
				txParams.maxPriorityFeePerGas = maxPriorityFeePerGas;
				txParams.gas = gasLimit;

				// Calculate gas cost using maxFeePerGas for worst-case scenario
				estimatedGasCostWei = BigInt(gasLimit) * BigInt(maxFeePerGas);
			} else {
				// Legacy gasPrice mode
				const gasPrice = await this.web3.eth.getGasPrice();

				txParams.gasPrice = gasPrice;
				txParams.gas = gasLimit;

				// Calculate gas cost using gasPrice
				estimatedGasCostWei = BigInt(gasLimit) * BigInt(gasPrice);
			}

			// 4. Check if native token balance is sufficient to pay gas fees
			const gasCostEth = this.web3.utils.fromWei(estimatedGasCostWei, 'ether');
			const ethBalance = await this.web3.eth.getBalance(this.userAccount.value);
			const ethBalanceEth = this.web3.utils.fromWei(ethBalance, 'ether');

			logger.log(
				`Gas Fee Estimate: ${gasCostEth} ${this.chainConfig.nativeCurrency.symbol}`
			);
			logger.log(
				`Account Balance: ${ethBalanceEth} ${this.chainConfig.nativeCurrency.symbol}`
			);

			if (BigInt(ethBalance) < estimatedGasCostWei) {
				throw new Error(
					`Insufficient ${this.chainConfig.nativeCurrency.symbol} for gas fees. ` +
						`Need: ${gasCostEth} ${this.chainConfig.nativeCurrency.symbol}, ` +
						`Available: ${ethBalanceEth} ${this.chainConfig.nativeCurrency.symbol}`
				);
			}

			console.log('==> transcation => ', txParams);

			const txResult = await this.web3.eth.sendTransaction(txParams);

			const txHashBytes = txResult.transactionHash;
			const txHash = Web3.utils.bytesToHex(txHashBytes);
			logger.log(`✅ transfer successful!`);
			logger.log(`Transaction Hash: ${txHash}`);

			return txHash;
		} catch (error: unknown) {
			const err = error as { code?: number; message?: string };
			if (err.code === 4001) {
				const cancelMsg = 'User cancelled transfer';
				logger.log(cancelMsg);
				throw new Error(cancelMsg);
			}
			const errorMsg =
				error instanceof Error ? error.message : 'transfer error';
			logger.error(`❌ transfer failed: ${errorMsg}`);
			throw error;
		}
	}

	disconnectWallet(): void {
		this.isConnected.value = false;
		this.userAccount.value = '';
		this.web3 = null;
		this.contract = null;
		logger.log('Wallet disconnected successfully');
	}

	getIsConnected() {
		return this.isConnected;
	}

	getUserAccount() {
		return this.userAccount;
	}

	getWeb3Instance() {
		return this.web3;
	}
}

export const web3Service = new Web3Service();

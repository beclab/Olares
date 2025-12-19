import { ref, Ref } from 'vue';
import axios from 'axios';
import { web3Service } from './web3Service';
import { logger } from '../logger';
import type {
	TransactionStatusResult,
	TxPollTask,
	VcPollTask,
	PollingStatus,
	VCPollResult,
	SubmitVCResult
} from '../types';

export class PollingService {
	private txPollTasks = new Map<string, TxPollTask>();
	private vcPollTasks = new Map<string, VcPollTask>();

	startTxPolling(
		txHash: string,
		options: {
			interval?: number;
			confirmations?: number;
			onSuccess?: (hash: string, result: TransactionStatusResult) => void;
			onError?: (error: Error) => void;
		}
	): Ref<PollingStatus> {
		const { interval = 5000, confirmations = 1, onSuccess, onError } = options;

		if (this.txPollTasks.has(txHash)) {
			this.stopTxPolling(txHash);
			logger.warn(`Restarting tx polling for: ${txHash}`);
		}

		const taskStatus = ref<PollingStatus>('polling');
		const checkTxStatus = async () => {
			try {
				const web3 = web3Service.getWeb3Instance();
				if (!web3) throw new Error('Web3 instance not available');

				const receipt = await web3.eth.getTransactionReceipt(txHash);
				if (!receipt) {
					logger.log(`Tx [${txHash}] not mined yet`);
					return;
				}

				if (!receipt.status) {
					taskStatus.value = 'failed';
					const error = new Error(`Tx [${txHash}] failed`);
					onError?.(error);
					logger.error(error.message);
					this.stopTxPolling(txHash);
					return;
				}

				const currentBlock = await web3.eth.getBlockNumber();
				const blockConfirmations = currentBlock - receipt.blockNumber;
				if (blockConfirmations >= confirmations) {
					taskStatus.value = 'completed';
					const result: TransactionStatusResult = {
						success: true,
						confirmed: true,
						receipt
					};
					onSuccess?.(txHash, result);
					logger.log(`Tx [${txHash}] confirmed (${blockConfirmations} blocks)`);
					this.stopTxPolling(txHash);
				} else {
					logger.log(
						`Tx [${txHash}] confirmed (${blockConfirmations}/${confirmations} blocks)`
					);
				}
			} catch (error) {
				taskStatus.value = 'failed';
				const err =
					error instanceof Error ? error : new Error('startTxPolling error');
				onError?.(err);
				logger.error(`❌ Tx polling error [${txHash}]: ${err.message}`);
				this.stopTxPolling(txHash);
			}
		};

		checkTxStatus();
		const timer = setInterval(checkTxStatus, interval);

		this.txPollTasks.set(txHash, {
			id: txHash,
			timer,
			status: taskStatus,
			interval,
			confirmations,
			onSuccess,
			onError
		});

		logger.log(`🚀 Started tx polling for: ${txHash}`);
		return taskStatus;
	}

	stopTxPolling(txHash: string): void {
		const task = this.txPollTasks.get(txHash);
		if (!task) return;

		clearInterval(task.timer);
		this.txPollTasks.delete(txHash);
		if (task.status.value === 'polling') {
			task.status.value = 'idle';
		}

		logger.log(`Stopped tx polling for: ${txHash}`);
	}

	stopAllTxPolling(): void {
		Array.from(this.txPollTasks.keys()).forEach((txHash) => {
			this.stopTxPolling(txHash);
		});
		logger.log(
			`Stopped all tx polling tasks (total: ${this.txPollTasks.size})`
		);
	}

	getTxPollStatus(txHash: string): Ref<PollingStatus> | null {
		const task = this.txPollTasks.get(txHash);
		return task ? task.status : null;
	}

	startVCPolling(
		orderId: string,
		options: {
			apiUrl: string;
			interval?: number;
			onSuccess?: (result: VCPollResult) => void;
			onError?: (error: Error) => void;
		}
	): Ref<PollingStatus> {
		const { apiUrl, interval = 3000, onSuccess, onError } = options;

		if (this.vcPollTasks.has(orderId)) {
			this.stopVCPolling(orderId);
			logger.warn(`Restarting VC polling for order: ${orderId}`);
		}

		const taskStatus = ref<PollingStatus>('polling');
		const pollVC = async () => {
			try {
				const response = await axios.get(apiUrl, {
					params: { orderId }
				});

				if (response.data.code !== 0) {
					throw new Error(
						response.data.message || `VC poll failed for order ${orderId}`
					);
				}

				const { vc } = response.data.data;
				if (vc) {
					taskStatus.value = 'completed';
					const result: VCPollResult = { success: true, vc };
					onSuccess?.(result);
					logger.log(`✅ VC received for order: ${orderId}`);
					this.stopVCPolling(orderId);
				} else {
					logger.log(`VC not ready for order: ${orderId}`);
				}
			} catch (error) {
				taskStatus.value = 'failed';
				const err = error instanceof Error ? error : new Error('pollVC error');
				onError?.(err);
				logger.error(`❌ VC polling error [${orderId}]: ${err.message}`);
			}
		};

		pollVC();
		const timer = setInterval(pollVC, interval);

		this.vcPollTasks.set(orderId, {
			id: orderId,
			timer,
			status: taskStatus,
			apiUrl,
			interval,
			onSuccess,
			onError
		});

		logger.log(`Started VC polling for order: ${orderId}`);
		return taskStatus;
	}

	stopVCPolling(orderId: string): void {
		const task = this.vcPollTasks.get(orderId);
		if (!task) return;

		clearInterval(task.timer);
		this.vcPollTasks.delete(orderId);
		if (task.status.value === 'polling') {
			task.status.value = 'idle';
		}

		logger.log(`⏹Stopped VC polling for order: ${orderId}`);
	}

	stopAllVCPolling(): void {
		Array.from(this.vcPollTasks.keys()).forEach((orderId) => {
			this.stopVCPolling(orderId);
		});
		logger.log(
			`Stopped all VC polling tasks (total: ${this.vcPollTasks.size})`
		);
	}

	getVcPollStatus(orderId: string): Ref<PollingStatus> | null {
		const task = this.vcPollTasks.get(orderId);
		return task ? task.status : null;
	}

	async submitVC(
		vc: string,
		submitUrl: string,
		onSuccess?: (result: SubmitVCResult) => void,
		onError?: (error: Error) => void
	): Promise<void> {
		const submitStatus = ref<PollingStatus>('polling');
		try {
			const response = await axios.post(submitUrl, {
				vc,
				timestamp: new Date().toISOString()
			});

			if (response.data.code !== 0) {
				throw new Error(response.data.message || 'VC submission failed');
			}

			submitStatus.value = 'completed';
			const result: SubmitVCResult = {
				success: true,
				message: response.data.message,
				transactionHash: response.data.data?.txHash
			};
			onSuccess?.(result);
			logger.log(`✅ VC submitted successfully`);
		} catch (error) {
			submitStatus.value = 'failed';
			const err = error instanceof Error ? error : new Error('submitVC error');
			onError?.(err);
			logger.error(`❌ VC submission error: ${err.message}`);
		}
	}

	resetAll(): void {
		this.stopAllTxPolling();
		this.stopAllVCPolling();
		logger.log(`🔄 Reset all polling tasks`);
	}
}

export const pollingService = new PollingService();

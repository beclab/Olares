import PaymentRecoverDialog from 'src/components/appcard/PaymentRecoverDialog.vue';
import PaymentQueryDialog from 'src/components/appcard/PaymentQueryDialog.vue';
import { paymentWithProduct } from 'src/payment/service/paymentService';
import { PAYMENT_STATUS, PaymentOrderData } from 'src/constant/constants';
import { notifyFailed, notifySuccess } from 'src/utils/settings/btNotify';
import { pollingService } from 'src/payment/service/pollingService';
import { useAppStore } from 'src/stores/market/appStore';
import { BtDialog } from '@bytetrade/ui';
import { QVueGlobals } from 'quasar';
import { defineStore } from 'pinia';
import {
	getAppPaymentStatus,
	getAppPurchase,
	recoverAppPurchase,
	startBackendPolling,
	submitTransaction
} from 'src/api/market/private/payment';

export const usePaymentStore = defineStore('payment', {
	state: () => {
		return {};
	},

	actions: {
		async fetchPaymentInfo(
			appId: string,
			sourceId: string,
			t: any,
			q: QVueGlobals
		) {
			if (!appId || !sourceId) {
				console.error(
					`appId:${appId} and sourceId:${sourceId} setPaymentStatus error`
				);
				return null;
			}
			let result;
			try {
				result = await getAppPaymentStatus(appId, sourceId);
			} catch (e) {
				console.error(e);
				notifyFailed(e.response.data.message || e.message);
				return;
			}
			if (!result) {
				notifyFailed('getAppPaymentStatus result null');
				return;
			}
			const appStore = useAppStore();
			if (result.status === PAYMENT_STATUS.SYNCING) {
				notifySuccess(t('The backend is querying data, please wait.'));
			}
			appStore.updateLocalStatus(appId, sourceId, {
				status: result.status
			});

			if (result.status !== PAYMENT_STATUS.PURCHASED) {
				q.dialog({
					component: PaymentRecoverDialog,
					componentProps: {
						appId,
						sourceId,
						tokenInfo: result.token_info[0]
					}
				});
			}

			return result;
		},
		async recoverAppPurchase(appId: string, sourceId: string, t: any) {
			let result;
			try {
				result = await recoverAppPurchase(appId, sourceId);
			} catch (e) {
				console.error(e);
				notifyFailed(e.response.data.message || e.message);
				return;
			}
			if (!result) {
				notifyFailed('recoverAppPurchase result null');
				return;
			}

			const appStore = useAppStore();
			if (result.status === PAYMENT_STATUS.SYNCING) {
				notifySuccess(t('The backend is querying data, please wait.'));
			}
			appStore.updateLocalStatus(appId, sourceId, {
				status: result.status
			});

			const localStatus = appStore.getLocalStatus(appId, sourceId)?.status;
			if (localStatus === PAYMENT_STATUS.SIGNATURE_REQUIRED) {
				BtDialog.show({
					title: t('Identity Verification Required'),
					message: t(
						'Please open the Larepass mobile app for verification to proceed with the process'
					),
					okText: t('Verify Now')
				})
					.then((res) => {
						if (res) {
							//Todo
						} else {
							//Todo;
						}
					})
					.catch((err) => {
						console.log('click error', err);
					});
				return;
			}

			if (localStatus === PAYMENT_STATUS.SIGNATURE_NEED_RESIGN) {
				BtDialog.show({
					title: t('Signature Required'),
					message: t(
						'Please open the Larepass mobile app to sign and complete the App purchase'
					),
					okText: t('Sign Now')
				})
					.then((res) => {
						if (res) {
							//Todo
						} else {
							//Todo;
						}
					})
					.catch((err) => {
						console.log('click error', err);
					});
				return;
			}
			return result;
		},
		async queryPaymentInfo(
			appId: string,
			sourceId: string,
			t: any,
			q: QVueGlobals
		) {
			let result;
			try {
				result = await getAppPurchase(appId, sourceId);
			} catch (e) {
				console.error(e);
				notifyFailed(e.response.data.message || e.message);
				return;
			}
			if (!result) {
				notifyFailed('getAppPurchase result null');
				return;
			}
			const appStore = useAppStore();
			if (result.status === PAYMENT_STATUS.SYNCING) {
				notifySuccess('The backend is querying data, please wait.');
			}
			appStore.updateLocalStatus(appId, sourceId, {
				status: result.status
			});

			const localStatus = appStore.getLocalStatus(appId, sourceId)?.status;

			const confirmPay = () => {
				if (
					localStatus !== PAYMENT_STATUS.NOT_BUY &&
					localStatus !== PAYMENT_STATUS.PAYMENT_REQUIRED &&
					localStatus !== PAYMENT_STATUS.NOTIFICATION_SENT &&
					localStatus !== PAYMENT_STATUS.PAYMENT_RETRY_REQUIRED
				) {
					return;
				}
				const paymentData = result.payment_data as PaymentOrderData;
				if (!paymentData) {
					notifyFailed('payment data empty');
					return;
				}
				paymentWithProduct(paymentData)
					.then(({ txHash, productId }: any) => {
						submitTransaction(appId, sourceId, productId, {
							txHash,
							productId
						})
							.then(() => {
								notifySuccess(
									'Transaction submitted. Awaiting confirmation...'
								);
							})
							.catch((error) => {
								const msg =
									error instanceof Error ? error.message : String(error);
								notifyFailed(msg || 'Payment failed');
							});

						pollingService.startTxPolling(txHash, {
							onSuccess: async (hash, result) => {
								console.info(
									`🎉 Transaction confirmed. Hash: ${hash}, Confirmations: ${result.confirmed}`
								);
								startBackendPolling(appId, sourceId, txHash, productId)
									.then(() => {
										notifySuccess(
											'Transaction confirmed. Waiting for backend data synchronization...'
										);
									})
									.catch((error) => {
										const msg =
											error instanceof Error ? error.message : String(error);
										notifyFailed(msg || 'Payment failed');
									});
							},
							onError: (error) => {
								console.error(
									`❌ Transaction polling failed: ${error.message}`
								);
								notifyFailed(error.message);
							}
						});
					})
					.catch((error) => {
						const msg = error instanceof Error ? error.message : String(error);
						notifyFailed(msg || 'Payment failed');
					});
			};

			if (localStatus === PAYMENT_STATUS.NOT_BUY) {
				BtDialog.show({
					title: t('Identity Verification Required'),
					message: t(
						'Please open the Larepass mobile app for verification to proceed with the process'
					),
					okText: t('Verify Now')
				})
					.then((res) => {
						if (res) {
							//Todo
						} else {
							//Todo;
						}
					})
					.catch((err) => {
						console.log('click error', err);
					});
				return;
			}

			if (localStatus === PAYMENT_STATUS.SIGNATURE_REQUIRED) {
				BtDialog.show({
					title: t('Identity Verification Required'),
					message: t(
						'Please open the Larepass mobile app for verification to proceed with the process'
					),
					okText: t('Verify Now')
				})
					.then((res) => {
						if (res) {
							//Todo
						} else {
							//Todo;
						}
					})
					.catch((err) => {
						console.log('click error', err);
					});
				return;
			}

			if (localStatus === PAYMENT_STATUS.SIGNATURE_NEED_RESIGN) {
				BtDialog.show({
					title: t('Signature Required'),
					message: t(
						'Please open the Larepass mobile app to sign and complete the App purchase'
					),
					okText: t('Sign Now')
				})
					.then((res) => {
						if (res) {
							//Todo
						} else {
							//Todo;
						}
					})
					.catch((err) => {
						console.log('click error', err);
					});
				return;
			}

			if (localStatus === PAYMENT_STATUS.PAYMENT_RETRY_REQUIRED) {
				q.dialog({
					component: PaymentQueryDialog
				}).onOk(async (res) => {
					if (res.status === 'unpaid') {
						confirmPay();
					} else if (res.status === 'paid') {
						if (
							result?.frontend_data?.txHash &&
							result?.frontend_data?.productId
						) {
							startBackendPolling(
								appId,
								sourceId,
								result.frontend_data.txHash,
								result.frontend_data.productId
							)
								.then(() => {
									notifySuccess('Success, waiting for data synchronization...');
								})
								.catch((error) => {
									const msg =
										error instanceof Error ? error.message : String(error);
									notifyFailed(msg || 'Payment failed');
								});
						} else {
							BtDialog.show({
								title: t('On-Chain Transaction Query'),
								cancel: true,
								prompt: {
									model: '',
									type: 'text', // optional
									name: t('Transaction Hash'),
									placeholder: t('Please enter Transaction Hash')
								}
							}).then((res) => {
								const productId = result.frontend_data.productId
									? result.frontend_data.productId
									: result.payment_data.price_config.paid.product_id;
								if (res) {
									startBackendPolling(appId, sourceId, res as string, productId)
										.then(() => {
											notifySuccess(
												'Success, waiting for data synchronization...'
											);
										})
										.catch((error) => {
											const msg =
												error instanceof Error ? error.message : String(error);
											notifyFailed(msg || 'Payment failed');
										});
								} else {
									notifyFailed(
										`Failed to obtain txHash ${res} and productId ${productId}`
									);
								}
							});
						}
					}
				});
				return;
			}

			confirmPay();
		}
	}
});

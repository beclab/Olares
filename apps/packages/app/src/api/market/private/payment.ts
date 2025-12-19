import { useCenterStore } from 'src/stores/market/center';
import axios from 'axios';

export async function getAppPaymentStatus(
	appId: string,
	sourceId: string
): Promise<any> {
	const store = useCenterStore();
	const url = `${store.appUrl}/sources/${sourceId}/apps/${appId}/payment-status`;
	const { data } = await axios.get(url);
	console.log(data);
	return data;
}

export async function getAppPurchase(appId: string, sourceId: string) {
	const store = useCenterStore();
	const url = `${store.appUrl}/sources/${sourceId}/apps/${appId}/purchase`;
	const { data } = await axios.post(url);
	console.log(data);
	return data;
}

export async function submitTransaction(
	appId: string,
	sourceId: string,
	productId: string,
	transaction: any
) {
	const store = useCenterStore();
	const url = `${store.appUrl}/payment/frontend-start`;
	const { data } = await axios.post(url, {
		app_id: appId,
		source_id: sourceId,
		product_id: productId,
		frontend_data: transaction
	});
	console.log(data);
	return data;
}

export async function startBackendPolling(
	appId: string,
	sourceId: string,
	txHash: string,
	productId: string
): Promise<any> {
	const store = useCenterStore();
	const url = `${store.appUrl}/payment/start-polling`;
	const { data } = await axios.post(url, {
		app_id: appId,
		source_id: sourceId,
		tx_hash: txHash,
		product_id: productId
	});
	console.log(data);
	return data;
}

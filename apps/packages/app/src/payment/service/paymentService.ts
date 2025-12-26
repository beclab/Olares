import type { OrderDataBase } from '../types';
import { RandomUtils } from '../utils/randomUtils';
import { web3Service } from './web3Service';
import { encryptService } from './encryptService';
import { WEB3_CONFIG } from '../config';
import { PaymentOrderData } from 'src/constant/constants';
import BigNumber from 'bignumber.js';

export async function paymentWithProduct(paymentInfo: PaymentOrderData) {
	if (!paymentInfo) {
		throw new Error('Payment info cannot be empty');
	}

	try {
		console.info('=== Starting product payment process ===');

		const orderData: OrderDataBase = {
			from: paymentInfo.from,
			product: paymentInfo.product,
			to: paymentInfo.to,
			nonce: RandomUtils.generateNonce()
		};
		console.debug('Built base order data:', orderData);

		const priceConfig = paymentInfo.price_config.paid;
		if (!priceConfig || !priceConfig.price?.length) {
			throw new Error('Invalid price config: missing paid price information');
		}
		//TODO first price
		const targetPrice = priceConfig.price[0];
		console.debug('Target price config:', targetPrice);

		const tokenInfo = paymentInfo.token_info.find(
			(item) =>
				item.token_symbol.toLowerCase() ===
					targetPrice.token_symbol.toLowerCase() &&
				item.chain.toLowerCase() === targetPrice.chain.toLowerCase()
		);
		if (!tokenInfo) {
			throw new Error(
				`No supported token found for [symbol: ${targetPrice.token_symbol}, chain: ${targetPrice.chain}]`
			);
		}
		console.debug('Matched token info:', {
			symbol: tokenInfo.token_symbol,
			chain: tokenInfo.chain,
			contract: tokenInfo.token_contract
		});

		const supportChain = WEB3_CONFIG.CHAIN.find(
			(item) => item.chainName.toLowerCase() === tokenInfo.chain.toLowerCase()
		);
		if (!supportChain) {
			throw new Error(
				`Chain "${
					tokenInfo.chain
				}" is not supported. Supported chains: ${WEB3_CONFIG.CHAIN.map(
					(c) => c.chainName
				).join(', ')}`
			);
		}
		console.debug('Matched supported chain:', {
			chainName: supportChain.chainName,
			chainId: supportChain.chainId
		});

		console.info('Connecting to wallet...');
		await web3Service.connectWallet(tokenInfo.token_contract, supportChain);
		const userAccount = web3Service.getUserAccount();
		if (!userAccount.value) {
			throw new Error('Failed to get user account: wallet not connected');
		}
		console.info(
			`✅ Wallet connected successfully. User account: ${userAccount.value}`
		);

		console.info('Querying user token balance...');
		const balance = await web3Service.queryBalance();
		console.debug(`User balance (smallest unit): ${balance}`);

		const balanceBN = new BigNumber(balance);
		if (balanceBN.isLessThanOrEqualTo(0)) {
			throw new Error(`User has no balance for ${tokenInfo.token_symbol}`);
		}

		// tokenInfo.token_amount should be in the smallest unit (e.g., wei)
		const requiredAmount = new BigNumber(tokenInfo.token_amount);

		// Check if user has enough balance
		if (balanceBN.isLessThan(requiredAmount)) {
			// Format for display
			const divisor = new BigNumber(10).exponentiatedBy(
				tokenInfo.token_decimals
			);
			const balanceFormatted = balanceBN.dividedBy(divisor).toString();
			const requiredFormatted = requiredAmount.dividedBy(divisor).toString();

			throw new Error(
				`Insufficient ${tokenInfo.token_symbol} balance. ` +
					`Need: ${requiredFormatted} ${tokenInfo.token_symbol}, ` +
					`Available: ${balanceFormatted} ${tokenInfo.token_symbol}`
			);
		}

		console.info('Signing order data...');
		const signedOrder = await web3Service.signOrder(orderData);
		console.info('✅ Order signed successfully');

		console.info('Encrypting order data...');
		await encryptService.processOrderData(
			signedOrder,
			paymentInfo.ras_public_key
		);
		const encryptedData = encryptService.getEncryptedData();
		if (!encryptedData.value) {
			throw new Error('Failed to get encrypted order data');
		}
		console.debug(
			`✅ Order data encrypted. Length: ${encryptedData.value.length}`
		);

		console.info(`Initiating transfer to ${tokenInfo.receive_wallet}...`);
		const txHash = await web3Service.transfer(
			tokenInfo.token_contract,
			tokenInfo.receive_wallet,
			tokenInfo.token_amount.toString(),
			encryptedData.value
		);
		if (!txHash) {
			throw new Error('Failed to get transaction hash after transfer');
		}
		const explorerUrl = `${supportChain?.blockExplorerUrls[0]}/tx/${txHash}`;
		console.info(`✅ Transfer initiated. Transaction hash: ${txHash}`);
		console.info(`Explorer link: ${explorerUrl}`);
		console.info('=== Product payment process completed ===');
		return {
			success: true,
			txHash,
			productId: priceConfig.product_id,
			orderData: signedOrder
		};
	} catch (error) {
		console.error(`❌ Product payment failed: ${error.message}`, error.stack);
		throw error instanceof Error
			? error
			: new Error('paymentWithProduct error');
	}
}
